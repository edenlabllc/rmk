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
  ssh-init-server-host: 10.1.1.10 # K3S init server host.
  ssh-private-key: /home/user/.ssh/id_rsa # Absolute path to SSH private key file. 
  ssh-user: user # SSH user.
# ...
```

## Prerequisites

> :warning: **Important**
>
> The On-Premise provider is fundamentally **different** from cloud providers like AWS, Azure, GCP.
>
> It requires **manual** preparation and configuration the underlying machines (bare-metal or VMs),
> **low-level** understanding of how Kubernetes and K3S behave in a **non-cloud** setup (control plane bootstrap,
> agents joining, resource planning, scheduling, etc.).
>
> Additionally, you must ensure correct **networking** rules (firewall, routing, node-to-node connectivity).

1. **OS**:

   * Only [Linux](https://docs.k3s.io/installation/requirements#operating-systems) machines with
     [systemd](https://systemd.io/) are supported.
   * Tested distributions:
     [RHEL 9](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9),
     [Ubuntu 22.04](https://releases.ubuntu.com/jammy/),
     [Debian 12](https://www.debian.org/releases/bookworm/).

2. **Disk and filesystem**:

   * Nodes should provide a **dedicated** disk or partition for container storage (optional but recommended).

3. **System user**:

   * Shared system user must exist on **all cluster nodes**.
   * Requires **sudo** privileges without password prompt.

4. **SSH access**:

   * SSH connectivity from the administrator machine to **all cluster nodes** must be available.
   * SSH authentication is done via a **private key**.
   * An **absolute path** to the private key must be specified
     during [configuration initialization](../configuration-management/init-onprem-provider.md#configuration),
     or RMK will attempt to use the default path (e.g., `~/.ssh/id_[ed25519|rsa|ecdsa|dsa]`).

5. **Networking**:

   * Firewall must allow full bidirectional connectivity between **all cluster nodes**.
   * Required [ports](https://docs.k3s.io/installation/requirements#networking) include (but are not limited to):
     * _6443/TCP_ ([Kubernetes API server](https://kubernetes.io/docs/concepts/overview/kubernetes-api/))
     * _10250/TCP_ ([Kubelet API](https://kubernetes.io/docs/concepts/architecture/#kubelet))
     * _8472/UDP_ ([Flannel](https://github.com/flannel-io/flannel) VXLAN overlay network)
     * _51820/UDP_, _51821/UDP_ ([WireGuard](https://www.wireguard.com/), if enabled)

6. **DNS resolution**:

   * If hostnames are used in the operator/K3S's configuration, they must resolve to correct internal IPs using
     [DNS](https://en.wikipedia.org/wiki/Domain_Name_System).
   * Alternatively, configure [/etc/hosts](https://en.wikipedia.org/wiki/Hosts_(file)) on **all cluster nodes**.

7. **K3S init server host**:

   * An IP address must be allocated for the [bootstrap](https://docs.k3s.io/datastore/ha-embedded) control
     plane node (used by other nodes when joining the cluster).

## Configuration

The 2 supported configuration scenarios are:

* **via RMK flags**:
  ```shell
  rmk config init --cluster-provider=onprem \
    --onprem-ssh-init-server-host=<k3s_init_server_ip> \
    --onprem-ssh-private-key=<ssh_private_key_path> \
    --onprem-ssh-user=<ssh_user>
  ```

* **via environment variables**: `RMK_ONPREM_SSH_INIT_SERVER_HOST`, `RMK_ONPREM_SSH_PRIVATE_KEY`, `RMK_ONPREM_SSH_USER`.
  ```shell
  export RMK_ONPREM_SSH_INIT_SERVER_HOST=<k3s_init_server_ip>
  export RMK_ONPREM_SSH_PRIVATE_KEY=<ssh_private_key_path>
  export RMK_ONPREM_SSH_USER=<ssh_user>
  rmk config init --cluster-provider=onprem
  ```  

If an environment variable or flag specifies a **custom** SSH private key, RMK will **copy** that key into the default
SSH location (using the **same** file name), and will use it for subsequent operations.

If CLI flags are provided, RMK will prioritize them over environment variables, as **CLI flags take precedence**.

## Reconfiguration of the On-Premise SSH private key

Modify the value of a specific flag if changes are needed:

```shell
rmk config init --onprem-ssh-private-key=<new_ssh_private_key_path>
```
