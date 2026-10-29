---
name: rmk-inspect
description: >-
  Inspect an RMK project using MCP tools. Use when the user asks about RMK
  config, releases, Helmfile DAGs, project.yaml, or the state of an rmk-test
  (or other) tenant project.
---

# Inspect an RMK project

Prefer RMK MCP tools over guessing from open files. Work in the project root
(directory that contains `project.yaml`).

## Steps

1. Call `project_status` to confirm `project.yaml` and the workdir.
2. Call `config_view` for the selected environment configuration.
3. Call `release_list` (default JSON) for releases; pass `selector` when the user names labels.
4. Call `release_show_dag` or `release_template` only when the user needs dependency order or rendered manifests.
5. Summarize clearly. Do **not** decrypt secrets or request SOPS keys.

## Notes

- If tools fail with “config not initialized”, tell the user to run `rmk config init` in the project.
- Optional tool argument `workdir` overrides cwd when the workspace root is not the RMK project.
