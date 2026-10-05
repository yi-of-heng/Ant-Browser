// Package controlops defines the control operations shared by antctl and ant-mcp.
// Both adapters use the same names, argument schemas and execution path.
package controlops

import "strings"

type Field struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

type Spec struct {
	Name        string
	CLI         string
	Description string
	Fields      []Field
}

var specs = []Spec{
	{"health", "health", "检查 Launch API", nil},
	{"list_proxy_nodes", "proxies list", "列出代理节点", nil},
	{"list_instances", "profiles list", "列出浏览器实例", nil},
	{"get_instance", "profiles get", "查询浏览器实例", []Field{{"instance_id", "string", "实例 ID", true}}},
	{"create_instance", "profiles create", "创建浏览器实例", []Field{{"name", "string", "实例名称", true}, {"proxy_id", "string", "代理节点 ID", false}, {"core_id", "string", "内核 ID", false}, {"group_id", "string", "分组 ID", false}, {"user_data_dir", "string", "用户数据目录", false}, {"tags", "string_array", "标签列表", false}, {"keywords", "string_array", "关键字列表", false}, {"auto_launch", "boolean", "创建后启动", false}}},
	{"update_instance", "profiles update", "修改浏览器实例", []Field{{"instance_id", "string", "实例 ID", true}, {"name", "string", "实例名称", false}, {"proxy_id", "string", "代理节点 ID，可传空字符串清除", false}, {"core_id", "string", "内核 ID", false}, {"group_id", "string", "分组 ID", false}, {"user_data_dir", "string", "用户数据目录", false}, {"tags", "string_array", "标签列表", false}, {"keywords", "string_array", "关键字列表", false}, {"auto_launch", "boolean", "修改后启动", false}}},
	{"start_instance", "profiles start", "启动浏览器实例", []Field{{"instance_id", "string", "实例 ID", true}}},
	{"stop_instance", "profiles stop", "停止浏览器实例", []Field{{"instance_id", "string", "实例 ID", true}}},
	{"delete_instance", "profiles delete", "删除已停止的浏览器实例", []Field{{"instance_id", "string", "实例 ID", true}, {"confirm", "boolean", "必须明确确认", true}}},
	{"create_instances_batch", "profiles create-batch", "批量创建浏览器实例；失败时返回已创建项", []Field{{"profiles", "profile_array", "ProfileInput 对象数组", true}}},
}

func Specs() []Spec { return append([]Spec(nil), specs...) }

func Find(name string) (Spec, bool) {
	for _, spec := range specs {
		if spec.Name == name {
			return spec, true
		}
	}
	return Spec{}, false
}

func FindCLI(group, command string) (Spec, bool) {
	key := strings.TrimSpace(group + " " + command)
	for _, spec := range specs {
		if spec.CLI == key {
			return spec, true
		}
	}
	return Spec{}, false
}

func Tools() []map[string]any {
	out := make([]map[string]any, 0, len(specs))
	for _, spec := range specs {
		out = append(out, map[string]any{"name": spec.Name, "description": spec.Description, "inputSchema": InputSchema(spec)})
	}
	return out
}

func InputSchema(spec Spec) map[string]any {
	properties := map[string]any{}
	required := []string{}
	for _, field := range spec.Fields {
		property := map[string]any{"description": field.Description}
		switch field.Type {
		case "string_array":
			property["type"] = "array"
			property["items"] = map[string]any{"type": "string"}
		case "profile_array":
			property["type"] = "array"
			property["items"] = map[string]any{
				"type": "object", "required": []string{"profileName"},
				"properties": map[string]any{
					"profileName": map[string]any{"type": "string"}, "proxyId": map[string]any{"type": "string"},
					"coreId": map[string]any{"type": "string"}, "groupId": map[string]any{"type": "string"},
					"userDataDir": map[string]any{"type": "string"},
					"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"keywords":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			}
		default:
			property["type"] = field.Type
		}
		properties[field.Name] = property
		if field.Required {
			required = append(required, field.Name)
		}
	}
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}
