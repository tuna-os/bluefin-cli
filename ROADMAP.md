# Bluefin CLI roadmap

**Last updated:** 2026-08-22

**Current release line:** v0.10.x

**Planning issue:** [#196](https://github.com/tuna-os/bluefin-cli/issues/196)

Bluefin CLI is a cross-platform environment setup and customization tool in the
[TunaOS](https://tunaos.org) ecosystem. This roadmap describes outcomes and
readiness gates rather than promising individual features or dates. Execution
work should be tracked in linked GitHub issues.

## Current baseline

The project currently ships:

- a persistent native TUI with fuzzy filtering and a command palette;
- shell setup, curated application bundles, wallpapers, and Starship themes;
- portable setup profiles and a cross-platform Brewfile workflow;
- self-update and diagnostics;
- release assets for Linux, macOS, and Windows, including deb and rpm packages;
- CI across Linux, macOS, and Windows, plus scheduled Windows validation.

The release pipeline and direct-download path are operational. Package-manager
availability is incomplete and remains the clearest adoption gap.

## Near term: make distribution dependable

**Outcome:** every installation path shown as available in the README installs
the current release, and unavailable paths are clearly marked.

- Publish and verify the Homebrew formula in `tuna-os/homebrew-tap`.
- Publish and verify the Scoop manifest in `tuna-os/scoop-bucket`.
- Bring Winget to the current release line and validate its update loop.
- Decide whether Chocolatey and AUR are supported commitments; either publish
  and verify them or keep them explicitly out of the supported matrix.
- Add a lightweight post-release check that records which channels received the
  release and makes partial publication visible.

**Exit evidence:** a release checklist links to successful installs or package
records for every supported channel, with no stale version claims in the README.

## Next: define the v1 readiness gate

**Outcome:** maintainers and users share a concrete definition of a stable
Bluefin CLI release.

- Document supported operating systems, architectures, shells, installation
  methods, and compatibility expectations.
- Define configuration and profile compatibility guarantees, including how
  breaking changes and migrations are communicated.
- Set release-health expectations: required CI checks, rollback or hotfix
  ownership, and the maximum acceptable lag across supported package channels.
- Resolve or explicitly defer known platform gaps before declaring v1.0.

**Exit evidence:** a published v1 checklist is complete, the support matrix is
current, and at least one release candidate passes all supported-platform and
distribution checks.

## Then: prove adoption and contributor sustainability

**Outcome:** growth and maintenance capacity are visible enough to guide
priorities after v1.

- Establish a privacy-respecting adoption baseline using release downloads,
  package-channel installs where available, and opt-in `countme` data.
- Publish a small recurring health snapshot covering active users, release
  reach, issue response, and external contributions.
- Create contributor-ready issues for bounded documentation, platform
  verification, and package-maintenance work.
- Document ownership and backup maintainers for release credentials and each
  supported distribution channel.

**Exit evidence:** two consecutive health snapshots are published, at least
three contributor-ready issues are maintained, and release/channel ownership
does not depend on one person.

## Later opportunities

These remain exploratory until the readiness gates above are met:

- remotely managed or user-defined bundle manifests;
- deeper desktop integration beyond the terminal experience;
- additional curated bundle categories driven by measured demand.

## How this roadmap is maintained

- Review it at each minor release or at least monthly.
- Link execution issues when an outcome moves into active work.
- Move completed outcomes into the baseline instead of leaving shipped work in
  future-looking sections.
- Record scope changes in the planning issue so roadmap edits remain auditable.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues and pull requests are welcome.
