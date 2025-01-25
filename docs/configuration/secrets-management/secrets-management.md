# Secrets management

## Overview

RMK utilizes [SOPS](https://github.com/mozilla/sops) and [Age](https://github.com/mozilla/sops#encrypting-using-age)
for secrets management, ensuring secure encryption, storage, and access to sensitive data. The tool ensures
seamless and automated secret management, reducing manual effort while maintaining security best practices.

The functionality in RMK is divided into two key areas:

1. **Working with secrets keys** – Managing encryption keys used for encrypting and decrypting secrets.
2. **Working with secrets files** – Handling encrypted YAML files that store sensitive configuration data.

### Secrets keys

This area focuses on integration with cloud providers, ensuring secure storage, retrieval, and local access to secrets.
Once the keys are generated, RMK automatically provisions all necessary cloud resources, securely stores the keys in the
provider’s Secrets Manager service, and downloads them to the local machine upon first use.

For each supported cloud provider, RMK integrates with the respective secrets management service:

- **AWS** – Integrates with [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/).
- **Azure** – Integrates with [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault).
- **GCP** – Integrates
  with [Google Cloud Secret Manager](https://cloud.google.com/security/products/secret-manager?hl=en).

Locally, secret keys are stored in a secure file within the user's home directory:

```shell
${HOME}/.rmk/sops-age-keys/<project_name>/.keys.txt
```

This ensures that each project has its own isolated key storage.

The secrets key file will look like this:

```shell
# created: 2025-01-23T20:47:30+01:00
# public key: age1rq0gx9zuwphw8kjx6ams84rgysqk5kdmhnysxs28r0x955xnzsdsslgtn0
AGE-SECRET-KEY-15K8LZB3MT0QWJ4N7X90H2A9C5L6E7FYZ3XGKP1DRN8SWV2QXT90H2A9C5L
```

### Secrets files

This area focuses on integration with Helmfile, Helm, and Kubernetes, ensuring automated and seamless secrets management
throughout the deployment process.

Normally, the secrets files can be committed to Git because they are encrypted with SOPS Age keys
using [symmetric-key algorithms](https://en.wikipedia.org/wiki/Symmetric-key_algorithm).

In a project repository, all secrets files are stored in the `etc/<scope>/<env>/secrets/` directories.
For example:

```shell
etc/deps/develop/secrets/postgres.yaml
etc/rmk-test/develop/secrets/app.yaml
```

> Secrets files are never inherited by projects, in contrast to the Helmfile values. Each project should have its own
> unique set of secrets for all deployed releases.

Below is an example of an encrypted secrets file, where sensitive values are encrypted
using [AES256-GCM](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard):

```yaml
username: ENC[AES256_GCM,data:A0jb1wU=,iv:RM8V1IOHvCrBv7f9f/GKobaBYyjcX9jcNQp6XPopNcU=,tag:T79VY3/yqlIffdbvYDwukQ==,type:str]
password: ENC[AES256_GCM,data:Kjo5hDSb+VmhdLLuq48oVg==,iv:5wpJBsiA5B82RRaguW8/TcKgGYbiZhihdIhXnPwyRG8=,tag:yQ5Chi949jBB1cSaFDVlOQ==,type:str]
sops:
  kms: [ ]
  gcp_kms: [ ]
  azure_kv: [ ]
  hc_vault: [ ]
  age:
    - recipient: age1rq0gx9zuwphw8kjx6ams84rgysqk5kdmhnysxs28r0x955xnzsdsslgtn0
      enc: |
        -----BEGIN AGE ENCRYPTED FILE-----
        YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSA1ZnBDR1pLNWt3TGVOVDlI
        TGdNOEdDZmEzQjFzaWZuWXVqN0RZMWxBcjN3CnZvdDRtSDZIaTlyenF4bG9wRzg3
        ZURpTGUrd3JIZGV6clpwTkVKeDT5ekkKLS0tIHUwMGVvWnFDY2FQWm8rcmg4Wnl3
        YlJtb1dIczAvbnRNZWtqZlJLdXB5K1EKZHC1YAnMnRJdXfin1KYsbBZBViSysroo
        8wLK53RXN4dgyLsLMmESAWqEqIGOgkns7gbP8N7efakI1aI239SlVg==
        -----END AGE ENCRYPTED FILE-----
  lastmodified: "2025-01-25T09:40:29Z"
  mac: ENC[AES256_GCM,data:ytSnoJOi6eIzWjETgJo8/ppttKbHiSDcxQRJfocW0SWC2kQhyXtM0Y9R/d9JXbJrupqEcFH3yS4NJQz4uFyButI78pOrFxuhxNIhL3YSghTrBZKZ71IpjTe6W/oqz4UUhio5r1VU6KKFcKRKIvZZIUUnlqhJToOLB/VcLxqIQgw=,iv:Gufcas0JD7RVCTPIycN46EUq8V5OzYu++qmtolFu7hA=,tag:46k/pE546i4h18sXudp6Qw==,type:str]
  pgp: [ ]
  unencrypted_suffix: _unencrypted
  version: 3.7.1
```

In the encrypted secrets file, only the values are encrypted, while the object keys remain in plaintext. This approach
allows teams to easily review changes
in [pull requests (PRs)](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests)
without exposing sensitive information.

The decrypted file will look like this:

```yaml
username: user1
password: password1
```

This is the human-readable version of the secrets after RMK decrypts them using the appropriate SOPS Age key.

## SOPS Age keys management

### Creating SOPS Age keys

Initially, a Kubernetes cluster admin generates the required SOPS Age keys using the following command:

```shell
rmk secret keys create
```

Alternatively, the keys for all secret scopes can be generated using the following command:

```shell
rmk project generate --create-sops-age-keys
```

This command will:

- Generate a set of Age private keys for each scope.
- Store them in the user’s home directory at:
  ```shell
  ${HOME}/.rmk/sops-age-keys/<project_name>/.keys.txt
  ```
- Create an Age private key for each scope where an empty SOPS configuration file exists at:
  ```shell
  etc/<scope>/<env>/secrets/.sops.yaml
  ```
- Automatically add `creation_rules` containing the public key and a `regex` for filtering secrets into
  the `.sops.yaml` file.

Example `.sops.yaml` configuration:

```yaml
creation_rules:
  - path_regex: .+\.yaml$
    age: 'age1rq0gx9zuwphw8kjx6ams84rgysqk5kdmhnysxs28r0x955xnzsdsslgtn0'
```

### Uploading SOPS Age keys to a remote storage

After generating the keys, they can be explicitly uploaded to a remote secrets storage supported by the cloud
provider:

```shell
rmk secret keys upload
```

> Users must have the necessary **permissions to upload** SOPS Age keys to the configured secrets storage service.

### Downloading SOPS Age keys from a remote storage

If the keys have already been created by another person (e.g., a cluster admin) and uploaded to the remote storage,
users can download them to their local environment using the following command:

```shell
rmk secret keys download
```

> Users must have **read permissions** to download keys.  
> Without proper permissions, users won't be able to encode/decode secrets or release services using RMK.

### Local storage of SOPS Age keys

Once downloaded, the directory:

```shell
${HOME}/.rmk/sops-age-keys/<project_name>
```

will contain all the necessary keys for secrets encryption and decryption.

## Batch secrets management using RMK secrets manager

### Generating all secrets from scratch

In case of a creating tenant from scratch all the needed directories, such as `etc/<scope>/<env>/secrets/` should first
be populated with an empty `.sops.yaml` file (acts as an indicator that this scope and environment will have the
secrets).
Also, before generating the secret files, you must create or complete a secret `.spec.yaml.gotmpl` file per each scope.
`.spec.yaml.gotmpl` is a template that runs the [sprig](https://masterminds.github.io/sprig) engine with additional
functions.

The additional functions which included in the template functions are:

- **{{\`{{ requiredEnv "ENV_NAME" }}\`}}:** The function requires an input of the specified mandatory variable.
- **{{\`{{ prompt "password" }}\`}}:** The function asks for interactive input. Useful for entering sensitive data.

Example of the files needed for the generation:

```shell
etc/deps/develop/secrets/.sops.yaml
etc/deps/develop/secrets/.spec.yaml.gotmpl
```

An example of `.spec.yaml.gotmpl` for the deps scope:

```yaml
generation-rules:
  - name: email-sender
    template: |
      envSecret:
        EMAIL_API_KEY: {{`{{ requiredEnv "EMAIL_API_KEY" }}`}}
  - name: postgres
    template: |
      rootUsername: root
      rootPassword: {{ randAlphaNum 16 }}
  - name: redis
    template: |
      auth:
        password: {{ randAlphaNum 16 }}
  # ...
```

Then run the following command to generate all the secrets is a batch manner:

```shell
rmk secret manager generate
```

After that the directories with the secret files will be generated.
The files will have the plain unencrypted YAML content.
After that, review the content of the new files and run the following command to encode them:

```shell
rmk secret manager encrypt
```

> The directories without the `.sops.yaml` or `.spec.yaml.gotmpl` files will be ignored.

Also, each of the `.sops.yaml` files will be updated automatically with the correct paths and public keys of the SOPS
age keys
used for encryption.

### Generating a single secret from scratch using the RMK secrets manager

To create a single secret from scratch (e.g., when a new service (release) is added), add a template of the new secret
to `.spec.yaml.gotmpl`. For example:

```yaml
generation-rules:
  # ...
  - name: new-release
    template: |
      envSecret:
        USERNAME: user
        PASSWORD: {{ randAlphaNum 16 }}
# ...
```

Then generate the new secret as the plain YAML and encrypt it using RMK for the needed scope and environment.
For example:

```shell
rmk secret manager generate --scope rmk-test --environment develop
rmk secret manager encrypt --scope rmk-test --environment develop
```

> At this moment, the `.sops.yaml` files has already been populated and therefor need no manual changes.
> The secrets generation process works in an idempotent, declarative mode, which means it will skip previously generated
> secret files.

### Rotating secrets for specific scope and environment

To force a new generation of the secrets for a specific scope and environment according to the `.spec.yaml.gotmpl` file,
run the following command (in this example, the scope is `rmk-test`, the environment is `production`):

```shell
rmk secret manager generate --scope rmk-test --environment production --force
```

> You might need to provide the required environment variables

To encode the generated secrets, run:

```shell
rmk secret manager encrypt --scope rmk-test --environment production
```

## Working with secrets

### Editing a single secret

For some environments where the `.spec.yaml.gotmpl` file and the manager commands were not used for some legacy reasons,
the `rmk secret edit` command can be executed. The command works in an idempotent mode, which means that it can be used
for both creating a new secret (e.g., when adding a new release/service) and editing an existing one:

```shell
rmk secret edit <path_to_new_file_or_existing_secret>
```

For example:

```shell
rmk secret edit etc/deps/develop/secrets/mongodb.yaml
```

An CLI editor will be displayed (e.g., [vim](https://www.vim.org/)). After the required changes are performed,
save and exit the editor. This will result in an encrypted and edited secret file (no need to encode it explicitly).

> Manual editing of the encrypted secrets files is strictly forbidden,
> because SOPS automatically controls the checksums of the secret files.

### Viewing an existing secret

To view the content of an existing secrets content, use the following command:

```shell
rmk secret view <file>
```

For example:

```shell
rmk secret view etc/deps/develop/secrets/postgres.yaml
```

> This might be useful when viewing credentials of the deployed services, e.g., a database or a web UI.

> All RMK commands related to the secrets management can be found under the [rmk secret](../../commands.md#secret)
> command category.
