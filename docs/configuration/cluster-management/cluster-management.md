# Cluster management

## Overview

RMK uses [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/introduction) and [K3D](https://k3d.io) for cluster
management.

RMK is suitable for both simple and complex Kubernetes deployments, enabling multi-level project inheritance through
native [Helmfile](https://helmfile.readthedocs.io/en/latest/) functionality.

The 2 scenarios are:

- **A cluster has already been provisioned using third-party tools/services**: An existing Kubernetes context will be
  used
  by RMK.
- **A cluster will be provisioned from scratch using RMK**: Any of the supported cluster providers for RMK, such as
  [AWS](../configuration-management/init-aws-provider.md),
  [Azure](../configuration-management/init-azure-provider.md),
  [GCP](../configuration-management/init-gcp-provider.md),
  [K3D](../configuration-management/init-k3d-provider.md) (local installation)
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

  The next command depends on whether a remote Kubernetes cluster provider
  (e.g., [AWS](../configuration-management/init-aws-provider.md),
  [Azure](../configuration-management/init-azure-provider.md),
  [GCP](../configuration-management/init-gcp-provider.md))
  or a local one (e.g., [K3D](../configuration-management/init-k3d-provider.md)) has been used:

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

## Using RMK cluster providers to provision and destroy target Kubernetes clusters

Currently, the following cluster providers are supported by RMK:

- **[AWS EKS](usage-aws-provider.md)**, **[Azure AKS](usage-azure-provider.md)**, **[GCP GKE](usage-gcp-provider.md)**:

  Configuration for managing remote clusters using Kubernetes Cluster API.

  Such Kubernetes clusters can be provisioned from scratch
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

- **[K3D](usage-k3d-provider.md)**:

  Configuration for managing single-machine clusters using K3D (suitable for both local development and minimal cloud
  deployments).

  Such Kubernetes clusters can be created from scratch and deleted via the
  [rmk cluster k3d create](../../commands.md#create-c) and
  [rmk cluster k3d delete](../../commands.md#delete-d) commands.

> When using the [rmk cluster capi](../../commands.md#capi-c) category commands, RMK **automatically switches** the
> Kubernetes context between the CAPI management cluster and the target Kubernetes cluster.

> **On-premise** support expected in **upcoming releases** This enhancement might include the introduction
> of additional Kubernetes Cluster API providers and
> [Kubernetes operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). The main infrastructure
> configuration can always be checked in the [cluster-deps](https://github.com/edenlabllc/cluster-deps.bootstrap.infra)
> repository.
