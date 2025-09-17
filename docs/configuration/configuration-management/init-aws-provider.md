# Initialization of AWS cluster provider

## List of main attributes of the RMK configuration

### With MFA

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
  config-source: /home/user/.aws/config_rmk-test-develop
  credentials-source: /home/user/.aws/credentials_rmk-test-develop
  user-name: user
  mfa-device: arn:aws:iam::123456789012:mfa/user # MFA device AWS ARN.
  profile: rmk-test-develop
  region: us-east-1
```

### Without MFA

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the project (tenant) name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant name, which is equivalent to the project name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
cluster-provider: aws # Selected cluster provider.
# ...
aws:
  account_id: "123456789012" # AWS account ID.
  config-source: /home/user/.aws/config_rmk-test-develop # Absolute path to the AWS profile config.
  credentials-source: /home/user/.aws/credentials_rmk-test-develop # Absolute path to the AWS profile credentials.
  user-name: user # AWS user name.
  profile: rmk-test-develop # AWS profile name.
  region: us-east-1 # AWS region of the current Kubernetes cluster.    
# ...
```

## Prerequisites

1. Having an account in AWS and a created user with access policies in IAM:
   [PowerUserAccess](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/PowerUserAccess.html),
   [SecretsManagerReadWrite](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/SecretsManagerReadWrite.html).
   > See the [useful link](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html).

2. Having an AWS access key pair.
   > See the [useful link](https://docs.aws.amazon.com/IAM/latest/UserGuide/access-key-self-managed.html).

3. Allocated [EC2 quotas](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html) for specific
   family VMs in the required region.

## Configuration

If an [AWS profile](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-files.html) with the correct name was
not created during the initial configuration, RMK will generate it and store the AWS config and credentials files at the
following path:

```shell
${HOME}/.aws/config_<project_name>-<project_branch>
${HOME}/.aws/credentials_<project_name>-<project_branch>
```

The 2 supported configuration scenarios are:

* **using RMK flags**:
  ```shell
  rmk config init --cluster-provider=aws \
    --aws-access-key-id=<aws_access_key_id> \
    --aws-region=us-east-1 \
    --aws-secret-access-key=<aws_secret_access_key>
  ```

* **using environment variables**: `AWS_ACCESS_KEY_ID`, `AWS_REGION`, `AWS_SECRET_ACCESS_KEY`.
  ```shell
  export AWS_ACCESS_KEY_ID=<aws_access_key_id>
  export AWS_REGION=us-east-1
  export AWS_SECRET_ACCESS_KEY=<aws_secret_access_key>
  rmk config init --cluster-provider=aws
  ```

If environment variables were set before running the command, RMK will create a profile based on their values.  
If flags are specified, RMK will prioritize them over environment variables, as **CLI flags take precedence**.

## Support for multi-factor authentication (MFA)

RMK automatically check for an [MFA](https://aws.amazon.com/iam/features/mfa/) device, when the following command
is executed:

```shell
rmk config init --cluster-provider=aws
```

To **set up an MFA device**, if it is required by the administrator, the following actions should be executed:

1. Sign in to the "AWS Management Console".
2. Go to the following page to set up security
   credentials: [My security credentials](https://console.aws.amazon.com/iam/home#/security_credentials)
3. Navigate to the "Multi-factor authentication (MFA)" section and set up an MFA device.
   If a device name is required, specify a name.
4. After that, sign out and sign in again to refresh AWS policies
   (might be required in case of an IAM policy based on the
   [aws:MultiFactorAuthPresent](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_mfa_configure-api-require.html)
   condition exists).
5. Finally, on the "My security credentials" page navigate to the "Access keys for CLI, SDK, & API access" section
   and create a new AWS access key, if needed.

> For the detailed documentation regarding the MFA setup in AWS, go to
> [AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_mfa_enable_virtual.html#enable-virt-mfa-for-own-iam-user).

You can also check the lifetime of the session token by running the [rmk config init](../../commands.md#init-i)
command:

```
2022-12-14T09:02:20.267+0100 INFO MFA remaining time for token validity: 11:59:48
```

## Reconfiguration of the AWS profile if wrong credentials has been input

Modify the value of a specific flag if changes are needed:

```shell
rmk config init --aws-access-key-id=<new_aws_access_key_id> --aws-secret-access-key=<new_aws_secret_access_key>
```
