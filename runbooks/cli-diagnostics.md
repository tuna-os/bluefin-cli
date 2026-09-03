# Operational Runbook: Bluefin CLI Diagnostics & Triage

## Overview

This runbook describes standard operating procedures for triaging and resolving failures in `bluefin-cli` local executions, package installation workflows, and shell environment setup.

## Symptoms & Triage

### 1. Installation or Package Fetch Failures
- **Symptom:** Subcommand `bluefin-cli install` fails or hangs while resolving Homebrew, Flatpak, or native package manager bundles.
- **Triage Steps:**
  1. Check network connectivity and package manager repository accessibility (`brew doctor`, `flatpak repair`, `apk update`, `winget list`).
  2. Verify environment permissions and user shell context.
  3. Re-run `bluefin-cli install` with verbose flags or `DEBUG=1`.

### 2. Shell Environment & Configuration Errors
- **Symptom:** Terminal prompt, MOTD, or Starship configuration fails to load after `bluefin-cli` initialization.
- **Triage Steps:**
  1. Validate `~/.config/bluefin-cli/` configuration file syntax.
  2. Test shell startup scripts independently (`~/.bashrc`, `~/.zshrc`, PowerShell `$PROFILE`).
  3. Run `bluefin-cli status` to report current profile state and configuration flags.

## Escalation Path

If CLI issues stem from upstream OS distribution shifts or package bundle schema changes:
1. File an issue using the `incident_report.md` template with full CLI output logs attached.
2. Tag relevant maintainers for distribution-specific fixes (Fedora Bluefin, Alpine, macOS, Windows).
