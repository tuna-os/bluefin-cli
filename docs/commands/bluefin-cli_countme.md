## bluefin-cli countme

Manage the anonymous count of installs

### Synopsis

bluefin-cli participates in Fedora's countme protocol to report
anonymous install counts alongside native Bluefin Linux installs.

Each week, bluefin-cli sends a single GET request to Fedora's mirrors,
with a User-Agent that identifies the platform (mac, wsl, powershell).
It sends no personal data, no IP addresses, and no machine identifiers.
The aggregate data is publicly available at:
  https://data-analysis.fedoraproject.org/csv-reports/countme/totals.csv

Opt out at any time:
  bluefin-cli countme --disable
  # or permanently via environment:
  export BLUEFIN_DISABLE_COUNTME=1

```
bluefin-cli countme [flags]
```

### Options

```
      --disable   Persistently opt out of anonymous usage counting
      --enable    Re-enable anonymous usage counting after opting out
  -h, --help      help for countme
      --status    Show current countme configuration (default behaviour)
```

### SEE ALSO

* [`bluefin-cli`](bluefin-cli.md) — A CLI tool to manage Homebrew and customize your shell

