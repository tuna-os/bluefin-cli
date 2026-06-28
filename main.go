package main

import (
	"fmt"
	"os"

	"github.com/tuna-os/bluefin-cli/cmd"
	"github.com/tuna-os/bluefin-cli/internal/config"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize configuration: %v\n", err)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
