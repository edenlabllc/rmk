# Quickstart

## Introduction

This guide demonstrates how to use RMK to prepare the structure of a new project, create a local cluster based on K3D,
and deploy your first application ([Nginx](https://nginx.org/)) using Helmfile releases.

## Prerequisites

- A
  prepared [project repository](configuration/project-management/preparation-of-project-repository.md#preparation-of-the-project-repository)
- Installed [RMK](index.md#installation)
- Fulfilled [requirements](index.md#requirements) for proper RMK operation

## Preparing a cluster

This example assumes the tenant is `rmk-test` and the current branch is `develop`.

1. [Initialize RMK configuration](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration)
   for the repository:

   ```shell
   rmk config init
   ```

   > The default cluster provider for the `init` command is `K3D`.

2. Generate the project structure based on
   the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file:

   ```shell
   rmk project generate --scope rmk-test --environments "develop.root-domain=localhost" --create-sops-age-keys
   ```

   > The `deps` scope is the default one and is added automatically during the project generation process.

3. Create a local K3D cluster:

   ```shell
   rmk cluster k3d create
   ```

   > Ensure that Docker is **installed and running**, as required in the [requirements](index.md#requirements).

4. Generate and encrypt secrets for the Helmfile releases, including Nginx:

   ```shell
   rmk secret manager generate --scope rmk-test --environment develop
   rmk secret manager encrypt --scope rmk-test --environment develop
   ```

5. Deploy ("sync") all Helmfile releases, including Nginx, to the local K3D cluster:

   ```shell
   rmk release sync
   ```

At this stage, the test application (Nginx) has been deployed via Helmfile to the local K3D cluster, and the project
structure is prepared.

## Checking the deployment

Verify the application's availability in the Kubernetes cluster:

```shell
kubectl --namespace rmk-test port-forward "$(kubectl --namespace rmk-test get pod --output name)" 8088:80
```

Then, open your browser and visit:

```shell
open http://localhost:8088
```

You should see the Nginx welcome page.

## Using the cluster by other team members

To allow other team members to use the project, commit the changes and push them to your VCS (e.g., GitHub):

```shell
git commit -am "Generate project structure, create secrets, deploy Nginx."
git push origin develop
```

Then, upload the private SOPS Age keys:

```shell
rmk secret keys upload
```

> The secret keys are Git-ignored and should **never be committed** to Git.

Once these steps are completed, other team members can deploy the project by pulling the latest changes and secret keys:

```shell
git checkout develop
git pull origin develop
rmk secret keys download
```

Then, they only need to execute steps 1, 3, and 5.
