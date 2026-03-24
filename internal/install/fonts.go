package install

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// DefaultFont returns the platform-appropriate default Nerd Font suggestion.
// Windows users get Caskaydia Cove (pairs well with Windows Terminal's Cascadia Code default).
// macOS / Linux users get JetBrains Mono.
func DefaultFont() NerdFont {
	if runtime.GOOS == "windows" {
		return CaskaydiaCove
	}
	for _, f := range NerdFonts {
		if f.Name == "JetBrains Mono Nerd Font" {
			return f
		}
	}
	return NerdFonts[0]
}

// NerdFont describes a Nerd Font with its installation identifiers per package manager.
type NerdFont struct {
	Name     string
	BrewCask string
	WingetID string // empty means not available on winget
	ScoopID  string
	Face     string // font face name for terminal configuration
}

// CaskaydiaCove is the default Windows font — pairs well with Windows Terminal's Cascadia Code default.
var CaskaydiaCove = NerdFont{
	Name:     "Caskaydia Cove Nerd Font",
	BrewCask: "font-caskaydia-cove-nerd-font",
	WingetID: "", // not on winget — use scoop
	ScoopID:  "CascadiaCode-NF-Mono",
	Face:     "CaskaydiaCove NF Mono",
}

// NerdFonts is the curated list of installable Nerd Fonts.
var NerdFonts = []NerdFont{
	CaskaydiaCove,
	{
		Name:     "JetBrains Mono Nerd Font",
		BrewCask: "font-jetbrains-mono-nerd-font",
		WingetID: "DEVCOM.JetBrainsMonoNerdFont",
		ScoopID:  "JetBrainsMono-NF",
		Face:     "JetBrainsMono NF Mono",
	},
	{
		Name:     "Fira Code Nerd Font",
		BrewCask: "font-fira-code-nerd-font",
		WingetID: "RyanLano.FiraCodeNerdFont",
		ScoopID:  "FiraCode-NF",
		Face:     "FiraCode NF Mono",
	},
	{
		Name:     "Hack Nerd Font",
		BrewCask: "font-hack-nerd-font",
		WingetID: "SourceFoundry.Hack.NerdFont",
		ScoopID:  "Hack-NF",
		Face:     "Hack NF Mono",
	},
	{
		Name:     "Meslo Nerd Font",
		BrewCask: "font-meslo-lg-nerd-font",
		WingetID: "Meslo.MesloLGNerdFont",
		ScoopID:  "Meslo-NF",
		Face:     "MesloLGS NF",
	},
	{
		Name:     "Ubuntu Nerd Font",
		BrewCask: "font-ubuntu-nerd-font",
		WingetID: "Ubuntu.UbuntuNerdFont",
		ScoopID:  "Ubuntu-NF",
		Face:     "Ubuntu NF Mono",
	},
}

// InstallFont installs a Nerd Font using the appropriate package manager for the current OS.
func InstallFont(font NerdFont) error {
	if runtime.GOOS == "windows" {
		return installFontWindows(font)
	}
	return installFontUnix(font)
}

func installFontWindows(font NerdFont) error {
	// Try Scoop first — it has a dedicated nerd-fonts bucket
	if _, err := exec.LookPath("scoop"); err == nil {
		fmt.Printf("  Installing %s via scoop...\n", font.Name)
		exec.Command("scoop", "bucket", "add", "nerd-fonts").Run() //nolint:errcheck
		cmd := exec.Command("scoop", "install", font.ScoopID)
		if err := cmd.Run(); err == nil {
			return nil
		}
		fmt.Printf("  Scoop failed for %s, trying winget...\n", font.Name)
	}

	if font.WingetID == "" {
		return fmt.Errorf("scoop not found and no winget package available for %s", font.Name)
	}

	if _, err := exec.LookPath("winget"); err != nil {
		return fmt.Errorf("neither scoop nor winget found — install scoop from https://scoop.sh")
	}

	fmt.Printf("  Installing %s via winget...\n", font.Name)
	cmd := exec.Command("winget", "install", "--id", font.WingetID, "--exact",
		"--accept-source-agreements", "--accept-package-agreements", "--silent")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", font.Name, err)
	}
	return nil
}

func installFontUnix(font NerdFont) error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not found")
	}
	fmt.Printf("  Installing %s via brew...\n", font.Name)
	cmd := exec.Command("brew", "install", "--cask", font.BrewCask)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", font.Name, err)
	}
	return nil
}

// IsFontInstalled returns true if the given font is already installed.
func IsFontInstalled(font NerdFont) bool {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("scoop"); err == nil {
			cmd := exec.Command("scoop", "list", font.ScoopID)
			if err := cmd.Run(); err == nil {
				return true
			}
		}
		if font.WingetID != "" {
			if _, err := exec.LookPath("winget"); err == nil {
				cmd := exec.Command("winget", "list", "--id", font.WingetID, "--exact")
				out, err := cmd.Output()
				return err == nil && strings.Contains(strings.ToLower(string(out)), strings.ToLower(font.WingetID))
			}
		}
		return false
	}
	cmd := exec.Command("brew", "list", "--cask", font.BrewCask)
	return cmd.Run() == nil
}
