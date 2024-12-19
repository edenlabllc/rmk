package cmd

const (
	codeOwners = `# These owners will be the default owners for everything in
# the repo and will be requested for review when someone opens a pull request.
`

	escape      = "`"
	escapeOpen  = `{{` + "`"
	escapeClose = "`" + `}}`

	gitignore = `**.dec
*/cluster/k3d/
dist/
.DS_Store
.PROJECT/
.deps/
.env
.helmfile/
.idea/
sops-age-keys/
`

	globals = `# This file defines the globals configuration list for values different releases, 
# is located in the environment directory of a specific releases scope: etc/<releases scope>/<environment>/globals.yaml.gotmpl.
# This configuration allows you to use the same values in value files of different releases only in the same scope.

# configs - enumeration of configurations divided into sets related to Kubernetes ConfigMaps
configs: {}

# envs - enumeration of environment variables divided into sets related to Kubernetes environment variables for container
envs: {}

# hooks - enumeration of environment variables divided into sets related to helmfile hooks arguments
hooks: {}
`

	helmfileEnvironments = `  {{ .EnvironmentName }}:
    missingFileHandler: Warn
    values:
      - etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/globals.yaml
      - etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/globals.yaml.gotmpl
      - etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/releases.yaml
      ` + escapeOpen + `{{- if eq (env "K3D_CLUSTER") "true" }}` + escapeClose + `
      - etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/values/k3d/releases.yaml
      ` + escapeOpen + `{{- end }}` + escapeClose + `
      - ` + escapeOpen + `{{ requiredEnv "PWD" }}` + escapeClose + `/etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/globals.yaml
      - ` + escapeOpen + `{{ requiredEnv "PWD" }}` + escapeClose + `/etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/globals.yaml.gotmpl
      - ` + escapeOpen + `{{ requiredEnv "PWD" }}` + escapeClose + `/etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/releases.yaml
      ` + escapeOpen + `{{- if eq (env "K3D_CLUSTER") "true" }}` + escapeClose + `
      - ` + escapeOpen + `{{ requiredEnv "PWD" }}` + escapeClose + `/etc/{{ .TenantName }}/` + escapeOpen + `{{ .Environment.Name }}` + escapeClose + `/values/k3d/releases.yaml
      ` + escapeOpen + `{{- end }}` + escapeClose
	helmDefaults = `---
helmDefaults:
  wait: true
  waitForJobs: true
  timeout: 3600
`

	helmfiles = `# The set of paths for inherited helmfiles is controlled through the version.yaml file using rmk.
# DO NOT EDIT field helmfiles values.
helmfiles: ` + escapeOpen + `{{ env "HELMFILE_` + escapeClose + `{{ .TenantNameEnvStyle }}` + escapeOpen + `_PATHS" }}` + escapeClose + `
`

	helmfileMissingFileHandler = `missingFileHandler: Warn
`

	helmfileCommonLabels = `commonLabels:
  scope: {{ .TenantName }}
  namespace: {{ .TenantName }}
  bin: ` + escapeOpen + `{{ env "HELMFILE_` + escapeClose + `{{ .TenantNameEnvStyle }}` + escapeOpen + `_HOOKS_DIR" }}/bin` + escapeClose + `
  repo: core-charts
  appChartVersion: 1.6.0
  host: ` + escapeOpen + `{{ env "ROOT_DOMAIN" }}` + escapeClose + `
`

	helmfileTemplates = `templates:
  release:
    createNamespace: true
    labels:
      app: "{{` + escape + `{{ .Release.Name }}` + escape + `}}"
    missingFileHandler: Warn
    values:
      - etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}-ingress-route.yaml.gotmpl
      - etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml.gotmpl
      - etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/k3d/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}-ingress-route.yaml.gotmpl
      - etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/k3d/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml.gotmpl
      - etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/k3d/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml
      {{- end }}
      - {{ requiredEnv "PWD" }}/etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}-ingress-route.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml
      {{- if eq (env "K3D_CLUSTER") "true" }}
      - {{ requiredEnv "PWD" }}/etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/k3d/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}-ingress-route.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/k3d/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml.gotmpl
      - {{ requiredEnv "PWD" }}/etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/values/k3d/values/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml
      {{- end }}
    secrets:
      - {{ requiredEnv "PWD" }}/etc/{{` + escape + `{{ .Release.Labels.scope }}` + escape + `}}/{{` + escape + `{{ .Environment.Name }}` + escape + `}}/secrets/{{` + escape + `{{ .Release.Name }}` + escape + `}}.yaml
`

	helmfileReleases = `releases:
  # TODO: Releases from group 1 are needed to deploy K3D clusters. 
  # TODO: If you do not inherit upstream repositories, you can leave these releases as is, 
  # TODO: or make sure that upstream repositories do not have the same releases to avoid conflicts.
  # Group 1
  - name: k3d-cluster
    namespace: kube-system
    chart: "{{"{{` + escape + `{{ .Release.Labels.repo }}` + escape + `}}"}}/k3d-cluster"
    version: 0.1.0
    labels:
      cluster: k3d
    installed: ` + escapeOpen + `{{ eq (env "K3D_CLUSTER" | default "false") "true" }}` + escapeClose + `
    inherit:
      - template: release

  # TODO: It is recommended to adapt this example considering security, performance and configuration management 
  # TODO: requirements specific to your application or infrastructure.  
  # Group 2
  - name: {{ .TenantName }}-app
    namespace: {{ .TenantName }}
    chart: "{{"{{` + escape + `{{ .Release.Labels.repo }}` + escape + `}}"}}/app"
    version: "{{"{{` + escape + `{{ .Release.Labels.appChartVersion }}` + escape + `}}"}}"
    installed: ` + escapeOpen + `{{ .Values | get (print "` + escapeClose + `{{ .TenantName }}-app` + escapeOpen + `" ".enabled") false }}` + escapeClose + `
    inherit:
      - template: release
    ` + escapeOpen + `{{- with .Values | get (print "hooks." "` + escapeClose + `{{ .TenantName }}-app` + escapeOpen + `") "" }}` + escapeClose + `
    hooks: # common-postuninstall-hook is needed only for the first service in namespace
      ` + escapeOpen + `{{- range $name, $hook := . }}` + escapeClose + `
      - name: ` + escapeOpen + `{{ $name }}` + escapeClose + `
        ` + escapeOpen + `{{- toYaml $hook | nindent 8 }}` + escapeClose + `
      ` + escapeOpen + `{{- end }}` + escapeClose + `
    ` + escapeOpen + `{{- end }}` + escapeClose + `
`

	readmeFile = `# {{ .RepoName }}

## Description

The repository designed for the rapid setup and deployment of the infrastructure required for the {{ .TenantName }} project. 
This project includes scripts, configurations, and instructions to automate the deployment of necessary services and dependencies.

## Getting Started

To get started with {{ .RepoName }}, ensure you have all the necessary tools and dependencies installed. 
Detailed information about requirements and installation instructions can be found in the [Requirements](#requirements) section.

### Requirements

- Git 
- GitHub PAT to access the repositories listed in the ` + "`dependencies`" + ` section of ` + "`project.yaml`" + `
- Note: K3D v5.x.x requires at least Docker v20.10.5 (runc >= v1.0.0-rc93) to work properly
- [RMK CLI](https://edenlabllc.github.io/rmk/latest)

### GitLab flow strategy

This repository uses the Environment branches with GitLab flow approach,
where each stable or ephemeral branches is a separate environment with its own Kubernetes cluster.

Release schema:
` + "```" + `text
develop ------> staging ------> production
   \            /     \           /
  release/vN.N.N-rc  release/vN.N.N
` + "```" + `

### Generating project structure

> Note: The generated project structure using the RMK tools is mandatory and is required for the interaction of the RMK with the code base. 
> All generated files have example content and can be supplemented according to project requirements.

After generating the project structure, a set of files is generated for the main project scope etc/{{ .TenantName }}
to demonstrate an example of configuring the {{ .TenantName }}-app release.
This example shows how the following options are configured and interact with each other:

- etc/{{ .TenantName }}/\<environment>/secrets/.spec.yaml.gotmpl
- etc/{{ .TenantName }}/\<environment>/values/rmk-test-app.yaml.gotmpl
- etc/{{ .TenantName }}/\<environment>/globals.yaml.gotmpl
- etc/{{ .TenantName }}/\<environment>/releases.yaml
- helmfile.yaml.gotmpl

{{ if .Dependencies }}
#### Inherited repositories
{{ range .Dependencies }}
- **{{ . }}**
{{ end }}
{{- end }}
{{- if .Scopes }}
#### Available scopes of variables
{{ range .Scopes }}
- **{{ . }}**
{{ end }}
{{- end }}
### Basic RMK commands for project management

#### Project generate

` + "```" + `shell
rmk project generate \
    --environment=develop.root-domain=localhost \
    --owners=user \
    --scopes={{ .TenantName }}
` + "```" + `

#### Initialization configuration

` + "```" + `shell
rmk config init
` + "```" + `

#### Create K3D cluster

` + "```" + `shell
rmk cluster k3d create
` + "```" + `

#### Release sync

` + "```" + `shell
rmk release sync
` + "```" + `

> Note: A complete list of RMK commands and capabilities can be found at the [link](https://edenlabllc.github.io/rmk/latest)
`

	releasesFile = `# This file defines the release list, is located in the environment directory
# of a specific releases scope: etc/<releases scope>/<environment>/releases.yaml.
# The absence of this file in the environment directory means that the entire list of releases will be installed.
# Set false to uninstall this release on sync.
`

	secretSpecFile = `# This file defines the generation-rules list for secrets, is located in the secrets directory
# of a specific environment of a specific releases scope: etc/<releases scope>/<environment>/secrets/.spec.yaml.gotmpl.
# This template allows you to generate a new set of secrets, thereby rotating existing secrets.

generation-rules: []
`

	sopsConfigFile = `creation_rules: []
`

	tenantGlobals = `# This file defines the globals configuration list for values different releases, 
# is located in the environment directory of a specific releases scope: etc/<releases scope>/<environment>/globals.yaml.gotmpl.
# This configuration allows you to use the same values in value files of different releases only in the same scope.

# TODO: It is recommended to adapt this example considering security, performance and configuration management 
# TODO: requirements specific to your application or infrastructure.
# configs - enumeration of configurations divided into sets related to Kubernetes ConfigMaps
configs:
  containerRegistryAuth:
    imagePullSecrets:
      - name: container-registry-credentials
  linkerd:
    # enable/disable linkerd sidecar injection: enabled|disabled
    inject: enabled

# envs - enumeration of environment variables divided into sets related to Kubernetes environment variables for container
envs:
  logger:
    LOG_LEVEL: "info"

# hooks - enumeration of environment variables divided into sets related to helmfile hooks arguments
hooks:
  {{ .TenantName }}-app:
    common-postuninstall-hook:
      events:
        - postuninstall
      showlogs: true
      command: "{{"{{` + escape + `{{ .Release.Labels.bin }}` + escape + `}}"}}/common-postuninstall-hook.sh"
      args:
        - "{{"{{` + escape + `{{ .Release.Namespace }}` + escape + `}}"}}"
`

	tenantReleasesFile = releasesFile + `
# TODO: It is recommended to adapt this example considering security, performance and configuration management 
# TODO: requirements specific to your application or infrastructure.
{{ .TenantName }}-app:
  enabled: true
  image:
    repository: nginx
    tag: latest
`

	tenantSecretSpecFile = `# This file defines the generation-rules list for secrets, is located in the secrets directory
# of a specific environment of a specific releases scope: etc/<releases scope>/<environment>/secrets/.spec.yaml.gotmpl.
# This template allows you to generate a new set of secrets, thereby rotating existing secrets.

# TODO: It is recommended to adapt this example considering security, performance and configuration management 
# TODO: requirements specific to your application or infrastructure.
generation-rules:
  - name: {{ .TenantName }}-app
    template: |
      envSecret:
        USERNAME: user
        PASSWORD: ` + escapeOpen + `{{ randAlphaNum 16 }}` + escapeClose + `
`

	tenantValuesExample = `# This value file is an introductory example configuration for running Nginx in Kubernetes via RMK.
# It combines several key components such as Deployment, Service, ConfigMap, Secrets and other options.
# The value file is intended to demonstrate the basic capabilities of RMK in deploying releases 
# and should not be used as is in a production environment.

# TODO: It is recommended to adapt this example considering security, performance and configuration management 
# TODO: requirements specific to your application or infrastructure.
replicaCount: 1
image:
  repository: '{{ .Values | get (printf "%s.image.repository" .Release.Name) }}'
  tag: '{{ .Values | get (printf "%s.image.tag" .Release.Name) }}'
imagePullSecrets:
  {{ .Values | get "configs.containerRegistryAuth.imagePullSecrets" list | toYaml | nindent 2 }}
jaegerAgent:
  enabled: false
podAnnotations:
  linkerd.io/inject: {{ "" | default .Values.configs.linkerd.inject }}
envFrom: []
#  - secretRef:
#      name: app-secret
command: []
#  - /app/server
env:
  #################################################################################
  # Logger settings
  #################################################################################
  # Log severity
  {{ .Values | get "envs.logger" | toYaml | nindent 2 }}
volumeMounts:
  - name: config
    mountPath: /etc/nginx/conf.d
volumesConfigMap:
  enable: true
  volumes:
    - name: config
      files:
        - name: default.conf
          data: |
            server {
                listen       80;
                server_name  localhost;

                location / {
                    root   /usr/share/nginx/html;
                    index  index.html index.htm;
                }

                error_page   500 502 503 504  /50x.html;
                location = /50x.html {
                    root   /usr/share/nginx/html;
                }
            }
ports:
  - name: http
    containerPort: 80
service:
  enable: true
  type: ClusterIP
  ports:
    - port: 80
      name: http
readinessProbe:
  timeoutSeconds: 10
  periodSeconds: 10
  failureThreshold: 3
  httpGet:
    path: /
    port: 80
livenessProbe:
  initialDelaySeconds: 60
  timeoutSeconds: 10
  periodSeconds: 10
  failureThreshold: 3
  httpGet:
    path: /
    port: 80
resources:
  limits:
    cpu: 100m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
`
)
