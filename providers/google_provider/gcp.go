package google_provider

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"
	"cloud.google.com/go/compute/apiv1/computepb"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2/apierror"
	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1beta1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"rmk/util"
)

const (
	GoogleClusterProvider   = "gcp"
	GoogleCredentialsPrefix = "gcp-credentials-"
	GoogleHomeDir           = ".config/gcloud"

	// List of APIs errors
	apiErrorAlreadyExists     = "alreadyExists"
	apiErrorNotFound          = "notFound"
	gRPCErrorAlreadyExists    = "AlreadyExists"
	gRPCErrorPermissionDenied = "PermissionDenied"
)

type GCPConfigure struct {
	AppCredentials     *auth.Credentials `yaml:"-"`
	AppCredentialsPath string            `yaml:"app-credentials-path"`
	Ctx                context.Context   `yaml:"-"`
	ProjectID          string            `yaml:"project-id"`
}

func NewGCPConfigure(ctx context.Context, appCredentialsPath string) *GCPConfigure {
	return &GCPConfigure{Ctx: ctx, AppCredentialsPath: appCredentialsPath}
}

func (gcp *GCPConfigure) ReadSACredentials() error {
	if !util.IsExists(gcp.AppCredentialsPath, true) {
		return fmt.Errorf("google application credentials JSON file for GCP not found")
	}

	data, err := os.ReadFile(gcp.AppCredentialsPath)
	if err != nil {
		return err
	}

	gcp.AppCredentials, err = credentials.DetectDefault(&credentials.DetectOptions{CredentialsJSON: data})
	if err != nil {
		return err
	}

	gcp.ProjectID, err = gcp.AppCredentials.ProjectID(gcp.Ctx)
	if err != nil {
		return err
	}

	return nil
}

func (gcp *GCPConfigure) CopySACredentials(fileSuffix string) error {
	if err := os.MkdirAll(util.GetHomePath(GoogleHomeDir), 0755); err != nil {
		return err
	}

	gcp.AppCredentialsPath = util.GetHomePath(GoogleHomeDir, GoogleCredentialsPrefix+fileSuffix+".json")

	return os.WriteFile(util.GetHomePath(GoogleHomeDir, GoogleCredentialsPrefix+fileSuffix+".json"),
		gcp.AppCredentials.JSON(), 0644)
}

func (gcp *GCPConfigure) GetGCPSecrets(tenant string) (map[string][]byte, error) {
	var secrets = make(map[string][]byte)

	if err := gcp.ReadSACredentials(); err != nil {
		return nil, err
	}

	client, err := secretmanager.NewClient(gcp.Ctx, option.WithCredentialsJSON(gcp.AppCredentials.JSON()))
	if err != nil {
		return nil, err
	}

	defer client.Close()

	listReq := &secretmanagerpb.ListSecretsRequest{
		Parent: fmt.Sprintf("projects/%s", gcp.ProjectID),
		Filter: "labels.resource-group=" + tenant + "-" + util.SopsRootName,
	}

	it := client.ListSecrets(gcp.Ctx, listReq)
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			var respError *apierror.APIError
			if errors.As(err, &respError) && respError.GRPCStatus().Code().String() == gRPCErrorPermissionDenied {
				zap.S().Warnf("permission denied to list GCP Secrets Manager secrets")
				return nil, nil
			} else {
				return nil, err
			}
		}

		accReq := &secretmanagerpb.AccessSecretVersionRequest{Name: resp.Name + "/versions/latest"}
		result, err := client.AccessSecretVersion(gcp.Ctx, accReq)
		if err != nil {
			var respError *apierror.APIError
			if errors.As(err, &respError) && respError.GRPCStatus().Code().String() == gRPCErrorPermissionDenied {
				zap.S().Warnf("permission denied to get access to GCP Secrets Manager secrets values")
				return nil, nil
			} else {
				return nil, err
			}
		}

		crc32c := crc32.MakeTable(crc32.Castagnoli)
		checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
		if checksum != *result.Payload.DataCrc32C {
			return nil, fmt.Errorf("data corruption detected for GCP Secrets Manager secrets value: %s", resp.Name)
		}

		secrets[filepath.Base(resp.Name)] = result.Payload.Data
	}

	return secrets, nil
}

