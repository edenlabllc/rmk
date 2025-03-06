# Backward compatibility

## Breaking change releases

---

### [v0.44.2](https://github.com/edenlabllc/rmk/releases/tag/v0.44.2) -> [v0.45.0](https://github.com/edenlabllc/rmk/releases/tag/v0.45.0)

#### Motivation

The main change in the [v0.45.0](https://github.com/edenlabllc/rmk/releases/tag/v0.45.0) RMK version is the
**replacement** of the technology stack for cluster provisioning, transitioning
from [Terraform](https://www.terraform.io/) to [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/).

This change was driven by several **key factors**:

1. **Maintaining open-source integrity**:

   Terraform's transition to a [BSL license](https://en.wikipedia.org/wiki/Business_Source_License) conflicts with our 
   commitment to keeping RMK fully [open-source (OSS)](https://en.wikipedia.org/wiki/Open-source_software). 
   By switching to Kubernetes Cluster API, we ensure that our customers' interests remain unaffected.

   > More details on the Terraform's license change are available at the [link](https://www.hashicorp.com/license-faq).

2. **Kubernetes-native solution**:

   We needed a provisioning approach that seamlessly integrates with Kubernetes across various environments.

   With the new version, we now support
   [AWS](configuration/configuration-management/init-aws-provider.md),
   [Azure](configuration/configuration-management/init-azure-provider.md),
   [GCP](configuration/configuration-management/init-gcp-provider.md),
   [K3D](configuration/configuration-management/init-k3d-provider.md) (local installation).

   > **On-premise** support is expected in [upcoming releases](index.md#roadmap).

3. **Simplified configuration management**:

   Cluster configurations are now stored in [Helm charts](https://helm.sh/docs/topics/charts/), aligning with the way
   installed components are managed. This ensures a unified format for all declarations.

4. **Seamless cluster upgrades**:

   Our new approach makes cluster updates Kubernetes-native with
   [pod status awareness](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) and
   [zero-downtime upgrades](https://en.wikipedia.org/wiki/Downtime#Service_levels).

> This transition marks a **significant step** in enhancing RMK’s provisioning capabilities, ensuring better
> scalability, openness, and ease of management. 
> 
> **Stay tuned** for more updates in [upcoming releases](index.md#roadmap)! 🚀

#### Deprecated features

For the [rmk config init](commands.md#init-i) command:

- **--artifact-mode**, **--aws-reconfigure-artifact-license**: removed the flag along with the functionality, no longer
  needed.
- **--aws-ecr-host**, **--aws-ecr-region**, **--aws-ecr-user-name**: removed the flag along with the functionality,
  replaced by the third-party Kubernetes-native
  [ecr-token-refresh](https://github.com/edenlabllc/ecr-token-refresh.operators.infra) operator.
- **--aws-reconfigure**: removed the flag, replaced [AWS CLI](https://aws.amazon.com/cli/) with
  [AWS SDK](https://github.com/aws/aws-sdk-go-v2).
- **--cloudflare-token**: removed the flag along with the functionality, replaced by the third-party
  Kubernetes-native [external-dns](https://github.com/kubernetes-sigs/external-dns) operator.
- **--cluster-provisioner-state-locking**: removed the flag (Terraform is no longer in use).
- **--config-from-environment**: removed the flag along with the functionality.
- **--root-domain**: removed the flag, replaced by the declarative configuration
  via [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml).
- **--s3-charts-repo-region**: removed the flag, replaced with the private repository configuration
  via [Helmfile](https://helmfile.readthedocs.io/en/latest/#configuration).

For the [rmk cluster](commands.md#cluster) command category:

- **container-registry**: removed the command along with all flags.
- **destroy**: removed the command along with all flags (Terraform is no longer in use).
- **list**: removed the command along with all flags (Terraform is no longer in use).
- **provision**: removed the command along with all flags (Terraform is no longer in use).
- **state**: removed the command along with all flags (Terraform is no longer in use).

#### Steps to migrate

##### Newly created project repositories

Before performing actions via RMK with this project repository, **simply update** to the
[v0.45.0](https://github.com/edenlabllc/rmk/releases/tag/v0.45.0) version.

```shell
rmk update
```

##### Previously created project repositories for the AWS cluster provider

To ensure a successful migration, the following the steps should be executed in the **specified order**:

1. **Ensure** you are using the previous [v0.44.2](https://github.com/edenlabllc/rmk/releases/tag/v0.44.2) RMK version.
   
   ```shell
   rmk update -v v0.44.2
   ```

2. **Download** private [SOPS Age keys](configuration/secrets-management/secrets-management.md#secret-keys) of the current
   RMK version if you haven't done it earlier.

   ```shell
   rmk config init --github-token=<github_personal_access_token>
   ```

   > Skip this step if you lack **administrator permissions** for the selected AWS account.

3. **Save** the old path to the private SOPS Age keys storage directory to an environment variable.

   ```shell
   RMK_SOPS_AGE_KEYS_PATH_OLD="$(rmk --log-format=json config view | yq '.config.SopsAgeKeys')"
   ```

   > Skip this step if you lack **administrator permissions** for the selected AWS account.

4. **Update** your current version to [v0.45.0](https://github.com/edenlabllc/rmk/releases/tag/v0.45.0).

   ```shell
   rmk update
   ```

5. **Add** root domain specification
   in [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) for project
   repository.

   ```yaml
   project:
     # ...
     spec:
       environments:
         develop:
           root-domain: <custom_root_domain_name> # or "*.edenlab.dev" for the Edenlab team
         production:
           root-domain: <custom_root_domain_name> # or "*.edenlab.dev" for the Edenlab team
         staging:
           root-domain: <custom_root_domain_name> # or "*.edenlab.dev" for the Edenlab team
   ```

   > If the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file has
   > the `spec.environments` section of a deprecated format already, e.g.:
   > 
   > ```yaml
   > project:
   >  # ...
   >  spec:
   >    environments:
   >      - develop # deprecated
   >      - production # deprecated
   >      - staging # deprecated
   > ```
   >
   > be sure to replace the scalar strings with the new `spec.environments` objects containing the respective root 
   > domains.

6. [Initialize](configuration/configuration-management/init-aws-provider.md#configuration-of-aws) a new configuration
   specifying the AWS cluster provider.

   ```shell
   rmk config init --cluster-provider=aws \
       --aws-access-key-id=<aws_access_key_id> \
       --aws-region=<aws_region> \
       --aws-secret-access-key=<aws_secret_access_key> \
       --github-token=<github_personal_access_token>
   ```

7. **Copy** private SOPS Age keys from the old path
   to the new directory.

   ```shell
   cp -f "${RMK_SOPS_AGE_KEYS_PATH_OLD}"/* "$(rmk --log-format=json config view | yq '.config.SopsAgeKeys')"
   unset RMK_SOPS_AGE_KEYS_PATH_OLD
   ```

   > Skip this step if you lack **administrator permissions** for the selected AWS account.

8. **Upload** the old private SOPS Age keys to [AWS Secret Manager](https://aws.amazon.com/secrets-manager/).

   ```shell
   rmk secret keys upload
   ```

   > Skip this step if you lack **administrator permissions** for the selected AWS account.

9. **Migrate** the code to the new repository structure, for example, compatible with cluster provisioning using  
   [Kubernetes Cluster API Provider AWS](configuration/cluster-management/usage-aws-provider.md).
