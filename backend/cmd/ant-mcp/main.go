// ant-mcp is a minimal local MCP stdio adapter over Launch API.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/controlclient"
)

type rpc struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
type result struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func main() {
	c := controlclient.New(envOr("ANT_BROWSER_URL", "http://127.0.0.1:19876"), os.Getenv("ANT_BROWSER_API_KEY"))
	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for in.Scan() {
		var req rpc
		if json.Unmarshal(in.Bytes(), &req) != nil {
			continue
		}
		if req.Method == "notifications/initialized" {
			continue
		}
		res := handle(context.Background(), c, req)
		data, _ := json.Marshal(res)
		fmt.Fprintln(out, string(data))
		out.Flush()
	}
}
func handle(ctx context.Context, c *controlclient.Client, req rpc) result {
	r := result{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		r.Result = map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "ant-browser", "version": "0.1.0"}}
	case "tools/list":
		r.Result = map[string]any{"tools": tools()}
	case "tools/call":
		r.Result = callTool(ctx, c, req.Params)
	default:
		r.Error = map[string]any{"code": -32601, "message": "method not found"}
	}
	return r
}
func tools() []map[string]any {
	return []map[string]any{
		{"name": "list_instances", "description": "列出浏览器实例", "inputSchema": map[string]any{"type": "object"}},
		{"name": "list_proxy_nodes", "description": "列出代理节点及其 ID", "inputSchema": map[string]any{"type": "object"}},
		{"name": "create_instance", "description": "创建实例，可用 proxy_id 绑定代理节点", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"name": map[string]string{"type": "string"}, "proxy_id": map[string]string{"type": "string"}, "core_id": map[string]string{"type": "string"}, "group_id": map[string]string{"type": "string"}}, "required": []string{"name"}}},
		{"name": "update_instance", "description": "更新实例配置；仅传需要修改的字段", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"instance_id": map[string]string{"type": "string"}, "name": map[string]string{"type": "string"}, "proxy_id": map[string]string{"type": "string"}, "core_id": map[string]string{"type": "string"}, "group_id": map[string]string{"type": "string"}}, "required": []string{"instance_id"}}},
		{"name": "start_instance", "description": "启动实例", "inputSchema": schema("instance_id")},
		{"name": "stop_instance", "description": "停止实例", "inputSchema": schema("instance_id")},
		{"name": "delete_instance", "description": "删除已停止实例，需要 confirm=true", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"instance_id": map[string]string{"type": "string"}, "confirm": map[string]string{"type": "boolean"}}, "required": []string{"instance_id", "confirm"}}},
	}
}
func schema(name string) map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{name: map[string]string{"type": "string"}}, "required": []string{name}}
}
func callTool(ctx context.Context, c *controlclient.Client, raw json.RawMessage) map[string]any {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if json.Unmarshal(raw, &p) != nil {
		return textErr("invalid tool arguments")
	}
	a := p.Arguments
	var value any
	var err error
	switch p.Name {
	case "list_instances":
		var items []browser.Profile
		items, err = c.ListProfiles(ctx)
		value = safeProfiles(items)
	case "list_proxy_nodes":
		var items []browser.Proxy
		items, err = c.ListProxies(ctx)
		value = safeProxies(items)
	case "create_instance":
		var item browser.Profile
		item, err = c.CreateProfile(ctx, input(a), false)
		value = safeProfile(item)
	case "update_instance":
		id := str(a, "instance_id")
		if id == "" {
			return textErr("instance_id is required")
		}
		var current browser.Profile
		current, err = c.GetProfile(ctx, id)
		in := inputFrom(current)
		applyPatch(&in, a)
		if err == nil {
			var item browser.Profile
			item, err = c.UpdateProfile(ctx, id, in, false)
			value = safeProfile(item)
		}
	case "start_instance":
		var item browser.Profile
		item, err = c.StartProfile(ctx, str(a, "instance_id"))
		value = safeProfile(item)
	case "stop_instance":
		err = c.StopProfile(ctx, str(a, "instance_id"))
		value = map[string]any{"ok": err == nil}
	case "delete_instance":
		if !boolArg(a, "confirm") {
			return textErr("delete requires confirm=true")
		}
		err = c.DeleteProfile(ctx, str(a, "instance_id"))
		value = map[string]any{"ok": err == nil}
	default:
		return textErr("unknown tool: " + p.Name)
	}
	if err != nil {
		return textErr(err.Error())
	}
	data, _ := json.Marshal(value)
	return map[string]any{"content": []map[string]string{{"type": "text", "text": string(data)}}, "structuredContent": value}
}
func input(a map[string]any) browser.ProfileInput {
	return browser.ProfileInput{ProfileName: str(a, "name"), ProxyId: str(a, "proxy_id"), CoreId: str(a, "core_id"), GroupId: str(a, "group_id")}
}
func inputFrom(p browser.Profile) browser.ProfileInput {
	return browser.ProfileInput{ProfileName: p.ProfileName, UserDataDir: p.UserDataDir, CoreId: p.CoreId, ProxyId: p.ProxyId, ProxyConfig: p.ProxyConfig, GroupId: p.GroupId, FingerprintArgs: p.FingerprintArgs, LaunchArgs: p.LaunchArgs, Tags: p.Tags, Keywords: p.Keywords, MemoryLimitMB: p.MemoryLimitMB, RestoreLastSession: p.RestoreLastSession}
}

// MCP results intentionally omit proxyConfig, userDataDir and launch args.
// Those fields can contain credentials, local paths, or fingerprint details;
// an agent only needs stable IDs and runtime state for orchestration.
func safeProfile(p browser.Profile) map[string]any {
	return map[string]any{
		"profileId": p.ProfileId, "profileName": p.ProfileName, "coreId": p.CoreId,
		"proxyId": p.ProxyId, "groupId": p.GroupId, "tags": p.Tags,
		"running": p.Running, "debugReady": p.DebugReady, "debugPort": p.DebugPort,
		"pid": p.Pid, "launchCode": p.LaunchCode, "lastError": p.LastError,
	}
}
func safeProfiles(items []browser.Profile) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, safeProfile(item))
	}
	return out
}
func safeProxies(items []browser.Proxy) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"proxyId": item.ProxyId, "proxyName": item.ProxyName, "groupName": item.GroupName, "preferredKernel": item.PreferredKernel, "lastTestOk": item.LastTestOk, "lastLatencyMs": item.LastLatencyMs, "lastTestedAt": item.LastTestedAt})
	}
	return out
}
func applyPatch(in *browser.ProfileInput, a map[string]any) {
	if v := str(a, "name"); v != "" {
		in.ProfileName = v
	}
	if v := str(a, "proxy_id"); v != "" {
		in.ProxyId = v
	}
	if v := str(a, "core_id"); v != "" {
		in.CoreId = v
	}
	if v := str(a, "group_id"); v != "" {
		in.GroupId = v
	}
}
func str(a map[string]any, k string) string   { v, _ := a[k].(string); return strings.TrimSpace(v) }
func boolArg(a map[string]any, k string) bool { v, _ := a[k].(bool); return v }
func textErr(msg string) map[string]any {
	return map[string]any{"isError": true, "content": []map[string]string{{"type": "text", "text": msg}}}
}
func envOr(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}

var _ = io.EOF
