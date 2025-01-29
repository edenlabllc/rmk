# Quickstart

## Introduction

This guide demonstrates **how to use RMK** to prepare the structure of a new project, create a local cluster based on
[K3D](configuration/configuration-management/init-k3d-provider.md),
and deploy your first application ([Nginx](https://nginx.org/)) using Helmfile releases.

All of this will be done in just **five steps**.

## Prerequisites

- Fulfilled [requirements](index.md#requirements) for proper RMK operation
- Installed [RMK](index.md#installation)

This example assumes the [project](configuration/project-management/requirement-for-project-repository.md) (tenant) name
is `rmk-test`, the
current [Git branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches)
is `develop`, the
configured [Git remote](https://docs.github.com/en/get-started/getting-started-with-git/managing-remote-repositories)
is `origin`.

## Steps

1. Generate
   the [project structure](configuration/project-management/requirement-for-project-repository.md#expected-repository-structure),
   the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file, and [SOPS
   Age keys](configuration/secrets-management/secrets-management.md#secret-keys):

   ```shell
   rmk project generate --scope rmk-test --environment "develop.root-domain=localhost" --create-sops-age-keys
   ```

   > The `deps` scope is the default one and is **added unconditionally** during the project generation process, no need
   to specify it explicitly.

   <details>
      <summary>Example output</summary>
      ```text
      2025-01-29T14:20:02.954+0100	INFO	file /home/user/rmk-test/etc/deps/develop/values/aws-cluster.yaml.gotmpl generated
      2025-01-29T14:20:02.956+0100	INFO	file /home/user/rmk-test/etc/deps/develop/values/azure-cluster.yaml.gotmpl generated
      2025-01-29T14:20:02.956+0100	INFO	file /home/user/rmk-test/etc/deps/develop/values/gcp-cluster.yaml.gotmpl generated
      2025-01-29T14:20:02.957+0100	INFO	file /home/user/rmk-test/etc/deps/develop/globals.yaml.gotmpl generated
      2025-01-29T14:20:02.957+0100	INFO	file /home/user/rmk-test/etc/deps/develop/releases.yaml generated
      2025-01-29T14:20:02.957+0100	INFO	file /home/user/rmk-test/etc/deps/develop/secrets/.spec.yaml.gotmpl generated
      2025-01-29T14:20:02.957+0100	INFO	file /home/user/rmk-test/etc/deps/develop/secrets/.sops.yaml generated
      2025-01-29T14:20:02.957+0100	INFO	file /home/user/rmk-test/etc/rmk-test/develop/globals.yaml.gotmpl generated
      2025-01-29T14:20:02.958+0100	INFO	file /home/user/rmk-test/etc/rmk-test/develop/releases.yaml generated
      2025-01-29T14:20:02.958+0100	INFO	file /home/user/rmk-test/etc/rmk-test/develop/secrets/.spec.yaml.gotmpl generated
      2025-01-29T14:20:02.958+0100	INFO	file /home/user/rmk-test/etc/rmk-test/develop/values/rmk-test-app.yaml.gotmpl generated
      2025-01-29T14:20:02.958+0100	INFO	file /home/user/rmk-test/etc/rmk-test/develop/secrets/.sops.yaml generated
      2025-01-29T14:20:02.958+0100	INFO	file /home/user/rmk-test/.gitignore generated
      2025-01-29T14:20:02.959+0100	INFO	file /home/user/rmk-test/helmfile.yaml.gotmpl generated
      2025-01-29T14:20:02.959+0100	INFO	file /home/user/rmk-test/README.md generated
      2025-01-29T14:20:02.986+0100	INFO	generate age key for scope: deps
      2025-01-29T14:20:02.986+0100	INFO	update SOPS config file: /home/user/rmk-test/etc/deps/develop/secrets/.sops.yaml
      2025-01-29T14:20:03.000+0100	INFO	generate age key for scope: rmk-test
      2025-01-29T14:20:03.001+0100	INFO	update SOPS config file: /home/user/rmk-test/etc/rmk-test/develop/secrets/.sops.yaml
      ```
   </details>

2. [Initialize RMK configuration](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration)
   for the repository:

   ```shell
   rmk config init
   ```

   > The default cluster provider is K3D.

   <details>
      <summary>Example output</summary>
      ```text
      2025-01-29T14:22:44.548+0100	INFO	loaded config file by path: /home/user/.rmk/config/rmk-test-develop.yaml
      2025-01-29T14:22:44.550+0100	INFO	RMK will use values for develop environment
      2025-01-29T14:22:44.553+0100	INFO	starting package download: cluster-deps.bootstrap.infra-v0.1.0
      2025-01-29T14:22:45.790+0100	INFO	downloaded: cluster-deps.bootstrap.infra-v0.1.0
      2025-01-29T14:22:45.793+0100	INFO	starting package download: helmfile.hooks.infra-v1.29.1
      2025-01-29T14:22:46.598+0100	INFO	downloaded: helmfile.hooks.infra-v1.29.1
      2025-01-29T14:22:46.864+0100	INFO	time spent on initialization: 2s
      ```
   </details>   

3. Create a local [K3D](configuration/configuration-management/init-k3d-provider.md) cluster:

   ```shell
   rmk cluster k3d create
   ```

   > Ensure that Docker is **running**.

   <details>
      <summary>Example output</summary>
      ```text
      INFO[0000] Using config file /var/folders/_d/y2s0znsj5l117xk90392xc540000gn/T/k3d-config.51481123.yaml (k3d.io/v1alpha5#simple)
      INFO[0000] portmapping '8080:80' targets the loadbalancer: defaulting to [servers:*:proxy agents:*:proxy]
      INFO[0000] portmapping '8443:443' targets the loadbalancer: defaulting to [servers:*:proxy agents:*:proxy]
      INFO[0000] portmapping '9111:9000' targets the loadbalancer: defaulting to [servers:*:proxy agents:*:proxy]
      INFO[0000] Prep: Network
      INFO[0000] Created network 'k3d-rmk-test-develop'
      INFO[0000] Created image volume k3d-rmk-test-develop-images
      INFO[0000] Starting new tools node...
      INFO[0000] Starting node 'k3d-rmk-test-develop-tools'
      INFO[0001] Creating node 'k3d-rmk-test-develop-server-0'
      INFO[0001] Creating LoadBalancer 'k3d-rmk-test-develop-serverlb'
      INFO[0002] Pulling image 'ghcr.io/k3d-io/k3d-proxy:5.7.3'
      INFO[0015] Using the k3d-tools node to gather environment information
      INFO[0016] Starting new tools node...
      INFO[0016] Starting node 'k3d-rmk-test-develop-tools'
      INFO[0019] Starting cluster 'rmk-test-develop'
      INFO[0019] Starting servers...
      INFO[0022] Starting node 'k3d-rmk-test-develop-server-0'
      INFO[0047] All agents already running.
      INFO[0047] Starting helpers...
      INFO[0047] Starting node 'k3d-rmk-test-develop-serverlb'
      INFO[0053] Injecting records for hostAliases (incl. host.k3d.internal) and for 3 network members into CoreDNS configmap...
      INFO[0056] Cluster 'rmk-test-develop' created successfully!
      INFO[0056] You can now use it like this:
      kubectl cluster-info
      ```
   </details>

4. [Generate and encrypt secrets](configuration/secrets-management/secrets-management.md#batch-secrets-management) for
   the Helmfile releases, including Nginx:

   ```shell
   rmk secret manager generate --scope rmk-test --environment develop
   rmk secret manager encrypt --scope rmk-test --environment develop
   ```

   <details>
      <summary>Example output</summary>
      ```text
      TBD
      ```
   </details>

5. Deploy ([sync](configuration/release-management/release-management.md#synchronization-of-all-releases)) all
   Helmfile releases, including Nginx, to the local K3D cluster:

   ```shell
   rmk release sync
   ```

   <details>
      <summary>Example output</summary>
      ```text
      Release "rmk-test-app" does not exist. Installing it now.
      NAME: rmk-test-app
      LAST DEPLOYED: Wed Jan 29 14:23:54 2025
      NAMESPACE: rmk-test
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      NOTES:
      The app will be available by url:
      rmk-test-app rmk-test 1 2025-01-29 14:23:54.839083 +0100 CET deployed app-1.6.0
      ```
   </details>

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

To allow other team members to use an existing project, the initial user should commit the changes and push them to
the [version control system (VCS)](https://github.com/resources/articles/software-development/what-is-version-control),
e.g., [GitHub](https://github.com):

```shell
git commit -am "Generate RMK project structure, create SOPS secrets, deploy Nginx release."
git push origin develop
```

After that, other team members can set up the project by pulling the latest changes:

```shell
git checkout develop
git pull origin develop
```

> By design, [SOPS Age keys](configuration/secrets-management/secrets-management.md#secret-keys) are **Git-ignored** and
> **never committed** to the repository.  
> Therefore, when using a local K3D cluster, secret keys are **not shared** and should be **recreated** on another
> machine before proceeding with the [steps](#steps):
>
> ```shell
> rmk secret keys create
> ```
>
> If **sharing** the secret keys is required, **consider switching** from a K3D provider to a
>
supported [cloud provider](configuration/configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers).

Finally, the team members should follow all the [steps](#steps) **except the 1st one**, as the project has already been
generated.
