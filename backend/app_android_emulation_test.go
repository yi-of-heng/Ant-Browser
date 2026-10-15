package backend

import "testing"

func TestParseAndroidEmulationConfigDefaults(t *testing.T) {
	config, enabled, err := parseAndroidEmulationConfig([]string{"--fingerprint-platform=android"})
	if err != nil || !enabled {
		t.Fatalf("parse defaults = config=%#v enabled=%v err=%v", config, enabled, err)
	}
	if config.Model != "Pixel 7" || config.Width != 412 || config.Height != 915 || config.DPR != 2.625 || config.TouchPoints != 5 {
		t.Fatalf("unexpected defaults: %#v", config)
	}
	if config.Platform != "Android" || config.PlatformVer != "14" || config.UserAgent == "" {
		t.Fatalf("unexpected identity defaults: %#v", config)
	}
}

func TestParseAndroidEmulationConfigCustomValues(t *testing.T) {
	config, enabled, err := parseAndroidEmulationConfig([]string{
		"--fingerprint-platform=android",
		"--fingerprint-platform-version=15",
		"--fingerprint-mobile-model=SM-S928B",
		"--fingerprint-mobile-viewport=430,932",
		"--fingerprint-mobile-dpr=3.5",
		"--fingerprint-mobile-touch-points=10",
		"--fingerprint-mobile-ua=custom-android-ua",
	})
	if err != nil || !enabled {
		t.Fatalf("parse custom = config=%#v enabled=%v err=%v", config, enabled, err)
	}
	if config.Model != "SM-S928B" || config.Width != 430 || config.Height != 932 || config.DPR != 3.5 || config.TouchPoints != 10 || config.UserAgent != "custom-android-ua" {
		t.Fatalf("unexpected custom config: %#v", config)
	}
}

func TestParseAndroidEmulationConfigRejectsInvalidViewport(t *testing.T) {
	_, enabled, err := parseAndroidEmulationConfig([]string{
		"--fingerprint-platform=android",
		"--fingerprint-mobile-viewport=10,10",
	})
	if !enabled || err == nil {
		t.Fatalf("expected invalid viewport error, enabled=%v err=%v", enabled, err)
	}
}

func TestParseAndroidEmulationConfigIgnoresDesktopProfiles(t *testing.T) {
	config, enabled, err := parseAndroidEmulationConfig([]string{"--fingerprint-platform=windows"})
	if err != nil || enabled || config != nil {
		t.Fatalf("desktop profile should be ignored: config=%#v enabled=%v err=%v", config, enabled, err)
	}
}

func TestAndroidLaunchPlanRemovesInternalArgsAndAddsGlobalFallbacks(t *testing.T) {
	plan := buildBrowserFingerprintLaunchPlanForHost("android-profile", []string{
		"--fingerprint-platform=android",
		"--fingerprint-mobile-model=Pixel 7",
		"--fingerprint-mobile-viewport=412,915",
		"--fingerprint-mobile-dpr=2.625",
	}, "144.0.7559.132", "darwin")
	for _, arg := range plan.launchArgs {
		if browserFingerprintArgKey(arg) == "--fingerprint-mobile-model" || browserFingerprintArgKey(arg) == "--fingerprint-mobile-viewport" {
			t.Fatalf("internal Android arg leaked into launch args: %q", arg)
		}
	}
	if browserArgWithKey(plan.launchArgs, "--user-agent") == "" || browserArgWithKey(plan.launchArgs, "--force-device-scale-factor") == "" || browserArgWithKey(plan.launchArgs, "--touch-events") == "" {
		t.Fatalf("missing Android global fallback args: %#v", plan.launchArgs)
	}
}
