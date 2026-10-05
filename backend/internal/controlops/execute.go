package controlops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/controlclient"
)

// Execute validates and runs one operation. The returned value is safe to
// present to scripts/agents: sensitive profile and proxy fields are redacted.
func Execute(ctx context.Context, c *controlclient.Client, name string, args map[string]any) (any, error) {
	spec, ok := Find(name)
	if !ok {
		return nil, fmt.Errorf("unknown operation: %s", name)
	}
	if args == nil {
		args = map[string]any{}
	}
	if err := validate(spec, args); err != nil {
		return nil, err
	}
	id := stringArg(args, "instance_id")
	scriptID := stringArg(args, "script_id")
	switch name {
	case "health":
		err := c.Health(ctx)
		return map[string]any{"ok": err == nil}, err
	case "list_proxy_nodes":
		items, err := c.ListProxies(ctx)
		return SafeProxies(items), err
	case "list_instances":
		items, err := c.ListProfiles(ctx)
		return SafeProfiles(items), err
	case "get_instance":
		item, err := c.GetProfile(ctx, id)
		return SafeProfile(item), err
	case "create_instance":
		item, err := c.CreateProfile(ctx, newProfile(args), boolArg(args, "auto_launch"))
		return SafeProfile(item), err
	case "update_instance":
		current, err := c.GetProfile(ctx, id)
		if err != nil {
			return nil, err
		}
		input := InputFromProfile(current)
		patchProfile(&input, args)
		item, err := c.UpdateProfile(ctx, id, input, boolArg(args, "auto_launch"))
		return SafeProfile(item), err
	case "start_instance":
		item, err := c.StartProfile(ctx, id)
		return SafeProfile(item), err
	case "stop_instance":
		err := c.StopProfile(ctx, id)
		return map[string]any{"ok": err == nil}, err
	case "delete_instance":
		if !boolArg(args, "confirm") {
			return nil, errors.New("delete requires confirm=true")
		}
		err := c.DeleteProfile(ctx, id)
		return map[string]any{"ok": err == nil}, err
	case "copy_instance":
		item, err := c.CopyProfile(ctx, id, stringArg(args, "name"), stringArg(args, "mode"), boolArg(args, "auto_launch"))
		return SafeProfile(item), err
	case "create_instances_batch":
		profiles, err := decodeProfiles(args["profiles"])
		if err != nil {
			return nil, err
		}
		result := map[string]any{"created": 0, "items": []map[string]any{}}
		items := make([]map[string]any, 0, len(profiles))
		for i, profile := range profiles {
			if strings.TrimSpace(profile.ProfileName) == "" {
				result["created"], result["items"], result["failedIndex"] = len(items), items, i
				return result, fmt.Errorf("profiles[%d].profileName is required", i)
			}
			item, err := c.CreateProfile(ctx, profile, false)
			if err != nil {
				result["created"], result["items"], result["failedIndex"] = len(items), items, i
				return result, fmt.Errorf("profiles[%d]: %w", i, err)
			}
			items = append(items, SafeProfile(item))
		}
		result["created"], result["items"] = len(items), items
		return result, nil
	case "list_automation_scripts":
		return c.ListAutomationScripts(ctx)
	case "get_automation_script":
		return c.GetAutomationScript(ctx, scriptID)
	case "run_automation_script":
		input := map[string]any{"scriptId": scriptID}
		for _, key := range []string{"selector", "target_input", "params"} {
			if value, ok := args[key]; ok {
				apiKey := map[string]string{"target_input": "targetInput"}[key]
				if apiKey == "" {
					apiKey = key
				}
				input[apiKey] = value
			}
		}
		for _, key := range []string{"use_script_selector", "use_script_params"} {
			if value, ok := args[key]; ok {
				input[camelCase(key)] = value
			}
		}
		if value, ok := args["timeout_ms"]; ok {
			input["timeoutMs"] = value
		}
		return c.RunAutomationScript(ctx, input)
	case "list_automation_runs":
		limit, _ := integerArg(args["limit"])
		return c.ListAutomationRuns(ctx, limit)
	}
	return nil, fmt.Errorf("unhandled operation: %s", name)
}

