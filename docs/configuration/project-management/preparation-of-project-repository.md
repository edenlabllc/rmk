# Preparation of the project repository

## Prerequisites

- Create a remote repository (Git) for a project in your Version Control System (e.g., [GitHub](https://github.com))
  according to the [requirements](requirement-for-project-repository.md#requirement-for-project-repository). For
  example: `rmk-test.bootstrap.infra`
- Clone the existing project repository:
  
  ```shell
  git clone <github_repo_url>
  ```
  
  Alternatively, initialize a new repository manually:
  
  ```shell
  git init
  git remote add <github_repo_name> <github_repo_url>
  git commit --allow-empty --message "Initial commit"
  ```
  
  > RMK requires a Git branch with at least one commit and a configured `origin` remote  
  > to correctly resolve the project name and environment.

- Checkout the required branch. For example: `develop`.

## Automatic generation of the project structure from scratch

RMK supports automatic generation of a complete project structure  
based on the project specification defined in the [project.yaml](#projectyaml) file.

Use the following command with the recommended flags:

```bash
rmk project generate \
  --environment="develop.root-domain=<custom_root_domain_name>" \
  --owner=gh-user1 \
  --scope=<upstream_project_name> \
  --scope=<downstream_project_name>
```

> Add the `--create-sops-age-keys` flag if you want to generate the project structure along with **new**
> SOPS Age private keys. See the [following page](../secrets-management/secrets-management.md#secret-keys) for details.
>
> Add one or more `--sops-age-key=<vals_backend_reference_scopeN>` flags if you want to define Vals backend
> references to **previously created** SOPS Age keys, the keys will **automatically be fetched** by other users later 
> during 
> [configuration initialization](../secrets-management/helmfile-vals-integration.md#configuration-initialization). 
> This functionality is available for `k3d` and `onprem` cluster providers **only**, as they do not use 
> any third-party secret storage out of the box. See the 
> [following page](../secrets-management/helmfile-vals-integration.md#sops-age-key-fetching-via-vals-backend-references-in-projectyaml) 
> for details.

This command will create a default project structure and configure an example release based
on [Nginx](https://nginx.org/). See the [Quickstart](../../quickstart.md) guide for a simple usage example.

> If the `project.yaml` file does not exist, it will be created automatically.

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
      - name: <upstream_project_name>.bootstrap.infra
        # Required, dependencies upstream project's repository version in `SemVer2` format, also can be a branch name or a commit hash.
        version: <SemVer2>
        # Required, dependencies upstream project's repository URL.
        url: git::https://github.com/<github_repo_owner>/{{.Name}}.git?ref={{.Version}}    
    # Optional, needed if you want automatic generation of the project structure from scratch.
    spec:
      # Required, list of available environments with specific root domain name (Git branches). 
      environments:
        - develop:
            root-domain: <custom_name>.example.com
        - production:
            root-domain: <custom_name>.example.com
        - staging:
            root-domain: <custom_name>.example.com
      # Optional, list of owners of the project.
      owners:
        - <gh_user1>
        - <gh_user2>
      # Required, list of available scope of the project.
      scopes:
        - <upstream_project_name>
        - <downstream_project_name>
      # Optional, list of SOPS Age keys defined as vals backend references.
      # Each reference must follow the vals format:
      #   ref+BACKEND://PATH[?PARAMS][#FRAGMENT][+]
      # The uploaded key must follow the naming convention: <project_name>-<scope>
      # Examples:
      #   ref+awssecrets://rmk-test-deps?region=us-east-1
      #   ref+azurekeyvault://rmk-test-deps
      #   ref+gcpsecrets://rmk-test-rmk-test
      sops-age-keys:
        - ref+<vals_backend>://<project_name>-<scope0>?<vals_backend_parameters>
        - ref+<vals_backend>://<project_name>-<scope1>?<vals_backend_parameters>
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
        url: https://github.com/<github_repo_owner>/helm-diff
      # ...
    # Optional, contains a map of the Helmfile hooks repositories with shell scripts.
    hooks:
      # Optional, Helmfile hooks repository name.
      helmfile.hooks.infra:
        # Required, Helmfile hooks repository version in the `SemVer2` format.
        version: <SemVer2>
        # Required, Helmfile hooks repository URL.
        url: git::https://github.com/<github_repo_owner>/{{.Name}}.git?ref={{.Version}}
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
  <summary>Example of the full <code>project.yaml</code> file</summary>

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
      - gh-user1
      - gh-user2
    scopes:
      - deps
      - rmk-test
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

The project file's `inventory` section supports placeholders, they are required for correct URL formation.

* **{{.Name}}:** Replaced with the key's value.
* **{{.Version}}:** Replaced with the `version` field.
* **{{.HelmfileTenant}}:** Replaced with the tenant (project) name for the Helmfile selected from the list.
* **{{.Os}}:** Replaced with the values from the `os-linux`, `os-mac` fields according to the specific operating system,
  where RMK is run.

> The field `rename` of the boolean type is required to correct the name of the binary file of the downloaded tool
> according to the value of the `name` field. This is mainly required for the cases, when the artifact is not the
> archive.
> For example:
>
> - The initial file name after the download: `helmfile_darwin_amd64`.
> - After applying the `rename` instruction it gets a value of the `name` field: `helmfile`.
