# Cluster management

RMK uses [Cluster API](https://cluster-api.sigs.k8s.io/introduction) and [K3D](https://k3d.io) for cluster management.

RMK is suitable for both simple and complex Kubernetes deployments, enabling multi-level project inheritance through native Helmfile functionality.

The 2 scenarios are:

- **A cluster has already been provisioned via 3rd-party tools/services:** An existing Kubernetes context will be used by RMK.
- **A cluster will be provisioned from scratch using RMK**: Any of the supported cluster providers for RMK, such as AWS, K3D, etc. will be utilized.

## Switch the context to an existing Kubernetes cluster

Switching to an existing Kubernetes cluster depends on how it has been provisioned:

* **Using a 3rd party tool:**
  
  Create a context with the name strictly matching the following:
  
  ```
  <project_name>-<environment>
  ```
  
  For example, if you are in the `project1` repository in the `develop` branch, any of the following Kubernetes context will be accepted:
  
  ```
  project1-develop
  ```
  
* **Using RMK cluster provider**:

  Checkout to the branch from which the K8S cluster was previously created. 

  An [initialization](../configuration-management.md#initialization-of-rmk-configuration) might be required,
  if the RMK configuration for this cluster has not been created before:
  
  ```shell
  rmk config init
  ```
   
  The next command depends on whether a remote cluster provider (e.g., AWS) or a local one (e.g., K3D) has been used:

  * **AWS:**

    ```shell
    # --force might required to refresh the credentials after a long period of inactivity
    rmk cluster switch --force
    ```
  
  * **K3D:**

    Explicit switching to the Kubernetes context is not required, if a K3D cluster has been created already. 
    RMK will switch implicitly, when running any of the `rmk release` commands.
  
Finally, run an RMK release command to verify the preparation of the Kubernetes context, e.g.:

```shell
rmk release list
```

## Use RMK cluster providers to provision and destroy Kubernetes clusters

Currently, the following cluster providers are supported by RMK:

- AWS EKS: Configuration for managing AWS EKS
  clusters using Cluster API. Kubernetes clusters can be provisioned from scratch and destroyed 
  via the `rmk cluster provision`, `rmk cluster destroy` commands.
- Azure: ...
- GCP: ...
- K3D: Configuration for managing
  single-machine clusters using K3D (suitable for both local development and minimal cloud deployments). 
  Kubernetes clusters can be created from scratch and deleted via the `rmk cluster k3d create`, `rmk cluster k3d delete` commands.

Support for other cloud providers and on-premise will be implemented in the future.
This enhancement might include the introduction of new RMK commands, Cluster API providers, K8S operators.
The main infrastructure configuration can always be checked in the [cluster-deps](https://github.com/edenlabllc/cluster-deps.bootstrap.infra) repository. 

### Provision or destroy AWS EKS Kubernetes clusters

> AWS users must have the `AdministratorAccess` permissions to be able to provision and destroy EKS clusters.

Before provisioning the K8S cluster, modify the core configurations for the on-demand cluster. 
The core configurations are divided into two types:

- **variables** (common AWS cluster management):

  _Path:_ `...`  

[//]: # (  TODO ACTUALIZE)

  _Frequently changed values:_

  ```terraform
  # k8s user list
  k8s_master_usernames = [] # list of AWS IAM users for K8S cluster management
  k8s_cluster_version  = "1.27" # current version of K8S (EKS) control plane
  # ...
  ```

  > Full list of input variables: `...`
  [//]: # (  TODO ACTUALIZE)

- **worker-groups** (resources for AWS worker nodes):

  _Path:_ `...`

  [//]: # (  TODO ACTUALIZE)

  _Frequently changed values:_

  ```terraform
  worker_groups = [
    {
      instance_type        = "t3.xlarge"
      additional_userdata  = "t3.xlarge"
      asg_desired_capacity = 1
      asg_max_size         = 1
      asg_min_size         = 1
      ami_id               = "ami-0dd8af8522cf16846"
    },
    # ...
  ]
  ```
  
  - `instance_type`: [AWS EC2 instance type](https://aws.amazon.com/ec2/instance-types)
  - `asg_desired_capacity`: Number of nodes of a specific group.
  - `ami_id`: Identifier of AWS AMI image for EKS.
    > Each AWS region requires its own AMI image ID. To determine the appropriate ID for a specific region, run the following command:
    > 
    > ```shell
    > AWS_PROFILE=$(rmk --lf=json config view | yq '.config.Profile') \
    > AWS_CONFIG_FILE="${HOME}/.aws/config_$(rmk --lf=json config view | yq '.config.Profile')" \
    > AWS_SHARED_CREDENTIALS_FILE="${HOME}/.aws/credentials_$(rmk --lf=json config view | yq '.config.Profile')" \
    > AWS_PAGER= \
    > aws ssm get-parameter \
    >   --name /aws/service/eks/optimized-ami/<eks_control_plane_version>/amazon-linux-2/recommended/image_id \
    >   --region "$(rmk --lf=json config view | yq '.config.Region')" \
    >   --output text \
    >   --query Parameter.Value
    > ```
    > > Replace `<eks_control_plane_version>` with a correct version.

To start provisioning a Kubernetes cluster, run the commands:

```shell
# prepare only plan
rmk cluster provision --plan
# prepare plan and launch it
rmk cluster provision
```

> When the K8S cluster is ready, RMK automatically switches the kubectl context to the newly created K8S cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster destroy
```

### Create or delete K3D Kubernetes clusters

RMK supports managing single-node Kubernetes clusters using [K3D](https://k3d.io).

The CLI will create a cluster according to the declarative instruction for K3D: 
`...`.

[//]: # (  TODO ACTUALIZE)

> Prerequisites:
> 
> 1. Create a separate feature branch: `feature/<issue_key>-<issue_number>-<issue_description>`.
> 2. [Initialize configuration](../configuration-management.md#initialization-of-rmk-configuration) for this branch with the `localhost` root domain name:
> 
> ```shell
> rmk config init --root-domain=localhost
> ```

#### Create K3D clusters

> By default, RMK will use `volume-host-path` as the current directory:

Run the following command:

```shell
rmk cluster k3d create
```

> When the Kubernetes cluster is ready, RMK automatically switches the kubectl context to the newly created Kubernetes cluster.

#### Delete K3D clusters

```shell
rmk cluster k3d delete
```
