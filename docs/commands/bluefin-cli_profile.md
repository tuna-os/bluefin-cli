## bluefin-cli profile

Export or import your bluefin-cli setup

### Synopsis

Capture this machine's setup (enabled shells, tool selection, theme
flavor) as a portable JSON document, and replay it on another machine:

  bluefin-cli profile export > my-setup.json
  bluefin-cli profile import my-setup.json

### Options

```
  -h, --help   help for profile
```

### SEE ALSO

* [bluefin-cli](bluefin-cli.md)	 — CLI tool for Homebrew and shell setup
* [bluefin-cli profile diff](bluefin-cli_profile_diff.md)	 — Show what import would change (drift from a saved profile)
* [bluefin-cli profile export](bluefin-cli_profile_export.md)	 — Write the current setup as JSON (stdout by default)
* [bluefin-cli profile import](bluefin-cli_profile_import.md)	 — Apply a setup file to this machine
* [bluefin-cli profile pull](bluefin-cli_profile_pull.md)	 — Get and apply the profile from its gist
* [bluefin-cli profile push](bluefin-cli_profile_push.md)	 — Send this machine's profile to a private GitHub gist (via gh)

