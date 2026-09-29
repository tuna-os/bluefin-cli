package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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
		if err := codeCommandLinks(docsDest); err != nil {
			return fmt.Errorf("failed to format command links: %w", err)
		}

		fmt.Println("Documentation generated successfully!")
		return nil
	},
}

func init() {
	docsCmd.Flags().StringVarP(&docsDest, "dest", "d", "./docs/commands", "Destination directory for generated docs")
	rootCmd.AddCommand(docsCmd)
}

// seeAlsoEntry matches a SEE ALSO entry that cobra writes, such as
// "* [bluefin-cli shell](bluefin-cli_shell.md)\t - Toggle ...".
var seeAlsoEntry = regexp.MustCompile("(?m)^\\* \\[([^\\]`]+)\\]\\(([^)]+)\\)\t - ")

// codeCommandLinks sets each command name in a SEE ALSO list as inline code,
// and puts an em dash between the name and its summary. A command name is a
// symbol, not prose. Without this, the STE check (.github/workflows/ste.yml)
// reads "bluefin-cli shell - Toggle" as one long noun cluster, and reports it
// again on every page that links to that command.
func codeCommandLinks(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		out := seeAlsoEntry.ReplaceAll(data, []byte("* [`$1`]($2) — "))
		if err := os.WriteFile(file, out, 0644); err != nil {
			return err
		}
	}
	return nil
}
