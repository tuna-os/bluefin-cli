package shell

import "regexp"

// Universal Blue's ublue-bling (projectbluefin/common) enabled its shell
// setup by appending a marked block to the rc file:
//
//	### bling.sh source start
//	test -f /usr/share/ublue-os/bling/bling.sh && source /usr/share/ublue-os/bling/bling.sh
//	### bling.sh source end
//
// Older releases wrote the source line alone, and users copied it by hand.
var (
	ublueBlingBlock = regexp.MustCompile(`(?m)^### bling\.(?:sh|fish) source start\n(?:.*\n)*?### bling\.(?:sh|fish) source end(?:\n|$)`)
	ublueBlingLine  = regexp.MustCompile(`(?m)^.*(?:source|\.) /usr/share/ublue-os/bling/bling\.(?:sh|fish).*(?:\n|$)`)
)

// stripUblueBling returns rc without the lines that run Universal Blue's
// bling.sh or bling.fish, which the shell experience replaces.
func stripUblueBling(rc string) string {
	return ublueBlingLine.ReplaceAllString(ublueBlingBlock.ReplaceAllString(rc, ""), "")
}
