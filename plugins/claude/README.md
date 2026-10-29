# RMK plugin for Claude

Packages the RMK MCP server for [Claude](https://claude.ai/) (Claude Code, claude.ai, Cowork) so the assistant can call
structured tools for project and release inspection (and gated `project generate`).

Requires an installed `rmk` binary on `PATH`. See [MCP server mode](../../docs/mcp.md).

This plugin starts a **local** stdio MCP server (`rmk mcp serve`). Local servers load in Claude Code and in
computer-local Cowork sessions; claude.ai chat may only list remote connectors from `.mcp.json`.

## What you get

- MCP server: `rmk mcp serve` (stdio)
- Skills: `rmk-inspect`, `rmk-generate`
- Commands: `inspect-project`, `list-releases`

## Install

### Claude Code

From the repository root:

```shell
claude --plugin-dir ./plugins/claude
```

Or add a marketplace / install the `plugins/claude` folder per [Claude plugin docs](https://claude.com/docs/plugins/build).

### claude.ai / Cowork (upload)

1. Zip the `plugins/claude` directory (manifest must be at `.claude-plugin/plugin.json` inside the archive).
2. Customize → Plugins → Add → Upload plugin.
3. Connect / enable the bundled RMK MCP entry where the UI offers it.
4. Open a session with cwd set to your RMK project (for example `rmk-test`).

Validate (if Claude Code is installed):

```shell
claude plugin validate ./plugins/claude
```

## Layout

```text
claude/
├── .claude-plugin/plugin.json
├── .mcp.json
├── skills/
├── commands/
└── README.md
```
