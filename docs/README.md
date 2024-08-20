# RMK CLI - Reduced Management for Kubernetes

[![Release](https://img.shields.io/github/v/release/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/releases/latest)
[![Software License](https://img.shields.io/github/license/edenlabllc/rmk.svg?style=for-the-badge)](LICENSE)
[![Powered By: Edenlab](https://img.shields.io/badge/powered%20by-edenlab-8A2BE2.svg?style=for-the-badge)](https://edenlab.io)

Command line tool for reduced management and provisioning of Kubernetes clusters and environments, Helm secrets and releases.

* [RMK CLI - Reduced Management for Kubernetes](#rmk-cli---reduced-management-for-kubernetes)
  * [Overview](#overview)
    * [Advantages](#advantages)
    * [Edenlab LLC use cases](#edenlab-llc-use-cases)
      * [Related OSS repositories](#related-oss-repositories)
  * [Requirements](#requirements)
  * [Quickstart](quickstart.md)
  * [Installation](#installation)
  * [Update](#update)
    * [General update process](#general-update-process)
    * [Update to specific version](#update-to-specific-version)
  * Configuration
    * [RMK configuration management](configuration/rmk-configuration-management.md)
    * Project management
      * [Requirement for project repository](configuration/project-management/requirement-for-project-repository.md)
      * [Preparation of project repository](configuration/project-management/preparation-of-project-repository.md)
      * [Dependencies management and Project inheritance](configuration/project-management/dependencies-management-and-project-inheritance.md)
    * [Cluster management](configuration/cluster-management/cluster-management.md)
      * [Exported environment variables](configuration/cluster-management/exported-environment-variables.md)
    * [Release management](configuration/release-management/release-management.md)
    * [Secrets management](configuration/secrets-management/secrets-management.md)
  * [Commands](commands.md)
  * [Development and release flow](development-and-release-flow.md)
  * [Features](#features)
  * [Supported Kubernetes providers](#supported-kubernetes-providers)
  * [Roadmap](#roadmap)
  * [License](#license)
  * [Code of Conduct](#code-of-conduct)

## Overview

This tool has been designed and developed initially by [Edenlab LLC](https://edenlab.io/) as the main CLI
for managing [Kodjin FHIR Server](https://kodjin.com) on Kubernetes clusters in different environments.

It is a wrapper around many popular CI/CD and DevOps CLI tools, including:

- [Helmfile](https://helmfile.readthedocs.io/en/latest/)
- [Helm](https://helm.sh/)
- [kubectl](https://kubernetes.io/reference/kubectl/)
- [SOPS](https://getsops.io/)
- [Terraform](https://www.terraform.io/)
- [K3D](https://k3d.io/)

The main goal of the tool is to simplify ("reduce") management of Kubernetes clusters and releases.

**RMK** is an abbreviation which stands for "**R**educed **M**anagement for **K**ubernetes".

### Advantages

RMK simplifies the start of any level of complexity of a project using Kubernetes due to the following advantages:
- **Respects the [GitOPS](https://www.gitops.tech/) approach:** Each Git branch is used as a unique identifier for determining the environment, cluster name, 
  set of configurations and other attributes required for setting up the wrapped tools for project management in the Kubernetes environment.
- **Respects the [GitLabFlow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) workflow**: Supports the standard _develop_, _staging_, _production_ and different ephemeral (e.g, _feature_, _release_) environments.
- **Provides a transparent project structure with a basic set of configurations**: Allows you to correctly reuse configurations between projects 
  and inherit project configurations from other repositories, e.g., establish parent-child ("upstream-downstream") project relationships.
- **Allows a diverse team to work in the [DevOPS](https://www.atlassian.com/devops) methodology without blocking each other**: Each team or multiple teams 
  can develop and release their projects separately, later on the result of their work can be combined in a single project.
- **Supports versioning of projects in a CI/CD pipeline**: Each project can be versioned and referenced by static or dynamic tags (e.g., [SemVer2](https://semver.org/)), 
  which guarantees stable, well-tested and predictable releases.
- **Integrates with any CI/CD tool easily**: The tool is a self-sufficient binary that strictly follows the [12 factor app](https://12factor.net/) methodology.
- **Calls the CLI tools directly instead of using their libraries/SDKs**: RMK executes the tools directly in a way that a typical person would do it, 
  passing correct sets of CLI arguments and flags to the commands based on a project configuration structure.
  This decouples the updating of RMK itself from the wrapped CLI tools, allowing developers to utilize recent functionality and fixes.

### Edenlab LLC use cases

At [Edenlab LLC](https://edenlab.io/), RMK is used for deploying the [Kodjin FHIR Server](https://kodjin.com).

A classic Kodjin installation uses 3-level inheritance:
- **Dependencies (upstream#1)**: Core components like DBs, search engines, caches, load balancers/proxies, operators
  etc.
- **Kodjin (downstream#1)**: Kodjin FHIR API ([REST](https://en.wikipedia.org/wiki/REST))
- **Target installation (downstream#2)**: Products based on Kodjin, such as UI components, user portals and middleware services.

The additional components used by Kodjin are:
- **\*.provisioner.infra:** Repositories for Kubernetes cluster provisioning.
- **helmfile.hooks.infra:** Shell scrips used as [Helmfile hooks](https://helmfile.readthedocs.io/en/latest/#hooks) in
  deps/Kodjin/any other tenant.
- **core.charts.infra:** Helm charts used by the Kodjin services.

The examples of Kubernetes providers, to which Kodjin has been installed, are:
- [Amazon Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks/)
- [Open Telekom Cloud - Cloud Container Engine (CCE)](https://www.open-telekom-cloud.com/en/products-services/core-services/cloud-container-engine)
- [Rancher Kubernetes Platform](https://www.rancher.com/)
- [Kubermatic Kubernetes Platform (KKP)](https://www.kubermatic.com/)
- on-premise installations deployed using [Ansible Kubespray](https://github.com/kubernetes-sigs/kubespray)
- single-machine [K3D](https://k3d.io/) clusters

#### Related OSS repositories

- [AWS cluster provider for RMK](https://github.com/edenlabllc/aws.provisioner.infra)
- [Azure cluster provider for RMK](https://github.com/edenlabllc/azure.provisioner.infra)
- [K3D cluster provider for RMK](https://github.com/edenlabllc/k3d.provisioner.infra)
- [Helmfile hooks](https://github.com/edenlabllc/helmfile.hooks.infra)

## Requirements

Currently, RMK only supports Unix-like operating systems (OS):
* **OS:**
    * **MacOS**: amd64, arm64 (M1, M2 require [Rosetta](https://support.apple.com/en-us/HT211861))
    * **Linux**: amd64
* **Software:**
    * **Python** >= 3.9
    * **[AWS CLI](https://aws.amazon.com/cli/)**
    * _For managing local clusters using K3D:_ Version _v5.x.x_ requires [Docker](https://www.docker.com/) => v20.10.5 ([runc](https://github.com/opencontainers/runc) >= v1.0.0-rc93) to work
      properly.

> If this is your first project repository managed by RMK, ensure that the above tools are specified in the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file.

## Installation

To install RMK, run the self-installer script using the following command:

```shell
curl -sL "https://edenlabllc-rmk.s3.eu-north-1.amazonaws.com/rmk/s3-installer" | bash
```

Alternatively, you can go directly to https://github.com/edenlabllc/rmk/releases and download the binary.

As another option, the binary can be [built from source](development-and-release-flow.md#building-from-source).

## Update

### General update process

To update RMK to the latest version, run the following command:

```shell
rmk update
```

### Update to specific version

You can update to a specific RMK version to maintain backward compatibility or when updating to the latest version is not possible. 
> This may be necessary due to specific version requirements or when a bug has been detected. 

To update to a specific version, use the following command:

```shell
rmk update --version vX.X.X 
```

## Features

- **[Reduced and simplified management of Kubernetes projects:](#overview)** Deploy to Kubernetes using Helmfile/Helm, use popular DevOps tools together in a single CI/CD pipeline.
- **[Time-proven project structure:](configuration/project-management/preparation-of-project-repository.md)** Define the project structure using the [GitLabFlow](https://about.gitlab.com/topics/version-control/what-is-gitlab-flow/) methodology.
- **[Hierarchies between different projects:](configuration/project-management/dependencies-management-and-project-inheritance.md)** Define upstream-downstream relationships between sibling projects to reuse releases and services across different installations.
- **[Batch secret management:](configuration/secrets-management/secrets-management.md#generating-all-secrets-from-scratch-in-a-batch-manner-using-the-rmk-secrets-manager)** Template, generate, and encode project secrets for all environments in a batch manner.
- **[Clone environments with one click:](configuration/rmk-configuration-management.md#initialization-of-rmk-configuration-for-feature-or-release-clusters)** Use the special `--config-from-environment` (`--cfe`) flag to create an environment based on an existing one.
- **[Automatic detection of Multi-Factor Authentication](configuration/rmk-configuration-management.md#support-for-multi-factor-authentication-mfa) ([MFA](https://en.wikipedia.org/wiki/Multi-factor_authentication)):** Automatically detect and use an MFA device if one is defined by an [IAM](https://aws.amazon.com/iam/) user (must be supported by the cluster provider, e.g., [AWS](https://aws.amazon.com/)).
- **[Push-based release and downstream project updates:](configuration/release-management/release-management.md#release-update-and-integration-into-the-cd-pipeline)** Easily integrate with CI/CD solutions via webhooks or workflow dispatch events 
  to update release and service version declarations, automatically commit the changes to Git.
- **[Project structure generation:](configuration/project-management/preparation-of-project-repository.md#automatic-generation-of-the-project-structure-from-scratch)** Generate a complete Kubernetes-based project structure from scratch using RMK, following the best practices.
- **[Documentation generation:](commands.md#doc)** Generate the full command documentation in the Markdown format with one click.
- **[Support for different types of code sources:](configuration/rmk-configuration-management.md#use-upstream-artifact-for-the-downstream-projects-repository)** Use Git when the _artifact-mode_ is _none_, S3 when the _artifact-mode_ is _online_, 
  switch to fully offline installations when the _artifact-mode_ is _offline_.

## Supported Kubernetes providers

By design, RMK can work with any Kubernetes provider.

Among the providers are:

- [Amazon Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks/)
- [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)
- [Azure Kubernetes Service (AKS)](https://azure.microsoft.com/en-us/products/kubernetes-service/)
- [Red Hat OpenShift](https://redhat.com/en/technologies/cloud-computing/openshift)
- [VMware Tanzu Kubernetes Grid](https://tanzu.vmware.com/kubernetes-grid)
- [Rancher Kubernetes Platform](https://www.rancher.com/)
- [Open Telekom Cloud - Cloud Container Engine (CCE)](https://www.open-telekom-cloud.com/en/products-services/core-services/cloud-container-engine)
- [Kubermatic Kubernetes Platform (KKP)](https://www.kubermatic.com/)
- on-premise installations deployed using [Ansible Kubespray](https://github.com/kubernetes-sigs/kubespray)
- single-machine [K3D](https://k3d.io/) clusters

## Roadmap

- **Guidelines for contributors:** Create comprehensive guidelines for contributors, including instructions for creating pull requests (PRs).
- **Integration with Helmfile [vals](https://github.com/helmfile/vals)**: Integrate RMK with the **vals** tool for enhanced values and secret management.
- **Major update of the AWS [EKS](https://aws.amazon.com/eks/) cluster provider:** Update the AWS EKS cluster provider to the latest versions to utilize all the supported features of the [Terraform](https://www.terraform.io/) CLI and modules.
- **Implementation of additional RMK cluster providers:** Implement support for additional cluster providers for popular Kubernetes services such as [GKE](https://cloud.google.com/kubernetes-engine), [AKS](https://azure.microsoft.com/en-us/products/kubernetes-service/), etc.
- **Offline artifact mode:** Implement the **offline** artifact mode to install artifacts in fully isolated offline environments.
- **Web documentation generator:** Add an HTML documentation generator based on the **.md** files.
- **Automatic testing of RMK during the CI/CD pipeline:** Ensure that changes to the RMK codebase do not introduce errors or regressions during the CI/CD.

Check the [issues](https://github.com/edenlabllc/rmk/issues) for more information.

## License

RMK is open source software (OSS) licensed under the [Apache 2.0 License](../LICENSE).

## Code of Conduct

This project adheres to the Contributor Covenant [Сode of Сonduct](CODE_OF_CONDUCT.md). 
By participating, you are expected to uphold this code.
Please refer to our [Contributing Guidelines](CONTRIBUTING.md) for further information.
