package controlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ant-chrome/backend/internal/browser"
)

// Client is the small, typed client shared by antctl and the MCP adapter.
// It deliberately talks to Launch API instead of opening the application DB.
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func New(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:19876"
	}
	h := &http.Client{Timeout: 30 * time.Second}
	return &Client{BaseURL: baseURL, APIKey: strings.TrimSpace(apiKey), HTTP: h}
}

type profilesResponse struct {
	Items []browser.Profile `json:"items"`
}
type profileResponse struct {
	Profile browser.Profile `json:"profile"`
}
type proxyResponse struct {
	Items []browser.Proxy `json:"items"`
}

type automationListResponse struct {
	Data struct {
		Items []map[string]any `json:"items"`
	} `json:"data"`
}

type automationItemResponse struct {
	Data struct {
		Item map[string]any `json:"item"`
	} `json:"data"`
}

type automationRunResponse struct {
	Data map[string]any `json:"data"`
}

func (c *Client) do(ctx context.Context, method, path string, input any, output any) error {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Set("X-Ant-Api-Key", c.APIKey)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("Launch API 请求失败: %w", err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var payload struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(data, &payload)
		if payload.Error == "" {
			payload.Error = strings.TrimSpace(string(data))
		}
		return fmt.Errorf("Launch API HTTP %d: %s", res.StatusCode, payload.Error)
	}
	if output != nil && len(data) > 0 {
		if err := json.Unmarshal(data, output); err != nil {
			return fmt.Errorf("解析 Launch API 响应失败: %w", err)
		}
	}
	return nil
}

func (c *Client) Health(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/api/health", nil, nil)
}

func (c *Client) ListProfiles(ctx context.Context) ([]browser.Profile, error) {
	var out profilesResponse
	if err := c.do(ctx, http.MethodGet, "/api/profiles", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) GetProfile(ctx context.Context, id string) (browser.Profile, error) {
	var out profileResponse
	if err := c.do(ctx, http.MethodGet, "/api/profiles/"+pathEscape(id), nil, &out); err != nil {
		return browser.Profile{}, err
	}
	return out.Profile, nil
}

type ProfileWrite struct {
	Profile    *browser.ProfileInput `json:"profile"`
	LaunchCode string                `json:"launchCode,omitempty"`
	AutoLaunch bool                  `json:"autoLaunch,omitempty"`
}

func (c *Client) CreateProfile(ctx context.Context, input browser.ProfileInput, autoLaunch bool) (browser.Profile, error) {
	var out profileResponse
	if err := c.do(ctx, http.MethodPost, "/api/profiles", ProfileWrite{Profile: &input, AutoLaunch: autoLaunch}, &out); err != nil {
		return browser.Profile{}, err
	}
	return out.Profile, nil
}

func (c *Client) UpdateProfile(ctx context.Context, id string, input browser.ProfileInput, autoLaunch bool) (browser.Profile, error) {
	var out profileResponse
	if err := c.do(ctx, http.MethodPut, "/api/profiles/"+pathEscape(id), ProfileWrite{Profile: &input, AutoLaunch: autoLaunch}, &out); err != nil {
		return browser.Profile{}, err
	}
	return out.Profile, nil
}

func (c *Client) DeleteProfile(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/profiles/"+pathEscape(id), nil, nil)
}

func (c *Client) CopyProfile(ctx context.Context, id, name, mode string, autoLaunch bool) (browser.Profile, error) {
	var out struct {
		Profile browser.Profile `json:"profile"`
	}
	body := map[string]any{"autoLaunch": autoLaunch}
	if strings.TrimSpace(name) != "" {
		body["name"] = strings.TrimSpace(name)
	}
	if strings.TrimSpace(mode) != "" {
		body["mode"] = strings.TrimSpace(mode)
	}
	if err := c.do(ctx, http.MethodPost, "/api/profiles/"+pathEscape(id)+"/copy", body, &out); err != nil {
		return browser.Profile{}, err
	}
	return out.Profile, nil
}
func (c *Client) StartProfile(ctx context.Context, id string) (browser.Profile, error) {
	var out struct {
		OK          bool   `json:"ok"`
		ProfileId   string `json:"profileId"`
		ProfileName string `json:"profileName"`
		DebugPort   int    `json:"debugPort"`
		DebugReady  bool   `json:"debugReady"`
		Pid         int    `json:"pid"`
		LaunchCode  string `json:"launchCode"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/launch", map[string]string{"profileId": id}, &out); err != nil {
		return browser.Profile{}, err
	}
	p, err := c.GetProfile(ctx, id)
	if err == nil {
		p.Running = true
		if out.DebugPort > 0 {
			p.DebugPort = out.DebugPort
			p.DebugReady = out.DebugReady
			p.Pid = out.Pid
		}
		return p, nil
	}
	return browser.Profile{
		ProfileId:   out.ProfileId,
		ProfileName: out.ProfileName,
		DebugPort:   out.DebugPort,
		DebugReady:  out.DebugReady,
		Pid:         out.Pid,
		Running:     true,
	}, nil
}
func (c *Client) StopProfile(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, "/api/profiles/"+pathEscape(id)+"/stop", nil, nil)
}

func (c *Client) ListProxies(ctx context.Context) ([]browser.Proxy, error) {
	var out proxyResponse
	if err := c.do(ctx, http.MethodGet, "/api/proxies", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) ListAutomationScripts(ctx context.Context) ([]map[string]any, error) {
	var out automationListResponse
	if err := c.do(ctx, http.MethodGet, "/api/automation/scripts", nil, &out); err != nil {
		return nil, err
	}
	return out.Data.Items, nil
}

func (c *Client) GetAutomationScript(ctx context.Context, id string) (map[string]any, error) {
	var out automationItemResponse
	if err := c.do(ctx, http.MethodGet, "/api/automation/scripts/"+pathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Data.Item, nil
}

func (c *Client) RunAutomationScript(ctx context.Context, input map[string]any) (map[string]any, error) {
	var out automationRunResponse
	if err := c.do(ctx, http.MethodPost, "/api/automation/scripts/run", input, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) ListAutomationRuns(ctx context.Context, limit int) ([]map[string]any, error) {
	path := "/api/automation/scripts/runs"
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var out automationListResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Data.Items, nil
}

func pathEscape(value string) string { return strings.ReplaceAll(strings.TrimSpace(value), "/", "%2F") }
