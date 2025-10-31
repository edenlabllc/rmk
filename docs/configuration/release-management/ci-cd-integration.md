# Integration into the CI/CD pipeline

## Release update

The [rmk release update](../../commands.md#update-u-2) command automates the process of **updating and
delivering releases** according to the version changes of artifacts, e.g., container images, following
the [GitOps](https://www.gitops.tech) methodology.

Since RMK is a binary file that can be **downloaded and installed** on any Unix-based operating system,
it can be **integrated** with almost any CI/CD system: GitHub Actions, GitLab, Drone CI, Jenkins, etc.

## Example of integration with GitHub Actions

> **Prerequisites:**
>
> - The project repository has already
    been [generated and prepared](../project-management/preparation-of-project-repository.md) using RMK.

Create the following workflow in your project repository at `.github/workflows/release-update.yaml`.
An example content of the [GitHub Actions](https://github.com/features/actions)' workflow:

```yaml
name: Release update

on:
    workflow_dispatch:
        inputs:
            image_repository_full_name:
                description: Image repository full name of application.
                required: true
            version:
                description: Current application version.
                required: true

jobs:
    release-update:
        runs-on: ubuntu-22.04
        steps:
            - name: Checkout main repository
              uses: actions/checkout@v4
              with:
                  ref: ${{ github.ref }}
                  fetch-depth: 0
            
            - name: Release update
              env:
                  AWS_REGION: us-east-1
                  AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
                  AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
                  RMK_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN_REPO }}
                  RMK_SLACK_WEBHOOK: ${{ secrets.RMK_SLACK_WEBHOOK }}
                  RMK_SLACK_CHANNEL: rmk-test-cd-notifications
                  RMK_RELEASE_UPDATE_REPOSITORY: ${{ github.event.inputs.image_repository_full_name }}
                  RMK_RELEASE_UPDATE_TAG: ${{ github.event.inputs.version }}
              run: |
                  curl -sL "https://edenlabllc-rmk-tools-infra.s3.eu-north-1.amazonaws.com/rmk/s3-installer" | bash
                  
                  rmk config init --cluster-provider=aws --progress-bar=false --slack-notifications
                  rmk release update --skip-ci --deploy
```

In this example, we have prepared a `GitHub Action` that expects two input parameters:

- `image_repository_full_name`
- `version`

As soon as a request with these parameters is sent to this action,
RMK will be executed, first analyzing all the `releases.yaml` files to match the `image_repository_full_name` and will
replace the tag field with the corresponding version if the versions differ. After that, it will
**automatically commit** the changes to the current branch in the `releases.yaml` files where changes have been found.
Then, it will **synchronize the releases** where the version changes were found.

An example of the `releases.yaml` file:

```yaml
# ...
foo:
    enabled: true
    image:
        repository: 123456789012.dkr.ecr.us-east-1.amazonaws.com/app.foo
        tag: v0.1.0
bar:
    enabled: true
    image:
        repository: 123456789012.dkr.ecr.us-east-1.amazonaws.com/app.bar
        tag: v0.1.1
# ...
```

To fully automate code delivery in the CI pipeline, add a step that triggers the deployment after building and pushing
the container image. This step should pass the full image repository name and tag via an API call
using [cURL](https://en.wikipedia.org/wiki/CURL) or [GitHub CLI](https://cli.github.com/). This ensures **seamless
deployment** to the infrastructure environment.
