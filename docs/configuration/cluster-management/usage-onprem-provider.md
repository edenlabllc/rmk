# Using On-Premise cluster provider

> On-Premise users must ensure that all target machines are accessible over the network and properly configured for 
> [K3S installation](https://docs.k3s.io/installation).

### Requirements
Before you can provision an On-Premise Kubernetes cluster, make sure the following prerequisites are satisfied:

1. Shared user account

   * Must exist on all nodes.
   * Requires sudo privileges without password prompt.
   * Used by RMK for provisioning.

2. SSH access

   * SSH connectivity from the administrator machine to all cluster nodes must be available.
   * Authentication is done via a private key.
   * An absolute path to the SSH private key must be specified during [rmk config init](../configuration-management/init-onprem-provider.md#configuration), 
     or RMK will attempt to use the default path (~/.ssh/id_rsa).

3. Networking

   * Firewall must allow full bidirectional connectivity between all cluster nodes.
   * [Required ports include](https://docs.k3s.io/installation/requirements#networking) (but are not limited to):
     * 6443 (Kubernetes API server)
     * 10250 (Kubelet API)
     * 8472/UDP (Flannel VXLAN overlay)
     * 51820/UDP, 51821/UDP (WireGuard, if enabled)
   * Ensure no firewall rules block node-to-node traffic.

4. K3S init server host 

   * An IP address must be allocated for the [init (bootstrap)](https://docs.k3s.io/datastore/ha-embedded) control plane node.
   * This IP is used by other nodes when joining the cluster.

5. OS requirements

   * [Only Linux nodes](https://docs.k3s.io/installation/requirements#operating-systems) with systemd are supported.
   * Tested distributions: RHEL 9, Ubuntu 22.04, Debian 12.

6. Disk and filesystem

   * Nodes must provide a dedicated disk or partition for container storage (optional but recommended).

7. DNS / Host resolution

   * If hostnames are used in configuration, they must resolve to correct internal IPs.
   * Alternatively, configure /etc/hosts on all nodes.

### Configuration

Before provisioning the Kubernetes cluster, add override for
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/onprem-cluster.yaml.gotmpl)
file to scope `deps` for the target Kubernetes cluster.

```yaml
# A complete list of all options can be found here https://github.com/edenlabllc/on-premise-configurator.operators.infra/blob/develop/watches.yaml
templates:
  machineSpec: &machineSpec
    bootstrap:
      # DataSecretName is the name of the secret that stores the bootstrap data script.
      # If nil, the Machine should remain in the Pending state.
      dataSecretName: ""

  machineRemoteServerSpec: &machineRemoteServerSpec
    # providerID will automatically be set by edenlabllc/on-premise-configurator.operators.infra
    k3sAirGapEnabled: false
    k3sServerConfigYAML: |
      node-taint:
        - node-role.kubernetes.io/control-plane=:NoSchedule
      tls-san:
        - 192.0.2.10
        - 192.0.2.11
        - 192.0.2.12
    k3sRole: server
    sshUser: user

  machineRemoteAgentSpec: &machineRemoteAgentSpec
    # providerID will automatically be set by edenlabllc/on-premise-configurator.operators.infra
    k3sAirGapEnabled: false
    k3sRole: agent
    sshUser: user

  controlPlaneHostInternal: &controlPlaneHostInternal "192.0.2.10" # e.g. 192.0.2.10

# ...
controlPlane:
  endpoint:
    host: *controlPlaneHostInternal
    port: 6443

  spec:
    k3sHAMode: true
    ## Kubernetes version
    version: v1.31.9+k3s1

  ## The control plane machines (k3sRole=server)
  machines:
    k3s-control-plane-0:
      enabled: true
      # ...
      remote:
        spec:
          address: *controlPlaneHostInternal
          k3sInitServer: true
          <<: *machineRemoteServerSpec
      spec:
        <<: *machineSpec

    k3s-control-plane-1:
      enabled: true
      # ...
      remote:
        spec:
          address: 192.0.2.11
          k3sApiEndpoint: *controlPlaneHostInternal
          k3sInitServer: false
          <<: *machineRemoteServerSpec
      spec:
        <<: *machineSpec

    k3s-control-plane-2:
      enabled: true
      # ...
      remote:
        spec:
          address: 192.0.2.12
          k3sApiEndpoint: *controlPlaneHostInternal
          k3sInitServer: false
          <<: *machineRemoteServerSpec
      spec:
        <<: *machineSpec

## The worker machines (k3sRole=agent)
machines: #{}
  load-balancer:
    enabled: true
    # ...
    remote:
      spec:
        address: 192.0.2.100
        k3sAgentConfigYAML: |
          node-label:
            - svccontroller.k3s.cattle.io/lbpool=pool1
            - svccontroller.k3s.cattle.io/enablelb=true
        k3sApiEndpoint: *controlPlaneHostInternal
        <<: *machineRemoteAgentSpec
    spec:
      <<: *machineSpec

  app-stateless:
    enabled: true
    # ...
    remote:
      spec:
        address: 192.0.2.101
        # agent config is optional
        # k3sAgentConfigYAML: |
        k3sApiEndpoint: *controlPlaneHostInternal
        <<: *machineRemoteAgentSpec
    spec:
      <<: *machineSpec

  app-stateful:
    enabled: true
    # ...
    remote:
      spec:
        address: 192.0.2.102
        k3sAgentConfigYAML: |
          node-label:
            - db=mydb
          node-taint:
            - key=mydb:NoSchedule
        k3sApiEndpoint: *controlPlaneHostInternal
        <<: *machineRemoteAgentSpec
    spec:
      <<: *machineSpec
```

Using the example above and the example from
the [cluster-deps repository](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/onprem-cluster.yaml.gotmpl)
you can add the required number of machines depending on the requirements for distribution into individual roles.

> For the On-Premise provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - Create secrets with SSH private key `capop-ssh-identity-secret` in the CAPI Management cluster.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the cluster is ready, RMK **automatically switches** the Kubernetes context to the newly created cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will **delete** the previously created SSH key and the context
> for the target Kubernetes cluster.
