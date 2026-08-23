# Contributing

Thanks for your interest in contributing! This project is part of the [TunaOS](https://tunaos.org) ecosystem.

## Getting Started

1. Fork the repo and clone it locally.
2. Install Go 1.25.8 or later and [`just`](https://just.systems/). Install
   Podman if you want to use the container-based recipes.
3. Open an issue to discuss your change before submitting a PR.

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

For the containerized integration suite, run:

```bash
just test
```

If a change affects commands or flags, regenerate the command reference and
include the resulting Markdown changes:

```bash
just gen-docs
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
