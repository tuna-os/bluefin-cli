package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [shell] [on|off]",
	Short: "Toggle shell experience enhancements",
	Long: `Enable or disable shell experience enhancements (modern aliases and tool initialization).
	
The Shell Experience provides:
  - Modern ls replacement with eza (ll, ls aliases)
  - bat for cat with syntax highlighting
  - ugrep for faster grep
  - Initialization for atuin, starship, and zoxide`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runShellMenu()
		}

		// Args provided
		selectedShell := args[0]
		enable := true // default to on
		if len(args) > 1 {
			enable = args[1] == "on"
		}

		return shell.Toggle(selectedShell, enable)
	},
}

var shellConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure individual shell experience tools",
	Long:  `Enable or disable specific shell experience components interactively.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return configureShellTools()
	},
}

func runShellMenu() error {
	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Main Menu > Shell")

		currentShellPath := os.Getenv("SHELL")
		currentShell := filepath.Base(currentShellPath)
		if currentShell == "" || currentShell == "." {
			if env.IsWindows() {
				currentShell = "powershell"
			} else {
				currentShell = "bash"
			}
		}

		status := shell.CheckStatus()
		isEnabled := status[currentShell]
		toggleLabel := fmt.Sprintf("Enable for current shell (%s)", currentShell)
		if isEnabled {
			toggleLabel = fmt.Sprintf("Disable for current shell (%s)", currentShell)
		}

		var action string
		componentsLabel := "Configure Components ❯"
		motdLabel := "📰 MOTD Settings ❯"
		shellsLabel := "Enable/Disable for other shells ❯"
		if env.IsWindows() {
			componentsLabel = "Configure Components >"
			motdLabel = "MOTD Settings >"
			shellsLabel = "Enable/Disable for other shells >"
		}

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose an option").
					Options(
						huh.NewOption(toggleLabel, "toggle_current"),
						huh.NewOption(componentsLabel, "components"),
						huh.NewOption(motdLabel, "motd"),
						huh.NewOption(shellsLabel, "shells"),
						huh.NewOption("Exit to Main Menu", "exit"),
					).
					Value(&action),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
			return nil
		}

		switch action {
		case "toggle_current":
			if err := shell.Toggle(currentShell, !isEnabled); err != nil {
				return err
			}
			tui.Pause()
		case "shells":
			if err := shellShellsMenu(); err != nil {
				return err
			}
		case "components":
			if err := configureShellTools(); err != nil {
				return err
			}
		case "motd":
			if err := runMotdMenu(); err != nil {
				return err
			}
		case "exit":
			return nil
		}
	}
}

func shellShellsMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Main Menu > Shell > Shells")

	status := shell.CheckStatus()
	installedShells := shell.GetInstalledShells()

	var selected []string
	for _, s := range installedShells {
		if status[s] {
			selected = append(selected, s)
		}
	}

	initialSelected := make(map[string]bool)
	for _, sh := range selected {
		initialSelected[sh] = true
	}

	// Build options dynamically based on installed shells
	var options []huh.Option[string]
	for _, s := range installedShells {
		options = append(options, huh.NewOption(s, s))
	}

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Manage other shells").
				Description("Selected = ON, Deselected = OFF").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
		return nil // Interrupted - go back to main menu
	}

	finalSelected := make(map[string]bool)
	for _, sh := range selected {
		finalSelected[sh] = true
	}

	for _, shName := range installedShells {
		wasEnabled := initialSelected[shName]
		isEnabled := finalSelected[shName]

		if wasEnabled != isEnabled {
			if err := shell.Toggle(shName, isEnabled); err != nil {
				return err
			}
			tui.Pause()
		}
	}
	return nil
}

func configureShellTools() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Main Menu > Shell > Components")

	currentShellPath := os.Getenv("SHELL")
	currentShell := filepath.Base(currentShellPath)
	if currentShell == "" || currentShell == "." {
		if env.IsWindows() {
			currentShell = "powershell"
		} else {
			currentShell = "bash"
		}
	}

	cfg, err := shell.LoadConfig(currentShell)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	availableTools := shell.ToolsForShell(currentShell)

	var selected []string
	for _, tool := range availableTools {
		if cfg.IsEnabled(tool.Name) {
			selected = append(selected, tool.Name)
		}
	}

	var options []huh.Option[string]
	for _, tool := range availableTools {
		label := fmt.Sprintf("%s (%s)", tool.Name, tool.Description)
		options = append(options, huh.NewOption(label, tool.Name))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select tools to enable").
				Description("Uncheck to disable specific tools").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	newCfg := shell.DefaultConfig(currentShell)
	selectedSet := make(map[string]bool)
	for _, s := range selected {
		selectedSet[s] = true
	}

	for _, tool := range availableTools {
		newCfg.SetEnabled(tool.Name, selectedSet[tool.Name])
	}

	if err := shell.SaveConfig(newCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Install any newly enabled tools
	shell.InstallTools(currentShell, newCfg)

	fmt.Println(tui.SuccessStyle.Render("Configuration saved! Tools installed/updated."))
	tui.Pause()
	return nil
}

func init() {
	describedTools := shell.Tools
	if env.IsWindows() {
		describedTools = shell.ToolsForShell("powershell")
	}

	// Generate dynamic long description
	var sb strings.Builder
	sb.WriteString("Enable or disable shell experience enhancements (modern aliases and tool initialization).\n\nThe Shell Experience provides:\n")
	for _, tool := range describedTools {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", tool.Name, tool.Description))
	}
	shellCmd.Long = sb.String()

	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellConfigCmd)
}
