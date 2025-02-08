# Using Azure cluster provider

> Azure service principal must have the
> [Contributor](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/privileged#contributor),
> [Key Vault Secrets Officer](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/security#key-vault-secrets-officer)
> roles to be able to provision and destroy [Azure AKS](https://azure.microsoft.com/en-us/products/kubernetes-service)
> clusters.

Before provisioning the Kubernetes cluster, add override for the
[configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/azure-cluster.yaml.gotmpl)
file to scope `deps` for the target Kubernetes cluster.

```yaml
controlPlane:
  spec:
    ## Kubernetes version
    version: v1.29.8

machinePools:
  system:
    enabled: true

  app:
    enabled: true
    replicas: 1
    spec:
      mode: User
      sku: Standard_B2ls_v2
      osDiskSizeGB: 30
      nodeLabels:
        db: app
      scaling:
        minSize: 1
        maxSize: 1
# ...
```

Using the example above and the example from
the [cluster-deps repository](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/azure-cluster.yaml.gotmpl)
you can add the required number of machine pools depending on the requirements for distribution into individual roles.

> For the Azure provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - Create secrets with private [SOPS Age keys](../secrets-management/secrets-management.md#secret-keys) in the
>   [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault), if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the cluster is ready, RMK **automatically switches** the Kubernetes context to the newly created cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the context for the target Kubernetes cluster.
