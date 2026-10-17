# ADR 0003: semantic-release + goreleaser delivery

**Status**: accepted (2026-07)

## Context
For months, releases failed because of stale tag collisions, publisher ID
mismatches, and silent version errors. Install must use one command on every OS,
and self-update must stay safe.

## Decision
semantic-release computes versions from conventional commits and invokes
Create one archive for each build (`standard`/`plus`) because publisher IDs must
match ARCHIVE IDs. Channels include archives, deb/rpm (nfpms), brew tap, scoop
bucket, AUR, and winget. Each publisher self-gates with
`skip_upload: {{ if envOrDefault "SECRET" "" }}auto{{ else }}true{{ end }}`
so a missing secret can never fail a release.
Repository `token`/`private_key` fields render in goreleaser's env-only mode:
plain `{{ .Env.X }}` only, no functions. `WINGET_TOKEN` must be a classic PAT
with the `public_repo` scope. It lets the workflow open PRs against upstream
`microsoft/winget-pkgs`; fine-grained PATs return 403 (see #110).
Self-update (`internal/update`) refuses archives that do not match the release's
`checksums.txt`.

## Consequences
A new channel needs config and a secret. Releases stay all-or-nothing green.
The project accepts `brews` until a Linux-capable replacement exists.
The `Automated Release` workflow is the sole GoReleaser owner. Semantic-release
tags are release outputs, not a second workflow trigger. Manual recovery paths
must cover a separate scope and must not overlap the automated release.
