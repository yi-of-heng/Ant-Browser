# Agent control

Ant Browser exposes two adapters over the same local Launch API:

- `antctl`: deterministic CLI for shell scripts and batch jobs.
- `ant-mcp`: MCP stdio server for an agent host such as Codex or Claude.

Both adapters use the shared operation catalog in
`backend/internal/controlops` and the typed HTTP client in
`backend/internal/controlclient`. They never open `app.db` directly. A tool
name and its arguments therefore have the same meaning in CLI and MCP.

## Build

```sh
go build -o ./build/antctl ./backend/cmd/antctl
go build -o ./build/ant-mcp ./backend/cmd/ant-mcp
```

## Operations

| Operation | CLI | MCP tool |
|---|---|---|
| Health check | `health` | `health` |
| List proxy nodes | `proxies list` | `list_proxy_nodes` |
| List instances | `profiles list` | `list_instances` |
| Get an instance | `profiles get INSTANCE_ID` | `get_instance` |
| Create an instance | `profiles create` | `create_instance` |
| Update an instance | `profiles update INSTANCE_ID` | `update_instance` |
| Start/stop an instance | `profiles start/stop INSTANCE_ID` | `start_instance` / `stop_instance` |
| Delete an instance | `profiles delete INSTANCE_ID --confirm` | `delete_instance` (`confirm=true`) |
| Copy an instance | `profiles copy INSTANCE_ID` | `copy_instance` |
| Batch create | `profiles create-batch FILE.json` | `create_instances_batch` |
| List/get/run automation | `automation scripts list/get/run` | `list_automation_scripts` / `get_automation_script` / `run_automation_script` |
| Automation run history | `automation runs list` | `list_automation_runs` |

The CLI uses kebab-case flags (`--proxy-id`); MCP uses the corresponding
snake-case argument (`proxy_id`). Sensitive connection strings, local paths,
launch arguments and fingerprint arguments are not returned by either adapter.

## CLI examples

```sh
export ANT_BROWSER_URL=http://127.0.0.1:19876
export ANT_BROWSER_API_KEY=API_KEY   # only needed when API auth is enabled

./build/antctl health
./build/antctl proxies list
./build/antctl profiles list
./build/antctl profiles get PROFILE_ID
./build/antctl profiles create --name "Google-01" --proxy-id NODE_ID --group-id google --tags oauth,google
./build/antctl profiles update PROFILE_ID --proxy-id NODE_ID --keywords account-1
./build/antctl profiles start PROFILE_ID
./build/antctl profiles stop PROFILE_ID
./build/antctl profiles delete PROFILE_ID --confirm
./build/antctl profiles copy PROFILE_ID --name "Google-01-cycle-002" --auto-launch
./build/antctl profiles create-batch ./profiles.json
./build/antctl automation scripts list
./build/antctl automation scripts run SCRIPT_ID --selector '{"profileId":"PROFILE_ID"}' --timeout-ms 300000
./build/antctl automation runs list --limit 20
```

`profiles create-batch` accepts a JSON array of `browser.ProfileInput` objects.
It returns the successfully created items and a `failedIndex` if a later item
fails; already-created items are intentionally not rolled back.

## MCP configuration

Register the built binary as a local stdio MCP server:

```json
{
  "mcpServers": {
    "ant-browser": {
      "command": "/ABSOLUTE/PATH/TO/Ant-Browser/build/ant-mcp",
      "env": {
        "ANT_BROWSER_URL": "http://127.0.0.1:19876",
        "ANT_BROWSER_API_KEY": "API_KEY"
      }
    }
  }
}
```

Keep the server on localhost and enable the Launch API key before exposing it
outside the machine. Proxy CRUD, proxy health testing, extension management,
cores, groups and backups remain outside this shared operation set. The
automation operations call the existing Launch API script runner; they do not
create a second automation runtime.
