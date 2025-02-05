# Initialization of Azure cluster provider

## List of main attributes of the RMK configuration

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the tenant name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant repository name.
environment: develop # Environment name.
root-domain: rmk-test-develop.edenlab.dev # Root domain name used across the cluster.
cluster-provider: azure # Selected cluster provider.
# ...
azure:
  key-vault:
    key-vault-name: kv-ecc1c839a7b9bf5e # Azure Key Vault autogenerate name.
    key-vault-uri: https://kv-ecc1c839a7b9bf5e.vault.azure.net/ # Azure Key Vault API URL.
    resource-group-name: rmk-test-sops-age-keys # Azure resource group name for Key Vault.
  location: eastus # Azure location of the current Kubernetes cluster.
  subscription-id: abcdef12-3456-7890-abcd-ef1234567890 # Azure subscription ID.
# ...
```

### Prerequisites

1. Having an subscription in Azure and a created service principal with access roles in IAM: Contributor, Key Vault Secrets Officer.
   [Useful links](https://learn.microsoft.com/en-us/entra/identity-platform/howto-create-service-principal-portal).

2. Enable the following resource providers:
   
   - Microsoft.Authorization
   - Microsoft.Compute
   - Microsoft.ContainerService
   - Microsoft.ManagedIdentity
   - Microsoft.Network

3. Allocated quotas for specific family VMs in the required region.

### Configuration of Azure

If an Azure service principal file `service-principal-credentials_<project_name>-<project_branch>.json` with 
the correct name has not been created previously during the first initialization of the configuration, RMK will start the creation process. 
RMK store the Azure service principle file by path: `${HOME}/.azure/service-principal-credentials_<project_name>-<project_branch>.json`.

The 3 supported configuration scenarios are:

* **through RMK flags**:
  ```shell
  rmk config init --cluster-provider=azure \ 
    --azure-client-id=<azure_client_id> \
    --azure-client-secret=<azure_client_secret> \
    --azure-location=eastus \
    --azure-subscription-id=<azure_subscription_id> \ 
    --azure-tenant-id=<azure_tenant_id>
  ```
  
* **through environment variables**: `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_LOCATION`, `AZURE_SUBSCRIPTION_ID`, `AZURE_TENANT_ID`.
  ```shell
  rmk config init --cluster-provider=azure
  ```
  
* **through stdin from az CLI**:
  ```shell
  az login
  az ad sp create-for-rbac --name rmk-test --role contributor --scopes="/subscriptions/<azure_subscription_id>" --output json | \
    rmk config init --cluster-provider=azure --azure-location=eastus --azure-service-principle
  ```

If the environment variables has been declared before the  `rmk config init --cluster-provider=azure` command was run,
RMK will create a Azure service principal file based on their values.
If flags will be declared, RMK will create a Azure service principal file based on values flags because flags has priority.

### Reconfiguration of the Azure service principal attributes if wrong credentials has been input

Change the value of a specific flag if adjustments are required.

```shell
rmk config init --azure-client-id=<new_azure_client_id> --azure-client-secret=<new_azure_client_secret>
```
