# Preparation of the project repository

## Prerequisites:

- Create a remote repository in your Version Control System (GitHub) according to the
  following [requirements](requirement-for-project-repository.md#requirement-for-project-repository).
- Clone the project repository. For example: **rmk-test.bootstrap.infra**
  OR `git init && git remote add <name> <url> && git commit -m "init commit"`
- Checkout the needed branch. For example: `develop|staging|production`.
- Make sure there is a file in the root of the repository named [project.yaml](#projectyaml), which contains the project
  configuration.
- [Initialize the configuration](../configuration-management/configuration-management.md#initialization-of-rmk-configuration).

## Automatic generation of the project structure from scratch

RMK supports automatic generation of the project structure from scratch, according to the presented project
specification described in [project.yaml](#projectyaml) file.

Use the following command:

```shell
rmk project generate
```

> Add the `--create-sops-age-keys` flag if you want to create the project structure along with SOPS age private keys.

This will create a default project structure and prepare an example release based on [Nginx](https://nginx.org/).

## project.yaml

The `project.yaml` file is the main configuration file of the repository, the file is used by RMK
and contains the following main sections:

[//]: # (  TODO ACTUALIZE)

* `project`: Optional, contains a list of dependencies of the upstream project's repositories and the project
  specification.

  ```yaml
  project:
    # Optional, needed if you want to add the dependencies with upstream projects to the downstream project.
    dependencies:
        # Required, dependencies upstream project's repository name.
      - name: <upstream_repository_prefix>.bootstrap.infra
        # Required, dependencies upstream project's repository version in `SemVer2` format, also can be a branch name or a commit hash.
        version: <SemVer2>
        # Required, dependencies upstream project's repository URL.
        url: git::https://github.com/<owner>/{{.Name}}.git?ref={{.Version}}    
    # Optional, needed if you want automatic generation of the project structure from scratch.
    spec:
      # Required, list of available environments of the project (Git branches). 
      environments:
        - develop
        - staging
        - production
      # Optional, list of owners of the project.
      owners:
        - <owner_1>
        - <owner_2>
      # Required, list of available scope of the project.
      scopes:
        - clusters
        - <upstream_project_name>
        - <downstream_project_name>
  # ... 
  ```

[//]: # (  TODO ACTUALIZE)

* `inventory`: Optional, contains a map of the extra configurations required to launch the project.

  ```yaml
  inventory:
    # Optional, contains a map of the Helm plugins repositories.
    helm-plugins:
      # Optional, Helm plugin name.
      diff:
        # Required, Helm plugin version in the `SemVer2` format.
        version: <SemVer2>
        # Required, Helm plugin repository URL.
        url: https://github.com/<owner>/helm-diff
      # ...
    # Optional, contains a map of the Helmfile hooks repositories with shell scripts.
    hooks:
      # Optional, Helmfile hooks repository name.
      helmfile.hooks.infra:
        # Required, Helmfile hooks repository version in the `SemVer2` format.
        version: <SemVer2>
        # Required, Helmfile hooks repository URL.
        url: git::https://github.com/<owner>/{{.Name}}.git?ref={{.Version}}
    # Optional, contains a map of the sources of binary file tools.
    tools:
      # Optional, tool name.
      terraform:
        # Required, tool version in `SemVer2` format.
        version: <SemVer2>
        # Required, tool source URL.
        url: https://releases.hashicorp.com/{{.Name}}/{{.Version}}/{{.Name}}_{{.Version}}_{{.Os}}_amd64.zip
        # Optional, tool checksum source URL.
        checksum: https://releases.hashicorp.com/{{.Name}}/{{.Version}}/{{.Name}}_{{.Version}}_SHA256SUMS
        # Optional, specific key overrides for the described OS name.
        os-linux: linux
        os-mac: darwin
      # ...
  ```

<details>
  <summary>Example of the full <code>project.yaml</code> file.</summary>

[//]: # (  TODO ACTUALIZE)

```yaml
project:
  dependencies:
    - name: deps.bootstrap.infra
      version: v2.17.0
      url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
  spec:
    environments:
      - develop
      - staging
      - production
    owners:
      - owner1
      - owner2
    scopes:
      - clusters
      - deps
      - project1
inventory:
  helm-plugins:
    diff:
      version: v3.8.1
      url: https://github.com/databus23/helm-diff
    secrets:
      version: v4.5.0
      url: https://github.com/jkroepke/helm-secrets
  hooks:
    helmfile.hooks.infra:
      version: v1.18.0
      url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
  tools:
    terraform:
      version: 1.0.2
      url: https://releases.hashicorp.com/{{.Name}}/{{.Version}}/{{.Name}}_{{.Version}}_{{.Os}}_amd64.zip
      checksum: https://releases.hashicorp.com/{{.Name}}/{{.Version}}/{{.Name}}_{{.Version}}_SHA256SUMS
      os-linux: linux
      os-mac: darwin
    kubectl:
      version: 1.27.6
      url: https://dl.k8s.io/release/v{{.Version}}/bin/{{.Os}}/amd64/{{.Name}}
      checksum: https://dl.k8s.io/release/v{{.Version}}/bin/{{.Os}}/amd64/{{.Name}}.sha256
      os-linux: linux
      os-mac: darwin
    helm:
      version: 3.10.3
      url: https://get.helm.sh/{{.Name}}-v{{.Version}}-{{.Os}}-amd64.tar.gz
      checksum: https://get.helm.sh/{{.Name}}-v{{.Version}}-{{.Os}}-amd64.tar.gz.sha256sum
      os-linux: linux
      os-mac: darwin
    helmfile:
      version: 0.157.0
      url: https://github.com/{{.Name}}/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}_{{.Version}}_{{.Os}}_amd64.tar.gz
      checksum: https://github.com/{{.Name}}/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}_{{.Version}}_checksums.txt
      os-linux: linux
      os-mac: darwin
    jq:
      version: 1.7
      url: https://github.com/jqlang/{{.Name}}/releases/download/{{.Name}}-{{.Version}}/{{.Name}}-{{.Os}}
      os-linux: linux-amd64
      os-mac: macos-amd64
      rename: true
    sops:
      version: 3.8.1
      url: https://github.com/getsops/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}-v{{.Version}}.{{.Os}}
      os-linux: linux.amd64
      os-mac: darwin
      rename: true
    age:
      version: 1.1.1
      url: https://github.com/FiloSottile/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}-v{{.Version}}-{{.Os}}-amd64.tar.gz
      os-linux: linux
      os-mac: darwin
    k3d:
      version: 5.6.0
      url: https://github.com/k3d-io/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}-{{.Os}}-amd64
      os-linux: linux
      os-mac: darwin
      rename: true
    yq:
      version: 4.35.2
      url: https://github.com/mikefarah/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}_{{.Os}}_amd64
      os-linux: linux
      os-mac: darwin
      rename: true
```

</details>

The project file supports placeholders, they are required for correct URL formation.

* **{{.Name}}:** Replaced with the `name` field.
* **{{.Version}}:** Replaced with the `version` field.
* **{{.HelmfileTenant}}:** Replaced with the tenant name for the Helmfile selected from the list.
* **{{.Os}}:** Replaced with the values from the `os-linux`, `os-mac` fields according to the specific operating system,
  where RMK is run.

> The field `rename` of the boolean type is required to correct the name of the binary file of the downloaded tool
> according to the value of the `name` field. This is mainly required for the cases, when the artifact is not the
> archive.
> For example:
>
> - The initial file name after the download: `helmfile_darwin_amd64`.
> - After applying the `rename` instruction it gets a value of the `name` field: `helmfile`.
