# MCP server mode

RMK can run as a local [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server so AI assistants
([Cursor](https://cursor.com/), [Claude Desktop](https://claude.ai/download), and other MCP clients) can call a small,
safe set of RMK tools over **stdio**.

This is local process integration only (the client starts `rmk mcp serve`). Remote/HTTP MCP is out of scope for this
release.

## Prerequisites

- RMK installed and available as a binary (or built from source).
- An RMK project directory (`project.yaml` at the root). Prefer setting the MCP client's `cwd` to that project root.

## Commands

```shell
# Start the MCP server (stdio JSON-RPC on stdin/stdout)
rmk mcp serve

# Show binary path, project detection, and client config snippets
rmk mcp status
```

`rmk mcp serve` keeps application logs on **stderr** so stdout stays reserved for the MCP protocol.

## Tools (MVP)

| Tool | Wraps | Notes |
|------|--------|--------|
| `project_status` | — | Detects `project.yaml` / workdir |
| `config_view` | `rmk config view` | Current RMK config (no decrypted secrets) |
| `release_list` | `rmk release list -s` | Default output `json`; optional selectors |
| `release_template` | `rmk release template -s` | Read-only render |
| `release_show_dag` | `rmk release show-dag -s` | Helmfile v1.1.x |
| `project_generate` | `rmk project generate` | Requires `confirm=true` |

Destructive cluster operations are **not** exposed.

Optional tool argument `workdir` overrides the server process working directory for that call.

## Cursor

Add to `~/.cursor/mcp.json` or project `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "rmk": {
      "command": "rmk",
      "args": ["mcp", "serve"],
      "cwd": "/absolute/path/to/your/rmk-project"
    }
  }
}
```

Use the absolute path from `rmk mcp status` if `rmk` is not on the client `PATH`.

## Claude Desktop

Add the same `mcpServers.rmk` block to Claude Desktop's `claude_desktop_config.json`, then restart Claude Desktop.

Typical config locations:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`

## Verify

1. Run `rmk mcp status` in a project directory.
2. Register the printed snippet in Cursor or Claude Desktop.
3. In the client, confirm tools such as `project_status` and `config_view` are listed and callable.

## Follow-ups

Marketplace plugins / MCP bundles for one-click install, vector knowledge retrieval, and broader tool coverage are
tracked separately on the [product roadmap](index.md#roadmap).
