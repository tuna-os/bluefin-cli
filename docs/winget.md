# Winget Publishing Guide

This guide explains how to publish `bluefin-cli` to Winget.

## Can Bluefin CLI be registered on Winget?

Yes. `bluefin-cli` is a good fit for a Winget `portable` package.

## 1. Prepare Release Assets

Winget needs a stable downloadable artifact and hash.

Required for each release tag:

- `bluefin-cli_<version>_windows_amd64.zip`
- Optional: `bluefin-cli_<version>_windows_arm64.zip`
- `checksums.txt`

The zip should contain `bluefin-cli.exe` at the top level.

## 2. Add Windows Build to Release Workflow

Update `.github/workflows/release.yml` to produce Windows zip artifacts alongside Linux/macOS builds.

Example build commands:

```bash
GOOS=windows GOARCH=amd64 go build -o bluefin-cli.exe .
zip -j dist/bluefin-cli_${CLEAN_VERSION}_windows_amd64.zip bluefin-cli.exe
rm bluefin-cli.exe
```

## 3. Create Winget Manifest

Use `wingetcreate`:

```powershell
winget install wingetcreate
wingetcreate new https://github.com/hanthor/bluefin-cli/releases/download/v1.2.3/bluefin-cli_1.2.3_windows_amd64.zip --identifier Hanthor.BluefinCLI --version 1.2.3
```

This generates:

- `Hanthor.BluefinCLI.yaml`
- `Hanthor.BluefinCLI.installer.yaml`
- `Hanthor.BluefinCLI.locale.en-US.yaml`

For a portable package, ensure installer fields include:

- `InstallerType: portable`
- `NestedInstallerType: portable`
- `NestedInstallerFiles` entry for `bluefin-cli.exe`
- `Commands` includes `bluefin-cli`

## 4. Submit to Winget Community Repo

Submit to `microsoft/winget-pkgs` using `wingetcreate submit` or manual PR.

```powershell
wingetcreate submit <path-to-manifest-folder>
```

## 4a. Automate Submission with GitHub Actions

This repo includes `.github/workflows/winget.yml`.

It supports:

- Automatic submission on GitHub Release `published`
- Manual submission via `workflow_dispatch`

Required repository secret:

- `WINGET_CREATE_GITHUB_TOKEN` (classic PAT with `public_repo` scope)

The workflow will:

1. Resolve version and release asset URL.
2. Download `wingetcreate`.
3. Run `wingetcreate update ... --submit` and fallback to `new` if needed.

## 5. Validate Installation

After merge and index propagation:

```powershell
winget install --id Hanthor.BluefinCLI --exact
bluefin-cli --help
bluefin-cli shell powershell on
```

## Notes

- Keep release asset names stable across versions to avoid manifest churn.
- Do not rewrite or delete old release assets referenced by existing Winget manifests.
- If package ID changes, users may need uninstall/reinstall rather than upgrade.
