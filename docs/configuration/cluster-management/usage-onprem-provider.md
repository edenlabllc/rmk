# Using On-Premise cluster provider

> Ensure that all target machines are accessible over the network via SSH and properly configured for
> [K3S installation](https://docs.k3s.io/installation).

Before provisioning the Kubernetes cluster, add override for
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/onprem-cluster.yaml.gotmpl)
file to the `deps` scope for the target Kubernetes cluster.

```yaml
# A complete list of all options can be found here: https://github.com/edenlabllc/on-premise-configurator.operators.infra/blob/develop/watches.yaml
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
      remote:
        spec:
          address: *controlPlaneHostInternal
          k3sInitServer: true
          <<: *machineRemoteServerSpec
      spec:
        <<: *machineSpec

    k3s-control-plane-1:
      enabled: true
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
      remote:
        spec:
          address: 192.0.2.12
          k3sApiEndpoint: *controlPlaneHostInternal
          k3sInitServer: false
          <<: *machineRemoteServerSpec
      spec:
        <<: *machineSpec

## The worker machines (k3sRole=agent)
machines:
  load-balancer:
    enabled: true
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
the [cluster-deps](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/onprem-cluster.yaml.gotmpl)
repository you can add the required number of machines depending on the requirements for distribution into individual roles.

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
