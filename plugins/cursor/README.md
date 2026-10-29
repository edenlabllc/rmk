# RMK plugin for Cursor

Packages the RMK MCP server for [Cursor](https://cursor.com/) so the agent can call structured tools for project and
release inspection (and gated `project generate`).

Requires an installed `rmk` binary on `PATH`. See [MCP server mode](../../docs/mcp.md).

## What you get

- MCP server: `rmk mcp serve` (stdio)
- Skills: `rmk-inspect`, `rmk-generate`
- Commands: `inspect-project`, `list-releases`

## Install (local / from this repo)

1. Ensure `rmk` works: `rmk mcp status` from your project (for example `rmk-test`).
2. In Cursor, add this folder as a plugin from the repo path:

   `plugins/cursor`

   Or copy / symlink `plugins/cursor` into your Cursor plugins location and enable it.
3. Open the RMK project as the workspace root so tool cwd resolves correctly.
4. Confirm MCP tools such as `project_status` and `config_view` are available.

Marketplace one-click install may be published later; until then use this directory or manual [MCP config](../../docs/mcp.md).

## Layout

```text
cursor/
├── .cursor-plugin/plugin.json
├── mcp.json
├── skills/
├── commands/
└── README.md
```
