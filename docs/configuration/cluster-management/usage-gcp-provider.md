# Using GCP cluster provider

> GCP service account must have the `Editor`, `Secret Manager Admin`, `Kubernetes Engine Admin`
> [roles](https://cloud.google.com/iam/docs/understanding-roles) to be able to provision and destroy
> [GCP GKE](https://cloud.google.com/kubernetes-engine?hl=en) clusters.

Before provisioning the Kubernetes cluster, add override
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/gcp-cluster.yaml.gotmpl)
file to scope `deps` for the target Kubernetes cluster.

```yaml
controlPlane:
  spec:
    version: "v1.30.5"

machinePools:
  app:
    enabled: true
    managed:
      spec:
        # MachineType is the name of a Google Compute Engine
        # (https://cloud.google.com/compute/docs/machine-types).
        # If unspecified, the default machine type is `e2-medium`.
        machineType: "e2-medium"
        management:
          # AutoUpgrade specifies whether node auto-upgrade is enabled for the node
          # pool. If enabled, node auto-upgrade helps keep the nodes in your node pool
          # up to date with the latest release version of Kubernetes.
          autoUpgrade: true
        # MaxPodsPerNode is constraint enforced on the max num of pods per node.
    replicas: 1
# ...
```

Using the example above and the example from
the [cluster-deps repository](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/gcp-cluster.yaml.gotmpl)
you can add the required number of machine pools depending on the requirements for distribution into individual roles.

> For the GCP provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - Create a [Cloud NAT](https://cloud.google.com/nat/docs/overview) for outbound traffic cluster nodes.
> - Create secrets with private [SOPS Age keys](../secrets-management/secrets-management.md#secret-keys) in the
>   [GCP Secret Manager](https://cloud.google.com/security/products/secret-manager?hl=en), 
>   if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the cluster is ready, RMK **automatically switches** the Kubernetes context to the newly created cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the previously created
> [Cloud NAT](https://cloud.google.com/nat/docs/overview) if this resource is no longer used by other clusters in the
> same region. Also deleting the context for the target Kubernetes cluster.
