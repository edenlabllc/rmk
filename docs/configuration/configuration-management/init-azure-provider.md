# Initialization of Azure cluster provider

## List of main attributes of the RMK configuration

```yaml
name: rmk-test-develop # RMK config name, a unique identifier which consists of the project (tenant) name and the abbreviated name of the Git branch.
tenant: rmk-test # Tenant name, which is equivalent to the project name.
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

1. Having an subscription in Azure and a created service principal with access roles in IAM:
   [Contributor](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/privileged#contributor),
   [Key Vault Secrets Officer](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/security#key-vault-secrets-officer).
   > See the
   > [useful link](https://learn.microsoft.com/en-us/entra/identity-platform/howto-create-service-principal-portal).

2. Enable the following resource
   providers: `Microsoft.Authorization`, `Microsoft.Compute`, `Microsoft.ContainerService`, `Microsoft.ManagedIdentity`, `Microsoft.Network`.

   > See the
   > [useful link](https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/resource-providers-and-types).

3. Allocated [quotas](https://learn.microsoft.com/en-us/azure/quotas/quotas-overview) for specific family VMs in the
   required region.

### Configuration of Azure

If
an [Azure service principal](https://learn.microsoft.com/en-us/entra/identity-platform/app-objects-and-service-principals?tabs=browser)
file was not created during the initial configuration, RMK will generate it automatically and store it at the following
path:

```shell
`${HOME}/.azure/service-principal-credentials_<project_name>-<project_branch>.json`.
```

The 3 supported configuration scenarios are:

* **via RMK flags**:
  ```shell
  rmk config init --cluster-provider=azure \ 
    --azure-client-id=<azure_client_id> \
    --azure-client-secret=<azure_client_secret> \
    --azure-location=eastus \
    --azure-subscription-id=<azure_subscription_id> \ 
    --azure-tenant-id=<azure_tenant_id>
  ```

* **via environment variables
  **: `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_LOCATION`, `AZURE_SUBSCRIPTION_ID`, `AZURE_TENANT_ID`.
  ```shell
  export AZURE_CLIENT_ID=<azure_client_id>
  export AZURE_CLIENT_SECRET=<azure_client_secret>
  export AZURE_LOCATION=eastus
  export AZURE_SUBSCRIPTION_ID=<azure_subscription_id>
  export AZURE_TENANT_ID=<azure_tenant_id>
  rmk config init --cluster-provider=azure
  ```

* **via `stdin` using output of the `az` CLI**:
  ```shell
  # login interactively: https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli-interactively#interactive-login
  az login
  az ad sp create-for-rbac --name rmk-test --role contributor --scopes="/subscriptions/<azure_subscription_id>" --output json | \
    rmk config init --cluster-provider=azure --azure-location=eastus --azure-service-principle
  ```

If environment variables were set before running the command, RMK will create an Azure service principal file based on
their values. If flags are specified, RMK will prioritize them over environment variables, as **CLI flags take
precedence**.

### Reconfiguration of the Azure service principal attributes if wrong credentials has been input

Modify the value of a specific flag if changes are needed:

```shell
rmk config init --azure-client-id=<new_azure_client_id> --azure-client-secret=<new_azure_client_secret>
```
