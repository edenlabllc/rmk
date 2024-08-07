package util

const (
	AWSClusterProvider         = "aws"
	AzureClusterProvider       = "azure"
	AWSECRHost                 = "288509344804.dkr.ecr.eu-north-1.amazonaws.com"
	AWSECRRegion               = "eu-north-1"
	AWSECRUserName             = "AWS"
	GitSSHPrivateKey           = ".ssh/id_rsa"
	GlobalsFileName            = "globals.yaml.gotmpl"
	HelmfileFileName           = "helmfile.yaml"
	HelmfileGoTmplName         = HelmfileFileName + ".gotmpl"
	HelpFlagFull               = "--help"
	K3DConfigPrefix            = "k3d"
	LocalClusterProvider       = K3DConfigPrefix
	RMKBin                     = "rmk"
	RMKBucketName              = "edenlabllc-rmk"
	RMKBucketRegion            = "eu-north-1"
	RMKConfig                  = "config"
	RMKDir                     = ".rmk"
	RMKSymLinkPath             = "/usr/local/bin/rmk"
	RMKToolsDir                = "tools"
	ReadmeFileName             = "README.md"
	RegionException            = "us-east-1"
	ReleasesFileName           = "releases.yaml"
	SecretSpecFile             = ".spec.yaml.gotmpl"
	SopsAgeKeyExt              = ".txt"
	SopsAgeKeyFile             = ".keys.txt"
	SopsRootName               = "sops-age-keys"
	SopsConfigFile             = ".sops.yaml"
	TenantBucketKey            = "tf.tfstate"
	TenantBucketName           = "cloud-provisioner"
	TenantDDBTablePrefix       = "cloud-provisioner-state-locks"
	TenantDomainSuffix         = ".edenlab.dev"
	TenantProjectCodeOwners    = "docs/CODEOWNERS"
	TenantProjectDIR           = ".PROJECT"
	TenantProjectFile          = "project.yaml"
	TenantProjectGitIgn        = ".gitignore"
	TenantValuesDIR            = "etc"
	TerraformVarHostedZoneName = "hosted_zone_name"
	TerraformVarsExt           = "tfvars"
	TerraformVarsFile          = "variables.auto." + TerraformVarsExt
	TerraformVarsPrefix        = "rmk_"
	TerraformWGFile            = "worker-groups.auto." + TerraformVarsExt
	ToolsBinDir                = "bin"
	ToolsTmpDir                = "tmp"
	ToolsVersionDir            = "version"

	ConfigNotInitializedErrorText = "RMK config not initialized, " +
		"please run command 'rmk config init' with specific parameters"
	// UnknownErrorText standard text for unknown errors
	UnknownErrorText = "unknown error when calling %s"
	//HelmPluginExist HelmSecretsIsNotEncrypted HelmSecretsAlreadyEncrypted - exception err text matching
	HelmPluginExist             = "Error: plugin already exists"
	HelmSecretsIsNotEncrypted   = "File is not encrypted: "
	HelmSecretsAlreadyEncrypted = "Already encrypted: "
	HelmSecretsOutputPrefix     = "[helm-secrets] "
	HelmSecretsError            = "Error: plugin \"secrets\" exited with error"

	CompletionZshDescription = `Run the following scripts to enable Zsh completion:

rmk completion zsh > ~/.local/bin/rmk-completion-zsh.sh
chmod +x ~/.local/bin/rmk-completion-zsh.sh
echo "PROG=rmk source ~/.local/bin/rmk-completion-zsh.sh" >> ~/.zshrc`

	// CompletionZshScript https://github.com/urfave/cli/blob/v2.27.1/autocomplete/zsh_autocomplete
	CompletionZshScript = `#compdef $PROG

_cli_zsh_autocomplete() {
  local -a opts
  local cur
  cur=${words[-1]}
  if [[ "$cur" == "-"* ]]; then
    opts=("${(@f)$(${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
  else
    opts=("${(@f)$(${words[@]:0:#words[@]-1} --generate-bash-completion)}")
  fi

  if [[ "${opts[1]}" != "" ]]; then
    _describe 'values' opts
  else
    _files
  fi
}

compdef _cli_zsh_autocomplete $PROG`
)
