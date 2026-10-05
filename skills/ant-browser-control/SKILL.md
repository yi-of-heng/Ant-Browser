---
name: ant-browser-control
description: Use when an agent needs to inspect, invoke, extend, or troubleshoot Ant Browser's Launch API, antctl CLI, ant-mcp server, browser profiles, proxy bindings, or automation scripts. Prefer this skill before adding orchestration around Ant Browser so the agent uses the existing public control path.
---

# Ant Browser Control

Use Ant Browser as the browser/profile execution layer and keep orchestration in
the caller. Read `references/api-surface.md` when a task involves API, CLI, or
MCP capability discovery.

## Non-negotiable architecture

- Route control through the local Launch API; do not open or mutate the SQLite
  database from an agent, script, CLI command, or MCP tool.
- Prefer stable `profileId` selectors for automation. Use tags/group/keywords
  only when a deliberate multi-profile selector is required.
- Keep API credentials in environment variables (`ANT_BROWSER_API_KEY`), never
  in source, profile names, script params, or committed files.
- Treat `userDataDir`, proxy connection strings, fingerprint args, launch args,
  cookies, and script params as sensitive. Do not print them unless the user
  explicitly needs them.
- Before starting or stopping a profile, query its current status when the
  workflow can race with another controller.

## Fast capability check

```bash
antctl health
antctl profiles list
antctl proxies list
antctl automation scripts list
```

For MCP, call `tools/list`; the operation names should match the shared
catalog in `backend/internal/controlops`.

## Profile replacement workflow

For a new persistent instance while retaining the source profile:

1. `POST /api/profiles/{profileId}/copy` (or `antctl profiles copy`).
2. Use the returned `profile.profileId`; never infer it from the name.
3. Start it or request `autoLaunch`.
4. Run the automation script against that new ID.
5. Stop the old profile only after the replacement is ready.
6. Never delete the old profile unless the user explicitly requests retention
   cleanup.

`auto_fingerprint` is the safe default copy mode. `regular` intentionally
retains the source fingerprint args and should only be selected deliberately.
Copying creates a fresh user-data directory, but the copied profile retains the
source proxy binding. Rotating a provider port later changes the exit IP seen
by any profile using that port, including stopped historical profiles if they
are started again.

## Automation workflow

- List scripts with `GET /api/automation/scripts` or
  `antctl automation scripts list`.
- Run a script with `POST /api/automation/scripts/run` or
  `antctl automation scripts run SCRIPT_ID`.
- Pass `selector: {"profileId": "..."}` for an existing profile.
- A script configured with target mode `create` can clone a template, but use
  the explicit profile-copy endpoint when the caller must persist the new ID
  and coordinate old-profile shutdown.
- The Launch API executes automation; a recurring scheduler/slot supervisor is
  an external responsibility unless the user asks to add one.

## Validation after changes

Run focused tests first:

```bash
go test ./backend/cmd/antctl ./backend/cmd/ant-mcp \
  ./backend/internal/controlops ./backend/internal/controlclient \
  ./backend/internal/launchcode
go build -o build/antctl ./backend/cmd/antctl
go build -o build/ant-mcp ./backend/cmd/ant-mcp
```

Then check MCP `tools/list` and a local `antctl health`. Full repository tests
may contain unrelated backup-import failures; report those separately rather
than weakening this control surface.
