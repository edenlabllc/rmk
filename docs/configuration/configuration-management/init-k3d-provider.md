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

### Configuration of K3D

K3D as a cluster provider promote in RMK by default. For initialization RMK configuration for K3D cluster provider run 
the following command:

```shell
rmk config init
```

> K3D can be used to provision **local** clusters for **develop** environments.
