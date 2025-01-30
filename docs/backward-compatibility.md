# Backward compatibility

## Breaking change releases

**[v0.44.2](https://github.com/edenlabllc/rmk/releases/tag/v0.44.2) -> [v0.45.0](https://github.com/edenlabllc/rmk/releases/tag/v0.45.0)**

### Motivations

Migrating RMK `v0.45.0` from **Terraform** to **Cluster API**

We are replaced the technology stack for provisioning clusters in RMK `v0.45.0`, transitioning from Terraform to [Cluster API](https://cluster-api.sigs.k8s.io/). 
This shift was driven by several key factors:

Why We Switched to Cluster API?

1. Maintaining Open-Source Integrity
   Terraform's transition to a BSL license conflicts with our commitment to keeping RMK fully open-source (OSS). 
   By switching to Cluster API, we ensure that our customers' interests remain unaffected.
   More details on the [Terraform license change](https://www.hashicorp.com/license-faq).

2. A More Native Kubernetes Solution
   We needed a provisioning approach that seamlessly integrates with Kubernetes across various environments. 
   With the new RMK `v0.45.0`, we now support:

   - AWS
   - Azure
   - GCP
   - On-Premise (support is expected in upcoming releases)
   - K3D (local installation)
   
3. Simplified Configuration Management
   Cluster configurations are now stored in Helm charts, aligning with the way installed components are managed. 
   This ensures a unified format for all declarations.

4. Seamless Cluster Upgrades
   Our new approach makes cluster updates easier and Kubernetes-native, leveraging:

   - Pod status awareness
   - Zero downtime upgrades

> This transition marks a significant step in enhancing RMK’s provisioning capabilities, ensuring better scalability, 
> openness, and ease of management. Stay tuned for more updates in upcoming releases! 🚀

### Deprecated Features

For command `rmk config init`:

- **--artifact-mode**, **--aws-reconfigure-artifact-license** - deprecated along with the functionality. No need to use.
- **--aws-ecr-host**, **--aws-ecr-region**, **--aws-ecr-user-name** - deprecated along with the functionality. 
  Replaced by a third-party Kubernetes native solution [ecr-token-refresh](https://github.com/edenlabllc/ecr-token-refresh.operators.infra) operator.
- **--aws-reconfigure** - deprecated, refusal to use AWS CLI. Replaced with using [AWS SDK](https://github.com/aws/aws-sdk-go-v2).
- **--cloudflare-token** - deprecated along with the functionality. Replaced by a third-party Kubernetes native solution [external-dns](https://github.com/kubernetes-sigs/external-dns).
- **--cluster-provisioner-state-locking** - deprecated, refusal to use Terraform.
- **--config-from-environment** - deprecated along with the functionality. 
- **--root-domain** - deprecated. Replaced by a declarative configuration via [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml).
- **--s3-charts-repo-region** - deprecated. Configuration of private repositories for the chart is available in the [Helmfile](https://helmfile.readthedocs.io/en/latest/#configuration).

For command category `rmk cluster`:

- **container-registry** - deprecated command with full list of flags, refusal to use Terraform.
- **destroy** - deprecated command with full list of flags, refusal to use Terraform.
- **list** - deprecated command with full list of flags, refusal to use Terraform. 
- **provision** - deprecated command with full list of flags, refusal to use Terraform.
- **state** - deprecated command with full list of flags, refusal to use Terraform.

### How to migrate to RMK v0.45.0 version from currently

#### For newly created project repositories

Before performing actions via RMK with this project repository, simply update to `v0.45.0` version.

```shell
rmk update
```

#### For previously created project repositories for AWS cluster provider

> For correct migration, be sure to follow the steps in strict order.

1. On the current RMK version download private Age keys if early not do it.

   ```shell
   rmk config init --github-token=<GitHub_PAT>
   ```

2. Save the path to the private Age keys storage directory to an environment variable.

   ```shell
   export RMK_OLD_PATH_AGE_KEYS="$(rmk --log-format=json config view | yq '.config.SopsAgeKeys')"
   ```

   > Skip this step if you lack administrator permissions for the selected AWS account.

3. Update your current version to `v0.45.0`.

   ```shell
   rmk update
   ```

4. Add root domain specification in project.yaml for project repository.

   ```yaml
   project:
     # ...
     spec:
       environments:
         develop:
           root-domain: <custom_root_domain_name> # or <*.edenlab.dev> if you member Edenlab team
         production:
           root-domain: <custom_root_domain_name> # or <*.edenlab.dev> if you member Edenlab team
         staging:
           root-domain: <custom_root_domain_name> # or <*.edenlab.dev> if you member Edenlab team
   ```

   > Skip this step if the project.yaml file was previously modified to match the new specification.

5. Initialize a new configuration specifying the AWS cluster provider. 

   ```shell
   rmk config init --cluster-provider=aws \
       --aws-access-key-id=<aws_access_key_id> \
       --aws-region=<aws_region> \
       --aws-secret-access-key=<aws_secret_access_key> \
       --github-token=<GitHub_PAT>
   ```

6. Copy from old path private Age keys in new directory.

   ```shell
   cp -f "${RMK_OLD_PATH_AGE_KEYS}"/* $(rmk --log-format=json config view | yq '.config.SopsAgeKeys')
   ```

   > Skip this step if you lack administrator permissions for the selected AWS account.

7. Upload old private Age keys to AWS Secret Manager.

   ```shell
   rmk secret keys upload
   ```
   
   > Skip this step if you lack administrator permissions for the selected AWS account. 
