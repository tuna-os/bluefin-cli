package motd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

const motdMarker = "# bluefin-cli motd"

var defaultTips = []string{
	"Use `brew search` and `brew install` to install packages. Homebrew will take care of updates automatically",
	"`tldr vim` will give you the basic rundown on commands for a given tool",
	"Performance profiling tools are built-in: try `top`, `htop`, and other debugging tools",
	"Switch shells safely: change your shell in Terminal settings instead of system-wide",
	"Container development is OS-agnostic - your devcontainers work on Linux, macOS, and Windows",
	"Use `docker compose` for multi-container development if devcontainers don't fit your workflow",
	"Bluefin separates the OS from your development environment - embrace the cloud-native workflow",
	"Check out DevPod for open-source, client-only development environments that work with any IDE",
	"Develop with devcontainers! Use `devcontainer.json` files in your projects for isolated, reproducible environments",
	"VS Code comes with devcontainers extension pre-installed - perfect for containerized development",
	"Use `eza -l --icons` for a beautiful file listing with icons and colors",
	"The `bat` command is like `cat` but with syntax highlighting and Git integration",
	"Navigate directories faster with `zoxide` - just use `z <partial-name>` to jump around",
	"Search your shell history with `atuin` using Ctrl+R for a better history search experience",
	"Customize your prompt with `starship config` to modify colors, icons, and modules",
}

var defaultTemplate = `# 󱍢 Welcome to Bluefin CLI
󱋩 %s:%s

|  Command | Description |
| ------- | ----------- |
| ` + "`bluefin-cli shell bash on`" + `  | Enable shell experience for bash  |
| ` + "`bluefin-cli status`" + ` | Show current configuration |
| ` + "`bluefin-cli help`" + ` | Show all available commands |
| ` + "`brew help`" + ` | Manage command line packages |
| ` + "`brew search <query>`" + ` | Search for packages |

%s

- **󰊤** [GitHub Issues](https://github.com/tuna-os/bluefin-cli/issues)
- **󰈙** [Documentation](https://github.com/tuna-os/bluefin-cli)
`

type ImageInfo struct {
	ImageName     string `json:"image-name"`
	ImageTag      string `json:"image-tag"`
	ImageFlavor   string `json:"image-flavor"`
	ImageVendor   string `json:"image-vendor"`
	FedoraVersion string `json:"fedora-version"`
}

type Config struct {
	TipsDirectory   string `json:"tips-directory"`
	CheckOutdated   string `json:"check-outdated"`
	ImageInfoFile   string `json:"image-info-file"`
	DefaultTheme    string `json:"default-theme"`
	TemplateFile    string `json:"template-file"`
	ThemesDirectory string `json:"themes-directory"`
}

// Toggle enables or disables MOTD for shells
// Deprecated: Use 'bluefin-cli init' instead
func Toggle(target string, enable bool) error {
	fmt.Println(tui.InfoStyle.Render("ℹ Note: shell integration is now handled via 'bluefin-cli init'"))

	// For backward compatibility cleanup, we could implemented removal here,
	// but for now, we just inform the user.
	return nil
}

// Show displays the MOTD
func Show() error {
	// Get OS info
	info := getImageInfo()

	// Get configuration (defaults if file missing)
	config, err := loadConfig()
	if err != nil {
		// If error loading config, just use defaults
		config = DefaultConfig()
	}

	// Get random tip
	// First check if user has custom tips in the config directory
	var tip string
	if config.TipsDirectory != "" {
		// Try to find custom tips
		customTip := getRandomTipFromDir(config.TipsDirectory)
		if customTip != "" {
			tip = customTip
		}
	}

	// Fallback to default tips if no custom tip found
	if tip == "" {
		tip = getRandomDefaultTip()
	}

	// Format template
	content := fmt.Sprintf(defaultTemplate, info.ImageName, info.ImageTag, tip)

	// Render with glow if available
	if glowPath, err := exec.LookPath("glow"); err == nil {
		cmd := exec.Command(glowPath, "-s", "dark", "-w", "80", "-")
		cmd.Stdin = strings.NewReader(content)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to plain text
	fmt.Println(content)
	return nil
}

// SetTheme sets the MOTD theme
func SetTheme(theme string) error {
	configDir, err := env.EnsureConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "motd.json")

	config, err := loadConfig()
	if err != nil {
		config = DefaultConfig()
	}

	config.DefaultTheme = theme

	// Convert to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("✓ MOTD theme set to: %s", theme)))
	return nil
}

func DefaultConfig() Config {
	return Config{
		TipsDirectory:   "",
		CheckOutdated:   "false",
		ImageInfoFile:   "", // No longer used?
		DefaultTheme:    "slate",
		TemplateFile:    "",
		ThemesDirectory: "",
	}
}

func loadConfig() (Config, error) {
	configDir, err := env.GetConfigDir()
	if err != nil {
		return DefaultConfig(), err
	}
	configPath := filepath.Join(configDir, "motd.json")

	var config Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}

func getImageInfo() ImageInfo {
	// Detect OS information
	info := ImageInfo{
		ImageFlavor:   "homebrew",
		ImageVendor:   "bluefin-cli",
		FedoraVersion: "N/A",
	}

	switch runtime.GOOS {
	case "darwin":
		info.ImageName = "macOS"
		if output, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			info.ImageTag = strings.TrimSpace(string(output))
		}
	case "linux":
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "NAME=") {
					info.ImageName = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
				} else if strings.HasPrefix(line, "VERSION_ID=") {
					info.ImageTag = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
				}
			}
		}
	}

	if info.ImageName == "" {
		info.ImageName = runtime.GOOS
	}
	if info.ImageTag == "" {
		info.ImageTag = "unknown"
	}

	return info
}

func getRandomTipFromDir(tipsDir string) string {
	files, err := filepath.Glob(filepath.Join(tipsDir, "*.md"))
	if err != nil || len(files) == 0 {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tipFile := files[r.Intn(len(files))]

	content, err := os.ReadFile(tipFile)
	if err != nil {
		return ""
	}

	return "💡 **Tip:** " + string(content)
}

func getRandomDefaultTip() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tip := defaultTips[r.Intn(len(defaultTips))]
	return "💡 **Tip:** " + tip
}

// CheckStatus returns whether MOTD is enabled for each shell
// Now acts as a check for the legacy configuration
func CheckStatus() map[string]bool {
	status := make(map[string]bool)
	shells := []string{"bash", "zsh", "fish"}

	for _, shell := range shells {
		var configFile string
		switch shell {
		case "bash":
			configFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
		case "zsh":
			configFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
		case "fish":
			configFile = filepath.Join(os.Getenv("HOME"), ".config/fish/config.fish")
		}

		content, err := os.ReadFile(configFile)
		if err != nil {
			status[shell] = false
			continue
		}

		// Check for new marker (part of shell experience) OR old motd marker
		status[shell] = strings.Contains(string(content), motdMarker) || strings.Contains(string(content), "# bluefin-cli shell-config")
	}

	return status
}
