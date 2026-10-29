---
name: rmk-generate
description: >-
  Generate an RMK project layout via MCP. Use when the user asks to scaffold
  or regenerate project directories with rmk project generate.
---

# Generate an RMK project

Use the `project_generate` MCP tool only. Do not invent file trees by hand.

## Steps

1. Confirm the target workdir (project root). Call `project_status` first.
2. Collect optional `scope` and `environment` lists from the user when needed.
3. Call `project_generate` with `confirm` set to **true** only after the user explicitly asks to generate.
4. Never set `create_sops_age_keys` unless the user clearly requests key creation.
5. Report tool output; on failure, suggest `rmk mcp status` and a local `rmk project generate` check.

## Safety

- `confirm=false` or omitting confirm must not be overridden by the agent.
- Do not expose or create secrets unless the user explicitly opts in.
