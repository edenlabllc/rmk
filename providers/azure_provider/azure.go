package azure_provider

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	"golang.org/x/net/context"

	"rmk/util"
)

const (
	AzureClusterProvider = "azure"
	AzureHomeDir         = "." + AzureClusterProvider
	AzurePrefix          = "service-principal-credentials_"
)

type AzureRawServicePrincipal struct {
	AppId       string `json:"appId"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	Tenant      string `json:"tenant"`
}

type AzureClient struct {
	ARMClient *armcontainerservice.ManagedClustersClient `json:"-" yaml:"-"`
	Ctx       context.Context                            `json:"-" yaml:"-"`
}

type AzureConfigure struct {
	AzureClient    `json:"-" yaml:"-"`
	ClientID       string `json:"client-id,omitempty" yaml:"-"`
	ClientSecret   string `json:"client-secret,omitempty" yaml:"-"`
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

func (ac *AzureConfigure) ValidateSPCredentials() error {
	if len(ac.ClientID) == 0 {
		return getTagStructName(ac, "ClientID")
	}

	if len(ac.ClientSecret) == 0 {
		return getTagStructName(ac, "ClientSecret")
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

func (ac *AzureConfigure) NewAzureManagedClustersClient(ctx context.Context, fileName string) error {
	if err := ac.ReadSPCredentials(fileName); err != nil {
		return err
	}

	cred, err := azidentity.NewClientSecretCredential(ac.TenantID, ac.ClientID, ac.ClientSecret, nil)
	if err != nil {
		return err
	}

	factory, err := armcontainerservice.NewClientFactory(ac.SubscriptionID, cred, nil)
	if err != nil {
		return err
	}

	ac.ARMClient = factory.NewManagedClustersClient()
	ac.Ctx = ctx

	return nil
}

func (ac *AzureConfigure) GetAzureClusterContext(groupName, clusterName string) ([]byte, error) {
	credentials, err := ac.ARMClient.ListClusterAdminCredentials(ac.Ctx, groupName, clusterName, nil)
	if err != nil {
		return nil, err
	}

	if len(credentials.CredentialResults.Kubeconfigs) == 1 {
		return credentials.CredentialResults.Kubeconfigs[0].Value, nil
	}

	return nil, fmt.Errorf("context for cluster %s not found", clusterName)
}
