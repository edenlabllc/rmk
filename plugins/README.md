# RMK AI client plugins

Thin [Claude](https://claude.com/docs/plugins/overview) and [Cursor](https://cursor.com/docs/plugins) plugins that wrap the
RMK [MCP](https://modelcontextprotocol.io/) server (`rmk mcp serve`). Product logic stays in the RMK binary; plugins only
package MCP config plus optional skills and commands.

```text
plugins/
├── README.md          # this file
├── claude/            # Claude Code / claude.ai / Cowork plugin
└── cursor/            # Cursor IDE plugin
```

## Prerequisites

1. Install [RMK](https://github.com/edenlabllc/rmk/releases/latest) so `rmk` is on your `PATH` (or pass an absolute path — see `rmk mcp status`).
2. Open an RMK project as the workspace / working directory (`project.yaml` at the root). Example tenant name in docs: `rmk-test`.

## Install

- **Cursor:** see [cursor/README.md](cursor/README.md)
- **Claude:** see [claude/README.md](claude/README.md)

Manual MCP registration (without plugins) remains documented in [docs/mcp.md](../docs/mcp.md).

## Design

| Layer | Responsibility |
|-------|----------------|
| `rmk mcp serve` | Tools (`config_view`, `release_list`, …) |
| Plugin `mcp.json` / `.mcp.json` | Start that server over stdio |
| Skills / commands | Guide the agent to call the right tools |

Do not vendor a second MCP server implementation inside these plugins.
