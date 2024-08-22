# Quickstart

This guide demonstrates how to use `RMK` to prepare the structure of a new project in five steps,
create a local cluster based on `K3D`, deploy your first application ([Nginx](https://nginx.org/)) using `Helmfile` releases.

> Prerequisites:
> 
> - An active AWS user with access keys and the `AdministratorAccess` permissions.
> - A prepared [project repository](configuration/project-management/preparation-of-project-repository.md#preparation-of-the-project-repository)
> - Installed [RMK](README.md#installation)
> - The fulfilled [requirements](README.md#requirements) for proper RMK operation.

0. Create a [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) 
   file in the root of the project repository with the following content:

```yaml
project:
  spec:
    environments:
      - develop
    owners:
      - owner
    scopes:
      - rmk-test
inventory:
  clusters:
    k3d.provisioner.infra:
      version: v0.2.0
      url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
  helm-plugins:
    diff:
      version: v3.8.1
      url: https://github.com/databus23/helm-diff
    helm-git:
      version: v0.15.1
      url: https://github.com/aslafy-z/helm-git
    secrets:
      version: v4.5.0
      url: https://github.com/jkroepke/helm-secrets
  hooks:
    helmfile.hooks.infra:
      version: v1.18.0
      url: git::https://github.com/edenlabllc/{{.Name}}.git?ref={{.Version}}
  tools:
    terraform:
      version: 1.0.2
      url: https://releases.hashicorp.com/{{.Name}}/{{.Version}}/{{.Name}}_{{.Version}}_{{.Os}}_amd64.zip
      checksum: https://releases.hashicorp.com/{{.Name}}/{{.Version}}/{{.Name}}_{{.Version}}_SHA256SUMS
      os-linux: linux
      os-mac: darwin
    kubectl:
      version: 1.27.6
      url: https://dl.k8s.io/release/v{{.Version}}/bin/{{.Os}}/amd64/{{.Name}}
      checksum: https://dl.k8s.io/release/v{{.Version}}/bin/{{.Os}}/amd64/{{.Name}}.sha256
      os-linux: linux
      os-mac: darwin
    helm:
      version: 3.10.3
      url: https://get.helm.sh/{{.Name}}-v{{.Version}}-{{.Os}}-amd64.tar.gz
      checksum: https://get.helm.sh/{{.Name}}-v{{.Version}}-{{.Os}}-amd64.tar.gz.sha256sum
      os-linux: linux
      os-mac: darwin
    helmfile:
      version: 0.157.0
      url: https://github.com/{{.Name}}/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}_{{.Version}}_{{.Os}}_amd64.tar.gz
      checksum: https://github.com/{{.Name}}/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}_{{.Version}}_checksums.txt
      os-linux: linux
      os-mac: darwin
    sops:
      version: 3.8.1
      url: https://github.com/getsops/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}-v{{.Version}}.{{.Os}}
      os-linux: linux.amd64
      os-mac: darwin
      rename: true
    age:
      version: 1.1.1
      url: https://github.com/FiloSottile/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}-v{{.Version}}-{{.Os}}-amd64.tar.gz
      os-linux: linux
      os-mac: darwin
    k3d:
      version: 5.6.0
      url: https://github.com/k3d-io/{{.Name}}/releases/download/v{{.Version}}/{{.Name}}-{{.Os}}-amd64
      os-linux: linux
      os-mac: darwin
      rename: true
```

1. Run the [RMK configuration initialization](configuration/configuration-management.md#initialization-of-rmk-configuration) command for the repository:

   ```shell
   rmk config init --root-domain=localhost --github-token=<github_personal_access_token>
   ```
   
   > When executing the command, properly fill in the AWS credentials and region. 
   > RMK will save the references for them in the system and use them for subsequent executions of this command. 
   > In our example, the AWS credentials are used to create an S3 bucket for storing private SOPS Age keys 
   > and distributing them among team members.

2. Generate the project structure according to the [project.yaml](configuration/project-management/preparation-of-project-repository.md#projectyaml) file:

   ```shell
   rmk project generate --create-sops-age-keys
   ```

3. Create a local K3D cluster:

   > Before running this step, ensure that Docker is installed in the system according to the [requirements](README.md#requirements).
   
   ```shell
   rmk cluster k3d create
   ```

4. Generate and encrypt secrets for the `Helmfile` release (Nginx):

   ```shell
   rmk secret manager generate
   rmk secret manager encrypt
   ```

5. Deploy the `Helmfile` release (Nginx) to the local `K3D` cluster:

   ```shell
   rmk release sync
   ```

At this stage, we have completed the deployment of our test application (Nginx) provided by the `Helmfile` release 
to the local `K3D` cluster, also the structure of the future project has been prepared. 

We can check the availability of the application in the Kubernetes cluster using the following command:

```shell
kubectl port-forward $(kubectl get pod --namespace rmk-test --output name) 8088:80 --namespace rmk-test
```

Open your browser and enter the http://localhost:8088 address, after which you will see the Nginx welcome page.

Next, you can commit your changes to a Git branch and push them to your VCS (e.g., GitHub). 
You can also upload the private SOPS Age keys using the following command: 

```shell
rmk secret keys upload
```

After that, your team members will be able to deploy this project on their own, skipping the 2nd and 4th steps.
