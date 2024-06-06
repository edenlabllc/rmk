# Secrets management

RMK uses [SOPS](https://github.com/mozilla/sops) and [Age](https://github.com/mozilla/sops#encrypting-using-age)
for secrets management.
All RMK commands related to the secrets management can be found under the [rmk secret](../../commands.md#secret) command category.

In a project repository, all secrets files are stored in the `etc/<scope>/<env>/secrets/` directories.
For example:

```
etc/deps/develop/secrets/mongodb.yaml
etc/kodjin/develop/secrets/kodjin-minio-config-buckets.yaml
```

Normally, the files are committed to Git because they are encrypted using SOPS age keys and symmetric-key algorithms.
The keys are stored remotely in an encrypted bucket of AWS S3, downloaded locally when first using RMK
on that machine.

> The secrets are never inherited by projects, in contrast to values. Each project should have its own unique set 
> of secrets for all deployed releases.

## SOPS age keys management

Initially, a Kubernetes cluster admin creates the keys using the following command:

```shell
rmk secret keys create
```

This command will create a set of `Age` private keys for each scope in the user's home directory 
at `$HOME/.rmk/sops-age-keys/<project_name>-sops-age-keys-<short_AWS_account_id>/*.txt`. 
The command will create a `Age` private key for the scope for which an empty 
SOPS configuration file `etc/<scope>/<env>/secrets/.sops.yaml` was created. 
Additionally, `creation_rules` containing the public key and regex for filtering secrets will be automatically added to this file. 
Example of a `.sops.yaml` file:

```yaml
creation_rules:
  - path_regex: .+\.yaml$
    age: 'age12ugnsdhxd44fa56q5w22mvf09e93p66xrcrnxmxla4dnqpyugpwqs2swag'
```

> The SOPS configuration files for all secret scopes can be generated using the `rmk project generate` command.

Then the generated keys might be uploaded explicitly to S3:

```shell
rmk secret keys upload
```

> AWS users must have the `AdministratorAccess` permissions to be able to upload the SOPS age keys to S3.

After the keys have been created by the cluster admin, they can be downloaded from S3:

```shell
rmk secret keys download
```

> AWS users must have read permissions for downloading keys from S3,
> otherwise they won't be able to encode/decode secrets and release services using RMK.

After that the `$HOME/.rmk/sops-age-keys/<project_name>-sops-age-keys-<short_AWS_account_id>` directory will have all the needed keys
needed for secrets encryption or decryption.

## Generating all secrets from scratch in a batch manner using the RMK secrets manager

In case of a creating tenant from scratch all the needed directories, such as `etc/<scope>/<env>/secrets/` should first
be populated with an empty `.sops.yaml` file (acts as an indicator that this scope and environment will have the secrets).
Also, before generating the secret files, you must create or complete a secret `.spec.yaml.gotmpl` file per each scope.
`.spec.yaml.gotmpl` is a template that runs the [sprig](https://masterminds.github.io/sprig) engine with additional functions.

The additional functions which included in the template functions are:

- **{{`{{ requiredEnv "ENV_NAME" }}`}}:** The function requires an input of the specified mandatory variable.
- **{{`{{ prompt "password" }}`}}:** The function asks for interactive input. Useful for entering sensitive data.


Example of the files needed for the generation:

```yaml
etc/deps/develop/secrets/.sops.yaml
etc/deps/develop/secrets/.spec.yaml.gotmpl
```

An example of `.spec.yaml.gotmpl` for the deps scope:

```gotemplate
{{- $minioSecretKey := randAlphaNum 16 -}}
generation-rules:
  - name: ecr-token-refresh-operator
      template: |
        envSecret:
          AWS_ACCESS_KEY_ID: {{`{{ requiredEnv "AWS_ACCESS_KEY_ID" }}`}}
          AWS_SECRET_ACCESS_KEY: {{`{{ requiredEnv "AWS_SECRET_ACCESS_KEY" }}`}}
  - name: elastic
      template: |
        esUserName: admin
        esPassword: {{ randAlphaNum 16 }}
  # ...
  - name: minio
      template: |
        rootUser: minio
        rootPassword: {{ $minioSecretKey }}
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

Also, each of the `.sops.yaml` files will be updated automatically with the correct paths and public keys of the SOPS age keys
used for encryption.

## Generating a single secret from scratch using the RMK secrets manager

To create a single secret from scratch (e.g., when a new service (release) is added), add a template of the new secret 
to `.spec.yaml.gotmpl`. For example:

```gotemplate
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
rmk secret manager generate --scope kodjin --environment develop
rmk secret manager encrypt --scope kodjin --environment develop
```

> At this moment, the `.sops.yaml` files has already been populated and therefor need no manual changes.
> The secrets generation process works in an idempotent, declarative mode, which means it will skip previously generated secret files.

## Editing a single secret

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

## Viewing an existing secret

To view the content of an existing secrets content, use the following command:

```shell
rmk secret view <file>
```

For example:

```shell
rmk secret view etc/deps/develop/secrets/minio.yaml
```

> This might be useful when viewing credentials of the deployed services, e.g., a database or a web UI.

## Rotate secrets for specific scope and environment

To force a new generation of the secrets for a specific scope and environment according to the `.spec.yaml.gotmpl` file,
run the following command (in this example, the scope is `kodjin`, the environment is `production`):

```shell
rmk secret manager generate --scope kodjin --environment production --force
```

> You might need to provide the required environment variables

To encode the generated secrets, run:

```shell
rmk secret manager encrypt --scope kodjin --environment production
```
