# Quickstart

## Introduction

This guide demonstrates **how to use RMK** to prepare the structure of a new project, create a local cluster based on
[K3D](configuration/configuration-management/init-k3d-provider.md),
and deploy your first application ([Nginx](https://nginx.org/)) using Helmfile releases.

All of this will be done in just **five steps**.

## Prerequisites

- Fulfilled [requirements](index.md#requirements) for proper RMK operation
- Installed [RMK](index.md#installation)

## Steps

This example assumes the tenant is `rmk-test` and the current branch is `develop`.

1. Generate
   the [project structure](configuration/project-management/requirement-for-project-repository.md#expected-repository-structure),
   the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file, and [SOPS
   Age keys](configuration/secrets-management/secrets-management.md#secret-keys):

   ```shell
   rmk project generate --scope rmk-test --environments "develop.root-domain=localhost" --create-sops-age-keys
   ```

   > The `deps` scope is the default one and is **added unconditionally** during the project generation process, no need
   to specify it explicitly.

2. [Initialize RMK configuration](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration)
   for the repository:

   ```shell
   rmk config init
   ```

   > The default cluster provider is K3D.

3. Create a local [K3D](configuration/configuration-management/init-k3d-provider.md) cluster:

   ```shell
   rmk cluster k3d create
   ```

   > Ensure that Docker is **running**.

4. [Generate and encrypt secrets](configuration/secrets-management/secrets-management.md#batch-secrets-management) for the Helmfile releases, including Nginx:

   ```shell
   rmk secret manager generate --scope rmk-test --environment develop
   rmk secret manager encrypt --scope rmk-test --environment develop
   ```

5. Deploy ("[sync](configuration/release-management/release-management.md#synchronization-of-all-releases)") all Helmfile releases, including Nginx, to the local K3D cluster:

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

Then, open your browser and visit [http://localhost:8080](http://localhost:8080), or run:

```shell
open http://localhost:8080
```

You should see the Nginx welcome page.

## Collaborating with other team members

### Working with an already generated project

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

> By design, [SOPS Age keys](configuration/secrets-management/secrets-management.md#secret-keys) are Git-ignored, *
*never committed** to Git.
> Therefor, when a local K3D cluster is used, the secret keys are **not shared** and **should be recreated** on other
> machine before proceeding with the [steps](#steps),
>
> ```shell
> rmk secret keys create
> ```
>
> If **sharing the secret keys** is required, consider switching from a K3D provider to any
> other [cloud provider](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers)
> supported by RMK.

Finally, the members should go through all the [steps](#steps) **except the 1st one**, because the **project has been
generated** already.
