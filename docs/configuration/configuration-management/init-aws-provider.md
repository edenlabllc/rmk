# Initialization of AWS cluster provider

### List of main attributes of the RMK configuration

Without MFA:
```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the tenant name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant repository name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
cluster-provider: aws # Selected cluster provider.
# ...
aws:
  account_id: "123456789012" # AWS account ID.
  config-source: /Users/user/.aws/config_rmk-test-develop # Absolute path to the AWS profile config.
  credentials-source: /Users/user/.aws/credentials_rmk-test-develop # Absolute path to the AWS profile credentials.
  user-name: user # AWS user name.
  profile: rmk-test-develop # AWS profile name.
  region: us-east-1 # AWS region of the current Kubernetes cluster.    
# ...
```

With MFA:
```yaml
name: rmk-test-develop
tenant: rmk-test
environment: develop
root-domain: rmk-test-develop.edenlab.dev
cluster-provider: aws
# ...
aws-mfa-profile: rmk-test-develop-mfa # AWS profile name for MFA.
aws-mfa-token-expiration: "1738006158" # Time expiration MFA token.
aws:
  account_id: "123456789012"
  config-source: /Users/test-mfa-user/.aws/config_rmk-test-develop
  credentials-source: /Users/test-mfa-user/.aws/credentials_rmk-test-develop
  user-name: test-mfa-user
  mfa-device: arn:aws:iam::123456789012:mfa/test-mfa-user # MFA device AWS ARN.
  profile: rmk-test-develop
  region: us-east-1
```

### Prerequisites

1. Having an account in AWS and a created user with access policies in IAM: PowerUserAccess, SecretsManagerReadWrite.
[Useful links](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html).

2. Having an AWS access key.
[Useful links](https://docs.aws.amazon.com/IAM/latest/UserGuide/access-key-self-managed.html).

3. Allocated quotas for specific family VMs (EC2) in the required region.

### Configuration of AWS

If an AWS profile with the correct name has not been created previously during the first initialization of the configuration,
RMK will start the creation process.
RMK store the AWS config and credentials files by path:
- `${HOME}/.aws/config_<project_name>-<project_branch>`
- `${HOME}/.aws/credentials_<project_name>-<project_branch>`

The 2 supported configuration scenarios are:

* **through RMK flags:**:
  ```shell
  rmk config init --cluster-provider=aws \
    --aws-access-key-id=<aws_access_key_id> \
    --aws-region=us-east-1 \
    --aws-secret-access-key=<aws_secret_access_key>
  ```
  
* **through environment variables**: `AWS_ACCESS_KEY_ID`, `AWS_REGION`, `AWS_SECRET_ACCESS_KEY`.
  ```shell
  rmk config init --cluster-provider=aws
  ```

If the environment variables has been declared before the  `rmk config init --cluster-provider=aws` command was run, 
RMK will create a profile based on their values. 
If flags will be declared, RMK will create a profile based on values flags because flags has priority.

### Support for Multi-Factor Authentication (MFA)

RMK automatically check for an MFA device, when the following command is executed: `rmk config init --cluster-provider=aws`.

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

### Reconfiguration of the AWS profile if wrong credentials has been input

Change the value of a specific flag if adjustments are required.

```shell
rmk config init --aws-access-key-id=<new_aws_access_key_id> --aws-secret-access-key=<new_aws_secret_access_key>
```
