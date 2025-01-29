# Quickstart

## Introduction

This guide demonstrates **how to use RMK** to prepare the structure of a new project, create a local cluster based on
K3D,
and **deploy** your first **application** ([Nginx](https://nginx.org/)) using Helmfile releases.

All of this will be done in just **five steps**.

## Prerequisites

- Fulfilled [requirements](index.md#requirements) for proper RMK operation
- Installed [RMK](index.md#installation)

## Steps

This example assumes the tenant is `rmk-test` and the current branch is `develop`.

1. Generate the project structure and SOPS age keys based on
   the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file:

   ```shell
   rmk project generate --scope rmk-test --environments "develop.root-domain=localhost" --create-sops-age-keys
   ```

   > The `deps` scope is the default one and is **added unconditionally** during the project generation process.

2. [Initialize RMK configuration](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration)
   for the repository:

   ```shell
   rmk config init
   ```

   > The default cluster provider for the `init` command is `K3D`.

3. Create a local K3D cluster:

   ```shell
   rmk cluster k3d create
   ```

   > Ensure that Docker is **running**.

4. Generate and encrypt secrets for the Helmfile releases, including Nginx:

   ```shell
   rmk secret manager generate --scope rmk-test --environment develop
   rmk secret manager encrypt --scope rmk-test --environment develop
   ```

5. Deploy ("sync") all Helmfile releases, including Nginx, to the local K3D cluster:

   ```shell
   rmk release sync
   ```

At this stage, the project structure is **prepared**, and the application (Nginx) has been successfully **deployed** via
Helmfile to the local K3D cluster and is now **running**.

## Verifying the deployment

Verify the application's availability in the Kubernetes cluster:

```shell
kubectl --namespace rmk-test port-forward "$(kubectl --namespace rmk-test get pod --output name)" 8080:80
```

Then, open your browser and visit [http://localhost:8080](http://localhost:8080), alternatively, run:

```shell
open http://localhost:8080
```

You should see the Nginx welcome page.

## Collaborating with other team members

### Working with an existing project

To allow other team members to use an existing project, the initial person should commit the changes and push them to
your VCS (e.g., GitHub):

```shell
git commit -am "Generate RMK project structure, create SOPS secrets, deploy Nginx release."
git push origin develop
```

After that, other team members can deploy the project by pulling the latest changes:

```shell
git checkout develop
git pull origin develop
```

> The [secret keys](configuration/secrets-management/secrets-management.md#secret-keys) are Git-ignored, they
> should **never be committed** to Git.
>
> By default, when a local cluster is created via the K3D cluster provider, the private SOPS Age keys **are not shared**
> and should be recreated, then all the secrets should be regenerated and re-encoded:
> ```shell
> rmk secret keys create
> ```
> Otherwise, [initialize any cloud provider](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers), 
> supported by RMK.

Finally, they need to execute all the steps **except the 1st one**, because the project has been generated already.
