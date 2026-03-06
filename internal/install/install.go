package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
)

const (
	commonBaseURL   = "https://raw.githubusercontent.com/projectbluefin/common/main/system_files"
	defaultBrewPath = "shared/usr/share/ublue-os/homebrew"
)

type BundleSpec struct {
	File        string
	Description string
	Path        string // Optional: override defaultBrewPath
}

var bundles = map[string]BundleSpec{
	"ai": {
		File:        "ai-tools.Brewfile",
		Description: "AI tools: Goose, Codex, Gemini, Ramalama, etc.",
	},
	"cli": {
		File:        "cli.Brewfile",
		Description: "CLI essentials: GitHub CLI, chezmoi, etc.",
	},
	"cncf": {
		File:        "cncf.Brewfile",
		Description: "Cloud Native Computing Foundation tools.",
	},
	"experimental-ide": {
		File:        "experimental-ide.Brewfile",
		Description: "Experimental IDE tools.",
	},
	"fonts": {
		File:        "fonts.Brewfile",
		Description: "Development fonts: Fira Code, JetBrains Mono, etc.",
	},
	"full-desktop": {
		File:        "full-desktop.Brewfile",
		Description: "Full GNOME Desktop apps.",
		Path:        "bluefin/usr/share/ublue-os/homebrew",
	},
	"ide": {
		File:        "ide.Brewfile",
		Description: "IDE tools: VS Code, JetBrains Toolbox, etc.",
	},
	"k8s": {
		File:        "k8s-tools.Brewfile",
		Description: "Kubernetes tools: kubectl, k9s, kubectx, etc.",
	},
}

func Bundle(nameOrPath string) error {
	if runtime.GOOS == "windows" {
		return BundleWindows(nameOrPath)
	}

	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Please install Homebrew first: https://brew.sh")
	}

	brewfilePath, cleanup, err := GetBrewfile(nameOrPath)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := EnsureBbrew(); err != nil {
		return err
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("🍺 Opening %s in bbrew...", brewfilePath)))

	if err := RunBbrew(brewfilePath); err != nil {
		return fmt.Errorf("bbrew failed: %w", err)
	}

	return nil
}

func GetBrewfile(nameOrPath string) (string, func(), error) {
	if strings.Contains(nameOrPath, "/") || strings.Contains(nameOrPath, "\\") {
		if _, err := os.Stat(nameOrPath); os.IsNotExist(err) {
			return "", func() {}, fmt.Errorf("Brewfile not found: %s", nameOrPath)
		}
		return nameOrPath, func() {}, nil
	}

	bundle, ok := bundles[nameOrPath]
	if !ok {
		return "", func() {}, fmt.Errorf("unknown bundle: %s (available: ai, cli, cncf, experimental-ide, fonts, full-desktop, ide, k8s, all)", nameOrPath)
	}

	if nameOrPath == "full-desktop" {
		if err := EnsureFlathub(); err != nil {
			return "", func() {}, err
		}
	}

	path := defaultBrewPath
	if bundle.Path != "" {
		path = bundle.Path
	}

	url := fmt.Sprintf("%s/%s/%s", commonBaseURL, path, bundle.File)
	tmpDir := os.TempDir()
	brewfilePath := filepath.Join(tmpDir, bundle.File)

	fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Downloading %s bundle...", nameOrPath)))

	if err := downloadFile(url, brewfilePath); err != nil {
		return "", func() {}, fmt.Errorf("failed to download bundle: %w", err)
	}

	cleanup := func() {
		os.Remove(brewfilePath)
	}

	return brewfilePath, cleanup, nil
}

func MergeBrewfiles(paths []string) (string, func(), error) {
	if len(paths) == 0 {
		return "", func() {}, fmt.Errorf("no brewfiles to merge")
	}

	tmpDir := os.TempDir()
	mergedPath := filepath.Join(tmpDir, "merged.Brewfile")

	f, err := os.Create(mergedPath)
	if err != nil {
		return "", func() {}, err
	}
	defer f.Close()

	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			return "", func() {}, err
		}
		if _, err := f.Write(content); err != nil {
			return "", func() {}, err
		}
		if _, err := f.WriteString("\n"); err != nil {
			return "", func() {}, err
		}
	}

	cleanup := func() {
		os.Remove(mergedPath)
	}

	return mergedPath, cleanup, nil
}

func CheckBbrew() error {
	_, err := exec.LookPath("bbrew")
	return err
}

func EnsureBbrew() error {
	if err := CheckBbrew(); err == nil {
		return nil
	}

	fmt.Println(infoStyle.Render("🍺 bbrew not found, installing..."))
	cmd := exec.Command("brew", "install", "Valkyrie00/homebrew-bbrew/bbrew")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install bbrew: %w", err)
	}
	return nil
}

func RunBbrew(brewfilePath string) error {
	cmd := exec.Command("bbrew", "-f", brewfilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ListBundles() {
	fmt.Println(titleStyle.Render("📦 Available Bundles"))
	fmt.Println()

	for name, bundle := range bundles {
		fmt.Printf("  %s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render(name+":"),
			bundle.Description)
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Usage:"))
	fmt.Println("  bluefin-cli install <bundle-name>")
	fmt.Println("  bluefin-cli install /path/to/Brewfile")
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsGnome() bool {
	xdgCurrentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	return strings.Contains(strings.ToUpper(xdgCurrentDesktop), "GNOME")
}

func CheckFlatpak() error {
	_, err := exec.LookPath("flatpak")
	return err
}

func EnsureFlathub() error {
	if err := CheckFlatpak(); err != nil {
		return fmt.Errorf("flatpak not found. Please install flatpak first: https://flatpak.org/setup/")
	}

	cmd := exec.Command("flatpak", "remote-list")
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "flathub") {
		return nil
	}

	fmt.Println(infoStyle.Render("Adding Flathub remote..."))
	addCmd := exec.Command("flatpak", "remote-add", "--if-not-exists", "flathub", "https://dl.flathub.org/repo/flathub.flatpakrepo")
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	return addCmd.Run()
}
