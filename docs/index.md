# RMK CLI - Reduced Management for Kubernetes

[![Release](https://img.shields.io/github/v/release/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/releases/latest)
[![Software License](https://img.shields.io/github/license/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/blob/master/LICENSE)
[![Powered By: Edenlab](https://img.shields.io/badge/powered%20by-edenlab-8A2BE2.svg?style=for-the-badge)](https://edenlab.io)

Command-line tool for simplified management and provisioning of [Kubernetes](https://kubernetes.io/) clusters and
environments,
[Helm](https://helm.sh/) secrets and releases, built according to best practices
in [CI/CD](https://www.redhat.com/en/topics/devops/what-is-ci-cd) and [DevOps](https://www.atlassian.com/devops).

* [Overview](#overview)
  * [Efficiency in numbers](#efficiency-in-numbers)
  * [Advantages](#advantages)
  * [Edenlab LLC use cases](#edenlab-llc-use-cases)
    * [Related repositories](#related-repositories)
      * [GitHub](#github)
      * [Helm charts](#helm-charts)
* [Requirements](#requirements)
  * [Operating systems (OS)](#operating-systems-os)
  * [Software](#software)
* [Quickstart](quickstart.md)
* [Installation](#installation)
* [Update](#update)
  * [General update process](#general-update-process)
  * [Update to specific version](#update-to-specific-version)
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
  * [Release management](configuration/release-management/release-management.md)
  * [Secrets management](configuration/secrets-management/secrets-management.md)
* [Commands](commands.md)
* [Features](#features)
* [Supported Kubernetes cluster providers](#supported-kubernetes-cluster-providers)
* [Roadmap](#roadmap)
* [Development and release](development-and-release.md)
* [License](#license)
* [Code of Conduct](#code-of-conduct)

## Overview

**RMK** stands for "**R**educed **M**anagement for **K**ubernetes".
The main goal of the [CLI](https://en.wikipedia.org/wiki/Command-line_interface) tool is to simplify (**"reduce"**) the
management of Kubernetes clusters and releases,
serving as a "Swiss knife" for daily CI/CD and DevOps tasks while allowing efficient control
with a minimal set of CLI commands.

RMK serves as a wrapper for various popular CI/CD and DevOps CLI tools, including:

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

RMK has been designed to be used by different IT specialists, among them are DevOps engineers, software developers,
SREs,
cloud architects, system analytics, software testers and even managers with minimal technical background.

### Efficiency in numbers

Initially, it has been developed by [Edenlab LLC](https://edenlab.io/) as the main CLI for provisioning and managing
[Kodjin FHIR Server](https://kodjin.com) on Kubernetes clusters in different environments.

Since 2021, RMK has been an integral part of the company’s Kubernetes infrastructure, used regularly for automated
provisioning and destroy of temporary Kubernetes clusters for development and testing purposes, both manually and within
CI/CD pipelines.

**:rocket: Proven at scale**:

- **220+** clusters handled **monthly** (based on a 5-day workweek).
- **2,600+** clusters handled **annually**.
- **10,000+** clusters orchestrated **since 2021**.

Beyond internal use, RMK is also leveraged by various external clients to streamline their CI/CD workflows, ensuring
fast and
efficient Kubernetes environment management.

### Advantages

RMK simplifies the setup and management of Kubernetes-based projects of any complexity level due to the following
advantages:

- **Includes everything needed for daily CI/CD and DevOps tasks**:
  Utilizes [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/)
  for provisioning clusters across various environments and providers,
  and [Helmfile](https://helmfile.readthedocs.io/en/latest/)/[Helm](https://helm.sh/)
  along with other tools for efficient release, secrets management across multiple clusters.
- **Integrates with any [CI/CD](https://www.redhat.com/en/topics/devops/what-is-ci-cd) tool easily**: A self-sufficient,
  portable binary that strictly follows the [12 factor app](https://12factor.net/) methodology.
- **Supports [versioning](https://en.wikipedia.org/wiki/Software_versioning) of projects in a CI/CD pipeline**: Each
  project can be versioned and referenced by static or dynamic tags (e.g., [SemVer2](https://semver.org/)),
  which guarantees stable, well-tested and predictable releases.
- **Provides a transparent project structure, generation from scratch, dependency management:** Enables rapid project
  setup, efficient configuration reuse across projects, inheritance from other projects, e.g., "parent-child" or "
  upstream-downstream" relationships.
- **Respects the [GitOps](https://www.gitops.tech/) approach:** Each Git branch is used as a unique identifier for
  determining the environment, cluster name,
  set of configurations and other attributes required for setting up the wrapped tools for project management in the
  Kubernetes environment.
- **Respects the [GitLab Flow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) workflow**: Follows
  a standard branching model with _develop_, _staging_, _production_, and ephemeral branches (e.g., _feature_,
  _release_) for temporary environments.
- **Respects the [DevOps](https://www.atlassian.com/devops) methodology:** Allows diverse teams to work without blocking
  each other. Each team or multiple teams can develop and release their projects separately, later on the result of
  their work can be combined in a single project.
- **Calls the CLI tools directly instead of using their libraries/SDKs**: RMK executes the tools directly in a way that
  a typical person would do it,
  passing correct sets of CLI arguments and flags to the commands based on a project configuration structure.
  This decouples the updating of RMK itself from the wrapped CLI tools, allowing developers to utilize recent
  functionality and fixes.

### Edenlab LLC use cases

At [Edenlab LLC](https://edenlab.io/), RMK is utilized to deploy the [Kodjin FHIR Server](https://kodjin.com)
across various cloud providers and on-premise environments.

Examples of Kubernetes providers where Kodjin has already been deployed include:

- [Amazon Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks/)
- [Azure Kubernetes Service (AKS)](https://azure.microsoft.com/en-us/products/kubernetes-service/)
- [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)
- [Open Telekom Cloud - Cloud Container Engine (CCE)](https://www.open-telekom-cloud.com/en/products-services/core-services/cloud-container-engine)
- [Rancher Kubernetes Platform](https://www.rancher.com/)
- [Kubermatic Kubernetes Platform (KKP)](https://www.kubermatic.com/)
- [On-premise](https://en.wikipedia.org/wiki/On-premises_software) deployments
- Single-machine [K3D](https://k3d.io/) clusters

A standard Kodjin-based cluster follows a 4-level inheritance structure:

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
- **Target tenant (downstream#2)**:
  Encompasses products built on top of Kodjin, including UI components, user portals, and middleware services, such as
  the
  e.g., [Kodjin Demo FHIR Server](https://demo.kodjin.com/)

Each project ("tenant") follows a
standard [GitLab Flow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) branching model.

#### Related repositories

##### GitHub

- **[cluster-deps.bootstrap.infra](https://github.com/edenlabllc/cluster-deps.bootstrap.infra)**:
  [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) and system components required for provisioning of
  Kubernetes clusters for different providers.
- **[helmfile.hooks.infra](https://github.com/edenlabllc/helmfile.hooks.infra)**:
  A collection of shell scripts used as [Helmfile hooks](https://helmfile.readthedocs.io/en/latest/#hooks) in
  dependencies, Kodjin, or any other tenant,
  e.g.,
  check [cluster-deps global configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/globals.yaml.gotmpl#L16).
- **[aws-iam-provisioner.operators.infra](https://github.com/edenlabllc/aws-iam-provisioner.operators.infra):** K8S
  operator for automatic provisioning of IAM roles on the fly for the Kubernetes clusters managed
  using [Kubernetes Cluster API Provider AWS](https://cluster-api-aws.sigs.k8s.io/getting-started).
- **[ebs-snapshot-provision.operators.infra](https://github.com/edenlabllc/ebs-snapshot-provision.operators.infra):**
  K8S operator for automatic provisioning of Amazon EBS snapshots to be used in existing K8S clusters.
- **[ecr-token-refresh.operators.infra](https://github.com/edenlabllc/ecr-token-refresh.operators.infra):** K8S operator
  for automatic refresh of the Amazon ECR authorization token before it expires.
- **[secrets-sync.operators.infra](https://github.com/edenlabllc/secrets-sync.operators.infra):** K8S operator for
  automatically copying of existing K8S secrets between namespaces.

##### Helm charts

- **[core-charts](https://edenlabllc-core-charts-infra.s3.eu-north-1.amazonaws.com/)**: A publicly
  accessible, [S3-based](https://aws.amazon.com/s3/) [Helm chart repository](https://helm.sh/docs/topics/chart_repository/)
  used by Kodjin, or any other tenant, e.g.,
  check [cluster-deps Helmfile](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/helmfile.yaml.gotmpl#L49).

## Requirements

### Operating systems (OS)

Currently, RMK only supports Unix-like operating systems (OS):

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

Alternatively, you can go directly to https://github.com/edenlabllc/rmk/releases and download the binary.

As another option, the binary can be [built from source](development-and-release.md#building-from-source).

## Update

### General update process

To update RMK to the latest version, run the following command:

```shell
rmk update
```

### Update to specific version

You can update to a specific RMK version to maintain backward compatibility or when updating to the latest version is
not possible.

> This may be necessary due to specific version requirements or when a bug has been detected.

To update to a specific version, use the following command:

```shell
rmk update --version vX.X.X 
```

## Features

- **[Simplified management of Kubernetes clusters:](#overview)**
  Provision Kubernetes clusters across various environments and providers, streamline release management 
  with [Helmfile](https://helmfile.readthedocs.io/en/latest/)/[Helm](https://helm.sh/) and other popular DevOps tools.
- **[Time-proven project structure:](configuration/project-management/preparation-of-project-repository.md)** 
  Define the project structure using the [GitLab Flow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) methodology.
- **[Hierarchies between different projects:](configuration/project-management/dependencies-management-and-project-inheritance.md)** 
  Define upstream-downstream relationships between sibling projects to reuse releases and services across different installations.
- **[Initial project structure generation:](configuration/project-management/preparation-of-project-repository.md#automatic-generation-of-the-project-structure-from-scratch)** 
  Generate a complete Kubernetes-based project structure from scratch using RMK, following industry best practices.
- **[Batch secret management:](configuration/secrets-management/secrets-management.md#generating-all-secrets-from-scratch)**
  Template, generate, and encode project secrets for all environments in a batch manner.
- **[Self-documentation generation:](commands.md#doc)**
  Generate the full command documentation in the [Markdown](https://en.wikipedia.org/wiki/Markdown) format with one click.
- **[Push-based release and downstream project updates:](configuration/release-management/release-management.md#release-update-and-integration-into-the-cd-pipeline)**
  Easily integrate with CI/CD solutions via [webhooks](https://www.redhat.com/en/topics/automation/what-is-a-webhook) 
  or [workflow dispatch events](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#workflow_dispatch)
  to update release and service version declarations, automatically commit the changes to Git.
- **[Automatic detection of Multi-Factor Authentication (MFA) for AWS](configuration/configuration-management/init-aws-provider.md#support-for-multi-factor-authentication-mfa):** 
  Automatically detect and use an ([MFA](https://en.wikipedia.org/wiki/Multi-factor_authentication)) device 
  if one is defined by an [IAM](https://aws.amazon.com/iam/) user (must be supported by the cluster provider, e.g., [AWS](https://aws.amazon.com/)).

## Supported Kubernetes cluster providers

### Provisioned by RMK

RMK currently supports the provisioning of the following Kubernetes clusters:

- [Amazon Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks/)
- [Azure Kubernetes Service (AKS)](https://azure.microsoft.com/en-us/products/kubernetes-service/)
- [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)
- Single-machine [K3D](https://k3d.io/) clusters

> Check the [Roadmap](#roadmap) section for more information.

### Provisioned using 3rd-party tools and services

By design, RMK can work with any existing Kubernetes cluster, provided it has been provisioned in advance by a 3rd
party. The CLI tool simply requires an existing 
[Kubernetes context](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)
to connect to and manage the cluster.

## Roadmap

- **Integration with Helmfile [vals](https://github.com/helmfile/vals)**: Integrate RMK with **vals** for advanced
  values and secrets management.
- **Implementation of on-premise [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) provider:** Implement
  support for provisioning and destroy of on-premise Kubernetes clusters using Cluster API.
- **Automatic testing of RMK during the CI/CD pipeline:** Ensure that changes to the RMK codebase do not introduce
  errors or regressions during the CI/CD.
- **Guidelines for contributors:** Create comprehensive guidelines for contributors, including instructions for creating
  pull requests (PRs).

> Check [GitHub issues](https://github.com/edenlabllc/rmk/issues) for more information.

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
