# Initialization of GCP cluster provider

## List of main attributes of the RMK configuration

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the project (tenant) name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant name, which is equivalent to the project name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
cluster-provider: gcp # Selected cluster provider.
# ...
gcp-region: us-east1 # GCP region of the current Kubernetes cluster.
gcp:
  app-credentials-path: /Users/alexalex/.config/gcloud/gcp-credentials-rmk-test-develop.json # Absolute path to GCP service account file.   
  project-id: project-name # GCP project name. Got from GCP service account file.
# ...
```

## Prerequisites

1. Having an account and project in GCP and a created service account with access roles in IAM: `Editor`, `Secret
   Manager Admin`, `Kubernetes Engine Admin`.
   > See the
   > [useful link](https://cloud.google.com/iam/docs/understanding-roles).

2. Enable the following services
   API: `Kubernetes Engine API (container.googleapis.com)`, `Cloud Compute Engine API (compute.googleapis.com)`, `Cloud NAT API (servicenetworking.googleapis.com)`, `IAM Service API (iam.googleapis.com)`, `Cloud Logging API (logging.googleapis.com)`, `Cloud Monitoring API (monitoring.googleapis.com)`.
   > See the
   > [useful link](https://cloud.google.com/apis?hl=en).

3. Allocated [quotas](https://cloud.google.com/docs/quotas/overview) for specific family VMs in the required region.

## Configuration

If a [GCP service account key](https://cloud.google.com/iam/docs/keys-create-delete#creating) file with the correct name
was not created during the initial configuration, RMK will generate it automatically and store it at the following path:

```shell
${HOME}/.config/gcloud/gcp-credentials-<project_name>-<project_branch>.json
```

The 2 supported configuration scenarios are:

* **via RMK flags**:
  ```shell
  rmk config init --cluster-provider=gcp 
    --gcp-region=us-east1 \
    --google-application-credentials <path_to_exported_google_service_account_file>
  ```

* **via environment variables**: `GCP_REGION`, `GOOGLE_APPLICATION_CREDENTIALS`.
  ```shell
  export GCP_REGION=us-east1
  export GOOGLE_APPLICATION_CREDENTIALS=<path_to_exported_google_service_account_file>
  rmk config init --cluster-provider=gcp
  ```  

If environment variables were set before running the command, RMK will create a GCP service account file based on their
values. If flags are specified, RMK will prioritize them over environment variables, as **CLI flags take precedence**.

## Reconfiguration of the GCP service account attributes if wrong credentials has been input

Modify the value of a specific flag if changes are needed:

```shell
rmk config init --google-application-credentials <path_to_new_exported_google_service_account_file>
```
