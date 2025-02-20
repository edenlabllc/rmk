# RMK CLI - Reduced Management for Kubernetes

[![Release](https://img.shields.io/github/v/release/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/releases/latest)
[![Software License](https://img.shields.io/github/license/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/blob/master/LICENSE)
[![Powered By: Edenlab](https://img.shields.io/badge/powered%20by-edenlab-8A2BE2.svg?style=for-the-badge)](https://edenlab.io)

Command-line tool for simplified management and provisioning of [Kubernetes](https://kubernetes.io/) clusters and
environments,
[Helm](https://helm.sh/) secrets and releases, built according to best practices
in [CI/CD](https://www.redhat.com/en/topics/devops/what-is-ci-cd) and [DevOps](https://www.atlassian.com/devops).

* [Overview](#overview)
* [Advantages](#advantages)
* [Supported Kubernetes cluster providers](#supported-kubernetes-cluster-providers)
  * [Provisioned by RMK](#provisioned-by-rmk)
  * [Provisioned using third-party tools and services](#provisioned-using-third-party-tools-and-services)
* [Edenlab LLC use cases](#edenlab-llc-use-cases)
  * [Efficiency in numbers](#efficiency-in-numbers)
  * [Managing clusters](#managing-clusters)
  * [Related repositories](#related-repositories)
    * [GitHub](#github)
    * [Helm charts](#helm-charts)
* [Requirements](#requirements)
  * [Operating systems (OS)](#operating-systems-os)
  * [Software](#software)
* [Installation](#installation)
* [Update](#update)
  * [General update process](#general-update-process)
  * [Updating to specific version](#update-to-specific-version)
* [Quickstart](quickstart.md)
* [Backward compatibility](backward-compatibility.md)
* Configuration
  * [Configuration management](configuration/configuration-management/configuration-management.md)
    * [Initialization of AWS cluster provider](configuration/configuration-management/init-aws-provider.md)
    * [Initialization of Azure cluster provider](configuration/configuration-management/init-azure-provider.md)
    * [Initialization of GCP cluster provider](configuration/configuration-management/init-azure-provider.md)
    * [Initialization of K3D cluster provider](configuration/configuration-management/init-k3d-provider.md)
  * Project management
    * [Requirement for project repository](configuration/project-management/requirement-for-project-repository.md)
    * [Preparation of project repository](configuration/project-management/preparation-of-project-repository.md)
    * [Dependencies management and project inheritance](configuration/project-management/dependencies-management-and-project-inheritance.md)
  * [Cluster management](configuration/cluster-management/cluster-management.md)
    * [Exported environment variables](configuration/cluster-management/exported-environment-variables.md)
    * [Using AWS cluster provider](configuration/cluster-management/usage-aws-provider.md)
    * [Using Azure cluster provider](configuration/cluster-management/usage-azure-provider.md)
    * [Using GCP cluster provider](configuration/cluster-management/usage-azure-provider.md)
    * [Using K3D cluster provider](configuration/cluster-management/usage-k3d-provider.md)
  * [Release management](configuration/release-management/release-management.md)
  * [Secrets management](configuration/secrets-management/secrets-management.md)
* [Commands](commands.md)
* [Roadmap](#roadmap)
* [Development and release](development-and-release.md)
* [License](#license)
* [Code of Conduct](#code-of-conduct)

## Overview

**RMK** stands for "**R**educed **M**anagement for **K**ubernetes".

The main goal of the [CLI](https://en.wikipedia.org/wiki/Command-line_interface) tool is to simplify (**reduce**) the
management of Kubernetes clusters and releases, serving as a "Swiss knife" for daily CI/CD and DevOps tasks while 
allowing **efficient control** with a minimal set of CLI commands.

RMK serves as a **wrapper** for various popular CI/CD and DevOps CLI tools, including:

- [Helmfile](https://helmfile.readthedocs.io/en/latest/)
- [Helm](https://helm.sh/)
- [kubectl](https://kubernetes.io/docs/reference/kubectl/)
- [clusterctl](https://cluster-api.sigs.k8s.io/clusterctl/overview)
- [K3D](https://k3d.io/)
- [SOPS](https://getsops.io/)
- [Age](https://age-encryption.org/)

It leverages [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) for cluster provisioning and management across
different environments, such as [cloud providers](https://en.wikipedia.org/wiki/Cloud_computing)
and [on-premise](https://en.wikipedia.org/wiki/On-premises_software) deployments.

RMK has been designed to be **used by different IT specialists**, among them are DevOps engineers, software developers,
SREs,
cloud architects, system analytics, software testers and even managers with minimal technical background.

## Advantages

RMK **simplifies** the setup and management of Kubernetes-based projects of any complexity due to the following advantages:

- **[Time-proven](#efficiency-in-numbers) CI/CD solution**: Tested and validated across multiple cloud providers and 
  real customers, RMK leverages [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) for cluster provisioning
  and [Helmfile](https://helmfile.readthedocs.io/en/latest/)/[Helm](https://helm.sh/) for efficient release and secrets
  management.
- **Seamless integration with [CI/CD](https://www.redhat.com/en/topics/devops/what-is-ci-cd) platforms**: A
  self-sufficient, portable binary that follows the [12-factor app](https://12factor.net/) methodology and can
  easily be integrated with any CI/CD solution.
- **Built-in [versioning](https://en.wikipedia.org/wiki/Software_versioning) for CI/CD pipelines**: Supports static and
  dynamic tags (e.g., [SemVer2](https://semver.org/)) for project and releases to guarantee stable, well-tested, and
  predictable deployments.
- **Transparent [project structure](configuration/project-management/preparation-of-project-repository.md) and
  [dependency management](configuration/project-management/dependencies-management-and-project-inheritance.md)**:
  Enables rapid project setup and hierarchical project inheritance, e.g., "parent-child" or "upstream-downstream"
  relationships) between sibling projects to enable release configuration reuse.
- **[Batch](configuration/secrets-management/secrets-management.md#generating-all-secrets-from-scratch) secret
  management**: Automates templating, generation, and encryption of secrets across all environments
  in batch mode.
- **Adheres to the [GitOps](https://www.gitops.tech/) approach**: Uses Git branches as unique identifiers for
  environments, clusters, configurations, and project management in Kubernetes.
- **Follows the [GitLab Flow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) model**: Implements
  a standard branching strategy (`develop`, `staging`, `production`) and ephemeral branches (`feature/*`, `release/*`) for
  temporary environments.
- **Aligns with the [DevOps](https://www.atlassian.com/devops) methodology**: Enables multiple teams to develop and
  release independently while seamlessly integrating their work into a single project.
- **Directly executes the wrapped [CLI tools](#overview)**: Calls CLI tools as a user would, passing the correct
  arguments and flags
  based on the project configuration, ensuring RMK updates remain decoupled from CLI tool updates for continued access
  to new features and fixes.

## Supported Kubernetes cluster providers

### Provisioned by RMK

RMK currently supports the **provisioning** of the following Kubernetes clusters:

- [Amazon Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks/)
- [Azure Kubernetes Service (AKS)](https://azure.microsoft.com/en-us/products/kubernetes-service/)
- [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)
- Single-machine [K3D](https://k3d.io/) clusters

> Please see the [Roadmap](#roadmap) section for more details on upcoming features.

### Provisioned using third-party tools and services

By design, RMK can work with **any existing Kubernetes cluster**, provided it has been provisioned in advance by a third
party. The CLI tool simply requires an existing
[Kubernetes context](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)
to connect to and manage the cluster.

## Edenlab LLC use cases

### Efficiency in numbers

Initially, it has been developed by [Edenlab LLC](https://edenlab.io/) as the main CLI for provisioning and managing
[Kodjin FHIR Server](https://kodjin.com) on Kubernetes clusters in different environments.

**Since 2021**, RMK has been an **integral part** of the company’s Kubernetes infrastructure, used regularly for automated
provisioning and destroy of temporary Kubernetes clusters for development and testing purposes, both manually and 
automatically within CI/CD pipelines.

**:rocket: Proven at scale**:

- **220+** clusters handled **monthly** (based on a 5-day workweek).
- **2,600+** clusters handled **annually**.
- **10,000+** clusters orchestrated **since 2021**.

Beyond internal use, RMK is also leveraged by various **external clients** to streamline their CI/CD workflows, ensuring
fast and
efficient Kubernetes environment management.

### Managing clusters

At [Edenlab LLC](https://edenlab.io/), RMK is utilized to deploy the [Kodjin FHIR Server](https://kodjin.com)
across various **cloud providers** and **on-premise** environments.

Examples of Kubernetes providers where Kodjin has already been deployed include:

- [Amazon Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks/)
- [Azure Kubernetes Service (AKS)](https://azure.microsoft.com/en-us/products/kubernetes-service/)
- [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)
- [Open Telekom Cloud - Cloud Container Engine (CCE)](https://www.open-telekom-cloud.com/en/products-services/core-services/cloud-container-engine)
- [Rancher Kubernetes Platform](https://www.rancher.com/)
- [Kubermatic Kubernetes Platform (KKP)](https://www.kubermatic.com/)
- [On-premise](https://en.wikipedia.org/wiki/On-premises_software) deployments
- Single-machine [K3D](https://k3d.io/) clusters

A standard Kodjin-based cluster follows a **4-level inheritance** structure:

- **[cluster-deps](https://github.com/edenlabllc/cluster-deps.bootstrap.infra) (upstream#1)**:
  Provides [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) and essential system components required by RMK
  for provisioning Kubernetes clusters across various providers.
- **Dependencies (upstream#2)**:
  Includes core components such as databases, search engines, caches, load balancers/proxies, and operators.
  etc., uses [cluster-deps](https://github.com/edenlabllc/cluster-deps.bootstrap.infra) as its primary project
  dependency.
- **[Kodjin](https://kodjin.com/) (downstream#1)**:
  A set of [Rust](https://www.rust-lang.org/) microservices that form the Kodjin FHIR
  API ([REST](https://en.wikipedia.org/wiki/REST)).
- **Target project (tenant) (downstream#2)**:
  Encompasses products built on top of Kodjin, including UI components, user portals, and middleware services, such as
  the
  e.g., [Kodjin Demo FHIR Server](https://demo.kodjin.com/)

Each project repository **follows** a
standard [GitLab Flow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) branching model.

### Related repositories

#### GitHub

- **[cluster-deps.bootstrap.infra](https://github.com/edenlabllc/cluster-deps.bootstrap.infra)**:
  [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) and system components required for provisioning of
  Kubernetes clusters for different providers.
- **[helmfile.hooks.infra](https://github.com/edenlabllc/helmfile.hooks.infra)**:
  A collection of shell scripts used as [Helmfile hooks](https://helmfile.readthedocs.io/en/latest/#hooks) in
  dependencies, Kodjin, or any other project,
  e.g.,
  check [cluster-deps global configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/globals.yaml.gotmpl#L16)).
- **[aws-iam-provisioner.operators.infra](https://github.com/edenlabllc/aws-iam-provisioner.operators.infra)**:
  Kubernetes operator for automatic provisioning of IAM roles on the fly for the Kubernetes clusters managed
  using [Kubernetes Cluster API Provider AWS](https://cluster-api-aws.sigs.k8s.io/getting-started).
- **[ebs-snapshot-provision.operators.infra](https://github.com/edenlabllc/ebs-snapshot-provision.operators.infra)**:
  Kubernetes operator for automatic provisioning of Amazon EBS snapshots to be used in existing Kubernetes clusters.
- **[ecr-token-refresh.operators.infra](https://github.com/edenlabllc/ecr-token-refresh.operators.infra)**:
  Kubernetes operator for automatic refresh of the Amazon ECR authorization token before it expires.
- **[secrets-sync.operators.infra](https://github.com/edenlabllc/secrets-sync.operators.infra)**:
  Kubernetes operator for automatically copying of existing Kubernetes secrets between namespaces.

#### Helm charts

- **[core-charts](https://edenlabllc-core-charts-infra.s3.eu-north-1.amazonaws.com/)**:
  A publicly accessible, [S3-based](https://aws.amazon.com/s3/) 
  [Helm chart repository](https://helm.sh/docs/topics/chart_repository/) used by Kodjin, or any other project, e.g.,
  check [cluster-deps Helmfile](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/helmfile.yaml.gotmpl#L49).

## Requirements

### Operating systems (OS)

Currently, RMK only supports [Unix-like](https://en.wikipedia.org/wiki/Unix) operating systems (OS):

- **MacOS**: amd64, arm64 ([M series](https://en.wikipedia.org/wiki/Apple_silicon#M_series) processors
  require [Rosetta](https://support.apple.com/en-us/HT211861))
- **Linux**: amd64

### Software

The following software is required to run RMK:

- **[Git](https://git-scm.com/)**
- **[for [local K3D clusters](configuration/configuration-management/init-k3d-provider.md)]**:
  Version [v5.X.X](https://k3d.io/stable/#learning)
  requires **[Docker](https://www.docker.com/)** >=
  v20.10.5 (**[runc](https://github.com/opencontainers/runc)** >= v1.0.0-rc93) to work properly.

## Installation

To install RMK, run the self-installer script using the following command:

```shell
curl -sL "https://edenlabllc-rmk.s3.eu-north-1.amazonaws.com/rmk/s3-installer" | bash
```

Alternatively, you can go directly to the [releases](https://github.com/edenlabllc/rmk/releases) and download the
binary or [build from source](development-and-release.md#building-from-source).

## Update

### General update process

To update RMK to the latest version, run the following command:

```shell
rmk update
```

### Updating to specific version

You can update to a specific RMK version to maintain backward compatibility or when updating to the latest version is
not possible.

> This may be necessary due to specific version requirements or when a bug has been detected.

To update to a specific version, use the following command:

```shell
rmk update --version vX.X.X 
```

## Roadmap

- **Integration with Helmfile [vals](https://github.com/helmfile/vals)**: Integrate RMK with _vals_ for advanced
  values and secrets management.
- **Implementation of on-premise [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) provider:** Implement
  support for provisioning and destroy of on-premise Kubernetes clusters.
- **Automatic testing of RMK during the CI/CD pipeline:** Ensure that changes to the RMK codebase do not introduce
  errors or regressions during the CI/CD across all cluster providers.
- **Guidelines for contributors:** Create comprehensive guidelines for contributors, including instructions for creating
  pull requests (PRs).

> Please refer to [GitHub issues](https://github.com/edenlabllc/rmk/issues) for more information.

## Development and release

The guidelines are available at the [link](development-and-release.md).

## License

RMK is open source software (OSS) licensed under
the [Apache 2.0 License](https://github.com/edenlabllc/rmk/blob/master/LICENSE).

## Code of Conduct

This project adheres to the Contributor
Covenant [Сode of Сonduct](https://github.com/edenlabllc/rmk/blob/master/docs/CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code.

Please refer to our [Contributing Guidelines](https://github.com/edenlabllc/rmk/blob/master/docs/CONTRIBUTING.md) for
further information.