func validate(spec Spec, args map[string]any) error {
	fields := map[string]Field{}
	for _, field := range spec.Fields {
		fields[field.Name] = field
		v, exists := args[field.Name]
		if field.Required && (!exists || v == nil) {
			return fmt.Errorf("%s is required", field.Name)
		}
	}
	for name, value := range args {
		field, ok := fields[name]
		if !ok {
			return fmt.Errorf("unknown argument: %s", name)
		}
		switch field.Type {
		case "string":
			v, ok := value.(string)
			if !ok || (field.Required && strings.TrimSpace(v) == "") {
				return fmt.Errorf("%s must be a non-empty string", name)
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("%s must be a boolean", name)
			}
		case "string_array":
			if _, err := decodeStrings(value); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		case "profile_array":
			if _, err := decodeProfiles(value); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		case "json_object":
			if _, err := decodeJSONObject(value); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		case "integer":
			if _, ok := integerArg(value); !ok {
				return fmt.Errorf("%s must be an integer", name)
			}
		}
	}
	return nil
}

func decodeJSONObject(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, errors.New("must be a JSON object")
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return nil, errors.New("must be a JSON object")
	}
	return object, nil
}

func decodeProfiles(v any) ([]browser.ProfileInput, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var profiles []browser.ProfileInput
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}
	if profiles == nil {
		return nil, errors.New("must be an array of ProfileInput objects")
	}
	return profiles, nil
}
func decodeStrings(v any) ([]string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, errors.New("must be an array of strings")
	}
	if items == nil {
		return []string{}, nil
	}
	return items, nil
}
func stringArg(args map[string]any, key string) string {
	v, _ := args[key].(string)
	return strings.TrimSpace(v)
}
func boolArg(args map[string]any, key string) bool { v, _ := args[key].(bool); return v }
func integerArg(value any) (int, bool) {
	switch number := value.(type) {
	case int:
		return number, true
	case int64:
		return int(number), true
	case float64:
		if number == float64(int(number)) {
			return int(number), true
		}
	case json.Number:
		parsed, err := number.Int64()
		if err == nil {
			return int(parsed), true
		}
	}
	return 0, false
}
func camelCase(value string) string {
	parts := strings.Split(value, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
func newProfile(args map[string]any) browser.ProfileInput {
	var p browser.ProfileInput
	patchProfile(&p, args)
	return p
}
func patchProfile(p *browser.ProfileInput, args map[string]any) {
	if v, ok := args["name"]; ok {
		p.ProfileName = strings.TrimSpace(v.(string))
	}
	if v, ok := args["proxy_id"]; ok {
		p.ProxyId = strings.TrimSpace(v.(string))
		if p.ProxyId == "" {
			p.ProxyConfig = ""
		}
	}
	if v, ok := args["core_id"]; ok {
		p.CoreId = strings.TrimSpace(v.(string))
	}
	if v, ok := args["group_id"]; ok {
		p.GroupId = strings.TrimSpace(v.(string))
	}
	if v, ok := args["user_data_dir"]; ok {
		p.UserDataDir = strings.TrimSpace(v.(string))
	}
	if v, ok := args["tags"]; ok {
		p.Tags, _ = decodeStrings(v)
	}
	if v, ok := args["keywords"]; ok {
		p.Keywords, _ = decodeStrings(v)
	}
}

func InputFromProfile(p browser.Profile) browser.ProfileInput {
	return browser.ProfileInput{ProfileName: p.ProfileName, UserDataDir: p.UserDataDir, CoreId: p.CoreId, RestoreLastSession: p.RestoreLastSession, FingerprintArgs: p.FingerprintArgs, ProxyId: p.ProxyId, ProxyConfig: p.ProxyConfig, MemoryLimitMB: p.MemoryLimitMB, LaunchArgs: p.LaunchArgs, Tags: p.Tags, Keywords: p.Keywords, GroupId: p.GroupId}
}
func SafeProfile(p browser.Profile) map[string]any {
	return map[string]any{"profileId": p.ProfileId, "profileName": p.ProfileName, "coreId": p.CoreId, "proxyId": p.ProxyId, "groupId": p.GroupId, "tags": p.Tags, "keywords": p.Keywords, "running": p.Running, "debugReady": p.DebugReady, "debugPort": p.DebugPort, "pid": p.Pid, "launchCode": p.LaunchCode, "lastError": p.LastError}
}
func SafeProfiles(items []browser.Profile) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, SafeProfile(item))
	}
	return out
}
func SafeProxies(items []browser.Proxy) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"proxyId": item.ProxyId, "proxyName": item.ProxyName, "groupName": item.GroupName, "preferredKernel": item.PreferredKernel, "lastTestOk": item.LastTestOk, "lastLatencyMs": item.LastLatencyMs, "lastTestedAt": item.LastTestedAt})
	}
	return out
}
