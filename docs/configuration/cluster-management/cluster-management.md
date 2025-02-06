# Cluster management

RMK uses [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/introduction) and [K3D](https://k3d.io) for cluster
management.

RMK is suitable for both simple and complex Kubernetes deployments, enabling multi-level project inheritance through
native [Helmfile](https://helmfile.readthedocs.io/en/latest/) functionality.

The 2 scenarios are:

- **A cluster has already been provisioned using third-party tools/services**: An existing Kubernetes context will be
  used
  by RMK.
- **A cluster will be provisioned from scratch using RMK**: Any of the supported cluster providers for RMK, such as
  [AWS](../../configuration/configuration-management/init-aws-provider.md),
  [Azure](../../configuration/configuration-management/init-azure-provider.md),
  [GCP](../../configuration/configuration-management/init-gcp-provider.md),
  [K3D](../../configuration/configuration-management/init-k3d-provider.md) (local installation)
  will be utilized.

## Switching the context to an existing Kubernetes cluster

Switching to an existing Kubernetes cluster depends on how it has been provisioned:

* **Using third-party tools/services**:

  Create a context with the name strictly matching the following:

  ```
  <project_name>-<environment>
  ```

  For example, if you are in
  the [`project1` repository](../project-management/requirement-for-project-repository.md#requirement-for-project-repository)
  in the `develop` branch, any of the following Kubernetes context will be accepted:

  ```
  project1-develop
  ```

* **Using RMK cluster providers**:

  Checkout to the branch from which the Kubernetes cluster was previously created.

  An [initialization](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers)
  might be required, if the RMK configuration for this cluster has not been created before:

  ```shell
  rmk config init --cluster-provider=<aws|azure|gcp|k3d>
  ```

  > The default value for the `--cluster-provider` argument is `k3d`.

  The next command depends on whether a remote Kubernetes cluster provider (e.g., AWS, Azure, GCP) or a local one (e.g.,
  K3D) has
  been used:

  **AWS, Azure, GCP**:

  ```shell
  # --force might be required to refresh the credentials after a long period of inactivity
  rmk cluster switch --force
  ```

  **K3D**:

  Explicit switching to the Kubernetes context is not required, if a K3D cluster has been created already.
  RMK will switch implicitly, when running any of the [rmk release](../../commands.md#release) commands.

Finally, run an RMK release command to verify the preparation of the Kubernetes context, e.g.:

```shell
rmk release list
```

## Using RMK to prepare CAPI management cluster

Before running provisioning and destroying of cloud provider target Kubernetes clusters, a local Kubernetes Cluster API
(CAPI) management cluster must be created:

```shell
rmk cluster capi create
```

> Only one local CAPI management cluster can exist on the cluster administrator machine.
> The cluster can contain all cloud cluster providers at once and work with them independently.

At the time of creation of the local CAPI management cluster, a Kubernetes K3D cluster with a specially
[lightweight configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/capi-cluster.yaml.gotmpl)
will be created.

After creating the CAPI management cluster, RMK will run the
[clusterctl](https://cluster-api.sigs.k8s.io/clusterctl/overview) tool that initializes the installation
of the cloud provider selected at the [rmk config init](../../commands.md#init-i) stage.
It will add specific credentials for the
[selected provider](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers).

The cloud provider version is fixed in the `clusterctl` initialization
[configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/clusterctl-config.yaml.gotmpl)
file, however it can be changed on demand.

**RMK supports the following key operations for the CAPI management cluster**:

- **Creating** a CAPI management cluster based on the cloud provider initialization configuration and provided
  credentials:

  ```shell
  rmk cluster capi create 
  ```

- **Updating**
  the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/clusterctl-config.yaml.gotmpl)
  of the installed provider, credentials, installing an additional cloud provider:

  ```shell
  rmk cluster capi update
  ```

  > RMK allows changing the provider for the same target Kubernetes cluster simply by changing the cloud provider
  > at the [rmk config init](../../commands.md#init-i) stage and update the provider initialization configuration via
  > [rmk cluster capi update](../../commands.md#update-u) command.
  > However, it is not recommended for `production` environments, only for `development` and `staging` or during
  > testing.

- **Deleting** an existing CAPI management cluster:

  ```shell
  rmk cluster capi delete
  ```

  > Important, deleting the CAPI management cluster **will not delete** the target Kubernetes cluster of the cloud
  > provider.

A full list of available commands for working with CAPI management clusters
and for provisioning target Kubernetes clusters can be found [here](../../commands.md#capi-c).

## Using RMK remote cluster providers to provision and destroy target Kubernetes clusters

Currently, the following cluster providers are supported by RMK:

- **AWS (EKS), Azure (AKS), GCP (GKE)**: 

  Configuration for managing clusters using Kubernetes Cluster API. Kubernetes clusters can be provisioned from scratch 
  and destroyed via the [rmk cluster capi provision](../../commands.md#provision-p) and
  [rmk cluster capi destroy](../../commands.md#destroy) commands. All configurations of the description of the provided
  target Kubernetes clusters are described in the values of the [Helmfile](https://helmfile.readthedocs.io/en/latest/) 
  [releases](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/tree/develop/etc/deps/develop/values),
  which in turn use the Helm charts we provide for each individual cloud provider.

  > The Kubernetes Cluster API provider is
  > a [Kubernetes operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/), meaning all configuration
  > changes are applied **declaratively**. Unlike Terraform, it **does not store** the state of managed resources but 
  > instead uses a **resource scanner** to match the current configuration. If a previously created target Kubernetes 
  > cluster needs to be destroyed, but the CAPI management cluster has no record of it, you must first run 
  > [rmk cluster capi provision](../../commands.md#provision-p) before executing 
  > [rmk cluster capi destroy](../../commands.md#destroy).

- **K3D**:

  Configuration for managing
  single-machine clusters using K3D (suitable for both local development and minimal cloud deployments).
  Such Kubernetes clusters can be created from scratch and deleted via the
  [rmk cluster k3d create](../../commands.md#create-c) and
  [rmk cluster k3d delete](../../commands.md#delete-d) commands.

> When using the [rmk cluster capi](../../commands.md#capi-c) category commands, RMK automatically switches the 
> Kubernetes context between the CAPI management cluster and the target Kubernetes cluster.

> Support for on-premise will be implemented in the future. This enhancement might include the introduction
> of Kubernetes Cluster API providers, Kubernetes operators. The main infrastructure configuration can always be checked in the
> [cluster-deps](https://github.com/edenlabllc/cluster-deps.bootstrap.infra) repository.

### Provisioning or destroying AWS EKS Kubernetes clusters

> AWS users must have the
> [PowerUserAccess](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/PowerUserAccess.html),
> [SecretsManagerReadWrite](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/SecretsManagerReadWrite.html)
> permissions to be able to provision and destroy EKS clusters.

Before provisioning the Kubernetes cluster, add override for
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/aws-cluster.yaml.gotmpl)
file to scope `deps` for the target Kubernetes cluster.

```yaml
# A complete list of all options can be found here https://capz.sigs.k8s.io/reference/v1beta1-api
controlPlane:
  spec:
    iamAuthenticatorConfig:
      # UserMappings is a list of user mappings
      mapUsers:
        # TODO: Add a list of users at the downstream project repository level
        - groups:
            - system:masters
          # UserARN is the AWS ARN for the user to map
          userarn: arn:aws:iam::{{ env "AWS_ACCOUNT_ID" }}:user/user1
          # UserName is a kubernetes RBAC user subject*/}}
          username: user1
    version: v1.29.8 # ^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.?(\.0|[1-9][0-9]*)?$

## The machine pools configurations
machinePools:
  app:
    enabled: true
    managed:
      spec:
        instanceType: t3.medium
        # Labels specifies labels for the Kubernetes node objects
        labels:
          db: app
        # Scaling specifies scaling for the ASG behind this pool
        scaling:
          maxSize: 1
          minSize: 1
    # Number of desired machines. Defaults to 1.
    replicas: 1
# ...
```

Using the example above and the example from
the [cluster-deps repository](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/aws-cluster.yaml.gotmpl)
you can add the required number of machine pools depending on the requirements for distribution into individual roles.

> For the AWS provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - create SSH key for cluster nodes.
> - create secrets with private [SOPS Age keys](../secrets-management/secrets-management.md#secret-keys) in the 
>   [AWS Secret Manager](https://aws.amazon.com/secrets-manager/), if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the cluster is ready, RMK automatically switches the Kubernetes context to the newly created cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the previously created SSH key and also delete the context
> for the target Kubernetes cluster.

### Provisioning or destroying Azure AKS Kubernetes clusters

> Azure service principal must have the
> [Contributor](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/privileged#contributor),
> [Key Vault Secrets Officer](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/security#key-vault-secrets-officer)
> roles to be able to provision and destroy AKS clusters.

Before provisioning the Kubernetes cluster, add override for the
[configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/azure-cluster.yaml.gotmpl)
file to scope `deps` for the target Kubernetes cluster.

```yaml
controlPlane:
  spec:
    ## Kubernetes version
    version: v1.29.8

machinePools:
  system:
    enabled: true

  app:
    enabled: true
    replicas: 1
    spec:
      mode: User
      sku: Standard_B2ls_v2
      osDiskSizeGB: 30
      nodeLabels:
        db: app
      scaling:
        minSize: 1
        maxSize: 1
# ...
```

Using the example above and the example from
the [cluster-deps repository](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/azure-cluster.yaml.gotmpl)
you can add the required number of machine pools depending on the requirements for distribution into individual roles.

> For the Azure provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - create secrets with private [SOPS Age keys](../secrets-management/secrets-management.md#secret-keys) in the 
>   [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault), if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the cluster is ready, RMK automatically switches the Kubernetes context to the newly created cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the context for the target Kubernetes cluster.

### Provisioning or destroying GCP GKE Kubernetes clusters

> GCP service account must have the `Editor`, `Secret Manager Admin`, `Kubernetes Engine Admin` 
> [roles](https://cloud.google.com/iam/docs/understanding-roles) to be able to provision and destroy GKE clusters.

Before provisioning the Kubernetes cluster, add override
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/gcp-cluster.yaml.gotmpl)
file to scope `deps` for the target Kubernetes cluster.

```yaml
controlPlane:
  spec:
    version: "v1.30.5"

machinePools:
  app:
    enabled: true
    managed:
      spec:
        # MachineType is the name of a Google Compute Engine
        # (https://cloud.google.com/compute/docs/machine-types).
        # If unspecified, the default machine type is `e2-medium`.
        machineType: "e2-medium"
        management:
          # AutoUpgrade specifies whether node auto-upgrade is enabled for the node
          # pool. If enabled, node auto-upgrade helps keep the nodes in your node pool
          # up to date with the latest release version of Kubernetes.
          autoUpgrade: true
        # MaxPodsPerNode is constraint enforced on the max num of pods per node.
    replicas: 1
# ...
```

Using the example above and the example from
the [cluster-deps repository](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/gcp-cluster.yaml.gotmpl)
you can add the required number of machine pools depending on the requirements for distribution into individual roles.

> For the GCP provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - create a [Cloud NAT](https://cloud.google.com/nat/docs/overview) for outbound traffic cluster nodes.
> - create secrets with private [SOPS Age keys](../secrets-management/secrets-management.md#secret-keys) in the 
>   [GCP Secret Manager](https://cloud.google.com/security/products/secret-manager?hl=en), if they have not been created
>   previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the cluster is ready, RMK automatically switches the Kubernetes context to the newly created cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the previously created 
> [Cloud NAT](https://cloud.google.com/nat/docs/overview) if this resource is no longer used by other clusters in the 
> same region. Also deleting the context for the target Kubernetes cluster.

### Creating or deleting K3D Kubernetes clusters

RMK supports managing single-node Kubernetes clusters using [K3D](https://k3d.io).

The CLI will create a cluster according to the
declarative [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/k3d-cluster.yaml.gotmpl)
for K3D.

> Prerequisites:
>
> 1. Create a separate feature branch: `feature/<issue_key>-<issue_number>-<issue_description>`.
> 2. [Initialize configuration](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-with-a-custom-root-domain)
     for this branch with the `localhost` root domain name.

#### Creating K3D clusters

Run the following command:

```shell
rmk cluster k3d create
```

> By default, RMK will use the current directory for the [--k3d-volume-host-path](../../commands.md#create-c) flag.

> When the Kubernetes cluster is ready, RMK automatically switches the Kubernetes context to the newly created cluster.
> You can create multiple local K3D clusters by separating them with environment Git branches.

#### Deleting K3D clusters

```shell
rmk cluster k3d delete
```
