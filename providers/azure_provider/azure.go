package azure_provider

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msgrapherror "github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"rmk/util"
)

const (
	AzureClusterProvider = "azure"
	AzureHomeDir         = "." + AzureClusterProvider
	AzurePrefix          = "service-principal-credentials_"
	AzureKeyVaultRole    = "Key Vault Secrets Officer"
)

type AzureRawServicePrincipal struct {
	AppId       string `json:"appId"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	Tenant      string `json:"tenant"`
}

type AzureKeyVault struct {
	KeyVaultName      string `json:"-" yaml:"key-vault-name,omitempty"`
	KeyVaultURI       string `json:"-" yaml:"key-vault-uri,omitempty"`
	ResourceGroupName string `json:"-" yaml:"resource-group-name,omitempty"`
}

type AzureClient struct {
	Credentials           *azidentity.ClientSecretCredential         `json:"-" yaml:"-"`
	Ctx                   context.Context                            `json:"-" yaml:"-"`
	GraphServiceClient    *msgraphsdk.GraphServiceClient             `json:"-" yaml:"-"`
	GroupsClient          *armresources.ResourceGroupsClient         `json:"-" yaml:"-"`
	ManagedClustersClient *armcontainerservice.ManagedClustersClient `json:"-" yaml:"-"`
	RoleAssignmentsClient *armauthorization.RoleAssignmentsClient    `json:"-" yaml:"-"`
	RoleDefinitionsClient *armauthorization.RoleDefinitionsClient    `json:"-" yaml:"-"`
	VaultsClient          *armkeyvault.VaultsClient                  `json:"-" yaml:"-"`
}

type AzureConfigure struct {
	AzureClient    `json:"-" yaml:"-"`
	AzureKeyVault  `json:"-" yaml:"key-vault,omitempty"`
	ClientID       string `json:"client-id,omitempty" yaml:"-"`
	ClientSecret   string `json:"client-secret,omitempty" yaml:"-"`
	Location       string `json:"location,omitempty" yaml:"location,omitempty"`
	SubscriptionID string `json:"subscription-id,omitempty" yaml:"subscription-id,omitempty"`
	TenantID       string `json:"tenant-id,omitempty" yaml:"-"`
}

func NewAzureConfigure() *AzureConfigure {
	return &AzureConfigure{}
}

func NewRawSP() *AzureRawServicePrincipal {
	return &AzureRawServicePrincipal{}
}

func (ac *AzureConfigure) MergeAzureRawSP(asp *AzureRawServicePrincipal) {
	ac.ClientID = asp.AppId
	ac.ClientSecret = asp.Password
	ac.TenantID = asp.Tenant
}

func getTagStructName(i interface{}, name string) error {
	if field, ok := reflect.TypeOf(i).Elem().FieldByName(name); ok {
		return fmt.Errorf("service principal option %s required", strings.TrimSuffix(field.Tag.Get("json"), ",omitempty"))
	} else {
		return fmt.Errorf("field with name %s not defined", name)
	}
}

func generateKeyVaultName(tenant string) string {
	return "kv-" + fmt.Sprintf("%x", sha1.Sum([]byte(tenant)))[0:16]
}

func (ac *AzureConfigure) ValidateSPCredentials() error {
	if len(ac.ClientID) == 0 {
		return getTagStructName(ac, "ClientID")
	}

	if len(ac.ClientSecret) == 0 {
		return getTagStructName(ac, "ClientSecret")
	}

	if len(ac.Location) == 0 {
		return getTagStructName(ac, "Location")
	}

	if len(ac.SubscriptionID) == 0 {
		return getTagStructName(ac, "SubscriptionID")
	}

	if len(ac.TenantID) == 0 {
		return getTagStructName(ac, "TenantID")
	}

	return nil
}

func (ac *AzureConfigure) ReadSPCredentials(fileSuffix string) error {
	data, err := os.ReadFile(util.GetHomePath(AzureHomeDir, AzurePrefix+fileSuffix+".json"))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &ac)
}

func (ac *AzureConfigure) WriteSPCredentials(fileSuffix string) error {
	data, err := json.MarshalIndent(ac, "", " ")
	if err != nil {
		return err
	}

	data = []byte(string(data) + "\n")

	if err := os.MkdirAll(util.GetHomePath(AzureHomeDir), 0755); err != nil {
		return err
	}

	return os.WriteFile(
		util.GetHomePath(AzureHomeDir, AzurePrefix+fileSuffix+".json"),
		data, 0644)
}

func (ac *AzureConfigure) NewAzureClient(ctx context.Context, fileName string) error {
	if err := ac.ReadSPCredentials(fileName); err != nil {
		return err
	}

	cred, err := azidentity.NewClientSecretCredential(ac.TenantID, ac.ClientID, ac.ClientSecret, nil)
	if err != nil {
		return err
	}

	ac.GraphServiceClient, err = msgraphsdk.NewGraphServiceClientWithCredentials(cred, nil)

	GroupFactory, err := armresources.NewClientFactory(ac.SubscriptionID, cred, nil)
	if err != nil {
		return err
	}

	ManagedClustersFactory, err := armcontainerservice.NewClientFactory(ac.SubscriptionID, cred, nil)
	if err != nil {
		return err
	}

	RoleFactory, err := armauthorization.NewClientFactory(ac.SubscriptionID, cred, nil)
	if err != nil {
		return err
	}

	VaultFactory, err := armkeyvault.NewClientFactory(ac.SubscriptionID, cred, nil)
	if err != nil {
		return err
	}

	ac.Ctx = ctx
	ac.Credentials = cred
	ac.GroupsClient = GroupFactory.NewResourceGroupsClient()
	ac.ManagedClustersClient = ManagedClustersFactory.NewManagedClustersClient()
	ac.RoleAssignmentsClient = RoleFactory.NewRoleAssignmentsClient()
	ac.RoleDefinitionsClient = RoleFactory.NewRoleDefinitionsClient()
	ac.VaultsClient = VaultFactory.NewVaultsClient()

	return nil
}

func (ac *AzureConfigure) GetAzureClusterContext(groupName, clusterName string) ([]byte, error) {
	var cpTitle = strings.ToUpper(AzureClusterProvider[:1]) + strings.ToLower(AzureClusterProvider[1:])

	credentials, err := ac.ManagedClustersClient.ListClusterAdminCredentials(ac.Ctx, groupName, clusterName, nil)
	if err != nil {
		return nil, fmt.Errorf("kubecontext for %s provider's %s cluster not found",
			cpTitle, clusterName)
	}

	if len(credentials.CredentialResults.Kubeconfigs) == 1 {
		return credentials.CredentialResults.Kubeconfigs[0].Value, nil
	}

	return nil, fmt.Errorf("kubecontext for %s provider's %s cluster not found",
		cpTitle, clusterName)
}

func (ac *AzureConfigure) createKeyVaultResourceGroup() error {
	params := armresources.ResourceGroup{
		Location: to.Ptr(ac.Location),
	}

	update, err := ac.GroupsClient.CreateOrUpdate(ac.Ctx, ac.ResourceGroupName, params, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 403 {
			zap.S().Warnf("permission denied to create Azure Resource Group: %s", ac.ResourceGroupName)
			return nil
		}

		return err
	}

	ac.ResourceGroupName = *update.Name

	zap.S().Infof("created Azure Resource Group: %s", ac.ResourceGroupName)

	return nil
}

func (ac *AzureConfigure) existsKeyVaultResourceGroup() (bool, error) {
	existence, err := ac.GroupsClient.CheckExistence(ac.Ctx, ac.ResourceGroupName, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 403 {
			zap.S().Warnf("permission denied to check existence of Azure Resource Group: %s",
				ac.ResourceGroupName)
			return false, nil
		}

		return false, err
	}

	return existence.Success, nil
}

func (ac *AzureConfigure) CreateAzureKeyVault(tenant string) error {
	if err := ac.createKeyVaultResourceGroup(); err != nil {
		return err
	}

	ac.KeyVaultName = generateKeyVaultName(tenant)

	params := armkeyvault.VaultCreateOrUpdateParameters{
		Location: to.Ptr(ac.Location),
		Properties: &armkeyvault.VaultProperties{
			EnableRbacAuthorization: to.Ptr(true),
			SKU: &armkeyvault.SKU{
				Family: to.Ptr(armkeyvault.SKUFamilyA),
				Name:   to.Ptr(armkeyvault.SKUNameStandard),
			},
			TenantID: to.Ptr(ac.TenantID),
		},
	}

	update, err := ac.VaultsClient.BeginCreateOrUpdate(ac.Ctx, ac.ResourceGroupName, ac.KeyVaultName, params, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 403 {
			zap.S().Warnf("permission denied to create Azure Key Vault: %s",
				ac.KeyVaultName)
			return nil
		}

		return err
	}

	result, err := update.PollUntilDone(ac.Ctx, nil)
	if err != nil {
		return err
	}

	ac.KeyVaultURI = *result.Properties.VaultURI

	zap.S().Infof("created Azure Key Vault: %s, %s", ac.KeyVaultName, ac.KeyVaultURI)

	return nil
}

func (ac *AzureConfigure) DefineAzureKeyVaultResourceGroup(tenant string) error {
	if len(ac.ResourceGroupName) > 0 {
		return nil
	}

	rgNameFromScope, err := ac.getResourceGroupNameByRoleAssignments()
	if err != nil {
		return err
	} else if len(rgNameFromScope) > 0 {
		ac.ResourceGroupName = rgNameFromScope
		return nil
	}

	ac.ResourceGroupName = tenant + "-" + util.SopsRootName

	return nil
}

func (ac *AzureConfigure) getResourceGroupNameByRoleAssignments() (string, error) {
	get, err := ac.GraphServiceClient.ServicePrincipalsWithAppId(to.Ptr(ac.ClientID)).Get(ac.Ctx, nil)
	if err != nil {
		var graphErr *msgrapherror.ODataError
		if errors.As(err, &graphErr) && graphErr.GetStatusCode() == 403 {
			zap.S().Warnf("permission denied to get Azure Service Principal by app ID")
			return "", nil
		}

		return "", err
	}

	pager := ac.RoleAssignmentsClient.NewListForSubscriptionPager(
		&armauthorization.RoleAssignmentsClientListForSubscriptionOptions{
			Filter:   to.Ptr(`assignedTo('` + *get.GetId() + `')`),
			TenantID: nil,
		})
	for pager.More() {
		page, err := pager.NextPage(ac.Ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.StatusCode == 403 {
				zap.S().Warnf("permission denied to list Azure Role Assignments for Service Principal by ID: %s",
					*get.GetId())
				return "", nil
			}

			return "", err
		}

		for _, v := range page.Value {
			id, err := ac.RoleDefinitionsClient.GetByID(ac.Ctx, *v.Properties.RoleDefinitionID, nil)
			if err != nil {
				var respErr *azcore.ResponseError
				if errors.As(err, &respErr) && respErr.StatusCode == 403 {
					zap.S().Warnf("permission denied to get Azure Role Definitions by ID: %s",
						*v.Properties.RoleDefinitionID)
					return "", nil
				}

				return "", err
			}

			if *id.Properties.RoleName == AzureKeyVaultRole {
				scopePrefix := "/subscriptions/" + ac.SubscriptionID + "/resourceGroups/"
				if strings.Contains(*v.Properties.Scope, scopePrefix) {
					return strings.TrimPrefix(*v.Properties.Scope, scopePrefix), nil
				}
			}
		}
	}

	return "", nil
}

func (ac *AzureConfigure) GetAzureKeyVault(tenant string) (bool, error) {
	if ok, err := ac.existsKeyVaultResourceGroup(); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	ac.KeyVaultName = generateKeyVaultName(tenant)

	resp, err := ac.VaultsClient.Get(ac.Ctx, ac.ResourceGroupName, ac.KeyVaultName, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			return false, nil
		}

		if errors.As(err, &respErr) && respErr.StatusCode == 403 {
			zap.S().Warnf("permission denied to get Azure Key Vault: %s", ac.KeyVaultName)
			return false, nil
		}

		return false, err
	}

	ac.KeyVaultURI = *resp.Properties.VaultURI

	return true, nil
}

func (ac *AzureConfigure) GetAzureSecrets() (map[string][]byte, error) {
	var secrets = make(map[string][]byte)

	client, err := azsecrets.NewClient(ac.KeyVaultURI, ac.Credentials, nil)
	if err != nil {
		return nil, err
	}

	listSecrets := client.NewListSecretsPager(nil)
	if err != nil {
		return nil, err
	}

	for listSecrets.More() {
		page, err := listSecrets.NextPage(ac.Ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.StatusCode == 403 {
				zap.S().Warnf("permission denied to list Azure Key Vault secrets")
				return nil, nil
			}

			return nil, err
		}

		for _, val := range page.Value {
			name := val.ID.Name()
			version := val.ID.Version()

			secret, err := client.GetSecret(ac.Ctx, name, version, nil)
			if err != nil {
				var respErr *azcore.ResponseError
				if errors.As(err, &respErr) && respErr.StatusCode == 403 {
					zap.S().Warnf("permission denied to get of Azure Key Vault secret: %s", name)
					return nil, nil
				}

				return nil, err
			}

			secrets[secret.ID.Name()] = []byte(*secret.Value)
		}
	}

	return secrets, nil
}

func (ac *AzureConfigure) SetAzureSecret(keyName, value string) error {
	client, err := azsecrets.NewClient(ac.KeyVaultURI, ac.Credentials, nil)
	if err != nil {
		return err
	}

	params := azsecrets.SetSecretParameters{
		Value: to.Ptr(value),
	}

	secret, err := client.SetSecret(ac.Ctx, keyName, params, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 403 {
			zap.S().Warnf("permission denied to create Azure Key Vault secret: %s", keyName)
			return nil
		}

		return err
	}

	zap.S().Infof("created Azure Key Vault secret: %s, %s", keyName, *secret.ID)

	return nil
}
