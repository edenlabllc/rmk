package onprem_provider

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/melbahja/goph"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"rmk/util"
)

const (
	OnPremClusterProvider = "onprem"
)

type OnPremConfigure struct {
	KubeAPIEndpoint         string `json:"kube-api-endpoint,omitempty" yaml:"kube-api-endpoint,omitempty"`
	SSHPassword             string `json:"ssh-password,omitempty" yaml:"-"`
	SSHPrivateKeyPassphrase string `json:"ssh-private-key-passphrase,omitempty" yaml:"-"`
	SSHPrivateKey           string `json:"ssh-private-key,omitempty" yaml:"ssh-private-key,omitempty"`
	SSHUser                 string `json:"ssh-user,omitempty" yaml:"ssh-user,omitempty"`
}

func NewOnPremConfigure() *OnPremConfigure {
	return &OnPremConfigure{}
}

func getTagStructName(i interface{}, name string) error {
	if field, ok := reflect.TypeOf(i).Elem().FieldByName(name); ok {
		return fmt.Errorf("SSH option onprem-%s required", strings.TrimSuffix(field.Tag.Get("json"), ",omitempty"))
	} else {
		return fmt.Errorf("field with name %s not defined", name)
	}
}

// ValidateSSHCredentials will validate the required parameters for SSH authentication
func (op *OnPremConfigure) ValidateSSHCredentials() error {
	if len(op.KubeAPIEndpoint) == 0 {
		return getTagStructName(op, "KubeAPIEndpoint")
	}

	if len(op.SSHUser) == 0 {
		return getTagStructName(op, "SSHUser")
	}

	return nil
}

func (op *OnPremConfigure) GetSSHPrivateKeyContent() []byte {
	readPrivateKey := func(privateKey string) []byte {
		data, err := os.ReadFile(privateKey)
		if err != nil {
			zap.S().Fatal(err)
		}

		return data
	}

	if util.IsExists(op.SSHPrivateKey, true) {
		return readPrivateKey(op.SSHPrivateKey)
	} else {
		if sshPrivateKey, err := util.FindSSHKey(); err != nil {
			zap.S().Fatal(err)
		} else {
			return readPrivateKey(sshPrivateKey)
		}
	}

	return nil
}

// SSHAuth - by default method used ssh private key for Auth
func (op *OnPremConfigure) SSHAuth(sshPassword, sshAgent bool) (goph.Auth, error) {
	if len(op.SSHPassword) > 0 && sshPassword {
		return goph.Password(op.SSHPassword), nil
	}

	if goph.HasAgent() && sshAgent {
		return goph.UseAgent()
	}

	if len(op.SSHPrivateKey) > 0 && util.IsExists(op.SSHPrivateKey, true) {
		return goph.Key(op.SSHPrivateKey, op.SSHPrivateKeyPassphrase)
	} else {
		if sshPrivateKey, err := util.FindSSHKey(); err != nil {
			return nil, err
		} else {
			return goph.Key(sshPrivateKey, op.SSHPrivateKeyPassphrase)
		}
	}
}

func (op *OnPremConfigure) SSHClient(sshAuth goph.Auth) (*goph.Client, error) {
	return goph.NewConn(&goph.Config{
		User:     op.SSHUser,
		Addr:     op.KubeAPIEndpoint,
		Port:     22,
		Auth:     sshAuth,
		Callback: util.VerifySSHHost,
	})
}

func (op *OnPremConfigure) GetOnPremClusterContext(clusterName string, sshAuth goph.Auth) ([]byte, error) {
	client, err := op.SSHClient(sshAuth)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	output, err := client.Run("sudo cat /etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return nil, fmt.Errorf("kubecontext for On-Premise provider's %s cluster not found", clusterName)
	}

	config, err := clientcmd.Load(output)
	if err != nil {
		return nil, err
	}

	return op.generateUserKubeconfig(clusterName, config)
}

func (op *OnPremConfigure) generateUserKubeconfig(clusterName string, config *api.Config) ([]byte, error) {
	cfg := op.generateBaseKubeConfig(clusterName, config)
	cfg.AuthInfos = map[string]*api.AuthInfo{
		op.getKubeConfigUserName(clusterName): {
			ClientCertificateData: config.AuthInfos["default"].ClientCertificateData,
			ClientKeyData:         config.AuthInfos["default"].ClientKeyData,
		},
	}

	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize kubeconfig to YAML: %w", err)
	}

	return out, nil
}

func (op *OnPremConfigure) generateBaseKubeConfig(clusterName string, config *api.Config) *api.Config {
	return &api.Config{
		APIVersion: api.SchemeGroupVersion.Version,
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   fmt.Sprintf("https://%s:6443", op.KubeAPIEndpoint),
				CertificateAuthorityData: config.Clusters["default"].CertificateAuthorityData,
			},
		},
		Contexts: map[string]*api.Context{
			clusterName: {
				Cluster:  clusterName,
				AuthInfo: op.getKubeConfigUserName(clusterName),
			},
		},
		CurrentContext: clusterName,
	}
}

func (op *OnPremConfigure) getKubeConfigUserName(clusterName string) string {
	return fmt.Sprintf("%s-capi-admin", clusterName)
}
