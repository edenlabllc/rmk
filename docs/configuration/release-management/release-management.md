# Release management

RMK uses [Helmfile](https://github.com/helmfile/helmfile) for the release management.

RMK uses a reduced set of the `Helmfile` commands without changing their behavior.
The full list of commands can be found in the [release category](../../commands.md#release).
Additionally, flags are provided for the commands, which allow extending capabilities and help during the command
execution debug.

For example:

```shell
rmk release build
rmk release list
rmk release template --skip-context-switch --selector app=myapp1
rmk release sync --helmfile-log-level=debug --selector app=myapp1 
rmk release destroy 
```

> The `--skip-context-switch` (`-s`) flag can be used with commands
> like [rmk release template](http://localhost:8000/rmk/commands/#template-t) to **skip switching** to a
> Kubernetes cluster.
>
> This is useful in situations where a **cluster has not been provisioned** yet, releases are being
> developed, but the user still wants to **preview intermediate results**, such as checking the templating of a release.

In a project repository, all the release values files are stored in the `etc/<scope>/<env>/values/` directories.
For example:

```
etc/deps/develop/values/
etc/rmk-test/staging/values/
```

> The release values are **inherited by the projects**, e.g., the upstream project's values ("deps") are included into
> the downstream project's values ("rmk-test").

All `releases.yaml` files controlling which releases are enabled/disabled are stored in the `etc/<scope>/<env>/`
directories.
For example:

```
etc/deps/develop/releases.yaml
```

> Similar to the [SOPS secrets files](../secrets-management/secrets-management.md#secret-files), the `releases.yaml`
> files are **not inherited by the projects** in contrast to the release values. Each project **should have its own**
> `releases.yaml` files for all deployed scopes and environments.

The release installation order is declared in `helmfile.yaml.gotmpl` file. For an example, see the `cluster-deps`
[Helmfile](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/helmfile.yaml.gotmpl).

Running any of the commands in the `release` category will trigger the Helmfile's dependency resolution mechanism
("[DAG](https://helmfile.readthedocs.io/en/latest/#dag-aware-installationdeletion-ordering-with-needs)").
Additionally, RMK **verifies** that the current Kubernetes context matches the Git branch and environment,
**preventing** releases from being synchronized to an **unintended** Kubernetes cluster (RMK will select a correct
automatically).

Among the user-defined `Helmfile` selectors, the following
[labels](https://helmfile.readthedocs.io/en/stable/#labels-overview) are available **by default**:

- `name`: Release name.
- `namespace`: Release namespace.
- `chart`: Chart name.

For example:

```shell
rmk release list --selector name=myapp1
rmk release sync --selector namespace=kube-system
rmk release destroy --selector chart=mychart1
```

## Examples of usage

### List of all available releases

```shell
rmk release list
```

### Viewing a specific release YAML after the Helm values template rendering

```shell
rmk release template --selector app=myapp1
```

### Synchronization of all releases

```shell
rmk release sync
```

### Synchronization of a specific scope of the releases

```shell
rmk release sync --selector scope=deps
```

### Synchronization of a specific release with passing the "--set" Helmfile argument

```shell
rmk release sync --selector app=myapp1 --helmfile-args="--set='values.key1=value1'"
```

### Destroy all releases

```shell
rmk release destroy
```

## Overriding release values for inherited upstream projects

It is possible to override any release value for
the [inherited upstream project repository](../project-management/dependencies-management-and-project-inheritance.md#dependencies-management-and-project-inheritance).
You can **override any element** separately in its YAML file.

For example, you have the following file in the `cluster-deps` upstream project,
e.g. [aws-cluster.yaml.gotmpl](https://github.com/edenlabllc/cluster-deps.bootstrap.infra/blob/develop/etc/deps/develop/values/aws-cluster.yaml.gotmpl).

```yaml
# ...

## The machine pools configurations
machinePools:
  app:
    enabled: false
    annotations: { }
    labels: { }
    managed:
      spec:
        # AdditionalTags is an optional set of tags to add to AWS resources managed by the AWS provider, in addition 
        # to the ones added by default.
        additionalTags: { }
        # DiskSize specifies the root disk size
        diskSize: 10
        # InstanceType specifies the AWS instance type
        instanceType: t3.medium
        # Labels specifies labels for the Kubernetes node objects
        labels:
          db: app
        # Scaling specifies scaling for the ASG behind this pool
        scaling:
          maxSize: 1
          minSize: 1
        # Taints specifies the taints to apply to the nodes of the machine pool
        taints:
          - effect: NoSchedule
            key: old-taint
            value: value1
        # ...
    # Number of desired machines.
    replicas: 1
    spec:
    # ...

# ...
```

Then, if you need to change the `machinePools.app.*` values in the `develop` environment of the `rmk-test` downstream
project, e.g., to enable the `app` nodes and increase their count to 10, you **don’t need to copy the entire file**.
Instead, you can **override only the specific values** by repeating their YAML path.

Your override file will be minimalistic and contain only the necessary changes:

```yaml
machinePools:
  app:
    enabled: true
    managed:
      spec:
        scaling:
          maxSize: 10
          minSize: 10
    replicas: 10
```

> If you want to **override** the `taints` field, which is an **array**, the correct way to do this is to
> **provide the whole array**:
>
> ```yaml
> taints:
>   # old taint
>   - effect: NoSchedule
>     key: old-taint
>     value: value1
>   # new taint
>   - effect: PreferNoSchedule
>     key: new-taint
>     value: value2
> ```
>
> You **cannot override a subset** of the array items:
> ```yaml
> taints:
>   # incorrect
>   - effect: NoSchedule
>     key: new-taint
>     value: value2
> ```

To check the final result, run the [rmk release template](http://localhost:8000/rmk/commands/#template-t) command and
see the final YAML.

## Release update and integration into the CD pipeline

The [rmk release update](http://localhost:8000/rmk/commands/#update-u_2) command automates the process of **updating and
delivering releases** according to the version changes of artifacts, e.g., container images, following
the [GitOps](https://www.gitops.tech) methodology.

Since RMK is a binary file that can be **downloaded and installed** on any Unix-based operating system,
it can be **integrated** with almost any CI/CD system: GitHub Actions, GitLab, Drone CI, Jenkins, etc.

### Example of integration with GitHub Actions

> **Prerequisites:**
>
> - The project repository has already
    been [generated and prepared](../project-management/preparation-of-project-repository.md) using RMK.

Create the following workflow in your project repository at `.github/workflows/release-update.yaml`.
An example content of the [GitHub Actions](https://github.com/features/actions)' workflow:

```yaml
name: Release update

on:
  workflow_dispatch:
    inputs:
      image_repository_full_name:
        description: Image repository full name of application.
        required: true
      version:
        description: Current application version.
        required: true

jobs:
  release-update:
    runs-on: ubuntu-22.04
    steps:
      - name: Checkout main repository
        uses: actions/checkout@v4
        with:
          ref: ${{ github.ref }}
          fetch-depth: 0

      - name: Release update
        env:
          AWS_REGION: us-east-1
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          RMK_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN_REPO }}
          RMK_SLACK_WEBHOOK: ${{ secrets.RMK_SLACK_WEBHOOK }}
          RMK_SLACK_CHANNEL: project-cd-notifications
          RMK_RELEASE_UPDATE_REPOSITORY: ${{ github.event.inputs.image_repository_full_name }}
          RMK_RELEASE_UPDATE_TAG: ${{ github.event.inputs.version }}
        run: |
          curl -sL "https://edenlabllc-rmk-tools-infra.s3.eu-north-1.amazonaws.com/rmk/s3-installer" | bash

          rmk config init --progress-bar=false --slack-notifications
          rmk release update --skip-ci --deploy
```

In this example, we have prepared a `GitHub Action` that expects two input parameters:

- `image_repository_full_name`
- `version`

As soon as a request with these parameters is sent to this action,
RMK will be executed, first analyzing all the `releases.yaml` files to match the `image_repository_full_name` and will
replace the tag field with the corresponding version if the versions differ. After that, it will 
**automatically commit** the changes to the current branch in the `releases.yaml` files where changes have been found. 
Then, it will **synchronize the releases** where the version changes were found.

An example of the `releases.yaml` file:

```yaml
# ...
foo:
  enabled: true
  image:
    repository: 123456789012.dkr.ecr.us-east-1.amazonaws.com/app.foo
    tag: v0.1.0
bar:
  enabled: true
  image:
    repository: 123456789012.dkr.ecr.us-east-1.amazonaws.com/app.bar
    tag: v0.1.1
# ...
```

To fully automate code delivery in the CI pipeline, add a step that triggers the deployment after building and pushing
the container image. This step should pass the full image repository name and tag via an API call
using [cURL](https://en.wikipedia.org/wiki/CURL) or [GitHub CLI](https://cli.github.com/). This ensures **seamless
deployment** to the infrastructure environment.
