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
