# Release publishing credentials

GoReleaser publishes the Homebrew formula to
[`tuna-os/homebrew-tap`](https://github.com/tuna-os/homebrew-tap). The
semantic-release workflow passes the repository secret
`HOMEBREW_TAP_TOKEN` to GoReleaser, and
`.goreleaser.yaml` skips the formula upload when that variable is empty.

## One-time administrator setup

An organization administrator should create a fine-grained GitHub token with:

- repository access restricted to `tuna-os/homebrew-tap`;
- repository permission `Contents: Read and write`.

Store the token in the `tuna-os/bluefin-cli` repository. Name the Actions
secret:

```text
HOMEBREW_TAP_TOKEN
```

The token must belong to an account that can push to the tap's `main` branch.
Do not put it in the workflow file, `.goreleaser.yaml`, or a fork.

## Verification

After adding the secret, merge a conventional commit that warrants a release
to `main` and check the `Automated Release` workflow's GoReleaser output for
the Homebrew publisher. The generated formula should land
in `tuna-os/homebrew-tap` under `Formula/` on `main`.

Without the secret, releases stay green and publish the other configured
channels. GoReleaser skips the Homebrew publisher. This lets forks and pull
requests test the setting without access to the organization credential.

## Scoop

GoReleaser publishes the Windows Scoop manifest to
[`tuna-os/scoop-bucket`](https://github.com/tuna-os/scoop-bucket). The
semantic-release workflow passes the repository secret
`SCOOP_BUCKET_TOKEN` to GoReleaser, and
`.goreleaser.yaml` skips the Scoop upload when that variable is empty.

## One-time administrator setup

An organization administrator should create a fine-grained GitHub token with:

- repository access restricted to `tuna-os/scoop-bucket`;
- repository permission `Contents: Read and write`.

Store the token in the `tuna-os/bluefin-cli` repository. Name the Actions
secret:

```text
SCOOP_BUCKET_TOKEN
```

The token must belong to an account that can push to the bucket's `main`
branch. Do not put it in the workflow file, `.goreleaser.yaml`, or a fork.

## Verification

After adding the secret, merge a conventional commit that warrants a release
to `main` and check the `Automated Release` workflow's GoReleaser output for
the Scoop publisher. The generated manifest should land
in `tuna-os/scoop-bucket` on `main`.

Without the secret, releases stay green and publish the other configured
channels. GoReleaser skips the Scoop publisher. This lets forks and pull
requests test the setting without access to the organization credential.
