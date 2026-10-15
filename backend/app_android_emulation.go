package backend

import (
	"fmt"
	"strconv"
	"strings"
)

// androidEmulationConfig is deliberately small and explicit. It models the
// browser-facing mobile surface without pretending to be a full Android OS.
// The values are kept in profile fingerprint args so copies remain stable.
type androidEmulationConfig struct {
	Model        string
	UserAgent    string
	Platform     string
	PlatformVer  string
	Width        int
	Height       int
	DPR          float64
	TouchPoints  int
	Brand        string
	BrandVersion string
	AcceptLang   string
}

func parseAndroidEmulationConfig(args []string) (*androidEmulationConfig, bool, error) {
	platform := normalizeFingerprintPlatform(browserArgValue(args, "--fingerprint-platform"))
	if platform != "android" {
		return nil, false, nil
	}

	model := strings.TrimSpace(browserArgValue(args, "--fingerprint-mobile-model"))
	if model == "" {
		model = "Pixel 7"
	}
	platformVersion := strings.TrimSpace(browserArgValue(args, "--fingerprint-platform-version"))
	if platformVersion == "" {
		platformVersion = "14"
	}
	brand := strings.TrimSpace(browserArgValue(args, "--fingerprint-brand"))
	if brand == "" {
		brand = "Chrome"
	}
	brandVersion := strings.TrimSpace(browserArgValue(args, "--fingerprint-brand-version"))
	if brandVersion == "" {
		brandVersion = "144.0.7559.132"
	}

	width, height := 412, 915
	if viewport := strings.TrimSpace(browserArgValue(args, "--fingerprint-mobile-viewport")); viewport != "" {
		parsedWidth, parsedHeight, err := parseAndroidViewport(viewport)
		if err != nil {
			return nil, true, err
		}
		width, height = parsedWidth, parsedHeight
	}

	dpr := 2.625
	if raw := strings.TrimSpace(browserArgValue(args, "--fingerprint-mobile-dpr")); raw != "" {
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil || parsed < 1 || parsed > 5 {
			return nil, true, fmt.Errorf("Android DPR 必须是 1 到 5 之间的数字")
		}
		dpr = parsed
	}

	touchPoints := 5
	if raw := strings.TrimSpace(browserArgValue(args, "--fingerprint-mobile-touch-points")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 20 {
			return nil, true, fmt.Errorf("Android 触控点必须是 1 到 20 的整数")
		}
		touchPoints = parsed
	}

	userAgent := strings.TrimSpace(browserArgValue(args, "--fingerprint-mobile-ua"))
	if userAgent == "" {
		userAgent = fmt.Sprintf("Mozilla/5.0 (Linux; Android %s; %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Mobile Safari/537.36", platformVersion, model, brandVersion)
	}

	return &androidEmulationConfig{
		Model:        model,
		UserAgent:    userAgent,
		Platform:     "Android",
		PlatformVer:  platformVersion,
		Width:        width,
		Height:       height,
		DPR:          dpr,
		TouchPoints:  touchPoints,
		Brand:        brand,
		BrandVersion: brandVersion,
		AcceptLang:   browserArgValue(args, "--accept-lang"),
	}, true, nil
}

func parseAndroidViewport(value string) (int, int, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("Android 视口必须是 宽,高，例如 412,915")
	}
	width, widthErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, heightErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if widthErr != nil || heightErr != nil || width < 240 || height < 320 || width > 3000 || height > 5000 {
		return 0, 0, fmt.Errorf("Android 视口尺寸不合法")
	}
	return width, height, nil
}

func applyAndroidEmulationCDP(debugPort int, config *androidEmulationConfig) error {
	if config == nil {
		return nil
	}

	// Apply the device metrics first so responsive pages immediately enter
	// mobile layout. CDP is used instead of JS injection so UA-CH and touch
	// emulation are changed through Chromium's own domains.
	if _, err := cdpCall(debugPort, "Emulation.setDeviceMetricsOverride", map[string]any{
		"width":             config.Width,
		"height":            config.Height,
		"deviceScaleFactor": config.DPR,
		"mobile":            true,
		"screenWidth":       config.Width,
		"screenHeight":      config.Height,
		"scale":             1,
		"screenOrientation": map[string]any{"type": "portraitPrimary", "angle": 0},
	}); err != nil {
		return fmt.Errorf("设置 Android 设备尺寸失败: %w", err)
	}
	if _, err := cdpCall(debugPort, "Emulation.setTouchEmulationEnabled", map[string]any{
		"enabled":        true,
		"maxTouchPoints": config.TouchPoints,
	}); err != nil {
		return fmt.Errorf("设置 Android 触摸事件失败: %w", err)
	}

	metadata := map[string]any{
		"brands":          []map[string]string{{"brand": config.Brand, "version": majorVersion(config.BrandVersion)}},
		"fullVersionList": []map[string]string{{"brand": config.Brand, "version": config.BrandVersion}},
		"platform":        config.Platform,
		"platformVersion": config.PlatformVer,
		"architecture":    "arm",
		"model":           config.Model,
		"mobile":          true,
		"bitness":         "64",
		"wow64":           false,
	}
	uaParams := map[string]any{
		"userAgent":         config.UserAgent,
		"platform":          config.Platform,
		"userAgentMetadata": metadata,
	}
	if config.AcceptLang != "" {
		uaParams["acceptLanguage"] = config.AcceptLang
	}
	if _, err := cdpCall(debugPort, "Network.setUserAgentOverride", uaParams); err != nil {
		return fmt.Errorf("设置 Android UA/Client Hints 失败: %w", err)
	}
	return nil
}

func majorVersion(value string) string {
	if idx := strings.IndexByte(value, '.'); idx >= 0 {
		return value[:idx]
	}
	return value
}
