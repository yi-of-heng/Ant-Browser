// antctl is the script-friendly client for Ant Browser's local Launch API.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"ant-chrome/backend/internal/controlclient"
	"ant-chrome/backend/internal/controlops"
)

func main() {
	base, key, args := globalFlags(os.Args[1:])
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	if err := run(context.Background(), controlclient.New(base, key), args); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func globalFlags(args []string) (string, string, []string) {
	f := flag.NewFlagSet("antctl", flag.ContinueOnError)
	f.SetOutput(os.Stderr)
	base := f.String("base-url", envOr("ANT_BROWSER_URL", "http://127.0.0.1:19876"), "Launch API 地址")
	key := f.String("api-key", os.Getenv("ANT_BROWSER_API_KEY"), "Launch API Key")
	_ = f.Parse(args)
	return *base, *key, f.Args()
}

func run(ctx context.Context, c *controlclient.Client, args []string) error {
	spec, consumed, ok := controlops.FindCLIPath(args)
	if !ok {
		usage()
		return errors.New("未知命令")
	}
	values, err := parseArgs(spec, args[consumed:])
	if err != nil {
		return err
	}
	result, err := controlops.Execute(ctx, c, spec.Name, values)
	if result != nil && (err == nil || spec.Name == "create_instances_batch") {
		data, marshalErr := json.MarshalIndent(result, "", "  ")
		if marshalErr != nil {
			return marshalErr
		}
		fmt.Println(string(data))
	}
	return err
}

func parseArgs(spec controlops.Spec, args []string) (map[string]any, error) {
	values := map[string]any{}
	for _, positional := range []string{"instance_id", "script_id"} {
		if !hasField(spec, positional) {
			continue
		}
		if len(args) == 0 || strings.HasPrefix(args[0], "-") {
			return nil, fmt.Errorf("用法: antctl %s %s [flags]", spec.CLI, strings.ToUpper(strings.TrimSuffix(positional, "_id"))+"_ID")
		}
		values[positional] = args[0]
		args = args[1:]
		break
	}
	if spec.Name == "create_instances_batch" {
		if len(args) != 1 {
			return nil, errors.New("用法: antctl profiles create-batch FILE.json")
		}
		data, err := os.ReadFile(args[0])
		if err != nil {
			return nil, err
		}
		var profiles []map[string]any
		if err := json.Unmarshal(data, &profiles); err != nil {
			return nil, fmt.Errorf("批量文件必须是 ProfileInput JSON 数组: %w", err)
		}
		values["profiles"] = profiles
		return values, nil
	}
	f := flag.NewFlagSet(spec.CLI, flag.ContinueOnError)
	f.SetOutput(os.Stderr)
	flagValues := map[string]*string{}
	boolValues := map[string]*bool{}
	intValues := map[string]*int{}
	for _, field := range spec.Fields {
		if field.Name == "instance_id" || field.Name == "script_id" {
			continue
		}
		name := strings.ReplaceAll(field.Name, "_", "-")
		switch field.Type {
		case "boolean":
			boolValues[field.Name] = f.Bool(name, false, field.Description)
		case "integer":
			intValues[field.Name] = f.Int(name, 0, field.Description)
		default:
			flagValues[field.Name] = f.String(name, "", field.Description)
		}
	}
	if err := f.Parse(args); err != nil {
		return nil, err
	}
	if f.NArg() != 0 {
		return nil, fmt.Errorf("多余参数: %s", strings.Join(f.Args(), " "))
	}
	f.Visit(func(found *flag.Flag) {
		key := strings.ReplaceAll(found.Name, "-", "_")
		if v, ok := boolValues[key]; ok {
			values[key] = *v
			return
		}
		if v, ok := intValues[key]; ok {
			values[key] = *v
			return
		}
		v := *flagValues[key]
		switch fieldType(spec, key) {
		case "string_array":
			parts := []string{}
			for _, part := range strings.Split(v, ",") {
				if trimmed := strings.TrimSpace(part); trimmed != "" {
					parts = append(parts, trimmed)
				}
			}
			values[key] = parts
		case "json_object":
			var object map[string]any
			if err := json.Unmarshal([]byte(v), &object); err != nil || object == nil {
				values[key] = json.RawMessage(v)
			} else {
				values[key] = object
			}
		default:
			values[key] = v
		}
	})
	return values, nil
}
func hasField(spec controlops.Spec, name string) bool {
	for _, field := range spec.Fields {
		if field.Name == name {
			return true
		}
	}
	return false
}
func fieldType(spec controlops.Spec, name string) string {
	for _, field := range spec.Fields {
		if field.Name == name {
			return field.Type
		}
	}
	return ""
}
func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func usage() {
	fmt.Fprintln(os.Stderr, "用法: antctl [--base-url URL] [--api-key KEY] COMMAND")
	for _, spec := range controlops.Specs() {
		fmt.Fprintf(os.Stderr, "  %-27s %s\n", spec.CLI, spec.Description)
	}
}
