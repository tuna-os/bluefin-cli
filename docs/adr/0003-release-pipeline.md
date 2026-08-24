# ADR 0003: semantic-release + goreleaser delivery

**Status**: accepted (2026-07)

## Context
Releases were broken for months (stale tag collision, publisher id
mismatches, silent version stamping failure); installs must be one command
on every OS and self-update safely.

## Decision
semantic-release computes versions from conventional commits and invokes
goreleaser. Per-build archives (`standard`/`plus`) because publisher `ids`
match ARCHIVE ids. Channels: archives, deb/rpm (nfpms), brew tap, scoop
bucket, AUR, winget — every publisher self-gates with
`skip_upload: {{ if envOrDefault "SECRET" "" }}auto{{ else }}true{{ end }}`
so a missing secret can never fail a release. Repository `token`/
`private_key` fields render in goreleaser's env-only mode: plain
`{{ .Env.X }}` only, no functions. `WINGET_TOKEN` must be a classic PAT with
`public_repo` scope to enable opening PRs against upstream `microsoft/winget-pkgs`
(fine-grained PATs return 403; see #110). Self-update (`internal/update`) refuses
archives that don't match the release's checksums.txt.

## Consequences
Adding a channel = config + secret; releases are all-or-nothing green.
`brews` deprecation is accepted until a Linux-capable replacement exists.
The `Automated Release` workflow is the sole GoReleaser owner. Tags created by
semantic-release are release outputs, not a second workflow trigger; manual
recovery paths must have an explicitly non-overlapping publishing scope.
