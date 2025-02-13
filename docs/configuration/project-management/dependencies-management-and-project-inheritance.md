# Dependencies management and project inheritance

## Overview

To work with the RMK project's repository, RMK needs to resolve and install additional dependencies that are described
in
the [project.yaml](preparation-of-project-repository.md#projectyaml) file.
The inheritance configuration of the upstream project's repository is defined in the `project.dependencies` section of
the [project.yaml](preparation-of-project-repository.md#projectyaml) file.
All inherited upstream project repositories will be loaded into the `.PROJECT` directory
in the root directory according to the sections described in
the [project.yaml](preparation-of-project-repository.md#projectyaml) file.

> To override inherited versions of dependencies and add-ons described in the inventory,
> you need to specify the entire block with all the required fields.
>
> ```yaml
> inventory:
>   # ...
>   hooks:
>     helmfile.hooks.infra:
>       version: v1.18.0
>       url: git::https://github.com/<owner>/{{.Name}}.git?ref={{.Version}}
>   # ...
> ```

> Dependency resolution occurs when executing almost any RMK command, except for those in the `rmk project` command
> category.

## Change dependency versions of the inherited project's repository

Find the `project` section in the [project.yaml](preparation-of-project-repository.md#projectyaml) file and change
the `version` value to the needed stable tag.
For example:

```yaml
project:
  dependencies:
    # ...
    - name: <upstream_repository_prefix>.bootstrap.infra
      version: v2.17.0 # e.g., a different version of the dependency is required by this project
      url: git::https://github.com/<owner>/{{.Name}}.git?ref={{.Version}}
    # ...
```

Then, in the `helmfiles` section of the `helmfile.yaml.gotmpl` file
the [{{ env "HELMFILE_<project_name\>_PATHS" }}](../cluster-management/exported-environment-variables.md)
environment variable will be used, this way RMK will manage the dependencies of the nested `Helmfile`s.

> The variable name is formed according to the following template: `HELMFILE_<project_name>_PATHS`.
> This mechanism is necessary for resolving circular dependencies correctly.

## Change inherited versions of Helmfile hooks

RMK allows to avoid controlling the versioning of the `Helmfile` hooks through
the [project.yaml](preparation-of-project-repository.md#projectyaml) file of the downstream project's repository,
instead of it, RMK allows inheriting these version hooks from the upstream project's repository.
It also supports multi-versioning of the `Helmfile` hooks as part of the inheritance from several upstream projects by a
downstream project.

In order for these features to work, you need to use
the [{{ env "HELMFILE_<current_project_name\>_HOOKS_DIR" }}](../cluster-management/exported-environment-variables.md)
variable in `helmfile.yaml.gotmpl`.
For example:

```yaml
commonLabels:
  # ...
  bin: {{ env "HELMFILE_RMK_TEST_HOOKS_DIR" }}/bin
# ...
```

Let's look at the following examples of the inheritance:

1. **Hook version inheritance from the upstream project's repository:**

   The [project.yaml](preparation-of-project-repository.md#projectyaml) file of the downstream project is the following:

   ```yaml
   project:
     dependencies:
       - name: rmk-test.bootstrap.infra
         version: v4.4.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
       # ...
   ```

   In this case, a version of the Helmfile hooks in the `inventory.hooks` section is not specified, however,
   it is indicated that the current project of the repository inherits `rmk-test.bootstrap.infra` with the `v4.4.0`
   version.
   In turn, `rmk-test.bootstrap.infra` inherits
   the [cluster-deps.bootstrap.infra](https://github.com/edenlabllc/cluster-deps.bootstrap.infra) repository.
   The [project.yaml](preparation-of-project-repository.md#projectyaml) file for the `rmk-test.bootstrap.infra`
   repository is also missing the version of the hooks:

   ```yaml
   project:
     dependencies:
       - name: cluster-deps.bootstrap.infra
         version: v0.1.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
       # ...
   ```

   Also, the [project.yaml](preparation-of-project-repository.md#projectyaml) file of the `cluster-deps.bootstrap.infra`
   repository will contain the version of the `Helmfile` hooks,
   which will finally be inherited by the downstream project's repository.

   ```yaml
   inventory:
     # ...
     hooks:
       helmfile.hooks.infra:
         version: v1.29.1
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
     # ...
   ```

   > There is no `project.dependencies` section in the [project.yaml](preparation-of-project-repository.md#projectyaml)
   file
   > of the `cluster-deps.bootstrap.infra` repository, since there is no inheritance.

   This configuration scheme is **the most common** and has the following inheritance scheme for the `Helmfile` hooks:

   ```textmate
   Project repo name:            cluster-deps.bootstrap.infra -> rmk-test.bootstrap.infra -----> <downstream_project>.bootstrap.infra
   Project repo version:         v0.1.0                          v4.4.0                          <downstream_project_version>
   Hooks repo name with version: helmfile.hooks.infra-v1.29.1 -> helmfile.hooks.infra-v1.29.1 -> helmfile.hooks.infra-v1.29.1
   ```

2. **Hook version inheritance from the upstream project's repository in case the `rmk-test` project has a fixed version
   of the `Helmfile` hooks specified in its [project.yaml](preparation-of-project-repository.md#projectyaml) file:**

   The [project.yaml](preparation-of-project-repository.md#projectyaml) file of the downstream project is the following:

   ```yaml
   project:
     dependencies:
       - name: rmk-test.bootstrap.infra
         version: v4.4.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
       # ...
   ```

   In this case, the version of the Helmfile hooks in the `inventory.hooks` section is not specified,
   however, it is indicated that the current project of the repository inherits `rmk-test.bootstrap.infra` with
   the `v4.4.0` version.
   In turn, `rmk-test.bootstrap.infra` inherits
   the [cluster-deps.bootstrap.infra](https://github.com/edenlabllc/cluster-deps.bootstrap.infra)
   repository which already has its own fixed version of `v1.29.0` of the `Helmfile` hooks in the `inventory.hooks`
   section:

   ```yaml
   project:
     dependencies:
       - name: cluster-deps.bootstrap.infra
         version: v0.1.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
   # ...
   inventory:
     # ...
     hooks:
       helmfile.hooks.infra:
         version: v1.29.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
     # ...
   ```

   The [project.yaml](preparation-of-project-repository.md#projectyaml) file of the `cluster-deps.bootstrap.infra`
   repository will contain the version of the `Helmfile` hooks,
   which will be inherited by the downstream project's repository:

   ```yaml
   inventory:
     # ...
     hooks:
       helmfile.hooks.infra:
         version: v1.29.1
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
     # ...
   ```

   This configuration scheme will look like this:

   ```textmate
   Project repo name:            cluster-deps.bootstrap.infra -> rmk-test.bootstrap.infra -----> <downstream_project>.bootstrap.infra
   Project repo version:         v0.1.0                          v4.4.0                          <downstream_project_version>
   Hooks repo name with version: helmfile.hooks.infra-v1.29.1 -> helmfile.hooks.infra-v1.29.0 -> helmfile.hooks.infra-v1.29.1
   ```

   > The downstream project's repository will inherit the latest version of `Helmfile` hooks, specifically from
   the `cluster-deps.bootstrap.infra` repository.
   > As a result, in the downstream project's repository, we will have the two loaded versions of `Helmfile` hooks:
   >
   > - One will be relevant for the `cluster-deps.bootstrap.infra` repository and the downstream project's repository.
   > - Another will be relevant for the `rmk-test.bootstrap.infra` repository.
   >
   > This mechanism allows for multi-versioning support of the `Helmfile` hooks at different levels of the inheritance.

3. **Hook version inheritance from the upstream project's repository in case the downstream project
   has a fixed version of [cluster-deps.bootstrap.infra](https://github.com/edenlabllc/cluster-deps.bootstrap.infra)
   specified in its [project.yaml](preparation-of-project-repository.md#projectyaml) file:**

   The [project.yaml](preparation-of-project-repository.md#projectyaml) file of the downstream project is the following:

   ```yaml
   project:
     dependencies:
       - name: cluster-deps.bootstrap.infra
         version: v0.2.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}       
       - name: rmk-test.bootstrap.infra
         version: v4.4.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
       # ...
   ```

   > The dependencies should be declared in the correct order of inheritance: the first one
   is `cluster-deps.bootstrap.infra`,
   > then `rmk-test.bootstrap.infra`, then other repositories (if needed).

   In this case, a version of the `Helmfile` hooks in the `inventory.hooks` section is not specified,
   however, it is indicated that the current project of the repository inherits `rmk-test.bootstrap.infra` with
   the `v4.4.0` version.
   In turn, `rmk-test.bootstrap.infra` inherits the `cluster-deps.bootstrap.infra` repository.
   The [project.yaml](preparation-of-project-repository.md#projectyaml) file for the `rmk-test.bootstrap.infra`
   repository is also missing the version of the hooks:

   ```yaml
   project:
     dependencies:
       - name: cluster-deps.bootstrap.infra
         version: v0.1.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
       # ...
   ```

   The [project.yaml](preparation-of-project-repository.md#projectyaml) file of the `cluster-deps.bootstrap.infra`
   repository
   of the `v0.2.0` version will contain the version of the `Helmfile` hooks, which will be inherited by the downstream
   projects:

   ```yaml
   inventory:
     # ...
     hooks:
       helmfile.hooks.infra:
         version: v1.30.0
         url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
     # ...
   ```

   This configuration scheme will look like this:

   ```textmate
   Project repo name:            cluster-deps.bootstrap.infra -> rmk-test.bootstrap.infra -----> <downstream_project>.bootstrap.infra
   Project repo version:         v0.2.0                          v4.4.0                          <downstream_project_version>
   Hooks repo name with version: helmfile.hooks.infra-v1.30.0 -> helmfile.hooks.infra-v1.30.0 -> helmfile.hooks.infra-v1.30.0
   ```

   > Since the downstream project's repositories inherit the `Helmfile` hooks from the `cluster-deps.bootstrap.infra`
   > repository, and we redefined the `cluster-deps.bootstrap.infra` dependency in the downstream project's,
   > all repositories will inherit this concrete version, and only it will be downloaded.

## Change inherited versions of Helm plugins, tools

The same inheritance method as for the `Helmfile` hooks is supported for `inventory` sections as `helm-plugins`
and `tools`.
If a specific version is not specified, the latest version from the upstream project's repository will always be used,
with one exception only: in this case, multi-versioning is not supported, and only one version will be downloaded.

> All add-ons versions in the inventory sections must be specified in the `SemVer2` format,
> as the inheritance mechanism relies on this format to distinguish the version order.
