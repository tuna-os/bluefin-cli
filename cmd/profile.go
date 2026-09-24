package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/profile"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Export or import your bluefin-cli setup",
	Long: `Capture this machine's setup (enabled shells, tool selection, theme
flavor) as a portable JSON document, and replay it on another machine:

  bluefin-cli profile export > my-setup.json
  bluefin-cli profile import my-setup.json`,
}

var profileExportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Write the current setup as JSON (stdout by default)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := profile.Export(currentShellName())
		if err != nil {
			return err
		}
		path := "-"
		if len(args) > 0 {
			path = args[0]
		}
		if err := p.Save(path); err != nil {
			return err
		}
		if path != "-" {
			fmt.Println(tui.SuccessStyle.Render("✓ Profile exported to " + path))
		}
		return nil
	},
}

var profileImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Apply a previously exported setup to this machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := profile.Load(args[0])
		if err != nil {
			return err
		}
		if err := p.Apply(); err != nil {
			return err
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Profile applied."))
		return nil
	},
}

func init() {
	profileDiffCmd.Flags().Bool("exit-code", false, "Exit non-zero when drift exists (for scripts)")
	profileCmd.AddCommand(profileExportCmd)
	profileCmd.AddCommand(profileImportCmd)
	profileCmd.AddCommand(profileDiffCmd)
	profileCmd.AddCommand(profilePushCmd)
	profileCmd.AddCommand(profilePullCmd)
	rootCmd.AddCommand(profileCmd)
}

var profileDiffCmd = &cobra.Command{
	Use:   "diff <file>",
	Short: "Show what import would change (drift from a saved profile)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		want, err := profile.Load(args[0])
		if err != nil {
			return err
		}
		current, err := profile.Export(currentShellName())
		if err != nil {
			return err
		}
		changes := profile.Diff(current, want)
		if len(changes) == 0 {
			fmt.Println(tui.SuccessStyle.Render("✓ No drift — this machine matches the profile."))
			return nil
		}
		fmt.Println("Import would change:")
		for _, c := range changes {
			fmt.Printf("  %s\n", c)
		}
		if exitCode, _ := cmd.Flags().GetBool("exit-code"); exitCode {
			return fmt.Errorf("%d difference(s)", len(changes))
		}
		return nil
	},
}

var profilePushCmd = &cobra.Command{
	Use:   "push",
	Short: "Sync this machine's profile to a private GitHub gist (via gh)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := profile.NewGitHubCLIClient()
		if err != nil {
			return err
		}
		p, err := profile.Export(currentShellName())
		if err != nil {
			return err
		}
		sync := profile.NewSync(client, profile.ConfigSyncIDStore{})
		wasConfigured := profile.ConfigSyncIDStore{}.Get() != ""
		id, err := sync.Push(p)
		if err != nil {
			return err
		}
		if wasConfigured {
			fmt.Println(tui.SuccessStyle.Render("✓ Profile pushed to gist " + id))
			return nil
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Profile pushed to new private gist " + id + " (saved in config)"))
		return nil
	},
}

var profilePullCmd = &cobra.Command{
	Use:   "pull [gist-id]",
	Short: "Fetch and apply the synced profile from its gist",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := profile.NewGitHubCLIClient()
		if err != nil {
			return err
		}
		id := ""
		if len(args) > 0 {
			id = args[0]
		}
		p, err := profile.NewSync(client, profile.ConfigSyncIDStore{}).Fetch(id)
		if err != nil {
			return err
		}
		if err := p.Apply(); err != nil {
			return err
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Profile pulled and applied."))
		return nil
	},
}
