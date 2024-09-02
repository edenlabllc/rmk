# Configuration management

To start working with Kubernetes clusters, RMK needs to initialize the configuration for the current environment.
At the time of configuration initialization launch, RMK prepares
the state in the form of the current environment config with all the required attributes for further work.
It also downloads and resolves and installs all necessary dependencies and tools described 
in the [project.yaml](project-management/preparation-of-project-repository.md#projectyaml) file in the root of the project repository.

## List of main attributes of the RMK configuration

Example of the configuration:

```yaml
name: kodjin-develop # RMK config name, a unique identifier which consists of the tenant name and the abbreviated name of the Git branch.
tenant: kodjin # Tenant name.
environment: develop # Environment name.
config-from: kodjin-develop # Configuration name from which the cluster configuration was inherited.
root-domain: kodjin-develop.edenlab.dev # Root domain name used across the cluster.
aws:
  profile: kodjin-develop # AWS profile name for the AWS CLI.
  region: eu-north-1 # AWS region of the current Kubernetes cluster.
  account_id: "123456789"
# ...
```

> All attributes can be overridden using RMK flags or environment variables.

## Initialization of RMK configuration

> Prerequisites:
> 
> - The `GITHUB_TOKEN` variable or `--github-token` flag are required: [GitHub Personal Access Tokens (PAT)](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-personal-access-token-classic).
>   > The token should have the `repo: full control` permissions.
> - [Project repository](project-management/requirement-for-project-repository.md) has already been created and initialized.
> - At least one Git branch for the environment exists already.

```shell
rmk config init
```

### Configuration of AWS profile

If an AWS profile with the correct name has not been created previously during the first initialization of the configuration,
RMK will start the creation process. The 2 supported configuration scenarios are:

* **through environment variables:** `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
* **interactive input**: the AWS credentials will be requested one by one.

If the environment variables has been declared before the  `rmk config init` command was run, RMK will create a profile
based on their values. Otherwise, the interactive mode will begin.

### Support for Multi-Factor Authentication (MFA)

RMK automatically check for an MFA device, when the following command is executed: `rmk config init`.

To set up an MFA device, if it is required by the administrator, the following actions should be executed:

1. First, sign in to the AWS Management Console.
2. Then, go to the following page to set up security
   credentials: [My security credentials](https://console.aws.amazon.com/iam/home#/security_credentials)
3. Navigate to the "Multi-factor authentication (MFA)" section and set up an MFA device. 
   If a device name is required, specify a name.
4. After that, sign out and sign in again to refresh AWS policies 
   (might be required in case of an IAM policy based on the `aws:MultiFactorAuthPresent` condition exists).
5. Finally, on the "My security credentials" page navigate to the "Access keys for CLI, SDK, & API access" section
   and create a new AWS access key, if needed.

> For the detailed documentation regarding the MFA setup in AWS, go to 
> [AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_mfa_enable_virtual.html#enable-virt-mfa-for-own-iam-user)

You can also check the lifetime of the session token by running the command: `rmk config init`

```
2022-12-14T09:02:20.267+0100 INFO MFA remaining time for token validity: 11:59:48
```

## Initialization of RMK configuration for feature or release clusters

When initializing the RMK configuration for feature or release clusters, you can use inheritance 
from a previously saved configuration that contains the necessary credentials to create a Kubernetes cluster.
Let's say you want to create or connect to the feature cluster with the credentials of the `develop` cluster,
in this case you must run the initialization command with the `--config-from-environment` flag. For example:

```shell
rmk config init --config-from-environment=<develop|staging|production|ffs-XXX|vX.X.X-rc|vX.X.X>
```

### Reconfiguration of the AWS profile if wrong credentials has been input

```shell
rmk config init --aws-reconfigure
```

### Initialization of RMK configuration with a custom root domain

```shell
rmk config init --root-domain="example.com"
```

### Deletion of RMK configuration

```shell
rmk config delete
```

> When deleting the current RMK configuration, the respective AWS profile files will be deleted as well.

## Use upstream artifact for the downstream project's repository

RMK supports downloading an upstream project's artifact using additional "license" AWS credentials. 
To switch RMK to the artifact usage mode, you need to use additional flags when initializing the RMK configuration 
for the current project. Additionally, before starting the initialization, you need to install the required version 
of the upstream project to which you want to update.
For example:

```yaml
project:
  dependencies:
    - name: deps.bootstrap.infra
      version: v2.17.0
      url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
      artifact-url: https://edenlabllc-{{.HelmfileTenant}}-artifacts-infra.s3.eu-north-1.amazonaws.com/{{.Version}}/{{.HelmfileTenant}}-{{.Version}}.tar.gz
    # ...
```

> The `artifact-url` field is required and contains the artifact URL generation template which consists 
> of the following [fields](project-management/preparation-of-project-repository.md#projectyaml).

Set the `version` field to the version of the upstream project for the current project. For example:

```shell
# artifact usage modes: none|online (default: "none")
rmk config init --artifact-mode=online
```

> Currently, only two artifact modes are supported:
> 
> - `none`: The standard mode of RMK which is used for development normally, the codebase will be downloaded from GitHub repositories.
>   The mode does not require the presence of the special "license" credentials.
> - `online`: Switches RMK to work with artifacts. In this mode, RMK will not use any credentials for GitHub 
>   (e.g., personal access tokens), but will request additional license AWS credentials to download and unpack 
>   the artifact from a repository like AWS S3.

To change the "license" AWS credentials when in the online artifact mode, use the following command:

```shell
rmk config init --aws-reconfigure-artifact-license
```
