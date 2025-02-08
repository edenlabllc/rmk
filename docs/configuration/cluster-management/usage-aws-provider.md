# Using AWS cluster provider

> AWS users must have the
> [PowerUserAccess](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/PowerUserAccess.html),
> [SecretsManagerReadWrite](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/SecretsManagerReadWrite.html)
> permissions to be able to provision and destroy [AWS EKS](https://aws.amazon.com/eks/) clusters.

Before provisioning the Kubernetes cluster, add override for
the [configuration](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/aws-cluster.yaml.gotmpl)
file to scope `deps` for the target Kubernetes cluster.

```yaml
# A complete list of all options can be found here https://capz.sigs.k8s.io/reference/v1beta1-api
controlPlane:
  spec:
    iamAuthenticatorConfig:
      # UserMappings is a list of user mappings
      mapUsers:
        # TODO: Add a list of users at the downstream project repository level
        - groups:
            - system:masters
          # UserARN is the AWS ARN for the user to map
          userarn: arn:aws:iam::{{ env "AWS_ACCOUNT_ID" }}:user/user1
          # UserName is a kubernetes RBAC user subject*/}}
          username: user1
    version: v1.29.8 # ^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.?(\.0|[1-9][0-9]*)?$

## The machine pools configurations
machinePools:
  app:
    enabled: true
    managed:
      spec:
        instanceType: t3.medium
        # Labels specifies labels for the Kubernetes node objects
        labels:
          db: app
        # Scaling specifies scaling for the ASG behind this pool
        scaling:
          maxSize: 1
          minSize: 1
    # Number of desired machines. Defaults to 1.
    replicas: 1
# ...
```

Using the example above and the example from
the [cluster-deps repository](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/aws-cluster.yaml.gotmpl)
you can add the required number of machine pools depending on the requirements for distribution into individual roles.

> For the AWS provider, before launching the actual provisioning of the cluster,
> RMK will perform the following preliminary steps:
>
> - create SSH key for cluster nodes.
> - create secrets with private [SOPS Age keys](../secrets-management/secrets-management.md#secret-keys) in the
    > [AWS Secret Manager](https://aws.amazon.com/secrets-manager/), if they have not been created previously.

To start provisioning a Kubernetes cluster, run the commands:

```shell
rmk cluster capi provision
```

> When the cluster is ready, RMK **automatically switches** the Kubernetes context to the newly created cluster.

To destroy a Kubernetes cluster, run the command:

```shell
rmk cluster capi destroy
```

> After the cluster is destroyed, RMK will delete the previously created SSH key and the context
> for the target Kubernetes cluster.
