package cmd

import (
	"fmt"
	"io"
	"strings"

	"charm.land/fang/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

// Fang v2.0.1 wraps stderr in a colorprofile.Writer, hiding the file from
// its non-terminal check. Its styled error text also changes the first word
// and reflows multiline messages. Keep the diagnostic itself verbatim.
func renderError(w io.Writer, styles fang.Styles, err error) {
	terminal := isTerminalErrorWriter(w)
	if terminal {
		_, _ = fmt.Fprintln(w, styles.ErrorHeader.String())
	}

	message := err.Error()
	_, _ = fmt.Fprint(w, message)
	if !strings.HasSuffix(message, "\n") {
		_, _ = fmt.Fprintln(w)
	}
	if isUsageError(message) {
		_, _ = fmt.Fprintln(w, "Try --help for usage.")
	}
}

func isTerminalErrorWriter(w io.Writer) bool {
	if wrapped, ok := w.(*colorprofile.Writer); ok {
		w = wrapped.Forward
	}
	f, ok := w.(term.File)
	return ok && term.IsTerminal(f.Fd())
}

// Fang's usage predicate is unexported. Include its Cobra/pflag prefixes as
// well as Cobra's argument-count and flag-group validation errors.
func isUsageError(message string) bool {
	for _, prefix := range []string{
		"flag needs an argument:",
		"unknown flag:",
		"unknown shorthand flag:",
		"unknown command",
		"invalid argument",
		"if any flags in the group ",
		"at least one of the flags in the group ",
		"required flag",
	} {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	// Cobra's positional argument validators include this phrase; avoid
	// treating arbitrary domain errors starting with "accepts" as usage.
	return (strings.HasPrefix(message, "accepts ") || strings.HasPrefix(message, "requires at least ")) &&
		strings.Contains(message, " arg(s), ")
}
