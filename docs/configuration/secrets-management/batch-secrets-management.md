# Batch secrets management

## Overview

RMK secrets manager automates secret management in Kubernetes in batch mode.

Key features include:

- Batch secret generation and encryption using [Golang templates](https://pkg.go.dev/text/template)
  and [Sprig](https://masterminds.github.io/sprig).
- Flexible configuration and process separation for **project**, **scope**, and **environment levels**.

## Generating all secrets from scratch

When creating a new project from scratch, all required directories, such as:

```shell
etc/<scope>/<environment>/secrets/
```

must first include an empty `.sops.yaml` file with no content.
This file acts as an indicator that secrets will be managed by RMK for this scope and environment.

To create an empty file in the specified directory, the
following commands can be utilized:

```shell
mkdir -p etc/<scope>/<environment>/secrets
touch etc/<scope>/<environment>/secrets/.spec.yaml.gotmpl

```

For example:

```shell
mkdir -p etc/deps/develop/secrets
touch etc/deps/develop/secrets/.spec.yaml.gotmpl
```

This ensures that the required directory structure exists before generating secrets.

After that, each scope requires a `.spec.yaml.gotmpl` template file to define the structure of the generated secrets.
This file is processed using the [Sprig](https://masterminds.github.io/sprig) templating engine. In addition to the
standard functions provided by the engine, RMK extends `.spec.yaml.gotmpl` with the following extra
template functions:

- `{{ requiredEnv "VAR_NAME" }}` – requires the specified environment variable as input.
- `{{ prompt "VAR_NAME" }}` – prompts the user for interactive input.

<details>
  <summary>Example <code>.spec.yaml.gotmpl</code> file</summary>

  ```yaml
  generation-rules:
    - name: email-sender
      template: |
        envSecret:
          EMAIL_API_KEY: {{ requiredEnv "EMAIL_API_KEY" }}
          EMAIL_SENDER: {{ prompt "EMAIL_SENDER" }}
    - name: postgres
      template: |
        rootUsername: root
        rootPassword: {{ randAlphaNum 16 }}
        appUsername: {{ requiredEnv "POSTGRES_USERNAME" }}
        appPassword: {{ prompt "POSTGRES_PASSWORD" }}
    - name: redis
      template: |
        auth:
          password: {{ randAlphaNum 16 }}
        cacheTTL: {{ requiredEnv "REDIS_TTL" }}
  ```

In this example:
  <ul>
    <li>The <code>name</code> field corresponds to a Helmfile/Helm release name, such as <code>email-sender</code>, <code>postgres</code>, or <code>redis</code>.</li>
    <li><code>EMAIL_API_KEY</code>, <code>POSTGRES_USER</code>, and <code>REDIS_TTL</code> must be set as environment variables before running the generation process.</li>
    <li>The user is prompted to enter the email sender address (<code>EMAIL_SENDER</code>) and the PostgreSQL application password (<code>POSTGRES_PASSWORD</code>) interactively.</li>
  </ul>
</details>

All variables defined in the template using the `requiredEnv` function must be exported before executing
the `rmk secret manager generate` command.

In our example, the required variables are (`EMAIL_API_KEY`, `POSTGRES_USER`, `REDIS_TTL`). They can be exported as
follows:

```shell
export EMAIL_API_KEY="dummy-email-api-key-XXX"
export POSTGRES_USER="user1"
export REDIS_TTL="3600"
```

After exporting the variables, run the following command to generate all
secret files in batch mode:

```shell
rmk secret manager generate
```

For the example above, secret files will be created for each of the three releases (`email-sender`, `postgres`,
and `redis`). These files will be generated in the `etc/deps/develop/secrets` directory with the following naming
pattern:

```shell
etc/deps/develop/secrets/email-sender.yaml
etc/deps/develop/secrets/postgres.yaml
etc/deps/develop/secrets/redis.yaml
```

The secrets generation process runs in an **idempotent** mode, skipping previously generated files and logging
a warning if they already exist.

> Each secret file will be generated in **plaintext YAML** format and **should be reviewed** before encryption and
> committing to Git.

## Encrypt all the generated secrets

Once the secrets have been verified, encrypt them using:

```shell
rmk secret manager encrypt
```

> Directories that do not contain a `.sops.yaml` or `.spec.yaml.gotmpl` file **will be ignored**.

Additionally, each `.sops.yaml` file will be automatically updated with the correct paths  
and the public keys of the secret keys used for encryption.

> Manual editing of the encrypted secrets files is **strictly forbidden**, because SOPS automatically controls the
> checksums of the secret files. To safely modify encrypted secrets, always use the
> specialized [edit](#creating-or-editing-a-secret) command.

## Create a new secret later

To generate and encode a new secret in addition to the previously generated ones (e.g., when a new service (release) is
added), a template for the new secret should be added to `.spec.yaml.gotmpl`.

For example, for a new `new-app` release:

```yaml
generation-rules:
  # ...
  - name: new-app
    template: |
      username: {{ requiredEnv "APP_USERNAME" }}
      password: {{ requiredEnv "APP_PASSWORD" }}
# ...
```

Then, generate the new secret as a plain YAML file and encrypt it using RMK for the required scope and environment.

For example, for the `rmk-test` scope and `develop` environment:

```shell
export APP_USERNAME="user1"
export APP_PASSWORD="password1"
rmk secret manager generate --scope rmk-test --environment develop
rmk secret manager encrypt --scope rmk-test --environment develop
```

Finally, a secret file will be created for the `new-app` release in the `etc/deps/develop/secrets` directory:

```shell
etc/deps/develop/secrets/new-app.yaml
```

## Rotating all the secrets for a specific scope and environment

To regenerate all the secrets for a specific scope and environment (e.g., when existing secrets have been compromised
and
need to be replaced) based on the `.spec.yaml.gotmpl` file, use the `--force` flag. This ensures that previously
generated secret files are **overwritten**.

For example, to rotate secrets for the `rmk-test` scope in the `production` environment, run:

```shell
# Export all required environment variables before generating
rmk secret manager generate --scope rmk-test --environment production --force
rmk secret manager encrypt --scope rmk-test --environment production
```

This process ensures that all secrets are freshly generated and securely encrypted before deployment.
