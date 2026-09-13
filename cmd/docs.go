package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsDest string

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation for bluefin-cli",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(docsDest, 0755); err != nil {
			return fmt.Errorf("failed to create docs directory: %w", err)
		}

		// Without this, cobra stamps "Auto generated ... on <date>" into the
		// footer of every page, so any regeneration is a whole-tree diff whose
		// real content change is buried in 40-odd date bumps. That churn is why
		// the docs kept drifting and being re-reported as stale
		// (tuna-os/bluefin-cli#238, #244, #246): nobody could see at a glance
		// whether a regeneration had actually changed anything. With the tag
		// off, `docs --dest` is deterministic, and CI can simply check that
		// regenerating produces no diff.
		rootCmd.DisableAutoGenTag = true

		fmt.Printf("Generating documentation in %s...\n", docsDest)
		if err := doc.GenMarkdownTree(rootCmd, docsDest); err != nil {
			return fmt.Errorf("failed to generate markdown: %w", err)
		}

		fmt.Println("Documentation generated successfully!")
		return nil
	},
}

func init() {
	docsCmd.Flags().StringVarP(&docsDest, "dest", "d", "./docs/commands", "Destination directory for generated docs")
	rootCmd.AddCommand(docsCmd)
}
