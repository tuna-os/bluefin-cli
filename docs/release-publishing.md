# Release and distribution policy

This document defines the versioned release line, required artifacts, and the
status of each distribution channel. The
[release-pipeline ADR](adr/0003-release-pipeline.md) records the implementation
decision; this page is the operational contract.

## Versioning and release line

`main` is the only release branch. On each push, semantic-release checks
headers for conventional commits and creates an immutable
`vMAJOR.MINOR.PATCH` tag and GitHub release when needed:

| Change | Version effect |
|---|---|
| `fix:` or `perf:` | patch |
| `feat:` | minor |
| a `BREAKING CHANGE` footer | major |
| `docs:`, `test:`, `chore:`, or `refactor:` alone | no release |

The current line is v0.11.x. Before v1.0, users that need a fixed interface
should pin an exact version. The project has not yet declared compatibility
guarantees in the roadmap's v1 gate. Install scripts accept
`BLUEFIN_CLI_VERSION` for that purpose.

Do not create, move, or reuse a release tag manually. Recovery must preserve the
existing tag and use a recovery workflow with a narrow scope for the affected
channel.

## Required GitHub release artifacts

GitHub Releases is the source distribution channel. A complete release contains:

- `bluefin-cli` and `bluefin-cli-plus` archives for Linux, macOS, and Windows
  on amd64 and arm64;
- deb and rpm packages for the standard Linux build on amd64 and arm64;
- `checksums.txt` with a SHA-256 digest for every published artifact;
- release notes from commits that meet release rules.

The Unix and Windows installers use `checksums.txt` to reject archives that do
not match it.
The built-in self-update path uses `checksums.txt` to reject the same archives.
Other consumers should compare the downloaded artifact's SHA-256 digest with its
entry before installation.

`checksums.txt` detects corruption or a mismatch between an artifact and the
published manifest. Because the manifest and artifacts share the same release,
it is not an independent publisher signature. Releases do not now claim signed
artifacts or provenance attestation; the v1 readiness gate still includes that
decision.

## Distribution channel contract

Channel states mean:

- **Supported**: maintained by this repository and expected to track each
  stable release.
- **Preview**: usable, but freshness or automated publication is not yet a
  release guarantee.
- **Unavailable**: not an advertised installation path, even when publisher
  configuration exists in the repository.

Status as of 2026-10-17:

| Channel | State | Contract |
|---|---|---|
| GitHub archives and one-line installers | Supported | Every stable release; archives are covered by `checksums.txt`. |
| GitHub deb and rpm assets | Supported | Every stable release; direct packages only, not a distro repository. |
| `ublue-os/homebrew-experimental-tap` | Preview | Source formula is synchronized manually and may lag. |
| `tuna-os/homebrew-tap` | Unavailable | Publisher is configured with the `HOMEBREW_TAP_SSH_KEY` deploy key; publication is not verified yet. |
| Winget (`Hanthor.BluefinCLI`) | Preview | Installable v0.8.1 is stale; normal publication is disabled pending token repair, with a manual resubmission workflow available. |
| `tuna-os/scoop-bucket` | Unavailable | Publisher is configured but skipped until `SCOOP_BUCKET_TOKEN` is installed and publication is verified. |
| Chocolatey | Unavailable | No package is published; the release invocation explicitly skips this publisher. |
| AUR (`bluefin-cli-bin`) | Unavailable | Publisher is configured but no package is published; it remains skipped without `AUR_KEY`. |

The README installation section must use the same states and versions. A
configured publisher is not a supported channel until a current package exists.
The team must verify an install. A maintainer and backup maintainer must own its
credentials and recovery path.

## Publishing flow

1. Merge a conventional commit to `main`.
2. The `Automated Release` workflow lets semantic-release calculate the next
   version and invokes GoReleaser once.
3. Confirm the GitHub release contains the complete artifact matrix and that
   every file has an entry in `checksums.txt`.
4. Confirm every supported downstream channel contains that version and record
   an installation check.
5. If a downstream publisher fails or skips publication, keep the GitHub
   release. Downgrade channel state until the publisher fixes and verifies
   publication. Never report a configured-but-empty channel as available.

Missing downstream credentials intentionally do not fail the core GitHub
release. This keeps direct distribution available. It does not make the channel
successful when the publisher skips publication.

## Channel credentials

Credentials belong in repository Actions secrets, never in workflow files,
GoReleaser configuration, forks, or documentation.

### Homebrew

Create an SSH key pair. Add the public key to `tuna-os/homebrew-tap` as a
deploy key with write access. Store the private key in this repository as
`HOMEBREW_TAP_SSH_KEY`. GoReleaser uses it to push over SSH to the tap's `main`
branch. A successful publication writes `Formula/bluefin-cli.rb`.

### Scoop

Create a fine-grained GitHub token. Restrict it to `tuna-os/scoop-bucket` with
`Contents: Read and write`, then store it in this repository as
`SCOOP_BUCKET_TOKEN`. The token's account must be able to push to the bucket's
`main` branch. A successful publication writes the `bluefin-cli` manifest.

### Winget

`WINGET_CREATE_GITHUB_TOKEN` must be a classic GitHub token with `public_repo`
scope. A fine-grained token cannot open the required pull request against
`microsoft/winget-pkgs`. Automated releases now skip Winget. After you repair
the token, use `.github/workflows/winget.yml` to resubmit an existing tag.
Verify the package. Then remove the release skip.

### AUR

`AUR_KEY` must contain a private SSH key. The AUR service must authorize the key
to publish `bluefin-cli-bin`. Enable it only after a maintainer records
ownership of the package and names a backup maintainer. Verify the generated
package before you change the channel state.
