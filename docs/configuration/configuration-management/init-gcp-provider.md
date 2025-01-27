# Initialization of GCP cluster provider

## List of main attributes of the RMK configuration

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the tenant name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant repository name.
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

### Prerequisites

1. Having an account and project in GCP and a created service account with access roles in IAM: **Editor**, **Secret Manager Admin**, **Kubernetes Engine Admin**.
   [Useful links](https://cloud.google.com/iam/docs/service-accounts-create#creating).

2. Enable the following services API:

   - Kubernetes Engine API (container.googleapis.com)
   - Cloud Compute Engine API (compute.googleapis.com)
   - Cloud NAT API (servicenetworking.googleapis.com)
   - IAM Service API (iam.googleapis.com)
   - Cloud Logging API (logging.googleapis.com)
   - Cloud Monitoring API (monitoring.googleapis.com)
   
3. Allocated quotas for specific family VMs in the required region.

### Configuration of GCP

If an GCP service account file `gcp-credentials-<project_name>-<project_branch>.json` with the correct name 
has not been created previously during the first initialization of the configuration, RMK will start the creation process.
RMK store the GCP service account file by path: `${HOME}/.config/gcloud/gcp-credentials-<project_name>-<project_branch>.json`.

The 2 supported configuration scenarios are:

* **through RMK flags**:
  ```shell
  rmk config init --cp=gcp 
    --gcp-region=us-east1 \
    --google-application-credentials <path_to_exported_GCP_service_accout_file>
  ```
  
* **through environment variables**: `GCP_REGION`, `GOOGLE_APPLICATION_CREDENTIALS`.
  ```shell
  rmk config init --cluster-provider=gcp
  ```  

If the environment variables has been declared before the  `rmk config init --cluster-provider=gcp` command was run,
RMK will create a GCP service account file based on their values.
If flags will be declared, RMK will create a GCP service account file based on values flags because flags has priority.

### Reconfiguration of the GCP service account attributes if wrong credentials has been input

Change the value of a specific flag if adjustments are required.

```shell
rmk config init --google-application-credentials <path_to_new_exported_GCP_service_accout_file>
```
