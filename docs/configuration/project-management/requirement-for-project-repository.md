# Requirement for project repository

1. The name of the project repository should consist of the following parts: `<project_name>`.`<custom_suffix>`.
   For example: `rmk-test.bootstrap.infra` or `rmk-test.infra`.
2. The project's repository exists within the [GitLab Flow](https://docs.gitlab.co.jp/ee/topics/gitlab_flow.html) only
   and therefor supports the following set of static branches:

   - `develop`
   - `staging`
   - `production`
   
   Each branch corresponds to its own environment with a separately deployed K8S cluster. RMK supports these branches 
   as well as the feature or release branches:

   - A feature branch should have the following naming: `feature/<issue_key>-<issue_number>-<issue_description>`.
     For example: `feature/FFS-1446-example`. RMK will use `<issue_key>` and `<issue_number>` as the feature cluster name.
   - A release branch should have the following naming: `release/<SemVer2>-rc` or `release/<SemVer2>`
     For example: `release/v1.0.0`. RMK will use the project name and the `<SemVer2>` tag as the release cluster name.

## Expected repository structure:

```yaml
etc/<upstream_project_name>/<environment>/secrets/
  .sops.yaml # The public key for the current set of secrets.
  .spec.yaml.gotmpl # The secrets template for generating new or rotating current secrets.
  <release name>.yaml # Values containing release secrets for a specific environment.
etc/<upstream_project_name>/<environment>/values/
  <release name>.yaml # Values containing release configuration for a specific environment.
  <release name>.yaml.gotmpl # Values containing the release configuration for a specific environment using the Golang templates.
etc/<upstream_project_name>/<environment>/
  releases.yaml # Release specification for installation of the charts.
  globals.yaml # Set of global values within a specific scope.
  globals.yaml.gotmpl # Set of global values within a specific scope using the Golang templates.
etc/<downstream_project_name>/<environment>/secrets/
  .sops.yaml  # -//-
  .spec.yaml.gotmpl # -//-
  <release name>.yaml # -//-
etc/<downstream_project_name>/<environment>/secrets/
  <release name>.yaml # -//-
  <release name>.yaml.gotmpl # -//-
etc/<downstream_project_name>/<environment>/
  releases.yaml # -//-
  globals.yaml # - // -
  globals.yaml.gotmpl # - // -
helmfile.yaml.gotmpl # Helmfile describing the release process for specific project releases using the Golang templates.
project.yaml # Project specification for the dependencies and inventory installed via RMK.
```

## Files for managing releases and their values at the scope level

### Requirement for `release.yaml`

```yaml
<release_name_foo>: # Required, release name from helmfile.yaml.gotmpl.
  enabled: true # Required, enable|disable release from helmfile.yaml.gotmpl.
  image: # Optional, needed when using a private container image with the automatic release update feature of RMK.
    repository: <full_container_images_repository_url>  
    tag: <container_images_tag>
<release_name_bar>: # -//-
  enabled: false # -//-
# ...
```

> releases.yaml cannot be used as a template, all the values must be defined.

### Requirement for `globals.yaml.gotmpl`

```yaml
# configs - enumeration of configurations divided into sets related to the Kubernetes ConfigMaps.
configs:
  auditLog: |
    {{- readFile (printf "%s/audit-log.json" "values/configs") | nindent 4 }}
  # ...

# envs - enumeration of environment variables divided into sets related to the Kubernetes environment variables for the containers.
envs:
  # The global environment variable used by multiple releases
  FOO: false
  # ...

# hooks - enumeration of environment variables divided into sets related to the Helmfile hooks arguments.
hooks:
  <release_name>:
    common-postuninstall-hook:
      events:
         - postuninstall
      showlogs: true
      command: "{{`{{ .Release.Labels.bin }}`}}/common-postuninstall-hook.sh"
      args:
         - "{{`{{ .Release.Namespace }}`}}"
  # ...
```

