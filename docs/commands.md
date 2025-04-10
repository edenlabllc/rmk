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

**BuiltBy:** goreleaser <br />
**Commit:** 23c4f52 <br />
**Date:** 2025-04-10T07:32:35Z <br />
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

#### capi, c

CAPI cluster management

##### create, c

Create CAPI management cluster

##### delete, d

Delete CAPI management cluster

##### destroy

Destroy K8S target (workload) cluster

##### list, l

List CAPI management clusters

##### provision, p

Provision K8S target (workload) cluster

##### update, u

Update CAPI management cluster

#### k3d, k

K3D cluster management

##### create, c

Create K3D cluster

**--k3d-volume-host-path, --kv**="": host local directory path for mount into K3D cluster (default: present working directory)

##### delete, d

Delete K3D cluster

##### import, i

Import images from docker to K3D cluster

**--k3d-import-image, --ki**="": list of images to import into running K3D cluster

##### list, l

List K3D clusters

##### start, s

Start K3D cluster

##### stop

Stop K3D cluster

#### switch, s

Switch Kubernetes context to project cluster

**--force, -f**: force update Kubernetes context from remote cluster

### completion

Completion management

#### zsh, z

View Zsh completion scripts

### config

Configuration management

#### init, i

Initialize configuration for current project and selected environment

**--aws-access-key-id, --awid**="": AWS access key ID for IAM user

**--aws-region, --awr**="": AWS region for current AWS account

**--aws-secret-access-key, --awsk**="": AWS secret access key for IAM user

**--aws-session-token, --awst**="": AWS session token for IAM user

**--azure-client-id, --azid**="": Azure client ID for Service Principal

**--azure-client-secret, --azp**="": Azure client secret for Service Principal

**--azure-key-vault-resource-group-name, --azkvrg**="": Azure Key Vault custom resource group name

**--azure-location, --azl**="": Azure location

**--azure-service-principle, --azsp**: Azure service principal STDIN content

**--azure-subscription-id, --azs**="": Azure subscription ID for current platform domain

**--azure-tenant-id, --azt**="": Azure tenant ID for Service Principal

**--cluster-provider, --cp**="": cluster provider for provisioning (default: "k3d")

**--gcp-region, --gr**="": GCP region

**--github-token, --ght**="": GitHub personal access token, required when using private repositories

**--google-application-credentials, --gac**="": path to GCP service account credentials JSON file

**--progress-bar, -p**: globally disable or enable progress bar for download process

**--slack-channel, --sc**="": channel name for Slack notifications

**--slack-message-details, --smd**="": additional details for Slack message body

**--slack-notifications, -n**: enable Slack notifications

**--slack-webhook, --sw**="": URL for Slack webhook

#### delete, d

Delete configuration for selected environment

#### list, l

List available configurations for current project

**--all, -a**: list all project configurations

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

**--environment, -e**="": list of project environments, root-domain config option must be provided: <environment>.root-domain=<domain-name>

**--owner, -o**="": list of project owners

**--scope, -s**="": list of project scopes

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

**--selector, -l**="": list of release labels, used as selector, selector can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### destroy, d

Destroy releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--selector, -l**="": list of release labels, used as selector, selector can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### list, l

List releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--output, -o**="": output format, available: short, yaml (default: "short")

**--selector, -l**="": list of release labels, used as selector, selector can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### rollback, r

Rollback specific releases to latest stable state

**--release-name, --rn**="": list release names for rollback status in Kubernetes

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### sync, s

Sync releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--selector, -l**="": list of release labels, used as selector, selector can take form of foo=bar or foo!=bar

**--skip-context-switch, -s**: skip context switch for not provisioned cluster

#### template, t

Template releases

**--helmfile-args, --ha**="": Helmfile additional arguments

**--helmfile-log-level, --hll**="": Helmfile log level severity, available: debug, info, warn, error (default: "error")

**--selector, -l**="": list of release labels, used as selector, selector can take form of foo=bar or foo!=bar

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

**--environment, -e**="": list of secret environments, used as selector

**--scope, -s**="": list of secret scopes, used as selector

##### encrypt, e

Encrypt secrets batch for selected scope and environment

**--environment, -e**="": list of secret environments, used as selector

**--scope, -s**="": list of secret scopes, used as selector

##### generate, g

Generate secrets batch for selected scope and environment

**--environment, -e**="": list of secret environments, used as selector

**--force, -f**: force overwriting current secrets after generating new

**--scope, -s**="": list of secret scopes, used as selector

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

**--version, -v**="": RMK special version (default: empty value corresponds latest version)

### help, h

Shows a list of commands or help for one command

