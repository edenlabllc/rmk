# Configuration management

## Overview

To start working with Kubernetes clusters, RMK needs to **initialize the configuration** for the current environment.

At the time of configuration initialization launch, RMK **prepares**
the state in the form of the current environment config with all the required attributes for further work.
It also **downloads** and **resolves** and installs all necessary dependencies and tools described
in the [project.yaml](../project-management/preparation-of-project-repository.md#projectyaml) file in the root of the
project repository.

## List of main attributes of the RMK configuration

Example of the configuration per cluster provider:

- [AWS](init-aws-provider.md#list-of-main-attributes-of-the-rmk-configuration)
- [Azure](init-azure-provider.md#list-of-main-attributes-of-the-rmk-configuration)
- [GCP](init-gcp-provider.md#list-of-main-attributes-of-the-rmk-configuration)
- [K3D](init-k3d-provider.md#list-of-main-attributes-of-the-rmk-configuration)

> All configuration attributes **can be overridden** using RMK flags or environment variables.

To view the available options of the created configuration, use the command:

```shell
rmk config view
```

## Understanding the behavior of the configuration initialization command

The [rmk config init](../../commands.md#init-i) command supports declarative behavior within a single
[project repository](../project-management/requirement-for-project-repository.md#requirement-for-project-repository)
and an environment that equal branch name.

This example assumes the [project](../project-management/requirement-for-project-repository.md)
(tenant) name is `rmk-test`,
the [Git branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches)
and environment are `develop`, the generated RMK configuration name is `rmk-test-develop`,
the selected [cluster provider](../cluster-management/cluster-management.md) is `aws`:

```shell
rmk config init --cluster-provider=aws \ 
  --github-token=<github_personal_access_token> \
  --aws-access-key-id=<aws_access_key_id> \
  --aws-region=us-east-1 \
  --aws-secret-access-key=<aws_secret_access_key>
```

In the example above, this configuration was applied for the first time, setting the options accordingly. After that,
there's no need to re-specify the entire list of values - simply update the required option as needed.

For example:

```shell
# no need to specify the AWS flags again 
rmk config init --github-token=<new_github_personal_access_token>
```

## Initialization of RMK configuration

### Prerequisites

- [Project repository](../project-management/requirement-for-project-repository.md) has already been created and
  initialized.
- At least one Git branch for the environment exists already.

### Command

```shell
rmk config init
```

### Initialization of RMK configuration for private GitHub repositories

#### Prerequisites

- The `GITHUB_TOKEN` variable or `--github-token` flag are required:
  [GitHub Personal Access Tokens (PAT)](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-personal-access-token-classic).
- The token should have the `repo: full control` permissions.

#### Command

```shell
rmk config init --github-token=<github_personal_access_token>
```

### Initialization of RMK configuration for different cluster providers

- [AWS](init-aws-provider.md#prerequisites)
- [Azure](init-azure-provider.md#prerequisites)
- [GCP](init-gcp-provider.md#prerequisites)
- [K3D](init-k3d-provider.md#list-of-main-attributes-of-the-rmk-configuration)

### Initialization of RMK configuration with a custom root domain

To change the root domain name, you need to edit
the [project.yaml](../project-management/preparation-of-project-repository.md#projectyaml)
file in the section `develop.root-domain`.

```yaml
project:
  spec:
    environments:
      develop:
        root-domain: test.example.com
```

Then run the following command:

```shell
rmk config init
```

## Deletion of RMK configuration

```shell
rmk config delete
```

> When deleting the current RMK configuration, the respective cluster providers files **will be deleted** as well.
