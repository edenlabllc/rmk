# NAME

RMK CLI - Reduced management for Kubernetes

## SYNOPSIS

rmk

```
[--help|-h]
[--log-format|--lf]=[value]
[--log-level|--ll]=[value]
[--version|-v]
```

## DESCRIPTION

Command line tool for reduced management of the provision of Kubernetes clusters in different environments and management of service releases.

**BuiltBy:** goreleaser \
**Commit:** 25e5886 \
**Date:** 2024-09-03T15:03:12Z \
**Target:** linux_amd64

**Usage**:

```
rmk [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

## GLOBAL OPTIONS

**--help, -h**: show help

**--log-format, --lf**="": log output format, available: console, json (default: "console")

**--log-level, --ll**="": log level severity, available: debug, info, error (default: "info")

**--version, -v**: print the version


## COMMANDS

### cluster

Cluster management

#### container-registry, c

Container registry management

##### login

Log in to container registry

**--get-token, -g**: get ECR token for authentication

##### logout

Log out from container registry

#### destroy, d

Destroy AWS cluster using Terraform

#### list, l

List all Terraform available workspaces

#### k3d, k

K3D cluster management

##### create, c

Create K3D cluster

**--k3d-volume-host-path, --kv**="": host local directory path for mount into K3D cluster (default: present working directory)

##### delete, d

Delete K3D cluster

##### import, i

Import images from docker to K3D cluster

**--k3d-import-image, --ki**="": list images for import into K3D cluster

##### list, l

List K3D clusters

##### start, s

Start K3D cluster

##### stop

Stop K3D cluster

#### provision, p

Provision AWS cluster using Terraform

**--plan, -p**: creates an execution Terraform plan

#### state, t

State cluster management using Terraform

##### delete, d

Delete resource from Terraform state

**--resource-address, --ra**="": resource address for delete from Terraform state

##### list, l

List resources from Terraform state

##### refresh, r

Update state file for AWS cluster using Terraform

#### switch, s

Switch Kubernetes context for tenant cluster

**--force, -f**: force update Kubernetes context from remote cluster

### completion

Completion management

#### zsh, z

View Zsh completion scripts

### config

Configuration management

#### init, i

Initialize configuration for current tenant and selected environment

**--artifact-mode, --am**="": choice of artifact usage model, available: none, online (default: "none")

**--aws-ecr-host, --aeh**="": AWS ECR host (default: "288509344804.dkr.ecr.eu-north-1.amazonaws.com")

**--aws-ecr-region, --aer**="": AWS region for specific ECR host (default: "eu-north-1")

**--aws-ecr-user-name, --aeun**="": AWS ECR user name (default: "AWS")

**--aws-reconfigure, -r**: force AWS profile creation

**--aws-reconfigure-artifact-license, -l**: force AWS profile creation for artifact license, used only if RMK config option artifact-mode has values: online, offline

**--cloudflare-token, --cft**="": Cloudflare API token for provision NS records

**--cluster-provider, --cp**="": select cluster provider to provision clusters (default: "aws")

**--cluster-provisioner-state-locking, -c**: disable or enable cluster provisioner state locking

**--config-from-environment, --cfe**="": inheritance of RMK config credentials from environments: develop, staging, production

**--github-token, --ght**="": personal access token for download GitHub artifacts

**--progress-bar, -p**: globally disable or enable progress bar for download process

**--root-domain, --rd**="": domain name for external access to app services via ingress controller

**--s3-charts-repo-region, --scrr**="": location constraint region of S3 charts repo (default: "eu-north-1")

**--slack-channel, --sc**="": channel name for Slack notification

**--slack-message-details, --smd**="": additional information for body of Slack message

**--slack-notifications, -n**: enable Slack notifications

**--slack-webhook, --sw**="": URL for Slack webhook

#### delete, d

Delete configuration for selected environment

#### list, l

List available configurations for current tenant

**--all, -a**: list all tenant configurations

#### view, v

View configuration for selected environment

### doc

Documentation management

**--help, -h**: show help

#### generate, g

Generate documentation by commands and flags in Markdown format

**--help, -h**: show help

##### help, h

Shows a list of commands or help for one command

#### help, h

Shows a list of commands or help for one command

### project

Project management

#### generate, g

Generate project directories and files structure

**--create-sops-age-keys, -c**: create SOPS age keys for generated project structure

#### update, u

Update project file with specific dependencies version

**--dependency, -d**="": specific dependency name for updating project file

**--skip-ci, -i**: add [skip ci] to commit message line to skip triggering other CI builds

**--skip-commit, -c**: only change a version in for project file without committing and pushing it

**--version, -v**="": specific dependency version for updating project file

### release

Release components list from state file (Helmfile)

#### build, b

Build releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--selector, -l**="": only run using releases that match labels. Labels can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### destroy, d

Destroy releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--output, -o**="": output format, available: short, yaml (default: "short")

**--selector, -l**="": only run using releases that match labels. Labels can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### list, l

List releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--output, -o**="": output format, available: short, yaml (default: "short")

**--selector, -l**="": only run using releases that match labels. Labels can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### rollback, r

Rollback specific releases to latest stable state

**--release-name, --rn**="": list release names for rollback status in Kubernetes

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### sync, s

Sync releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--selector, -l**="": only run using releases that match labels. Labels can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### template, t

Template releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--selector, -l**="": only run using releases that match labels. Labels can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### update, u

Update releases file with specific environment values

**--commit, -c**: only commit and push changes for releases file

**--deploy, -d**: deploy updated releases after committed and pushed changes

**--repository, -r**="": specific repository for updating releases file

**--skip-ci, -i**: add [skip ci] to commit message line to skip triggering other CI builds

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

**--tag, -t**="": specific tag for updating releases file

### secret

secrets management

#### manager, m

batch secrets management

##### decrypt, d

Decrypt secrets batch for selected scope and environment

**--environment, -e**="": specific environments for selected secrets

**--scope, -s**="": specific scopes for selected secrets

##### encrypt, e

Encrypt secrets batch for selected scope and environment

**--environment, -e**="": specific environments for selected secrets

**--scope, -s**="": specific scopes for selected secrets

##### generate, g

Generate secrets batch for selected scope and environment

**--environment, -e**="": specific environments for selected secrets

**--force, -f**: force overwriting current secrets after generating new

**--scope, -s**="": specific scopes for selected secrets

#### keys, k

SOPS age keys management

##### create, c

Create SOPS age keys

##### download, d

Download SOPS age keys from S3 bucket

##### upload, u

Upload SOPS age keys to S3 bucket

#### encrypt, e

Encrypt secret file

#### decrypt, d

Decrypt secret file

#### view, v

View secret file

#### edit

Edit secret file

### update

Update RMK CLI to a new version

**--release-candidate, -r**: force update RMK to latest release candidate version

**--version, -v**="": RMK special version. (default: empty value corresponds latest version)

### help, h

Shows a list of commands or help for one command

