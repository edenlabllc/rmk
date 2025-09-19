# Using K3D cluster provider

RMK supports managing single-node Kubernetes clusters using [K3D](https://k3d.io).

The CLI will create a cluster according to the
declarative [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/k3d-cluster.yaml.gotmpl)
for K3D.

#### Creating K3D clusters

Run the following command:

```shell
rmk cluster k3d create
```

> By default, RMK will use the current directory for the [--k3d-volume-host-path](../../commands.md#create-c-1) flag.

When the Kubernetes cluster is ready, RMK **automatically switches** the Kubernetes context to the newly created
cluster. You can create multiple local K3D clusters by **separating** them using Git branches.

#### Deleting K3D clusters

Run the following command:

```shell
rmk cluster k3d delete
```
