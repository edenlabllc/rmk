# RMK CLI - Reduced Management for Kubernetes

[![Release](https://img.shields.io/github/v/release/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/releases/latest)
[![Software License](https://img.shields.io/github/license/edenlabllc/rmk.svg?style=for-the-badge)](https://github.com/edenlabllc/rmk/blob/master/LICENSE)
[![Powered By: Edenlab](https://img.shields.io/badge/powered%20by-edenlab-8A2BE2.svg?style=for-the-badge)](https://edenlab.io)

Command line tool for reduced management and provisioning of Kubernetes clusters and environments, Helm secrets and releases.

Full documentation is available at https://edenlabllc.github.io/rmk/latest/

## Overview

This tool has been designed and developed initially by [Edenlab LLC](https://edenlab.io/) as the main CLI
for managing [Kodjin FHIR Server](https://kodjin.com) on Kubernetes clusters in different environments.

It is a wrapper around many popular CI/CD and DevOps CLI tools, including:

- [Helmfile](https://helmfile.readthedocs.io/en/latest/)
- [Helm](https://helm.sh/)
- [kubectl](https://kubernetes.io/docs/reference/kubectl/)
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

- **\*.provisioner.infra:** RMK cluster provider repositories for Kubernetes cluster provisioning.
- **helmfile.hooks.infra:** Shell scrips used as [Helmfile hooks](https://helmfile.readthedocs.io/en/latest/#hooks) in
  deps/Kodjin/any other tenant.
- **core.charts.infra:** Helm charts used by the Kodjin services.

The examples of Kubernetes providers, to which Kodjin has been installed already, are:

- [Amazon Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks/)
- [Azure Kubernetes Service (AKS)](https://azure.microsoft.com/en-us/products/kubernetes-service/)
- [Open Telekom Cloud - Cloud Container Engine (CCE)](https://www.open-telekom-cloud.com/en/products-services/core-services/cloud-container-engine)
- [Rancher Kubernetes Platform](https://www.rancher.com/)
- [Kubermatic Kubernetes Platform (KKP)](https://www.kubermatic.com/)
- on-premise installations deployed using [Ansible Kubespray](https://github.com/kubernetes-sigs/kubespray)
- single-machine [K3D](https://k3d.io/) clusters

### Related OSS repositories

- [AWS cluster provider for RMK](https://github.com/edenlabllc/aws.provisioner.infra)
- [Azure cluster provider for RMK](https://github.com/edenlabllc/azure.provisioner.infra)
- [K3D cluster provider for RMK](https://github.com/edenlabllc/k3d.provisioner.infra)
- [Helmfile hooks](https://github.com/edenlabllc/helmfile.hooks.infra)

## Development and release

The guidelines are available at https://edenlabllc.github.io/rmk/latest/development-and-release/

## License

RMK is open source software (OSS) licensed under the [Apache 2.0 License](https://github.com/edenlabllc/rmk/blob/master/LICENSE).

## Code of Conduct

This project adheres to the Contributor Covenant [Сode of Сonduct](https://github.com/edenlabllc/rmk/blob/master/docs/CODE_OF_CONDUCT.md). 
By participating, you are expected to uphold this code.
Please refer to our [Contributing Guidelines](https://github.com/edenlabllc/rmk/blob/master/docs/CONTRIBUTING.md) for further information.
