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

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/controlclient"
)

func main() {
	base, key, args := globalFlags(os.Args[1:])
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	c := controlclient.New(base, key)
	ctx := context.Background()
	if err := run(ctx, c, args); err != nil {
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
	if args[0] == "health" {
		return c.Health(ctx)
	}
	if args[0] == "proxies" && len(args) == 2 && args[1] == "list" {
		items, err := c.ListProxies(ctx)
		return printResult(items, err)
	}
	if args[0] != "profiles" || len(args) < 2 {
		usage()
		return errors.New("未知命令")
	}
	switch args[1] {
	case "list":
		items, err := c.ListProfiles(ctx)
		return printResult(items, err)
	case "get":
		if len(args) != 3 {
			return errors.New("用法: profiles get PROFILE_ID")
		}
		item, err := c.GetProfile(ctx, args[2])
		return printResult(item, err)
	case "start":
		if len(args) != 3 {
			return errors.New("用法: profiles start PROFILE_ID")
		}
		item, err := c.StartProfile(ctx, args[2])
		return printResult(item, err)
	case "stop":
		if len(args) != 3 {
			return errors.New("用法: profiles stop PROFILE_ID")
		}
		return c.StopProfile(ctx, args[2])
	case "delete":
		if len(args) != 3 {
			return errors.New("用法: profiles delete PROFILE_ID")
		}
		return c.DeleteProfile(ctx, args[2])
	case "create":
		return create(ctx, c, args[2:])
	case "update":
		return update(ctx, c, args[2:])
	case "create-batch":
		return createBatch(ctx, c, args[2:])
	default:
		usage()
		return errors.New("未知 profiles 命令")
	}
}

func create(ctx context.Context, c *controlclient.Client, args []string) error {
	f := flag.NewFlagSet("profiles create", flag.ContinueOnError)
	f.SetOutput(os.Stderr)
	in := profileFlags(f)
	auto := f.Bool("auto-launch", false, "创建后启动")
	if err := f.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*in.name) == "" {
		return errors.New("--name 不能为空")
	}
	item, err := c.CreateProfile(ctx, in.input(), *auto)
	return printResult(item, err)
}

func update(ctx context.Context, c *controlclient.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("用法: profiles update PROFILE_ID [flags]")
	}
	id := args[0]
	f := flag.NewFlagSet("profiles update", flag.ContinueOnError)
	f.SetOutput(os.Stderr)
	in := profileFlags(f)
	auto := f.Bool("auto-launch", false, "更新后启动")
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	current, err := c.GetProfile(ctx, id)
	if err != nil {
		return err
	}
	merged := inputFromProfile(current)
	in.applyTo(&merged)
	item, err := c.UpdateProfile(ctx, id, merged, *auto)
	return printResult(item, err)
}

func createBatch(ctx context.Context, c *controlclient.Client, args []string) error {
	if len(args) != 1 {
		return errors.New("用法: profiles create-batch FILE.json")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var inputs []browser.ProfileInput
	if err = json.Unmarshal(data, &inputs); err != nil {
		return fmt.Errorf("批量文件必须是 ProfileInput JSON 数组: %w", err)
	}
	results := make([]browser.Profile, 0, len(inputs))
	for _, in := range inputs {
		item, e := c.CreateProfile(ctx, in, false)
		if e != nil {
			return e
		}
		results = append(results, item)
	}
	return printResult(results, nil)
}

type profileFlagValues struct{ name, proxy, core, group, userData, tags *string }

func profileFlags(f *flag.FlagSet) *profileFlagValues {
	v := &profileFlagValues{}
	v.name = f.String("name", "", "实例名称")
	v.proxy = f.String("proxy-id", "", "代理节点 ID")
	v.core = f.String("core-id", "", "内核 ID")
	v.group = f.String("group-id", "", "分组 ID")
	v.userData = f.String("user-data-dir", "", "用户数据目录")
	v.tags = f.String("tags", "", "逗号分隔标签")
	return v
}
func (v *profileFlagValues) input() browser.ProfileInput {
	in := browser.ProfileInput{ProfileName: *v.name, ProxyId: *v.proxy, CoreId: *v.core, GroupId: *v.group, UserDataDir: *v.userData}
	in.Tags = split(*v.tags)
	return in
}
func (v *profileFlagValues) applyTo(in *browser.ProfileInput) {
	if *v.name != "" {
		in.ProfileName = *v.name
	}
	if *v.proxy != "" {
		in.ProxyId = *v.proxy
	}
	if *v.core != "" {
		in.CoreId = *v.core
	}
	if *v.group != "" {
		in.GroupId = *v.group
	}
	if *v.userData != "" {
		in.UserDataDir = *v.userData
	}
	if *v.tags != "" {
		in.Tags = split(*v.tags)
	}
}
func inputFromProfile(p browser.Profile) browser.ProfileInput {
	return browser.ProfileInput{ProfileName: p.ProfileName, UserDataDir: p.UserDataDir, CoreId: p.CoreId, RestoreLastSession: p.RestoreLastSession, FingerprintArgs: p.FingerprintArgs, ProxyId: p.ProxyId, ProxyConfig: p.ProxyConfig, MemoryLimitMB: p.MemoryLimitMB, LaunchArgs: p.LaunchArgs, Tags: p.Tags, Keywords: p.Keywords, GroupId: p.GroupId}
}
func split(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
func printResult(v any, err error) error {
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
	return nil
}
func envOr(k, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return fallback
}
func usage() {
	fmt.Fprintln(os.Stderr, `antctl profiles list|get|create|update|start|stop|delete|create-batch
       antctl proxies list
       全局参数: --base-url URL --api-key KEY`)
}
