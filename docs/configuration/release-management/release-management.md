# Release management

RMK uses [Helmfile](https://github.com/helmfile/helmfile) for the release management.

RMK uses a reduced set of the `Helmfile` commands without changing their behavior. 
The full list of commands can be found in the [release category](../../commands.md#release). 
Additionally, flags are provided for the commands, which allow extending capabilities and help during the command execution debug.

For example:

```shell
rmk release build
rmk release list --selector app=name
rmk release template --selector app=name --skip-context-switch
rmk release sync --helmfile-log-level=debug --selector app=name 
rmk release destroy 
```

> The `--skip-context-switch` (`-s`) flag can be used for the commands like `rmk release template` to skip switching to a Kubernetes cluster.
> This might be useful in the situations, when a cluster has not been provisioned yet and its releases and values are being developed.

In a project repository, all the release values files are stored in the `etc/<scope>/<env>/values/` directories.
For example:

```
etc/deps/develop/values/
etc/kodjin/staging/values/
```

> The release values are inherited by the projects, e.g., the upstream project's values are included into the downstream project's values.

All `releases.yaml` files controlling which releases are enabled/disabled are stored in the `etc/<scope>/<env>/`directories.
For example:

```
etc/deps/develop/releases.yaml
```

> The `releases.yaml` files are not inherited by the projects in contrast to the values. Each project should have its
> `releases.yaml` files for all deployed scopes and envs.
> Running any of the commands in the release category will trigger the dependency resolution mechanism,
> as well as the check for the Kubernetes context for the current environment to prevent releases 
> from being synchronized outside the environment context.

The release installation order is declared in `helmfile.yaml.gotmpl` file.

## Examples of Usage

### List of all available releases

```shell
rmk release list
```

### Viewing a specific release YAML after the Helm values template rendering

```shell
rmk release template --selector app=traefik
```

### Synchronization of a specific scope of the releases

```shell
rmk release sync --selector scope=deps
```

### Synchronization of a specific release with passing the "--set" Helmfile argument

```shell
rmk release sync --selector app=redis --helmfile-args="--set='values.name=foo'"
```

### Destroy all releases

```shell
rmk release destroy
```

Among the `Helmfile` selectors, the following [predefined keys](https://helmfile.readthedocs.io/en/stable/#labels-overview) 
are provided out of the box: 

- Release name.
- Release namespace.
- Chart name.

For example:

```shell
rmk release sync --selector namespace=kube-system
```

## Overriding release values for inherited projects

It is possible to override any release value for the [inherited project repository](../project-management/dependencies-management-and-project-inheritance.md#dependencies-management-and-project-inheritance).
You can override any element separately in its YAML file.

For example, you have the following file in the upstream project by the `.PROJECT/dependencies/deps.bootstrap.infra-<version>/etc/deps/develop/values/metrics-server.yaml` path:

```yaml
apiService:
  create: true
extraArgs:
  - --metric-resolution=10s
  - --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname
  - --cert-dir=/tmp
resources:
  requests:
    cpu: 100m
    memory: 200Mi
containerSecurityContext:
  readOnlyRootFilesystem: true
extraVolumes:
  - name: tmpdir
    emptyDir: {}
extraVolumeMounts:
  - name: tmpdir
    mountPath: /tmp
```

Then you want to change the `resources.requests.cpu` value of the `develop` environment in your downstream project. 
In this case, you don't need to copy the whole file but only change the concrete value by repeating the YAML path to it. 
So, your file with the override will look like this:

```yaml
resources:
  requests:
    cpu: 100m
```

> You cannot override a part of an element if it is an array. 
> If you want to override the name of the extraVolumeMounts field of the example file above, you cannot use the following content:
> ```yaml
> extraVolumeMounts:
>   - name: tmp
> ```
> The correct way to override is to provide the whole item of the array:
> ```yaml
> extraVolumeMounts:
>   - name: tmp
>     mountPath: /tmp
> ```

To check the final result, run the `rmk release template` command and see the final YAML.

## Release update and integration into the CD pipeline

The `rmk release update` command automates the process of updating and delivering releases 
according to the version changes of artifacts (container images) following the [GitOps](https://www.gitops.tech) methodology.

Since RMK is a binary file that can be downloaded and installed on any Unix-based CI/CD system, 
it can be integrated with almost any CI/CD system: GitHub Actions, GitLab, Drone CI, Jenkins, etc.

### Example of integration with GitHub Actions

> Prerequisites:
> 
> - The project repository has already been generated and [prepared](../project-management/preparation-of-project-repository.md) using RMK.

Create the following workflow in your project repository at `.github/workflows/release-update.yaml`. 
An example content of the GitHub Actions' workflow:

[//]: # (  TODO ACTUALIZE)

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
RMK will be executed, first analyzing all the `releases.yaml` files to match the `image_repository_full_name` and will replace the tag field 
with the corresponding version if the versions differ. 
After that, it will automatically commit the changes to the current branch in the `releases.yaml` files where changes have been found. 
Then, it will synchronize the releases where the version changes were found.

An example of the `releases.yaml` file:

```yaml
# ...
foo:
  enabled: true
  image:
    repository: 123456789.dkr.ecr.us-east-1.amazonaws.com/app.foo
    tag: v0.11.1
bar:
  enabled: true
  image:
    repository: 123456789.dkr.ecr.us-east-1.amazonaws.com/app.bar
    tag: v0.16.0
# ...
```

To make the code delivery process fully automatic on the CI pipeline's side, after building and pushing 
the container image to the image repository, add a step that triggers the deployment action and passes 
the container image repository's full name and its final tag. This can be done via an API call using [cURL](https://en.wikipedia.org/wiki/CURL) 
or through [GitHub CLI](https://cli.github.com/). This way, we achieve automatic code delivery to the infrastructure environment.
