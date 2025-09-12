# Initialization of On-Premise cluster provider

## List of main attributes of the RMK configuration

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the project (tenant) name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant name, which is equivalent to the project name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
cluster-provider: onprem # Selected cluster provider.
# ...
onprem:
  ssh-init-server-host: 10.1.1.10 #  K3S init server host.
  ssh-private-key: /Users/test-user/.ssh/id_rsa # Absolute path to SSH private key file. 
  ssh-user: test-user # SSH user.
# ...
```

## Prerequisites

1. A common user with passwordless sudo privileges must be available on all nodes where the Kubernetes cluster based on K3S will be deployed.

2. SSH access to all cluster nodes from the administrator’s machine must be configured using a private key.

3. Firewall rules must allow open network access between all cluster nodes.

4. An IP address must be available for the Kubernetes API server (K3S init server).

## Configuration

The absolute path to the SSH private key must be specified when running the rmk config init command. If it is not provided,
RMK will search in default SSH locations (e.g., ${HOME}/.ssh/id_[ed25519|rsa|ecdsa|dsa])

```shell
${HOME}/.ssh/id_[ed25519|rsa|ecdsa|dsa]
```

The 2 supported configuration scenarios are:

* **via RMK flags**:
  ```shell
  rmk config init --cluster-provider=onprem \
    --onprem-ssh-init-server-host=<k3s_init_server_ip> \
    --onprem-ssh-private-key=<ssh_private_key_path> \
    --onprem-ssh-user=<ssh_user_name>
  ```

* **via environment variables**: `RMK_ONPREM_SSH_INIT_SERVER_HOST`, `RMK_ONPREM_SSH_PRIVATE_KEY`, `RMK_ONPREM_SSH_USER`.
  ```shell
  export RMK_ONPREM_SSH_INIT_SERVER_HOST=<k3s_init_server_ip>
  export RMK_ONPREM_SSH_PRIVATE_KEY=<ssh_private_key_path>
  export RMK_ONPREM_SSH_USER=<ssh_user_name>
  rmk config init --cluster-provider=onprem
  ```  

If environment variables are set before running the command, RMK will place the SSH private key into the default 
SSH location using the same file name as specified in the flag or environment variable.  
If CLI flags are provided, RMK will prioritize them over environment variables, as **CLI flags take precedence**.

## Reconfiguration of the On-Premise SSH private key

Modify the value of a specific flag if changes are needed:

```shell
rmk config init --onprem-ssh-private-key=<ssh_private_key_path>
```
