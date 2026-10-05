# Ant Browser control surface

## Base URL and auth

Default base URL:

```text
http://127.0.0.1:19876
```

When API auth is enabled, send `X-Ant-Api-Key: $ANT_BROWSER_API_KEY`.
Keep the API local unless authentication and network access restrictions are
configured deliberately.

## Launch API

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/health` | Health check |
| GET | `/api/profiles` | List profiles |
| POST | `/api/profiles` | Create profile; supports `autoLaunch` |
| GET | `/api/profiles/{id}` | Get profile |
| PUT | `/api/profiles/{id}` | Update profile |
| POST | `/api/profiles/{id}/copy` | Copy profile; source is retained |
| GET | `/api/profiles/{id}/status` | Runtime status |
| POST | `/api/profiles/{id}/stop` | Stop profile |
| DELETE | `/api/profiles/{id}` | Delete a stopped profile |
| GET | `/api/proxies` | List proxy metadata |
| POST | `/api/launch` | Start by selector/profile ID |
| GET | `/api/launch/{code}` | Start by launch code |
| GET | `/api/runtime/active` | Active unified CDP target |
| POST | `/api/runtime/session` | Start/wait for a runtime session |
| POST | `/api/runtime/status` | Runtime status by selector |
| POST | `/api/runtime/stop` | Stop by selector |

Copy request:

```json
{
  "name": "slot-01-20261005-001",
  "mode": "auto_fingerprint",
  "autoLaunch": false
}
```

The response includes `sourceProfileId`, `profileId`, and a full `profile`
object. Agent-facing CLI/MCP output redacts sensitive profile fields.

## Automation API

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/automation/scripts` | List script metadata |
| GET | `/api/automation/scripts/{scriptId}` | Script detail |
| POST | `/api/automation/scripts/run` | Execute a script |
| GET | `/api/automation/scripts/runs?limit=20` | Recent run records |
| POST | `/api/automation/hooks/{path}` | Optional public script hook |

Run request example:

```json
{
  "scriptId": "SCRIPT_ID",
  "selector": {"profileId": "PROFILE_ID"},
  "params": {},
  "timeoutMs": 300000
}
```

`playwright-cdp` scripts start the selected profile when needed. The API
accepts timeouts from 1 second through 30 minutes.

## Shared CLI/MCP catalog

| Operation | CLI | MCP |
|---|---|---|
| Health | `health` | `health` |
| Proxy list | `proxies list` | `list_proxy_nodes` |
| Profile list/get | `profiles list/get ID` | `list_instances` / `get_instance` |
| Profile create/update | `profiles create/update ID` | `create_instance` / `update_instance` |
| Profile copy | `profiles copy ID` | `copy_instance` |
| Profile start/stop/delete | `profiles start/stop/delete ID` | `start_instance` / `stop_instance` / `delete_instance` |
| Batch create | `profiles create-batch FILE` | `create_instances_batch` |
| Automation scripts | `automation scripts list/get/run` | `list_automation_scripts` / `get_automation_script` / `run_automation_script` |
| Automation runs | `automation runs list` | `list_automation_runs` |

The CLI and MCP call `controlclient`, which calls these HTTP endpoints. They do
not contain a second database or automation runtime.
