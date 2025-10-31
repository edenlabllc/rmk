# Initialization of K3D cluster provider

## List of main attributes of the RMK configuration

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the project (tenant) name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant name, which is equivalent to the project name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
cluster-provider: k3d # Selected cluster provider.
# ...
```

## Prerequisites

1. Create a separate ephemeral **branch**, e.g.: `feature/<issue_key>-<issue_number>-<issue_description>`.
2. [Initialize configuration](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-with-a-custom-root-domain)
   for this branch with the `localhost` root domain name.

## Configuration

K3D is the default cluster provider in RMK. It is intended for provisioning **local** clusters,
primarily for **development** environments.

To initialize RMK configuration for a K3D cluster, run:

```shell
rmk config init
```

> If SOPS Age keys are **already created** and defined in 
> [project.yaml](../project-management/preparation-of-project-repository.md#projectyaml) using 
> [Vals](https://github.com/helmfile/vals),
> they will be **fetched automatically** during configuration initialization — just ensure that the required environment
> variables are **exported**. See the
> [following page](../secrets-management/helmfile-vals-integration.md#configuration-initialization)
> for details.
