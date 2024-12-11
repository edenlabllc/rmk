package google_provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"
	container "google.golang.org/api/container/v1beta1"
	"google.golang.org/api/option"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"rmk/util"
)

const (
	GoogleClusterProvider = "gcp"
	GoogleHomeDir         = ".config/gcloud"
	GooglePrefix          = "gcp-credentials-"
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

	gcp.AppCredentialsPath = util.GetHomePath(GoogleHomeDir, GooglePrefix+fileSuffix+".json")

	return os.WriteFile(util.GetHomePath(GoogleHomeDir, GooglePrefix+fileSuffix+".json"),
		gcp.AppCredentials.JSON(), 0644)
}

//func (gcp *GCPConfigure) CreateGateway() error {
//	if err := gcp.ReadSACredentials(); err != nil {
//		return err
//	}
//
//	client, err := compute.NewService(gcp.Ctx, option.WithAuthCredentials(gcp.AppCredentials))
//	if err != nil {
//		return err
//	}
//
//	client.BasePath = compute.CloudPlatformScope + "/v1"
//
//	fmt.Printf("%#v\n", *client)
//
//	req := client.Routers.List(gcp.ProjectID, "europe-central2")
//	if err := req.Pages(gcp.Ctx, func(page *compute.RouterList) error {
//		for _, router := range page.Items {
//			// process each `router` resource:
//			fmt.Printf("%#v\n", router)
//			// NAT Gateways are found in router.nats
//		}
//		return nil
//	}); err != nil {
//		log.Fatal(err)
//	}
//
//	return nil
//}

func (gcp *GCPConfigure) GetGCPClusterContext(clusterName string) ([]byte, error) {
	var cluster *container.Cluster

	if err := gcp.ReadSACredentials(); err != nil {
		return nil, err
	}

	client, err := container.NewService(gcp.Ctx, option.WithAuthCredentials(gcp.AppCredentials))
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
