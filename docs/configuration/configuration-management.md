# Configuration management

To start working with Kubernetes clusters, RMK needs to initialize the configuration for the current environment.
At the time of configuration initialization launch, RMK prepares
the state in the form of the current environment config with all the required attributes for further work.
It also downloads and resolves and installs all necessary dependencies and tools described 
in the [project.yaml](project-management/preparation-of-project-repository.md#projectyaml) file in the root of the project repository.

## List of main attributes of the RMK configuration

Example of the configuration:

[//]: # (  TODO ACTUALIZE)

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the tenant name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
aws:
  profile: rmk-test-develop # AWS profile name.
  region: us-east-1 # AWS region of the current Kubernetes cluster.
  account_id: "123456789012"
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
Let's say you want to create or connect to the feature cluster with the credentials of the `develop` cluster.

[//]: # (in this case you must run the initialization command with the `--config-from-environment` flag. For example:)

[//]: # (  TODO ACTUALIZE)

```shell
#rmk config init>
```

### Reconfiguration of the AWS profile if wrong credentials has been input

[//]: # (  TODO ACTUALIZE)

```shell
rmk config init
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
