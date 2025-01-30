# Preparation of the project repository

## Prerequisites

- Create a remote repository (Git) for a project in your Version Control System (e.g., [GitHub](https://github.com))
  according to the [requirements](requirement-for-project-repository.md#requirement-for-project-repository). For
  example: `rmk-test.bootstrap.infra`
- Clone the existing project repository:

  ```shell
  git clone <repo_url>
  ```

  Alternatively, initialize a new repository manually:

  ```shell
  git init
  git remote add <repo_name> <repo_url>
  git commit --allow-empty --message "Initial commit"
  ```

  > RMK requires a Git branch with at least one commit and a configured `origin` remote  
  > to correctly resolve the project/tenant name and environment.

- Checkout the required branch. For example: `develop`.

## Automatic generation of the project structure from scratch

RMK supports automatic generation of the project structure from scratch, according to the presented project
specification described in [project.yaml](#projectyaml) file.

Use the following command:

```shell
rmk project generate --environment="develop.root-domain=localhost" \
  --owner=gh_user --scope=<upstream_project_name> \
  --scope=<downstream_project_name> 
```

> Add the `--create-sops-age-keys` flag if you want to create the project structure along with SOPS age private keys.

This will create a default project structure, `project.yaml` file and prepare an example release based on [Nginx](https://nginx.org/).

> Check [Quickstart](../../quickstart.md) for a detailed example.

## project.yaml

The `project.yaml` file is the main configuration file of the repository, the file is used by RMK
and contains the following main sections:

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
      # Required, list of available environments with specific root domain name (Git branches). 
      environments:
        - develop:
            root-domain: localhost
        - production:
            root-domain: <custom_name>.example.com
        - staging:
            root-domain: <custom_name>.example.com
      # Optional, list of owners of the project.
      owners:
        - <owner_1>
        - <owner_2>
      # Required, list of available scope of the project.
      scopes:
        - <upstream_project_name>
        - <downstream_project_name>
  # ... 
  ```

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
      clusterctl:
        # Required, tool version in `SemVer2` format.
        version: <SemVer2>
        # Required, tool source URL.
        url: https://github.com/kubernetes-sigs/cluster-api/releases/download/v{{.Version}}/{{.Name}}-{{.Os}}-amd64
        # Optional, specific key overrides for the described OS name.
        os-linux: linux
        os-mac: darwin
        # Optional, an option that allows to rename the downloaded binary file by the tool name.
        rename: true
      # ...
  ```

<details>
  <summary>Example of the full <code>project.yaml</code> file.</summary>

```yaml
project:
  dependencies:
    - name: cluster-deps.bootstrap.infra
      version: v0.1.0
      url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
  spec:
    environments:
      - develop:
          root-domain: localhost
      - production:
          root-domain: localhost
      - staging:
          root-domain: localhost
    owners:
      - owner1
      - owner2
    scopes:
      - deps
      - project1
inventory:
  helm-plugins:
    diff:
      version: v3.8.1
      url: https://github.com/databus23/helm-diff
    helm-git:
      version: v0.15.1
      url: https://github.com/aslafy-z/helm-git
    secrets:
      version: v4.5.0
      url: https://github.com/jkroepke/helm-secrets
  hooks:
    helmfile.hooks.infra:
      version: v1.29.1
      url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
  tools:
    clusterctl:
      version: 1.7.4
      url: https://github.com/kubernetes-sigs/cluster-api/releases/download/v{{.Version}}/{{.Name}}-{{.Os}}-amd64
      os-linux: linux
      os-mac: darwin
      rename: true
    kubectl:
      version: 1.28.13
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
      version: 5.7.3
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
    aws-iam-authenticator:
      version: 0.6.27
      url: https://github.com/kubernetes-sigs/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}_{{.Version}}_{{.Os}}_amd64
      os-linux: linux
      os-mac: darwin
      rename: true
    gke-auth-plugin:
      version: 0.1.1
      url: https://github.com/traviswt/{{.Name}}/releases/download/{{.Version}}/{{.Name}}_{{.Os}}_x86_64.tar.gz
      os-linux: Linux
      os-mac: Darwin
```

</details>

The project file supports placeholders, they are required for correct URL formation.

* **{{.Name}}:** Replaced with the key's value.
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
