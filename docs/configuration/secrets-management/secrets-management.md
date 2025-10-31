# Secrets management

## Overview

RMK utilizes [SOPS](https://github.com/mozilla/sops) and [Age](https://github.com/mozilla/sops#encrypting-using-age)
for secrets management, ensuring **secure encryption, storage, and access to sensitive data**. The tool ensures
seamless and automated secret management, reducing manual effort while maintaining **security best practices**.

The functionality in RMK is divided into two key areas:

1. **Working with secret keys**: Managing encryption keys used for encrypting and decrypting secrets.
2. **Working with secret files**: Handling encrypted YAML files that store sensitive configuration data.

### Secret keys

This area focuses on integration with cloud providers, ensuring **secure storage, retrieval, and local access to secrets
**.
Once the keys are generated, RMK **automatically provisions** all necessary cloud resources, securely stores the keys in
the
provider’s Secrets Manager service, and downloads them to the local machine upon first use.

For each supported cloud provider, RMK integrates with the respective secrets management service:

- **AWS**: Integrates with [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/).
- **Azure**: Integrates with [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault).
- **GCP**: Integrates
  with [Google Cloud Secret Manager](https://cloud.google.com/security/products/secret-manager?hl=en).

Locally, secret keys are stored in a secure file within the user's home directory:

```shell
${HOME}/.rmk/sops-age-keys/<project_name>/
```

For example, the directory for the `rmk-test` project:

```shell
${HOME}/.rmk/sops-age-keys/rmk-test/
```

might have the following content:

- `.keys.txt`: the main merged file of all secret keys that SOPS uses.
- `rmk-test-deps.txt`: secret key for the `deps` scope.
- `rmk-test-rmk-test.txt`: secret key for the `rmk-test` scope.

A secret key will look like this:

```shell
# created: 2025-01-23T20:47:30+01:00
# public key: age1rq0gx9zuwphw8kjx6ams84rgysqk5kdmhnysxs28r0x955xnzsdsslgtn0
AGE-SECRET-KEY-15K8LZB3MT0QWJ4N7X90H2A9C5L6E7FYZ3XGKP1DRN8SWV2QXT90H2A9C5L
```

By design, **each project and scope has its own key**, ensuring isolated key storage.

> Secret keys are **not separated by environment name**. This allows secrets to be managed independently of the branch
> or environment currently in use.

### Secret files

This area focuses on integration with Helmfile, Helm, and Kubernetes, ensuring **automated and seamless secrets
management**
throughout the deployment process.

Normally, the secret files can be **committed to Git** because they are encrypted with secret keys
using [symmetric-key algorithms](https://en.wikipedia.org/wiki/Symmetric-key_algorithm).

In a project repository, all secret files are stored in the `etc/<scope>/<environment>/secrets/` directories.
For example:

```shell
etc/deps/develop/secrets/postgres.yaml
etc/rmk-test/develop/secrets/app.yaml
```

> Similar to the [releases.yaml](../project-management/requirement-for-project-repository.md#requirement-for-releasesyaml)
> files, secrets files are **never inherited by projects**, in contrast to the Helmfile values. Each project
> **should have its own** unique set of secrets for all deployed releases.

In the encrypted secret file, only the values are encrypted, while the object keys remain in plaintext. This approach
allows teams to easily review changes
in [pull requests (PRs)](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests)
without exposing sensitive information.

<details>
  <summary>Example of an encrypted secret file, where sensitive values are encrypted using <a href="https://en.wikipedia.org/wiki/Advanced_Encryption_Standard" target="_blank">AES256-GCM</a></summary>

  ```yaml
  username: ENC[AES256_GCM,data:A0jb1wU=,iv:RM8V1IOHvCrBv7f9f/GKobaBYyjcX9jcNQp6XPopNcU=,tag:T79VY3/yqlIffdbvYDwukQ==,type:str]
  password: ENC[AES256_GCM,data:Kjo5hDSb+VmhdLLuq48oVg==,iv:5wpJBsiA5B82RRaguW8/TcKgGYbiZhihdIhXnPwyRG8=,tag:yQ5Chi949jBB1cSaFDVlOQ==,type:str]
  sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
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
    pgp: []
    unencrypted_suffix: _unencrypted
    version: 3.7.1
  ```

</details>

The decrypted file will look like this:

```yaml
username: user1
password: password1
```

This is the human-readable version of the secrets after RMK decrypts them using the appropriate secret key.

## Secret keys management

### Creating secret keys

Initially, a Kubernetes cluster admin generates the required secret keys using the following command:

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
  ${HOME}/.rmk/sops-age-keys/<project_name>/
  ```
- Create an Age private key for each scope where an empty SOPS configuration file exists at:
  ```shell
  etc/<scope>/<environment>/secrets/.sops.yaml
  ```
- Automatically add `creation_rules` containing the public key and a `regex` for filtering secrets into
  the `.sops.yaml` file.

Example `.sops.yaml` configuration:

```yaml
creation_rules:
  - path_regex: .+\.yaml$
    age: 'age1rq0gx9zuwphw8kjx6ams84rgysqk5kdmhnysxs28r0x955xnzsdsslgtn0'
```

### Uploading secret keys to a remote storage

After generating the keys, they can be explicitly uploaded to a remote secrets storage supported by the cloud
provider:

```shell
rmk secret keys upload
```

> Users must have the necessary **[permissions to upload](../configuration-management/configuration-management.md#initialization-of-rmk-configuration-for-different-cluster-providers)** 
> secret keys to the configured secrets storage service.

### Downloading secret keys from a remote storage

If the keys have already been created by another person (e.g., a cluster admin) and uploaded to the remote storage,
users can download them to their local environment using the following command:

```shell
rmk secret keys download
```

> Users must have **read permissions** to download keys. Without proper permissions, users won't be able to
> encode/decode secrets or release services using RMK.

Once downloaded, the directory:

```shell
${HOME}/.rmk/sops-age-keys/<project_name>/
```

will contain all the necessary keys for secrets encryption and decryption.

## Working with a single secret

> All RMK commands related to the secrets management can be found under the [rmk secret](../../commands.md#secret)
> command category.

### Creating or editing a secret

The `rmk secret edit` command operates in an **idempotent** mode, meaning it can be used for both **creating** new
secrets (e.g., when adding a new release) and **modifying** existing ones.

To create or edit a secret, run:

```shell
rmk secret edit <path_to_new_file_or_existing_secret>
```

For example:

```shell
rmk secret edit etc/deps/develop/secrets/postgres.yaml
```

This command will open a CLI text editor (e.g., [vim](https://www.vim.org/)). After making the necessary changes, save
and exit the editor. The updated secret will be **automatically encrypted**, so no additional encoding is required.

> Manual editing of the encrypted secrets files is **strictly forbidden**, because SOPS automatically controls the
> checksums of the secret files.

### Viewing an existing secret

To view the decrypted content of an existing secret, use:

```shell
rmk secret view <path_to_existing_secret>
```

For example:

```shell
rmk secret view etc/deps/develop/secrets/postgres.yaml
```

This is useful for **inspecting credentials** of deployed services, such as database access details or authentication
credentials for a web UI.