func (gcp *GCPConfigure) SetGCPSecret(tenant, region, keyName string, value []byte) error {
	if err := gcp.ReadSACredentials(); err != nil {
		return err
	}

	client, err := secretmanager.NewClient(gcp.Ctx, option.WithCredentialsJSON(gcp.AppCredentials.JSON()))
	if err != nil {
		return err
	}

	defer client.Close()

	secretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", gcp.ProjectID),
		SecretId: keyName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_UserManaged_{
					UserManaged: &secretmanagerpb.Replication_UserManaged{
						Replicas: []*secretmanagerpb.Replication_UserManaged_Replica{{Location: region}},
					},
				},
			},
			Labels: map[string]string{"resource-group": tenant + "-" + util.SopsRootName},
		},
	}

	_, err = client.CreateSecret(gcp.Ctx, secretReq)
	if err != nil {
		var respError *apierror.APIError
		if errors.As(err, &respError) && respError.GRPCStatus().Code().String() == gRPCErrorPermissionDenied {
			zap.S().Warnf("permission denied to create GCP Secrets Manager secret: %s", keyName)
		} else if respError.GRPCStatus().Code().String() != gRPCErrorAlreadyExists {
			return err
		}
	}

	secretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent:  fmt.Sprintf("projects/%s/secrets/%s", gcp.ProjectID, keyName),
		Payload: &secretmanagerpb.SecretPayload{Data: value},
	}

	version, err := client.AddSecretVersion(gcp.Ctx, secretVersionReq)
	if err != nil {
		var respError *apierror.APIError
		if errors.As(err, &respError) && respError.GRPCStatus().Code().String() == gRPCErrorPermissionDenied {
			zap.S().Warnf("permission denied to add GCP Secrets Manager secret %s value", keyName)
		} else {
			return err
		}
	}

	zap.S().Infof("created GCP Secrets Manager secret: %s, %s", keyName, version.Name)

	return nil
}

func (gcp *GCPConfigure) CreateGCPCloudNATGateway(region string) error {
	if err := gcp.ReadSACredentials(); err != nil {
		return err
	}

	client, err := compute.NewService(gcp.Ctx, option.WithCredentialsJSON(gcp.AppCredentials.JSON()))
	if err != nil {
		return err
	}

	routerNat := &compute.RouterNat{
		AutoNetworkTier:               computepb.RouterNat_STANDARD.String(),
		EndpointTypes:                 []string{computepb.RouterNat_ENDPOINT_TYPE_VM.String()},
		Name:                          "default-nat-" + region,
		NatIpAllocateOption:           computepb.RouterNat_AUTO_ONLY.String(),
		SourceSubnetworkIpRangesToNat: computepb.RouterNat_ALL_SUBNETWORKS_ALL_IP_RANGES.String(),
		Type:                          computepb.RouterNat_PUBLIC.String(),
	}

	router := &compute.Router{
		Name:    "default-router-" + region,
		Nats:    []*compute.RouterNat{routerNat},
		Network: "projects/" + gcp.ProjectID + "/global/networks/default",
	}

	_, err = client.Routers.Insert(gcp.ProjectID, region, router).Context(gcp.Ctx).Do()
	if err != nil {
		var respError *googleapi.Error
		if errors.As(err, &respError) && respError.Code == http.StatusConflict && respError.Errors[0].Reason == apiErrorAlreadyExists {
			zap.S().Infof("GCP router %s with router NAT %s already exists for region %s",
				router.Name, routerNat.Name, region)
			return nil
		}

		return err
	}

	zap.S().Infof("created GCP router %s with router NAT %s", router.Name, routerNat.Name)

	return nil
}

