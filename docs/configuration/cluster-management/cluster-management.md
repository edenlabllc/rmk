# Cluster management

RMK uses [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/introduction) and [K3D](https://k3d.io) for cluster management.

RMK is suitable for both simple and complex Kubernetes deployments, enabling multi-level project inheritance through
native Helmfile functionality.

The 2 scenarios are:

- **A cluster has already been provisioned via 3rd-party tools/services**: An existing Kubernetes context will be used
  by RMK.
- **A cluster will be provisioned from scratch using RMK**: Any of the supported cluster providers for RMK,
  such as AWS, Azure, GCP, K3D will be utilized.

## Switch the context to an existing Kubernetes cluster

Switching to an existing Kubernetes cluster depends on how it has been provisioned:

* **Using a 3rd party tool**:

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

  Checkout to the branch from which the K8S cluster was previously created.

  An [initialization](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers)
  might be required, if the RMK configuration for this cluster has not been created before:

  ```shell
  rmk config init --cluster-provider=<aws|azure|gcp|k3d>
  ```
  
  > The default cluster provider is `k3d`.

  The next command depends on whether a remote cluster provider (e.g., AWS, Azure, GCP) or a local one (e.g., K3D) has
  been used:

    * **AWS | Azure | GCP**:

      ```shell
      # --force might required to refresh the credentials after a long period of inactivity
      rmk cluster switch --force
      ```

    * **K3D**:

      Explicit switching to the Kubernetes context is not required, if a K3D cluster has been created already.
      RMK will switch implicitly, when running any of the `rmk release` commands.

Finally, run an RMK release command to verify the preparation of the Kubernetes context, e.g.:

```shell
rmk release list
```

## Use RMK to prepare CAPI management cluster

Before running provisioning and destroying of cloud provider target clusters, a local CAPI management cluster must be
created:

```shell
rmk cluster capi create
```

> Only one local CAPI management cluster can exist on the cluster administrator machine.
> A CAPI management cluster can contain all cloud cluster providers at once and work with them independently.

At the time of creation of the local CAPI management cluster, a Kubernetes K3D cluster with a specially
[lightweight configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/capi-cluster.yaml.gotmpl)
will be created.

After creating the CAPI management cluster, RMK will launch the `clusterсtl` tool that initializes the installation
of the cloud provider selected at the `rmk config init` stage.
It will add specific credentials for the selected
provider [provided](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers)
at the `rmk config init` stage.

The cloud provider version is also fixed in the `clusterctl`
initialization [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/clusterctl-config.yaml.gotmpl)
file and can be changed on demand.

**RMK supports the following key operations for the CAPI management cluster**:

- Create a CAPI management cluster based on the cloud provider initialization configuration.
  Provide credentials for the selected provider:

  ```shell
  rmk cluster capi create 
  ```

- Updating
  the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/clusterctl-config.yaml.gotmpl)
  of the installed provider. Updating credentials. Installing an additional cloud provider:

  ```shell
  rmk cluster capi update
  ```

  > RMK allows changing the provider for the same target cluster simply by changing the cloud provider
  > at the `rmk config init` stage and update the provider initialization configuration via `rmk cluster capi update`
  command.
  > But we do not recommend using this for production environments. Only for development and testing.

- Deleting an existing CAPI management cluster:

  ```shell
  rmk cluster capi delete
  ```

  > Important, **deleting** the CAPI management cluster will not delete the target cluster of the cloud provider.

