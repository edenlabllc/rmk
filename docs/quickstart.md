# Quickstart

## Introduction

This guide demonstrates how to use `RMK` to prepare the structure of a new project
create a local cluster based on `K3D`, deploy your first application ([Nginx](https://nginx.org/)) using `Helmfile`
releases.

## Prerequisites

- A
  prepared [project repository](configuration/project-management/preparation-of-project-repository.md#preparation-of-the-project-repository)
- Installed [RMK](index.md#installation)
- The fulfilled [requirements](index.md#requirements) for proper RMK operation.

## Preparing a cluster

This example assume, the tenant is `rmk-test`, current branch is `develop`.

1. Checkout the needed branch:

   ```shell
   git checkout -b develop
   ```

2. [Initialize RMK configuration](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration)
   for the repository:

   ```shell
   rmk config init
   ```

   > The default cluster provider for the `init` command is `K3D`.

3. Generate the project structure according to
   the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file:

   ```shell
   rmk project generate --scope rmk-test --environments "develop.root-domain=localhost" --create-sops-age-keys
   ```
   
   > The `deps` scope is the default one, it is added unconditionally during the project generation process.

4. Create a local K3D cluster:

   ```shell
   rmk cluster k3d create
   ```

   > According to the [requirements](index.md#requirements), ensure that Docker is installed in and running.
   
5. Generate and encrypt secrets for the `Helmfile` releases including `Nginx`.

   ```shell
   rmk secret manager generate --scope rmk-test --environment develop
   rmk secret manager encrypt --scope rmk-test --environment develop
   ```

6. Deploy ("sync") all the `Helmfile` releases including `Nginx` to the local `K3D` cluster:

   ```shell
   rmk release sync
   ```

At this stage, we have completed the deployment of our test application (`Nginx`) provided by the `Helmfile` release
to the local `K3D` cluster, also the structure of the future project has been prepared.

## Check the deployment

We can check the availability of the application in the Kubernetes cluster using the following command:

```shell
kubectl --namespace rmk-test port-forward "$(kubectl --namespace rmk-test get pod --output name)" 8088:80
```

Open your browser and enter the [http://localhost:8088](http://localhost:8088) address, after which you will see the
`Nginx` welcome page.

```shell
open http://localhost:8088
```

## Using cluster by other team members

For other team members to use the project, commit the changes to a Git branch and push them to your VCS (e.g., GitHub):

```shell
git commit -am "Generate project structure, create secrets, deploy Nginx."
```

Then, upload the private SOPS Age keys using the following command:

```shell
rmk secret keys upload
```

> The secret keys are Git-ignored and never committed to Git.

After that, your team members will be able to deploy this project on their own.

```shell
git checkout develop
git pull origin develop
```

then executing steps 2,5,6
