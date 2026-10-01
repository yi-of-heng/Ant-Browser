# Agent control

Ant Browser now exposes two scriptable clients over the local Launch API:

- `antctl`: deterministic CLI for shell scripts and batch jobs.
- `ant-mcp`: MCP stdio server for an agent host such as Codex or Claude.

Both clients use the same HTTP API and never open `app.db` directly.

## Build

```sh
export PATH="/opt/homebrew/opt/go@1.22/bin:/opt/homebrew/bin:$PATH"
go build -o ./build/antctl ./backend/cmd/antctl
go build -o ./build/ant-mcp ./backend/cmd/ant-mcp
```

## CLI examples

```sh
./build/antctl health
./build/antctl proxies list
./build/antctl profiles list
./build/antctl profiles create --name "Google-01" --proxy-id NODE_ID --group-id google
./build/antctl profiles update PROFILE_ID --proxy-id NODE_ID
./build/antctl profiles start PROFILE_ID
./build/antctl profiles stop PROFILE_ID
./build/antctl profiles create-batch ./profiles.json
```

`profiles create-batch` accepts a JSON array of `browser.ProfileInput` objects.
The command exits on the first failed item and prints created profiles as JSON.

Use `--base-url` and `--api-key`, or set `ANT_BROWSER_URL` and
`ANT_BROWSER_API_KEY`.

## MCP configuration

Build `ant-mcp` and register it as a local stdio MCP server:

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

Current tools: `list_instances`, `list_proxy_nodes`, `create_instance`,
`update_instance`, `start_instance`, `stop_instance`, and
`delete_instance`. Deletion requires `confirm=true`.

MCP responses redact proxy connection strings, user-data paths, launch args,
and fingerprint arguments. Keep the server on localhost and enable the
Launch API key before exposing it outside the machine.
