// ant-mcp is a local MCP stdio adapter over the Launch API.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"ant-chrome/backend/internal/controlclient"
	"ant-chrome/backend/internal/controlops"
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
	in.Buffer(make([]byte, 64*1024), 16*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for in.Scan() {
		var req rpc
		if err := json.Unmarshal(in.Bytes(), &req); err != nil {
			fmt.Fprintln(out, `{"jsonrpc":"2.0","error":{"code":-32700,"message":"parse error"},"id":null}`)
			out.Flush()
			continue
		}
		if req.Method == "notifications/initialized" || req.ID == nil {
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
		r.Result = map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "ant-browser", "version": "0.2.0"}}
	case "tools/list":
		r.Result = map[string]any{"tools": controlops.Tools()}
	case "tools/call":
		r.Result = callTool(ctx, c, req.Params)
	default:
		r.Error = map[string]any{"code": -32601, "message": "method not found"}
	}
	return r
}
func callTool(ctx context.Context, c *controlclient.Client, raw json.RawMessage) map[string]any {
	var request struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &request); err != nil {
		return textErr("invalid tool arguments", nil)
	}
	value, err := controlops.Execute(ctx, c, request.Name, request.Arguments)
	if err != nil {
		return textErr(err.Error(), value)
	}
	data, _ := json.Marshal(value)
	return map[string]any{"content": []map[string]string{{"type": "text", "text": string(data)}}, "structuredContent": value}
}
func textErr(message string, partial any) map[string]any {
	out := map[string]any{"isError": true, "content": []map[string]string{{"type": "text", "text": message}}}
	if partial != nil {
		out["structuredContent"] = partial
	}
	return out
}
func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
