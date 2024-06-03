# Development and release flow

## Requirements for the availability of tools during development

- **[Golang](https://tip.golang.org/doc/install)** = v1.21.X
- **[GoReleaser](https://goreleaser.com/install)** = v1.23.0

## Building from source

To build RMK from source, run the following [GoReleaser](https://goreleaser.com/) command from the root of the repository:

```shell
goreleaser build --snapshot --clean
```

> You can also use this command for recompilation of RMK during development.

## Git workflow

In RMK development, we use the classic [GitFlow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow) workflow, 
embracing all its advantages and disadvantages.

### Git branch naming conventions

- `feature/RMK-<issue_number>-<issue_description>`
- `release/<SemVer2>`
- `hotfix/<SemVer2>`

For example:

- `feature/RMK-123-add-some-important-feature`
- `release/v0.42.0`
- `hotfix/v0.42.1`

## Release flow

After accumulating a certain set of features in the develop branch, 
a `release/<SemVer2>` branch is created for the next release version. 
Then a pull request (PR) is made from the `release/<SemVer2>` branch to the master branch. 
This triggers a CI process that will build and release an intermediate version, 
the `<SemVer2>-rc` release candidate. 

> This version is available for update from RMK via the `rmk update --release-candidate` command
and can be used for an intermediate or beta testing. 

After successful testing, the PR is merged into the master branch, 
triggering another CI process that will release the stable RMK version. 

> All CI processes are performed using GoReleaser, including the publishing of artifacts 
> for both the release and intermediate RMK versions.
