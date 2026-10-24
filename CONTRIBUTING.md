# Contributing

Thanks for your interest in this project! This project is part of the [TunaOS](https://tunaos.org) ecosystem.

<!-- hive-contribute-plea: donated-compute appeal, keep in sync across repos -->
## Contribute compute — no code needed

**[Donate time with your AI CLI](https://hub.tunaos.org/#contribute).**

Your machine runs one task at a time using your signed-in AI coding CLI.
The relay alternates between eligible **School** (OS) and **Reef** (apps)
hives after each task. If one hive has no work, the other can continue.

The guide covers prerequisites, verified `just` setup, credentials, provider
costs, and how to stop with Ctrl+C. Your subscription quota or API charges apply.
Never post account credentials or API keys in issues or pull requests.

## Getting Started

1. Fork the repo and clone it locally.
2. Install Go 1.26.0 or later (see the `go` directive in `go.mod`) and
   [`just`](https://just.systems/). Install
   Podman if you want to use the container-based recipes.
3. Open an issue to discuss your change before you submit a PR.

## Build and validate

Build both the standard and plus variants:

```bash
just build
```

Run the local test suite and the race-enabled configuration used in CI:

```bash
go test ./...
go test -tags extra -race ./...
```

Run the lint checks required by CI:
```bash
just lint
```
For the containerized integration suite, run:

```bash
just test
```

If a change affects commands or flags, regenerate the command reference from the
extra build. The extra build includes both published binaries. Include the new
Markdown in your PR:

```bash
go build -tags extra -o bluefin-cli .
./bluefin-cli docs --dest ./docs/commands
```

Run `just --list` to see the remaining development and inspection recipes.

## Pull Requests

- Keep PRs focused — one change per PR.
- Follow the existing code style and conventions.
- Update docs if your change affects usage.
- Run `gofmt` (or `just fmt`) on changed Go files.
- Confirm `go mod tidy -diff` is clean when dependencies change.

## Questions?

- [TunaOS Documentation](https://tunaos.org)
- [bluefin-cli GitHub Issues](https://github.com/tuna-os/bluefin-cli/issues)
