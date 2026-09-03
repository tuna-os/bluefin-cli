# Bluefin CLI Observability Posture & Operational Guidelines

## Overview

`bluefin-cli` is a local workstation environment setup tool and CLI utility written in Go. As a client-side command-line tool, it operates locally on user systems (Linux, macOS, WSL, Windows) without running as a persistent daemon or shipping telemetry to external servers.

## Current Telemetry Posture

- **Zero Exporter Footprint:** `bluefin-cli` does NOT collect, export, or send metrics, traces, or diagnostic data to remote endpoints or cloud backends.
- **Local Operational Logging:** Diagnostics, command status, and shell environment initialization errors are written locally to standard output / standard error streams (`stdout`/`stderr`) or local log files under standard user directories.
- **CountMe Integration:** Anonymous countme requests (if enabled) strictly conform to minimal privacy guidelines and do not track user behavior or system diagnostics.

## Operational Diagnostics & Triage

When diagnosing issues with `bluefin-cli` execution or installation:

1. **Terminal Diagnostics:** Run `bluefin-cli status` or set `DEBUG=1` in the environment to inspect verbosity and execution paths.
2. **Local Log Location:** Inspect local logs in `~/.config/bluefin-cli/logs/` or OS-specific standard application data locations.
3. **Exit Codes & Error Output:** CLI subcommands exit with non-zero status codes upon failure, emitting error details to `stderr`.

## Future Observability Recommendations

If backend or client-side telemetry is evaluated in future milestones:
- Maintain user opt-in control over telemetry collection.
- Preserve zero-external-flow defaults.
- Document any local log retention policies.
