# ADR 0003: semantic-release + goreleaser delivery

**Status**: accepted (2026-07)

## Context
For months, releases did not work (a stale tag collision, publisher ids that
did not match, a silent failure to stamp the version). An install must be one
command on every OS, and self-update must be safe.

## Decision
semantic-release computes versions from conventional commits and invokes
goreleaser. Each build has its own archive (`standard`/`plus`), because
publisher `ids` must match the ids of an ARCHIVE. Channels: archives, deb/rpm (nfpms), brew tap, scoop
bucket, AUR, winget — every publisher self-gates with
`skip_upload: {{ if envOrDefault "SECRET" "" }}auto{{ else }}true{{ end }}`
so a missing secret can never fail a release. Repository `token`/
`private_key` fields render in goreleaser's env-only mode: plain
`{{ .Env.X }}` only, no functions. `WINGET_TOKEN` must be a classic PAT with
`public_repo` scope, so that it can open PRs against upstream `microsoft/winget-pkgs`
(fine-grained PATs return 403; see #110). Self-update (`internal/update`) refuses
archives that don't match the release's checksums.txt.

## Consequences
Adding a channel = config + secret; releases are all-or-nothing green.
We accept the deprecation of `brews` until a replacement that works on Linux
exists. Only the `Automated Release` workflow runs GoReleaser. The tags that
semantic-release creates are outputs of a release, not a second workflow
trigger. A manual recovery path must publish only to channels that no other
path publishes to.