> globals.yaml.gotmpl is used in two cases:
> 
> 1. When values, configurations or environment variables need to be declared globally for multiple releases. 
> 2. When the current project is planned to be inherited by a downstream project and the overrides should be supported.

### Requirement for `helmfile.yaml.gotmpl`

The list of the `helmfile.yaml.gotmpl` sections that must be defined and remained unchanged for working with RMK correctly is:

```gotemplate
environments:
  local:
  develop:
    missingFileHandler: Warn
    values:
      - etc/<project_name>/{{ .Environment.Name }}/globals.yaml
      - etc/<project_name>/{{ .Environment.Name }}/globals.yaml.gotmpl
      - etc/<project_name>/{{ .Environment.Name }}/releases.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - etc/<project_name>/{{ .Environment.Name }}/values/k3d/releases.yaml
      {{- end }}
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/globals.yaml
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/globals.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/releases.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/values/k3d/releases.yaml
      {{- end }}
  production: 
    missingFileHandler: Warn
    values:
      - etc/<project_name>/{{ .Environment.Name }}/globals.yaml
      - etc/<project_name>/{{ .Environment.Name }}/globals.yaml.gotmpl 
      - etc/<project_name>/{{ .Environment.Name }}/releases.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - etc/<project_name>/{{ .Environment.Name }}/values/k3d/releases.yaml
      {{- end }}
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/globals.yaml
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/globals.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/releases.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/values/k3d/releases.yaml
      {{- end }}                        
  staging:
    missingFileHandler: Warn
    values:
      - etc/<project_name>/{{ .Environment.Name }}/globals.yaml
      - etc/<project_name>/{{ .Environment.Name }}/globals.yaml.gotmpl
      - etc/<project_name>/{{ .Environment.Name }}/releases.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - etc/<project_name>/{{ .Environment.Name }}/values/k3d/releases.yaml
      {{- end }}                     
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/globals.yaml
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/globals.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/releases.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - {{ requiredEnv "PWD" }}/etc/<project_name>/{{ .Environment.Name }}/values/k3d/releases.yaml
      {{- end }}
---
helmDefaults:
wait: true
waitForJobs: true
timeout: 3600
                                                                        
# The set of paths for the inherited Helmfiles is controlled through the project.yaml file using RMK.
# DO NOT EDIT the "helmfiles" field's values.
helmfiles: {{ env "HELMFILE_<project_name>_PATHS" }}

missingFileHandler: Warn

commonLabels:
  scope: <project_name>
  bin: {{ env "HELMFILE_<project_name>_HOOKS_DIR" }}/bin

templates:
  release:
    createNamespace: true
    labels:
      app: "{{`{{ .Release.Name }}`}}"
    missingFileHandler: Warn
    values:
      - etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/{{`{{ .Release.Name }}`}}.yaml.gotmpl
      - etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/{{`{{ .Release.Name }}`}}.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/k3d/values/{{`{{ .Release.Name }}`}}.yaml.gotmpl
      - etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/k3d/values/{{`{{ .Release.Name }}`}}.yaml
      {{- end }}
      - {{ requiredEnv "PWD" }}/etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/{{`{{ .Release.Name }}`}}.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/{{`{{ .Release.Name }}`}}.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - {{ requiredEnv "PWD" }}/etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/k3d/values/{{`{{ .Release.Name }}`}}.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/values/k3d/values/{{`{{ .Release.Name }}`}}.yaml
      {{- end }}
    secrets:
      - {{ requiredEnv "PWD" }}/etc/{{`{{ .Release.Labels.scope }}`}}/{{`{{ .Environment.Name }}`}}/secrets/{{`{{ .Release.Name }}`}}.yaml

releases:
  - name: <release_name_foo>
    installed: {{ .Values | get (print " <release_name_foo>" ".enabled") false }}
```

> You can use the [rmk project generate](preparation-of-project-repository.md#automatic-generation-of-the-project-structure-from-scratch) 
> command to view the full example of the contents of all the project files.
