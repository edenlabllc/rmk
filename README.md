# RMK CLI - Reduced Management for Kubernetes

[![Release](https://img.shields.io/github/v/release/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/releases/latest)
[![Software License](https://img.shields.io/github/license/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/blob/master/LICENSE)
[![Powered By: Edenlab](https://img.shields.io/badge/powered%20by-edenlab-8A2BE2.svg?style=for-the-badge)](https://edenlab.io)

Command-line tool for simplified management and provisioning of [Kubernetes](https://kubernetes.io/) clusters and
environments,
[Helm](https://helm.sh/) secrets and releases, built according to best practices
in [CI/CD](https://www.redhat.com/en/topics/devops/what-is-ci-cd) and [DevOps](https://www.atlassian.com/devops).

Full documentation is available at https://edenlabllc.github.io/rmk/latest/.

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

RMK **simplifies** the setup and management of Kubernetes-based projects of any complexity due to the following
advantages:

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
- **Transparent [project structure](docs/configuration/project-management/preparation-of-project-repository.md) and
  [dependency management](docs/configuration/project-management/dependencies-management-and-project-inheritance.md)**:
  Enables rapid project setup and hierarchical project inheritance, e.g., "parent-child" or "upstream-downstream"
  relationships) between sibling projects to enable release configuration reuse.
- **[Batch](docs/configuration/secrets-management/secrets-management.md#generating-all-secrets-from-scratch) secret
  management**: Automates templating, generation, and encryption of secrets across all environments
  in batch mode.
- **Adheres to the [GitOps](https://www.gitops.tech/) approach**: Uses Git branches as unique identifiers for
  environments, clusters, configurations, and project management in Kubernetes.
- **Follows the [GitLab Flow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) model**: Implements
  a standard branching strategy (`develop`, `staging`, `production`) and ephemeral branches (`feature/*`,
  `release/*`, `hotfix/*`) for
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
- [On-Premise](https://github.com/edenlabllc/on-premise-configurator.operators.infra) a custom-built
  infrastructure provider (operator) based on
  the [Ansible Operator SDK](https://sdk.operatorframework.io/docs/building-operators/ansible/)
  on top of [K3S](https://docs.k3s.io/), inspired by the [k3s-ansible](https://github.com/k3s-io/k3s-ansible) project.

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

**Since 2021**, RMK has been an **integral part** of the company’s Kubernetes infrastructure, used regularly for
automated provisioning and destroy of temporary Kubernetes clusters for development and testing purposes,
both manually and automatically within CI/CD pipelines.

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
- **[on-premise-configurator.operators.infra](https://github.com/edenlabllc/on-premise-configurator.operators.infra)**:
  Kubernetes Operator for declarative configuration of remote bare-metal or virtual machines via SSH using Ansible,
  with support for both isolated (air-gapped) and network-connected environments.
- **[secrets-sync.operators.infra](https://github.com/edenlabllc/secrets-sync.operators.infra)**:
  Kubernetes operator for automatically copying of existing Kubernetes secrets between namespaces.

#### Helm charts

- **[core-charts](https://edenlabllc-core-charts-infra.s3.eu-north-1.amazonaws.com/)**:
  A publicly accessible, [S3-based](https://aws.amazon.com/s3/)
  [Helm chart repository](https://helm.sh/docs/topics/chart_repository/) used by Kodjin, or any other project, e.g.,
  check [cluster-deps Helmfile](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/helmfile.yaml.gotmpl#L49).

## Roadmap

- :construction: 
  **Integration with Helmfile [vals](https://github.com/helmfile/vals)**: Integrate RMK with _vals_for advanced values and secrets management.
- :construction:
  **Enhanced automatic testing of RMK during the CI/CD pipeline:** Ensure that changes to the RMK codebase
  do not introduce errors or regressions during the CI/CD across all cluster providers.
- :construction:
  **Guidelines for contributors:** Create comprehensive guidelines for contributors, including instructions for 
  creating pull requests (PRs).
- :white_check_mark:
  _**Implementation of additional cloud [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) providers:**
  Implement support for other popular Kubernetes services such as
  [GKE](https://cloud.google.com/kubernetes-engine),
  [AKS](https://azure.microsoft.com/en-us/products/kubernetes-service/), etc._
- :white_check_mark:
  _**Implementation of on-premise [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) provider:**
  Implement support for provisioning and destroy of on-premise Kubernetes clusters._
- :white_check_mark: 
  _**Web documentation generator:** Add an HTML documentation [generator](https://www.mkdocs.org/) 
  based on the **.md** files._

> Please refer to [GitHub issues](https://github.com/edenlabllc/rmk/issues) for more information.

## Development and release

The guidelines are available at https://edenlabllc.github.io/rmk/latest/development-and-release/.

## License

RMK is open source software (OSS) licensed under
the [Apache 2.0 License](https://github.com/edenlabllc/rmk/blob/master/LICENSE).

## Code of Conduct

This project adheres to the Contributor
Covenant [Сode of Сonduct](https://github.com/edenlabllc/rmk/blob/master/docs/CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code.
Please refer to our [Contributing Guidelines](https://github.com/edenlabllc/rmk/blob/master/docs/CONTRIBUTING.md) for

further information.
