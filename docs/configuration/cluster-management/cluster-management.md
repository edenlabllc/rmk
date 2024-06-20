# Cluster management

RMK uses [Terraform](https://www.terraform.io/) and [K3D](https://k3d.io) for cluster management.

RMK is suitable for both simple and complex Kubernetes deployments, enabling multi-level project inheritance through native Helmfile functionality.

The 2 scenarios are:
- **A cluster has already been provisioned via a 3rd-party tool/service:** An existing Kubernetes context will be used by RMK.
- **A cluster will be provisioned from scratch using RMK**: Any of the supported cluster providers for RMK, such as AWS, K3D, etc. will be utilized.

## Switch the context to an existing Kubernetes cluster

Switching to an existing Kubernetes cluster depends on how it has been provisioned:

* **Using a 3rd party tool:**
  
  Create a context with the name matching the pattern:
  
  ```
  \b<project_name>-<environment>\b
  ```
  
  > The matching is **case-insensitive**. \
  > `\b` means the **ASCII word boundary** (`\w` on one side and `\W`, `\A`, or `\z` on the other).
  
  For example, if you are in the `project1` repository in the `develop` branch, any of the following Kubernetes contexts will be accepted:
  
  ```
  project1-develop
  Project1-Develop
  PROJECT1-DEVELOP
  project1-develop-cluster
  Project1-Develop-Cluster
  PROJECT1-DEVELOP-CLUSTER
  k3d-project1-develop
  arn:aws:eks:us-east-1:123456789000:cluster/PROJECT1-DEVELOP-CLUSTER
  ```
  
  > If there are **more than one** Kubernetes context which match the regular expression **simultaneously**, 
  > an **error** will be thrown indicating a conflict. For example, the following names will conflict:
  > ```shell
  > project1-develop
  > k3d-project1-develop
  > ```

* **Using RMK cluster provider**:

  Checkout to the branch from which the K8S cluster was previously created. 

  An [initialization](../rmk-configuration-management.md#initialization-of-rmk-configuration) might be required,
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
- [aws.provisioner.infra](https://github.com/edenlabllc/aws.provisioner.infra): Configuration for managing AWS EKS
  clusters using Terraform. Kubernetes clusters can be provisioned from scratch and destroyed 
  via the `rmk cluster provision`, `rmk cluster destroy` commands.
- [k3d.provisioner.infra](https://github.com/edenlabllc/k3d.provisioner.infra): Configuration for managing
  single-machine clusters using K3D (suitable for both local development and minimal cloud deployments). 
  Kubernetes clusters can be created from scratch and deleted via the `rmk cluster k3d create`, `rmk cluster k3d delete` commands.

Support for other cloud providers such as GCP, Azure will be implemented in the future.
This enhancement will include the introduction of new RMK commands and cluster providers, as well as the addition of _*.provisioner.infra_ repositories.

### Provision or destroy AWS EKS Kubernetes clusters

> AWS users must have the `AdministratorAccess` permissions to be able to provision and destroy EKS clusters.

Before provisioning the K8S cluster, modify the core configurations for the on-demand cluster. 
The core configurations are divided into two types:

- **variables** (common AWS cluster management): \
  _Path:_ `etc/clusters/aws/<environment>/values/variables.auto.tfvars` \
  _Frequently changed values:_
  ```terraform
  # k8s user list
  k8s_master_usernames = [] # list of AWS IAM users for K8S cluster management
  k8s_cluster_version  = "1.27" # current version K8S(EKS) control plane
  # ...
  ```

- **worker-groups** (resources for AWS worker nodes): \
  _Path:_ `etc/clusters/aws/<environment>/values/worker-groups.auto.tfvars` \
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
  > ```shell
  > AWS_PROFILE=$(rmk --lf=json config view | jq '.config.Profile' -r) \
  >   aws ssm get-parameter \
  >   --name /aws/service/eks/optimized-ami/<EKS_control_plane_version>/amazon-linux-2/recommended/image_id \
  >   --region <aws_region> \
  >   --query "Parameter.Value" \
  >   --output text
  > ```

Full list of input Terraform variables: `.PROJECT/inventory/clusters/aws.provisioner.infra-<version>/terraform/variables.tf`

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
`.PROJECT/inventory/clusters/k3d.provisioner.infra-<version>/k3d.yaml`.

> Prerequisites:
> 1. Create a separate feature branch: `feature/<issue_key>-<issue_number>-<issue_description>`.
> 2. [Initialize configuration](../rmk-configuration-management.md#initialization-of-rmk-configuration) for this branch with the `localhost` root domain name:
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