func (gcp *GCPConfigure) DeleteGCPCloudNATGateway(region string) error {
	if err := gcp.ReadSACredentials(); err != nil {
		return err
	}

	containerClient, err := container.NewService(gcp.Ctx, option.WithCredentialsJSON(gcp.AppCredentials.JSON()))
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", gcp.ProjectID, region)
	resp, err := containerClient.Projects.Locations.Clusters.List(parent).Context(gcp.Ctx).Do()
	if err != nil {
		return err
	}

	if len(resp.Clusters) > 0 {
		zap.S().Infof("skipped deleting GCP router %s because there are %d clusters in %s region",
			"default-router-"+region, len(resp.Clusters), region)
		return nil
	}

	computeClient, err := compute.NewService(gcp.Ctx, option.WithCredentialsJSON(gcp.AppCredentials.JSON()))
	if err != nil {
		return err
	}

	_, err = computeClient.Routers.Delete(gcp.ProjectID, region, "default-router-"+region).Context(gcp.Ctx).Do()
	if err != nil {
		var respError *googleapi.Error
		if errors.As(err, &respError) && respError.Code == http.StatusNotFound && respError.Errors[0].Reason == apiErrorNotFound {
			return nil
		}

		return err
	}

	zap.S().Infof("deleted GCP router %s for region %s", "default-router-"+region, region)

	return nil
}

func (gcp *GCPConfigure) GetGCPClusterContext(clusterName string) ([]byte, error) {
	var cluster *container.Cluster

	if err := gcp.ReadSACredentials(); err != nil {
		return nil, err
	}

	client, err := container.NewService(gcp.Ctx, option.WithCredentialsJSON(gcp.AppCredentials.JSON()))
	if err != nil {
		return nil, err
	}

	resp, err := client.Projects.Zones.Clusters.List(gcp.ProjectID, "-").Context(gcp.Ctx).Do()
	if err != nil {
		return nil, err
	}

	for _, val := range resp.Clusters {
		if val.Name == clusterName {
			cluster = val
			break
		}
	}

	if cluster != nil {
		return gcp.generateUserKubeconfig(cluster)
	}

	return nil, fmt.Errorf("kubecontext for %s provider's %s cluster not found",
		strings.ToUpper(GoogleClusterProvider), clusterName)
}

func (gcp *GCPConfigure) generateUserKubeconfig(cluster *container.Cluster) ([]byte, error) {
	var execEnvVars []api.ExecEnvVar

	userName := gcp.getKubeConfigUserName(cluster.Name)
	cfg, err := gcp.generateBaseKubeConfig(cluster)
	if err != nil {
		return nil, fmt.Errorf("creating base kubeconfig: %w", err)
	}

	execEnvVars = append(execEnvVars,
		api.ExecEnvVar{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcp.AppCredentialsPath},
	)

	// Version v1alpha1 was removed in Kubernetes v1.23.
	// Version v1 was released in Kubernetes v1.23.
	// Version v1beta1 was selected as it has the widest range of support
	// This should be changed to v1 once EKS no longer supports Kubernetes <v1.23
	execConfig := &api.ExecConfig{
		APIVersion:         "client.authentication.k8s.io/v1beta1",
		Args:               []string{},
		Command:            "gke-auth-plugin",
		Env:                execEnvVars,
		ProvideClusterInfo: true,
		InteractiveMode:    api.NeverExecInteractiveMode,
	}

	cfg.AuthInfos = map[string]*api.AuthInfo{
		userName: {
			Exec: execConfig,
		},
	}

	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize kubeconfig to YAML: %w", err)
	}

	return out, nil
}

func (gcp *GCPConfigure) generateBaseKubeConfig(cluster *container.Cluster) (*api.Config, error) {
	userName := gcp.getKubeConfigUserName(cluster.Name)

	certData, err := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return nil, fmt.Errorf("decoding cluster CA cert: %w", err)
	}

	cfg := &api.Config{
		APIVersion: api.SchemeGroupVersion.Version,
		Clusters: map[string]*api.Cluster{
			cluster.Name: {
				Server:                   "https://" + cluster.Endpoint,
				CertificateAuthorityData: certData,
			},
		},
		Contexts: map[string]*api.Context{
			cluster.Name: {
				Cluster:  cluster.Name,
				AuthInfo: userName,
			},
		},
		CurrentContext: cluster.Name,
	}

	return cfg, nil
}

func (gcp *GCPConfigure) getKubeConfigUserName(clusterName string) string {
	return fmt.Sprintf("%s-capi-admin", clusterName)
}