A full list of available commands for working with the CAPI management cluster
and for provisioning target clusters can be found [here](../../commands.md#capi-c).

## Use RMK remote cluster providers to provision and destroy Kubernetes clusters

Currently, the following cluster providers are supported by RMK:

- **AWS EKS | Azure AKS | GCP GKE**: Configuration for managing clusters using Kubernetes Cluster API.
  Kubernetes clusters can be provisioned from scratch and destroyed
  via the `rmk cluster capi provision`, `rmk cluster capi destroy` commands.
  All configurations of the description of the provided target clusters are described in the values of
  the `Helmfile` [releases](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/tree/develop/etc/deps/develop/values).
  Which in turn use the Helm charts we provide for each individual cloud provider.

  > Since the Kubernetes Cluster API provider is essentially a Kubernetes operator, all configuration changes are applied
  declarative.
  > It also does not store its state of managed resources like Terraform, but it does have a dynamic resource scanner
  > to match the current configuration. This leads to the fact that if it is necessary to destroy a previously
  > created target cluster and the CAPI management cluster did not previously contain this information on the target
  cluster,
  > then it will be necessary to first run `rmk cluster capi provision` command and then a `rmk cluster capi destroy`
  command.

- **K3D**: Configuration for managing
  single-machine clusters using K3D (suitable for both local development and minimal cloud deployments).
  Kubernetes clusters can be created from scratch and deleted via the `rmk cluster k3d create`, `rmk cluster k3d delete`
  commands.

> When using the `rmk cluster capi` category commands, RMK automatically switches the Kubernetes context between
> the CAPI management cluster and the target cluster.

Support for on-premise will be implemented in the future. This enhancement might include the introduction
of Kubernetes Cluster API providers, K8S operators. The main infrastructure configuration can always be checked in the
[cluster-deps](https://github.com/edenlabllc/cluster-deps.bootstrap.infra) repository.

### Provision or destroy AWS EKS Kubernetes clusters

> AWS users must have the `PowerUserAccess`, `SecretsManagerReadWrite` permissions to be able to provision and destroy
> EKS clusters.

Before provisioning the K8S cluster, add override
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/aws-cluster.yaml.gotmpl)
file to scope deps for the target cluster.

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

> For AWS provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
> 
> - create SSH key for cluster nodes.
> - create secrets with private SOPS Age keys in the AWS Secret Manager, if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the K8S cluster is ready, RMK automatically switches the kubectl context to the newly created K8S cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the previously created SSH key and also delete the Kubernetes context
> for the target cluster.

### Provision or destroy Azure AKS Kubernetes clusters

> Azure service principal must have the `Contributor`, `Key Vault Secrets Officer` roles to be able to provision and
> destroy AKS clusters.

Before provisioning the K8S cluster, add override
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/azure-cluster.yaml.gotmpl)
file to scope deps for the target cluster.

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

> For AWS provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
> 
> - create secrets with private SOPS Age keys in the Azure Key Vault, if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the K8S cluster is ready, RMK automatically switches the kubectl context to the newly created K8S cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the Kubernetes context for the target cluster.

### Provision or destroy GCP GKE Kubernetes clusters

> GCP service account must have the `Editor`, `Secret Manager Admin`, `Kubernetes Engine Admin` roles to be able to
> provision and destroy EKS clusters.

Before provisioning the K8S cluster, add override
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/gcp-cluster.yaml.gotmpl)
file to scope deps for the target cluster.

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

> For AWS provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - create Cloud NAT for outbound traffic cluster nodes.
> - create secrets with private SOPS Age keys in the GCP Secret Manager, if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the K8S cluster is ready, RMK automatically switches the kubectl context to the newly created K8S cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the previously created Cloud NAT if this Cloud NAT is no longer used
> by other clusters in the same region. Also deleting the Kubernetes context for the target cluster.

### Create or delete K3D Kubernetes clusters

RMK supports managing single-node Kubernetes clusters using [K3D](https://k3d.io).

The CLI will create a cluster according to the
declarative [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/k3d-cluster.yaml.gotmpl)
for K3D.

> Prerequisites:
>
> 1. Create a separate feature branch: `feature/<issue_key>-<issue_number>-<issue_description>`.
> 2. [Initialize configuration](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-with-a-custom-root-domain)
     for this branch with the `localhost` root domain name.

#### Create K3D clusters

> By default, RMK will use `volume-host-path` as the current directory:

Run the following command:

```shell
rmk cluster k3d create
```

> When the Kubernetes cluster is ready, RMK automatically switches the kubectl context to the newly created Kubernetes
> cluster.
> You can create multiple local K3D clusters by separating them with environment git branches.

#### Delete K3D clusters

```shell
rmk cluster k3d delete
```
