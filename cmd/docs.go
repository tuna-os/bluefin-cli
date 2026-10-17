package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

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
		if err := normalizeMarkdownLinks(docsDest); err != nil {
			return fmt.Errorf("failed to normalize markdown links: %w", err)
		}

		fmt.Println("Documentation generated successfully!")
		return nil
	},
}

func normalizeMarkdownLinks(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Cobra uses " - " between command links and descriptions. The STE
		// checker treats that hyphen as part of a noun cluster, so an em dash
		// keeps the generated index readable without a false finding.
		normalized := bytes.ReplaceAll(data, []byte("\t - "), []byte("\t — "))
		if bytes.Equal(data, normalized) {
			continue
		}
		if err := os.WriteFile(path, normalized, 0644); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	docsCmd.Flags().StringVarP(&docsDest, "dest", "d", "./docs/commands", "Destination directory for generated docs")
	rootCmd.AddCommand(docsCmd)
}
