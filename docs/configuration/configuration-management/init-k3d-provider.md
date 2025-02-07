# Initialization of K3D cluster provider

## List of main attributes of the RMK configuration

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the project (tenant) name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant name, which is equivalent to the project name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
cluster-provider: k3d # Selected cluster provider.
# ...
```

### Configuring K3D

K3D is the default cluster provider in RMK.

To initialize RMK configuration for a K3D cluster, run:

```shell
rmk config init
```

> K3D is intended for provisioning **local** clusters, primarily for **development** environments.

