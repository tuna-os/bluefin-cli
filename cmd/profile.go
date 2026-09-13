package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
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

// profileGistFile is the filename used inside the sync gist.
const profileGistFile = "bluefin-profile.json"

var profilePushCmd = &cobra.Command{
	Use:   "push",
	Short: "Sync this machine's profile to a private GitHub gist (via gh)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := exec.LookPath("gh"); err != nil {
			return fmt.Errorf("profile sync uses the GitHub CLI — install gh and run 'gh auth login'")
		}
		p, err := profile.Export(currentShellName())
		if err != nil {
			return err
		}
		tmpFile, err := os.CreateTemp("", "bluefin-profile-*.json")
		if err != nil {
			return err
		}
		tmp := tmpFile.Name()
		_ = tmpFile.Close()
		defer func() { _ = os.Remove(tmp) }()
		if err := p.Save(tmp); err != nil {
			return err
		}

		if id := viper.GetString("profile.gist_id"); id != "" {
			out, err := exec.Command("gh", "gist", "edit", id, "--filename", profileGistFile, tmp).CombinedOutput()
			if err != nil {
				return fmt.Errorf("updating gist %s: %v\n%s", id, err, out)
			}
			fmt.Println(tui.SuccessStyle.Render("✓ Profile pushed to gist " + id))
			return nil
		}

		// gh names the gist file after the file on disk, so the basename has to
		// stay profileGistFile — randomize the parent directory instead.
		tmpDir, err := os.MkdirTemp("", "bluefin-profile-")
		if err != nil {
			return err
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()
		named := filepath.Join(tmpDir, profileGistFile)
		if err := p.Save(named); err != nil {
			return err
		}
		out, err := exec.Command("gh", "gist", "create", "--desc", "bluefin-cli profile", named).CombinedOutput()
		if err != nil {
			return fmt.Errorf("creating gist: %v\n%s", err, out)
		}
		url := strings.TrimSpace(string(out))
		id := path.Base(url)
		viper.Set("profile.gist_id", id)
		if err := config.Save(); err != nil {
			return err
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
		if _, err := exec.LookPath("gh"); err != nil {
			return fmt.Errorf("profile sync uses the GitHub CLI — install gh and run 'gh auth login'")
		}
		id := viper.GetString("profile.gist_id")
		if len(args) > 0 {
			id = args[0]
		}
		if id == "" {
			return fmt.Errorf("no gist configured — run 'profile push' first or pass a gist id")
		}
		out, err := exec.Command("gh", "gist", "view", id, "--filename", profileGistFile, "--raw").Output()
		if err != nil {
			return fmt.Errorf("fetching gist %s: %w", id, err)
		}
		pullFile, err := os.CreateTemp("", "bluefin-profile-pull-*.json")
		if err != nil {
			return err
		}
		tmp := pullFile.Name()
		_ = pullFile.Close()
		defer func() { _ = os.Remove(tmp) }()
		if err := os.WriteFile(tmp, out, 0o600); err != nil {
			return err
		}

		p, err := profile.Load(tmp)
		if err != nil {
			return err
		}
		if err := p.Apply(); err != nil {
			return err
		}
		if len(args) > 0 && viper.GetString("profile.gist_id") == "" {
			viper.Set("profile.gist_id", id)
			_ = config.Save()
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Profile pulled and applied."))
		return nil
	},
}
