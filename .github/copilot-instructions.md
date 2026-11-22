# Copilot instructions

Read [`AGENTS.md`](../AGENTS.md) before you change this repository. It
describes the architecture, the two build variants and the development
workflow. [`CLAUDE.md`](../CLAUDE.md) lists known traps, and
[`CONTEXT.md`](../CONTEXT.md) defines the domain terms.

## Build and test

```bash
go build ./...                    # standard variant (bluefin-cli)
go build -tags extra ./...        # plus variant (bluefin-cli-plus)
go test -tags extra ./...         # the cmd tests need -tags extra
```

Both variants must compile before you commit.

## Command reference

`docs/commands/` has one page for each command. A generator writes these pages
from the Cobra command tree, so do not edit them by hand. After you change a
command or a flag, generate the pages from the plus variant and commit the
result:

```bash
go build -tags extra -o bluefin-cli .
./bluefin-cli docs --dest ./docs/commands
```

CI fails when `docs/commands/` is not current.
