package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func isPowerShellShell(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	return normalized == "powershell" || normalized == "pwsh"
}

func togglePowerShell(enable bool) error {
	configFiles, err := powerShellProfilePaths()
	if err != nil {
		return err
	}

	execPath := preferredInitExecutablePath()
	execDir := preferredInitExecutableDir(execPath)

	rcLine := fmt.Sprintf(`if (Get-Command bluefin-cli -ErrorAction SilentlyContinue) { Invoke-Expression ((& bluefin-cli init powershell | Out-String)) } %s`, shellMaker)
	if strings.TrimSpace(execPath) != "" {
		escapedExecPath := strings.ReplaceAll(execPath, "'", "''")
		rcLine = fmt.Sprintf(`if (Test-Path '%s') { Invoke-Expression ((& '%s' init powershell | Out-String)) } elseif (Get-Command bluefin-cli -ErrorAction SilentlyContinue) { Invoke-Expression ((& bluefin-cli init powershell | Out-String)) } %s`, escapedExecPath, escapedExecPath, shellMaker)
	} else if strings.TrimSpace(execDir) != "" {
		escapedExecDir := strings.ReplaceAll(execDir, "'", "''")
		rcLine = fmt.Sprintf(`$bluefinCliDir = '%s'; if (Test-Path "$bluefinCliDir\bluefin-cli.exe" -and ($env:PATH -notlike "*$bluefinCliDir*")) { $env:PATH = "$bluefinCliDir;$env:PATH" }; if (Get-Command bluefin-cli -ErrorAction SilentlyContinue) { Invoke-Expression ((& bluefin-cli init powershell | Out-String)) } %s`, escapedExecDir, shellMaker)
	}

	changedAny := false
	for _, configFile := range configFiles {
		changed, err := syncPowerShellProfile(configFile, rcLine, enable)
		if err != nil {
			return err
		}
		if changed {
			changedAny = true
		}
	}

	if enable {
		if changedAny {
			fmt.Println(successStyle.Render("✓ Enabled shell experience for powershell"))
		} else {
			fmt.Println(infoStyle.Render("powershell is already enabled for powershell"))
		}
	} else {
		if changedAny {
			fmt.Println(successStyle.Render("✓ Disabled shell experience for powershell"))
		} else {
			fmt.Println(infoStyle.Render("powershell is already disabled for powershell"))
		}
	}

	if enable {
		if cfg, err := LoadConfig("powershell"); err == nil {
			InstallTools(cfg)
		}
	}

	return nil
}

func powerShellProfilePath() (string, error) {
	paths, err := powerShellProfilePaths()
	if err != nil {
		return "", err
	}

	return paths[0], nil
}

func powerShellProfilePaths() ([]string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	profiles := []string{
		filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
		filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
	}

	return profiles, nil
}

func syncPowerShellProfile(configFile, rcLine string, enable bool) (bool, error) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			if !enable {
				return false, nil
			}

			if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
				return false, err
			}
			content = []byte("")
		} else {
			return false, err
		}
	}

	text := string(content)
	if enable && strings.Contains(text, shellMaker) && !hasLegacyPowerShellProfileLines(text) {
		return false, nil
	}

	cleanedText, removedManaged := stripManagedPowerShellProfileLines(text)

	if enable {
		hasManagedLine := strings.Contains(cleanedText, shellMaker)
		if hasManagedLine {
			if removedManaged {
				if err := os.WriteFile(configFile, []byte(cleanedText), 0644); err != nil {
					return false, err
				}
				return true, nil
			}
			return false, nil
		}

		output := cleanedText
		prefix := "\n"
		if len(output) == 0 || strings.HasSuffix(output, "\n") {
			prefix = ""
		}
		output += prefix + rcLine + "\n"

		if err := os.WriteFile(configFile, []byte(output), 0644); err != nil {
			return false, err
		}

		return true, nil
	}

	if !removedManaged {
		return false, nil
	}

	output := cleanedText
	if strings.TrimSpace(output) == "" {
		output = ""
	} else {
		output = strings.TrimRight(output, "\n") + "\n"
	}

	if err := os.WriteFile(configFile, []byte(output), 0644); err != nil {
		return false, err
	}

	return true, nil
}

func stripManagedPowerShellProfileLines(text string) (string, bool) {
	if text == "" {
		return text, false
	}

	lines := strings.Split(text, "\n")
	newLines := make([]string, 0, len(lines))
	removed := false
	for _, line := range lines {
		if isManagedPowerShellProfileLine(line) {
			removed = true
			continue
		}
		newLines = append(newLines, line)
	}

	return strings.Join(newLines, "\n"), removed
}

func isManagedPowerShellProfileLine(line string) bool {
	normalized := strings.ToLower(strings.TrimSpace(line))
	if normalized == "" {
		return false
	}

	if strings.Contains(normalized, strings.ToLower(shellMaker)) || strings.Contains(normalized, strings.ToLower(blingMarker)) {
		return true
	}

	if strings.Contains(normalized, "bluefin-cli") && strings.Contains(normalized, "init powershell") {
		return true
	}

	if strings.Contains(normalized, "get-command bluefin-cli") {
		return true
	}

	if strings.Contains(normalized, "test-path") && strings.Contains(normalized, "bluefin-cli.exe") {
		return true
	}

	return false
}

func hasLegacyPowerShellProfileLines(text string) bool {
	if text == "" {
		return false
	}

	lines := strings.Split(text, "\n")
	shellMarkerCount := 0
	for _, line := range lines {
		normalized := strings.ToLower(line)
		if strings.Contains(normalized, strings.ToLower(shellMaker)) {
			shellMarkerCount++
			continue
		}

		if isManagedPowerShellProfileLine(line) {
			return true
		}
	}

	return shellMarkerCount != 1
}

func preferredInitExecutablePath() string {
	candidates := []string{}

	if arg0 := strings.TrimSpace(os.Args[0]); arg0 != "" {
		if absPath, err := filepath.Abs(arg0); err == nil {
			candidates = append(candidates, absPath)
		} else {
			candidates = append(candidates, arg0)
		}
	}

	if execPath, err := os.Executable(); err == nil {
		candidates = append(candidates, execPath)
	}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "bluefin-cli.exe"),
			filepath.Join(cwd, "bluefin-cli"),
		)
	}

	for _, candidate := range candidates {
		cleaned := filepath.Clean(strings.TrimSpace(candidate))
		if cleaned == "" {
			continue
		}

		normalized := strings.ToLower(strings.ReplaceAll(cleaned, "/", "\\"))
		if strings.Contains(normalized, "\\go-build\\") {
			continue
		}

		if info, err := os.Stat(cleaned); err == nil && !info.IsDir() {
			return cleaned
		}
	}

	return ""
}

func preferredInitExecutableDir(execPath string) string {
	if strings.TrimSpace(execPath) != "" {
		dir := filepath.Dir(execPath)
		if dir != "" {
			return dir
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "bluefin-cli.exe")); err == nil {
			return cwd
		}
	}

	return ""
}

func checkPowerShellStatus() bool {
	configFiles, err := powerShellProfilePaths()
	if err != nil {
		return false
	}

	for _, configFile := range configFiles {
		content, err := os.ReadFile(configFile)
		if err != nil {
			return false
		}

		text := string(content)
		if !strings.Contains(text, shellMaker) && !strings.Contains(text, blingMarker) {
			return false
		}
	}

	return true
}
