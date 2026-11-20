## bluefin-cli motd show

Display the MOTD

### Synopsis

Display the Message of the Day with system information and a random tip.

Where uwelcome (the Bluefin and Universal Blue terminal greeting) is installed,
this shows uwelcome's greeting. Otherwise the message comes from the first
template that exists. The order is: template-file in motd.json,
/etc/bluefin-cli/motd.md, /usr/share/bluefin-cli/motd.md, then the built-in
one. Tips come the same way from tips-directory, /etc/bluefin-cli/tips and
/usr/share/bluefin-cli/tips.

With --if-enabled it prints nothing when the MOTD is off
('bluefin-cli motd toggle all off', or uwelcome's own switch), for login hooks
that greet every terminal.

```
bluefin-cli motd show [flags]
```

### Options

```
  -h, --help         help for show
      --if-enabled   Show nothing when the MOTD is off
```

### SEE ALSO

* [bluefin-cli motd](bluefin-cli_motd.md)	 — Set up daily shell messages

