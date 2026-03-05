package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hanthor/bluefin-cli/internal/env"
)

var (
	lookPathWSL = exec.LookPath
	execWSL     = exec.Command
	nowWSL      = time.Now

	registerWindowsTaskWSL = registerWindowsTask
	runWindowsTaskWSL      = runWindowsTask
	deleteWindowsTaskWSL   = deleteWindowsTask
	ensureStartupRunEntryWSL = ensureWindowsStartupRunEntry
	deleteStartupRunEntryWSL = deleteWindowsStartupRunEntry
)

const windowsThemeStateFile = "windows-theme.json"

const (
	taskThemeModeSync      = "BluefinCLI-ThemeModeSync"
	taskThemeModeSyncLogon = "BluefinCLI-ThemeModeSync-Logon"
	taskSetLightAt6AM      = "BluefinCLI-SetLightAt6AM"
	taskSetDarkAt6PM       = "BluefinCLI-SetDarkAt6PM"
	startupRunValueName    = "BluefinCLIThemeModeSync"
)

var windowsTaskDescriptions = map[string]string{
	taskThemeModeSync: "Keeps Bluefin wallpaper day/night variant in sync with current Windows light/dark mode.",
	taskThemeModeSyncLogon: "Legacy logon sync task (replaced by HKCU Run startup entry).",
	taskSetLightAt6AM: "Sets Windows light mode at 6:00 AM and refreshes Bluefin wallpaper variant.",
	taskSetDarkAt6PM:  "Sets Windows dark mode at 6:00 PM and refreshes Bluefin wallpaper variant.",
}

type windowsThemeState struct {
	SelectedTheme    string `json:"selectedTheme"`
	AutoApplyMonthly bool   `json:"autoApplyMonthly"`
	LastAppliedMonth string `json:"lastAppliedMonth"`
	MonthlyThemes    []string `json:"monthlyThemes,omitempty"`
}

func syncWallpapersToWindows(casks []string) error {
	if len(casks) == 0 {
		return nil
	}

	if _, err := lookPathWSL("powershell.exe"); err != nil {
		return fmt.Errorf("powershell.exe not available in WSL interop: %w", err)
	}

	if _, err := lookPathWSL("wslpath"); err != nil {
		return fmt.Errorf("wslpath not available: %w", err)
	}

	imagesByTheme := map[string][]string{
		"Bluefin": {},
		"Aurora":  {},
		"Bazzite": {},
	}
	monthlyThemes := make([]string, 0)

	for _, cask := range casks {
		normalized := normalizeCaskName(cask)
		themeName, ok := detectThemeName(normalized)
		if !ok {
			continue
		}

		images, err := findWallpaperImagesForCask(normalized, themeName)
		if err != nil {
			fmt.Println(infoStyle.Render(fmt.Sprintf("Could not find %s wallpapers to sync from WSL: %v", themeName, err)))
			continue
		}

		imagesByTheme[themeName] = append(imagesByTheme[themeName], images...)
		if supportsMonthlyWallpapers(images) {
			monthlyThemes = append(monthlyThemes, themeName)
		}
	}

	var registered int
	for _, themeName := range []string{"Bluefin", "Aurora", "Bazzite"} {
		images := uniqueStrings(imagesByTheme[themeName])
		if len(images) == 0 {
			continue
		}

		if err := registerWindowsTheme(themeName, images); err != nil {
			fmt.Println(infoStyle.Render(fmt.Sprintf("Skipping %s theme registration: %v", themeName, err)))
			continue
		}

		registered++
	}

	if registered > 0 {
		if err := ensureMonthlyThemeAutoRollover(uniqueStrings(monthlyThemes)); err != nil {
			fmt.Println(infoStyle.Render("Could not persist monthly rollover preferences: " + err.Error()))
		}
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Registered %d Windows theme(s) from WSL wallpapers", registered)))
		return nil
	}

	return fmt.Errorf("no Windows themes were registered; verify wallpaper files exist under ~/.local/share/backgrounds")
}

func ThemesFromWallpaperCasks(casks []string) []string {
	seen := map[string]struct{}{}
	themes := make([]string, 0, len(casks))

	for _, cask := range casks {
		normalized := normalizeCaskName(cask)
		themeName, ok := detectThemeName(normalized)
		if !ok {
			continue
		}
		if _, exists := seen[themeName]; exists {
			continue
		}
		seen[themeName] = struct{}{}
		themes = append(themes, themeName)
	}

	sort.Strings(themes)
	return themes
}

func normalizeCaskName(cask string) string {
	parts := strings.Split(cask, "/")
	return parts[len(parts)-1]
}

func detectThemeName(cask string) (string, bool) {
	name := strings.ToLower(cask)
	switch {
	case strings.Contains(name, "bluefin"):
		return "Bluefin", true
	case strings.Contains(name, "aurora"):
		return "Aurora", true
	case strings.Contains(name, "bazzite"):
		return "Bazzite", true
	default:
		return "", false
	}
}

func findWallpaperImagesForCask(cask, themeName string) ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve home directory: %w", err)
	}

	themeSlug := strings.ToLower(themeName)
	candidateDirs := []string{
		filepath.Join(homeDir, ".local", "share", "backgrounds", themeSlug),
		filepath.Join(homeDir, ".local", "share", "wallpapers", themeSlug),
		filepath.Join("/usr", "share", "backgrounds", themeSlug),
		filepath.Join("/usr", "local", "share", "backgrounds", themeSlug),
	}

	images := make([]string, 0)
	for _, dir := range candidateDirs {
		found, err := findImagesInDir(dir)
		if err != nil {
			continue
		}
		images = append(images, found...)
	}

	if len(images) == 0 {
		backgroundsRoot := filepath.Join(homeDir, ".local", "share", "backgrounds")
		found, err := findImagesByNameHint(backgroundsRoot, themeSlug)
		if err == nil {
			images = append(images, found...)
		}
	}

	if len(images) > 0 {
		return uniqueStrings(images), nil
	}

	prefixCmd := execWSL("brew", "--prefix")
	out, err := prefixCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve brew prefix: %w", err)
	}

	prefix := strings.TrimSpace(string(out))
	if prefix == "" {
		return nil, fmt.Errorf("brew prefix is empty")
	}

	caskDir := filepath.Join(prefix, "Caskroom", cask)
	fallbackImages, err := findImagesInDir(caskDir)
	if err != nil {
		return nil, fmt.Errorf("no wallpapers found in known install paths for %s", cask)
	}

	return uniqueStrings(fallbackImages), nil
}

func findImagesInDir(root string) ([]string, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	images := make([]string, 0)
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".bmp", ".webp":
			images = append(images, path)
		}

		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Strings(images)
	return images, nil
}

func findImagesByNameHint(root, hint string) ([]string, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	images := make([]string, 0)
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		lowerPath := strings.ToLower(path)
		if !strings.Contains(lowerPath, hint) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".bmp", ".webp":
			images = append(images, path)
		}

		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	return images, nil
}

func registerWindowsTheme(themeName string, linuxImagePaths []string) error {
	windowsPaths := make([]string, 0, len(linuxImagePaths))
	linuxToWindows := make(map[string]string, len(linuxImagePaths))
	for _, linuxPath := range linuxImagePaths {
		cmd := execWSL("wslpath", "-w", linuxPath)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		windowsPath := strings.TrimSpace(string(out))
		if windowsPath != "" {
			windowsPaths = append(windowsPaths, windowsPath)
			linuxToWindows[linuxPath] = windowsPath
		}
	}

	windowsPaths = uniqueStrings(windowsPaths)
	if len(windowsPaths) == 0 {
		return fmt.Errorf("no valid Windows paths for %s wallpapers", themeName)
	}

	primaryLinux := selectMonthlyWallpaper(linuxImagePaths, nowWSL())
	primaryWindows := ""
	if primaryLinux != "" {
		primaryWindows = linuxToWindows[primaryLinux]
	}

	scriptPath, err := writeThemeRegistrationScript()
	if err != nil {
		return err
	}
	defer os.Remove(scriptPath)

	args := []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", scriptPath, themeName, primaryWindows}
	args = append(args, windowsPaths...)

	cmd := execWSL("powershell.exe", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("PowerShell registration failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func ApplyWindowsTheme(themeName string) error {
	if _, err := lookPathWSL("powershell.exe"); err != nil {
		return fmt.Errorf("powershell.exe not available in WSL interop: %w", err)
	}

	themeFile := fmt.Sprintf(`$theme = Join-Path $env:LOCALAPPDATA "Microsoft\Windows\Themes\%s.theme"; if (-not (Test-Path -LiteralPath $theme)) { throw "Theme file not found: $theme" }; Start-Process -FilePath $theme`, themeName)
	cmd := execWSL("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", themeFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to apply %s theme: %w: %s", themeName, err, strings.TrimSpace(string(out)))
	}

	return nil
}

func ConfigureWindowsThemeAutomation(enableAutoDarkLightSwitch bool) error {
	if _, err := lookPathWSL("schtasks.exe"); err != nil {
		fmt.Println(infoStyle.Render("Skipping Windows task automation setup: schtasks.exe not available in WSL interop"))
		return nil
	}

	syncTaskCmd := `powershell.exe -NoProfile -NonInteractive -WindowStyle Hidden -ExecutionPolicy Bypass -File "%LOCALAPPDATA%\\BluefinCLI\\theme-mode-sync.ps1"`
	syncTaskRegistered := false
	if err := registerWindowsTaskWSL(taskThemeModeSync, []string{"/SC", "MINUTE", "/MO", "1"}, syncTaskCmd); err != nil {
		if isWindowsAccessDenied(err) {
			fmt.Println(infoStyle.Render("Could not register task " + taskThemeModeSync + ": access denied (continuing without this task)"))
		} else {
			return err
		}
	} else {
		syncTaskRegistered = true
	}

	if err := ensureStartupRunEntryWSL(startupRunValueName, syncTaskCmd); err != nil {
		if isWindowsAccessDenied(err) {
			fmt.Println(infoStyle.Render("Could not configure startup sync entry: access denied (continuing without startup sync)"))
		} else {
			fmt.Println(infoStyle.Render("Could not configure startup sync entry: " + err.Error()))
		}
	}

	if syncTaskRegistered {
		if err := runWindowsTaskWSL(taskThemeModeSync); err != nil {
			fmt.Println(infoStyle.Render("Could not immediately run theme mode sync task: " + err.Error()))
		}
	} else {
		fmt.Println(infoStyle.Render("Theme mode sync task was not registered; you can still switch themes manually."))
	}

	if !enableAutoDarkLightSwitch {
		_ = deleteWindowsTaskWSL(taskSetLightAt6AM)
		_ = deleteWindowsTaskWSL(taskSetDarkAt6PM)
		return nil
	}

	lightTaskCmd := `powershell.exe -NoProfile -NonInteractive -WindowStyle Hidden -ExecutionPolicy Bypass -File "%LOCALAPPDATA%\\BluefinCLI\\set-light-mode.ps1"`
	darkTaskCmd := `powershell.exe -NoProfile -NonInteractive -WindowStyle Hidden -ExecutionPolicy Bypass -File "%LOCALAPPDATA%\\BluefinCLI\\set-dark-mode.ps1"`

	if err := registerWindowsTaskWSL(taskSetLightAt6AM, []string{"/SC", "DAILY", "/ST", "06:00"}, lightTaskCmd); err != nil {
		if isWindowsAccessDenied(err) {
			fmt.Println(infoStyle.Render("Could not register task " + taskSetLightAt6AM + ": access denied (skipping auto 6:00 AM light-mode switch)"))
		} else {
			return err
		}
	}

	if err := registerWindowsTaskWSL(taskSetDarkAt6PM, []string{"/SC", "DAILY", "/ST", "18:00"}, darkTaskCmd); err != nil {
		if isWindowsAccessDenied(err) {
			fmt.Println(infoStyle.Render("Could not register task " + taskSetDarkAt6PM + ": access denied (skipping auto 6:00 PM dark-mode switch)"))
		} else {
			return err
		}
	}

	return nil
}

func isWindowsAccessDenied(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "access is denied")
}

func registerWindowsTask(taskName string, scheduleArgs []string, taskCommand string) error {
	description := windowsTaskDescriptions[taskName]
	if err := registerWindowsTaskPowerShell(taskName, scheduleArgs, taskCommand, description); err == nil {
		return nil
	}

	args := []string{"/Create", "/F", "/TN", taskName}
	args = append(args, scheduleArgs...)
	args = append(args, "/TR", taskCommand)

	cmd := execWSL("schtasks.exe", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to register task %s: %w: %s", taskName, err, strings.TrimSpace(string(out)))
	}

	if strings.TrimSpace(description) != "" {
		if err := setWindowsTaskDescription(taskName, description); err != nil {
			fmt.Println(infoStyle.Render("Could not set task description for " + taskName + ": " + err.Error()))
		}
	}

	return nil
}

func registerWindowsTaskPowerShell(taskName string, scheduleArgs []string, taskCommand, description string) error {
	if _, err := lookPathWSL("powershell.exe"); err != nil {
		return fmt.Errorf("powershell.exe not available in WSL interop: %w", err)
	}

	escapedTaskName := strings.ReplaceAll(taskName, "'", "''")
	escapedTaskCommand := strings.ReplaceAll(taskCommand, "'", "''")
	escapedDescription := strings.ReplaceAll(description, "'", "''")

	triggerScript, err := powershellTriggerScript(scheduleArgs)
	if err != nil {
		return err
	}

	script := fmt.Sprintf(`$ErrorActionPreference = "Stop"
%s
$action = New-ScheduledTaskAction -Execute 'cmd.exe' -Argument '/c %s'
Register-ScheduledTask -TaskName '%s' -Action $action -Trigger $trigger -Description '%s' -Force | Out-Null`, triggerScript, escapedTaskCommand, escapedTaskName, escapedDescription)

	cmd := execWSL("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to register task via PowerShell: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func powershellTriggerScript(scheduleArgs []string) (string, error) {
	if len(scheduleArgs) < 2 {
		return "", fmt.Errorf("invalid task schedule arguments")
	}

	args := make(map[string]string)
	for i := 0; i+1 < len(scheduleArgs); i += 2 {
		args[strings.ToUpper(scheduleArgs[i])] = scheduleArgs[i+1]
	}

	scheduleType := strings.ToUpper(args["/SC"])
	switch scheduleType {
	case "MINUTE":
		minutes := args["/MO"]
		if strings.TrimSpace(minutes) == "" {
			minutes = "1"
		}
		return fmt.Sprintf(`$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1)
$trigger.Repetition.Interval = (New-TimeSpan -Minutes %s)
$trigger.Repetition.Duration = ([TimeSpan]::MaxValue)`, minutes), nil
	case "DAILY":
		timeValue := args["/ST"]
		if strings.TrimSpace(timeValue) == "" {
			return "", fmt.Errorf("daily schedule missing /ST time")
		}
		escapedTime := strings.ReplaceAll(timeValue, "'", "''")
		return fmt.Sprintf(`$scheduledTime = [datetime]::Today.Add([TimeSpan]::Parse('%s'))
if ($scheduledTime -lt (Get-Date)) { $scheduledTime = $scheduledTime.AddDays(1) }
$trigger = New-ScheduledTaskTrigger -Daily -At $scheduledTime`, escapedTime), nil
	case "ONLOGON":
		return `$trigger = New-ScheduledTaskTrigger -AtLogOn`, nil
	default:
		return "", fmt.Errorf("unsupported task schedule type: %s", scheduleType)
	}
}

func setWindowsTaskDescription(taskName, description string) error {
	if _, err := lookPathWSL("powershell.exe"); err != nil {
		return fmt.Errorf("powershell.exe not available in WSL interop: %w", err)
	}

	escapedTaskName := strings.ReplaceAll(taskName, "'", "''")
	escapedDescription := strings.ReplaceAll(description, "'", "''")

	script := fmt.Sprintf(`$service = New-Object -ComObject "Schedule.Service"
$service.Connect()
$folder = $service.GetFolder("\")
$task = $folder.GetTask('%s')
if (-not $task) { throw "Task not found: %s" }
$definition = $task.Definition
$definition.RegistrationInfo.Description = '%s'
$null = $folder.RegisterTaskDefinition('%s', $definition, 6, $null, $null, 3, $null)`, escapedTaskName, escapedTaskName, escapedDescription, escapedTaskName)

	cmd := execWSL("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set task description: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func runWindowsTask(taskName string) error {
	cmd := execWSL("schtasks.exe", "/Run", "/TN", taskName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run task %s: %w: %s", taskName, err, strings.TrimSpace(string(out)))
	}

	return nil
}

func deleteWindowsTask(taskName string) error {
	cmd := execWSL("schtasks.exe", "/Delete", "/F", "/TN", taskName)
	if out, err := cmd.CombinedOutput(); err != nil {
		message := strings.ToLower(strings.TrimSpace(string(out)))
		if strings.Contains(message, "cannot find the file") || strings.Contains(message, "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete task %s: %w: %s", taskName, err, strings.TrimSpace(string(out)))
	}

	return nil
}

func ensureWindowsStartupRunEntry(valueName, command string) error {
	cmd := execWSL(
		"reg.exe",
		"ADD",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/V",
		valueName,
		"/T",
		"REG_SZ",
		"/D",
		command,
		"/F",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set startup run entry %s: %w: %s", valueName, err, strings.TrimSpace(string(out)))
	}

	return nil
}

func deleteWindowsStartupRunEntry(valueName string) error {
	cmd := execWSL(
		"reg.exe",
		"DELETE",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/V",
		valueName,
		"/F",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		message := strings.ToLower(strings.TrimSpace(string(out)))
		if strings.Contains(message, "unable to find") || strings.Contains(message, "cannot find") || strings.Contains(message, "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete startup run entry %s: %w: %s", valueName, err, strings.TrimSpace(string(out)))
	}

	return nil
}

func cleanupWindowsWallpaperSyncArtifacts() error {
	if _, err := lookPathWSL("powershell.exe"); err != nil {
		return fmt.Errorf("powershell.exe not available in WSL interop: %w", err)
	}

	cleanupScript := `$themes = Join-Path $env:LOCALAPPDATA "Microsoft\Windows\Themes"
$themeNames = @("Bluefin.theme", "Aurora.theme", "Bazzite.theme", "Bluefin-derived.theme", "Aurora-derived.theme", "Bazzite-derived.theme")
foreach($name in $themeNames) {
	$path = Join-Path $themes $name
	if(Test-Path -LiteralPath $path) {
		Remove-Item -LiteralPath $path -Force -ErrorAction SilentlyContinue
	}
}

$wallRoot = Join-Path $env:USERPROFILE "Pictures\Wallpapers"
foreach($dir in @("Bluefin", "Aurora", "Bazzite")) {
	$target = Join-Path $wallRoot $dir
	if(Test-Path -LiteralPath $target) {
		Remove-Item -LiteralPath $target -Recurse -Force -ErrorAction SilentlyContinue
	}
}

$cliDir = Join-Path $env:LOCALAPPDATA "BluefinCLI"
if(Test-Path -LiteralPath $cliDir) {
	Remove-Item -LiteralPath $cliDir -Recurse -Force -ErrorAction SilentlyContinue
}`

	cmd := execWSL("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", cleanupScript)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to cleanup Windows wallpaper artifacts: %w: %s", err, strings.TrimSpace(string(out)))
	}

	_ = deleteWindowsTaskWSL(taskThemeModeSync)
	_ = deleteWindowsTaskWSL(taskThemeModeSyncLogon)
	_ = deleteWindowsTaskWSL(taskSetLightAt6AM)
	_ = deleteWindowsTaskWSL(taskSetDarkAt6PM)
	_ = deleteStartupRunEntryWSL(startupRunValueName)

	statePath, err := windowsThemeStatePath()
	if err == nil {
		_ = os.Remove(statePath)
	}

	return nil
}

func SetWindowsThemePreference(themeName string, autoApplyMonthly bool) error {
	state, err := loadWindowsThemeState()
	if err != nil {
		return err
	}

	state.SelectedTheme = themeName
	if autoApplyMonthly || len(state.MonthlyThemes) > 0 {
		state.AutoApplyMonthly = true
		state.LastAppliedMonth = nowWSL().Format("2006-01")
	} else {
		state.AutoApplyMonthly = false
	}

	return saveWindowsThemeState(state)
}

func MaybeRollOverWindowsThemeOnInit() error {
	if !env.IsWSL() {
		return nil
	}

	state, err := loadWindowsThemeState()
	if err != nil {
		return err
	}

	if !state.AutoApplyMonthly || state.SelectedTheme == "" {
		if !state.AutoApplyMonthly || len(state.MonthlyThemes) == 0 {
			return nil
		}
	}

	themeForRollover := state.SelectedTheme
	if !containsString(state.MonthlyThemes, themeForRollover) {
		themeForRollover = preferredMonthlyTheme(state.MonthlyThemes)
	}

	if themeForRollover == "" {
		return nil
	}

	now := nowWSL()
	currentMonth := now.Format("2006-01")
	if state.LastAppliedMonth == currentMonth {
		return nil
	}

	if now.Day() != 1 {
		return nil
	}

	images, err := findWallpaperImagesForTheme(themeForRollover)
	if err != nil {
		return err
	}

	if err := registerWindowsTheme(themeForRollover, images); err != nil {
		return err
	}

	if err := ApplyWindowsTheme(themeForRollover); err != nil {
		return err
	}

	state.SelectedTheme = themeForRollover
	state.LastAppliedMonth = currentMonth
	return saveWindowsThemeState(state)
}

func findWallpaperImagesForTheme(themeName string) ([]string, error) {
	theme := strings.TrimSpace(themeName)
	if theme == "" {
		return nil, fmt.Errorf("theme name is empty")
	}

	cask := strings.ToLower(theme) + "-wallpapers"
	return findWallpaperImagesForCask(cask, theme)
}

func ThemeSupportsMonthly(themeName string) bool {
	images, err := findWallpaperImagesForTheme(themeName)
	if err != nil {
		return false
	}

	return supportsMonthlyWallpapers(images)
}

func selectMonthlyWallpaper(imagePaths []string, now time.Time) string {
	if len(imagePaths) == 0 {
		return ""
	}

	sorted := uniqueStrings(imagePaths)
	monthPrefix := fmt.Sprintf("%02d-", int(now.Month()))
	timeOfDay := "day"
	hour := now.Hour()
	if hour < 6 || hour >= 18 {
		timeOfDay = "night"
	}

	for _, path := range sorted {
		name := strings.ToLower(filepath.Base(path))
		if strings.HasPrefix(name, monthPrefix) && strings.Contains(name, timeOfDay) {
			return path
		}
	}

	for _, path := range sorted {
		name := strings.ToLower(filepath.Base(path))
		if strings.HasPrefix(name, monthPrefix) {
			return path
		}
	}

	return sorted[0]
}

func supportsMonthlyWallpapers(imagePaths []string) bool {
	months := map[string]struct{}{}
	for _, image := range imagePaths {
		name := strings.ToLower(filepath.Base(image))
		if len(name) < 3 {
			continue
		}
		prefix := name[:2]
		if name[2] != '-' {
			continue
		}
		switch prefix {
		case "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12":
			months[prefix] = struct{}{}
		}
	}

	return len(months) >= 2
}

func ensureMonthlyThemeAutoRollover(monthlyThemes []string) error {
	state, err := loadWindowsThemeState()
	if err != nil {
		return err
	}

	state.MonthlyThemes = uniqueStrings(monthlyThemes)
	if len(state.MonthlyThemes) == 0 {
		return saveWindowsThemeState(state)
	}

	state.AutoApplyMonthly = true
	if !containsString(state.MonthlyThemes, state.SelectedTheme) {
		state.SelectedTheme = preferredMonthlyTheme(state.MonthlyThemes)
	}

	if state.LastAppliedMonth == "" {
		state.LastAppliedMonth = nowWSL().Format("2006-01")
	}

	return saveWindowsThemeState(state)
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func preferredMonthlyTheme(themes []string) string {
	if len(themes) == 0 {
		return ""
	}

	if containsString(themes, "Bluefin") {
		return "Bluefin"
	}

	sorted := uniqueStrings(themes)
	return sorted[0]
}

func writeThemeRegistrationScript() (string, error) {
	script := `param(
    [Parameter(Mandatory = $true)]
    [string]$ThemeName,
    [Parameter(Mandatory = $false)]
    [string]$PrimaryWallpaper,
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ImagePaths
)

if (-not $ImagePaths -or $ImagePaths.Count -eq 0) {
    exit 0
}

$themesRoot = Join-Path $env:LOCALAPPDATA "Microsoft\Windows\Themes"
$wallpaperRoot = Join-Path $env:USERPROFILE "Pictures\Wallpapers"
$targetDir = Join-Path $wallpaperRoot $ThemeName

New-Item -Path $themesRoot -ItemType Directory -Force | Out-Null
New-Item -Path $targetDir -ItemType Directory -Force | Out-Null

$copied = @()
foreach ($src in $ImagePaths) {
    if (Test-Path -LiteralPath $src) {
        $dest = Join-Path $targetDir ([System.IO.Path]::GetFileName($src))
        Copy-Item -LiteralPath $src -Destination $dest -Force
        $copied += $dest
    }
}

if ($copied.Count -eq 0) {
    exit 0
}

$themePath = Join-Path $themesRoot ($ThemeName + ".theme")
$selectedWallpaper = $copied[0]
$isBluefin = $ThemeName -ieq "Bluefin"
if ($PrimaryWallpaper -and (Test-Path -LiteralPath $PrimaryWallpaper)) {
	$primaryName = [System.IO.Path]::GetFileName($PrimaryWallpaper)
	$candidate = Join-Path $targetDir $primaryName
	if (Test-Path -LiteralPath $candidate) {
		$selectedWallpaper = $candidate
	}
}

$baseTheme = Join-Path $env:WINDIR "Resources\Themes\aero.theme"
if (-not (Test-Path -LiteralPath $baseTheme)) {
	throw "Could not find base theme at $baseTheme"
}

$lines = Get-Content -LiteralPath $baseTheme

$displayNameSet = $false
$wallpaperSet = $false
$desktopSectionSeen = $false
$slideshowSectionSeen = $false
$intervalSet = $false
$shuffleSet = $false
$imagesRootSet = $false

for ($i = 0; $i -lt $lines.Count; $i++) {
	if ($lines[$i] -match '^\[Control Panel\\Desktop\]$') {
		$desktopSectionSeen = $true
	}

	if ($lines[$i] -match '^\[Slideshow\]$') {
		$slideshowSectionSeen = $true
	}

	if ($lines[$i] -match '^DisplayName=') {
		$lines[$i] = "DisplayName=$ThemeName"
		$displayNameSet = $true
	}

	if ($lines[$i] -match '^Wallpaper=') {
		$lines[$i] = "Wallpaper=$selectedWallpaper"
		$wallpaperSet = $true
	}

	if ($lines[$i] -match '^Interval=') {
		$lines[$i] = "Interval=86400000"
		$intervalSet = $true
	}

	if ($lines[$i] -match '^Shuffle=') {
		$lines[$i] = "Shuffle=0"
		$shuffleSet = $true
	}

	if ($lines[$i] -match '^ImagesRootPath=') {
		$lines[$i] = "ImagesRootPath=$targetDir"
		$imagesRootSet = $true
	}
}

if (-not $displayNameSet) {
	$lines += ""
	$lines += "[Theme]"
	$lines += "DisplayName=$ThemeName"
}

if (-not $wallpaperSet) {
	if (-not $desktopSectionSeen) {
		$lines += ""
		$lines += "[Control Panel\\Desktop]"
	}

	$lines += "Wallpaper=$selectedWallpaper"
	$lines += "TileWallpaper=0"
	$lines += "WallpaperStyle=10"
	$lines += "PicturePosition=4"
	$lines += "MultimonBackgrounds=0"
}

if ($isBluefin) {
	$filtered = @()
	$inSlideshow = $false
	foreach ($line in $lines) {
		if ($line -match '^\[Slideshow\]$') {
			$inSlideshow = $true
			continue
		}

		if ($inSlideshow -and $line -match '^\[') {
			$inSlideshow = $false
		}

		if ($inSlideshow) {
			continue
		}

		if ($line -match '^ImagesRootPath=' -or $line -match '^Interval=' -or $line -match '^Shuffle=') {
			continue
		}

		$filtered += $line
	}

	$lines = $filtered
} else {
	if (-not $slideshowSectionSeen) {
		$lines += ""
		$lines += "[Slideshow]"
	}

	if (-not $imagesRootSet) {
		$lines += "ImagesRootPath=$targetDir"
	}

	if (-not $intervalSet) {
		$lines += "Interval=86400000"
	}

	if (-not $shuffleSet) {
		$lines += "Shuffle=0"
	}
}

Set-Content -Path $themePath -Value $lines -Encoding Unicode

$syncDir = Join-Path $env:LOCALAPPDATA "BluefinCLI"
New-Item -Path $syncDir -ItemType Directory -Force | Out-Null
$syncScriptPath = Join-Path $syncDir "theme-mode-sync.ps1"

$syncScript = @'
param()

$themeRegPath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes"
$personalizePath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize"
$desktopPath = "HKCU:\Control Panel\Desktop"

function Set-WallpaperPath {
	param([string]$Path)

	if (-not (Test-Path -LiteralPath $Path)) {
		return
	}

	Set-ItemProperty -Path $desktopPath -Name Wallpaper -Value $Path -ErrorAction SilentlyContinue

	Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public class WallpaperInterop {
	[DllImport("user32.dll", SetLastError=true, CharSet=CharSet.Unicode)]
	public static extern bool SystemParametersInfo(int action, int param, string vparam, int winIni);
}
"@ -ErrorAction SilentlyContinue | Out-Null

	[WallpaperInterop]::SystemParametersInfo(20, 0, $Path, 3) | Out-Null
}

try {
	$currentTheme = (Get-ItemProperty -Path $themeRegPath -Name CurrentTheme -ErrorAction Stop).CurrentTheme
} catch {
	return
}

$themeName = [System.IO.Path]::GetFileNameWithoutExtension($currentTheme)
if ($themeName -notin @("Bluefin", "Aurora", "Bazzite")) {
	return
}

$themeDir = Join-Path (Join-Path $env:USERPROFILE "Pictures\Wallpapers") $themeName
if (-not (Test-Path -LiteralPath $themeDir)) {
	return
}

$images = Get-ChildItem -LiteralPath $themeDir -File | Sort-Object Name
if (-not $images -or $images.Count -eq 0) {
	return
}

$appsUseLight = 1
try {
	$appsUseLight = (Get-ItemProperty -Path $personalizePath -Name AppsUseLightTheme -ErrorAction Stop).AppsUseLightTheme
} catch {
}

$target = $null
$currentWallpaper = $null
try {
	$currentWallpaper = (Get-ItemProperty -Path $desktopPath -Name Wallpaper -ErrorAction Stop).Wallpaper
} catch {
}

if ($currentWallpaper -and (Test-Path -LiteralPath $currentWallpaper)) {
	$currentName = [System.IO.Path]::GetFileNameWithoutExtension($currentWallpaper)
	$currentExt = [System.IO.Path]::GetExtension($currentWallpaper)
	if ($currentName -match '^(.*?)-(day|night)$') {
		$stem = $matches[1]
		$desiredSuffix = if ($appsUseLight -eq 0) { 'night' } else { 'day' }
		$candidate = Join-Path $themeDir ($stem + '-' + $desiredSuffix + $currentExt)
		if (Test-Path -LiteralPath $candidate) {
			$target = Get-Item -LiteralPath $candidate
		}
	}
}

if (-not $target) {
	if ($appsUseLight -eq 0) {
		$target = $images | Where-Object { $_.BaseName -match '(night|dark)' } | Select-Object -First 1
	} else {
		$target = $images | Where-Object { $_.BaseName -match '(day|light)' } | Select-Object -First 1
	}
}

if (-not $target) {
	$target = $images | Select-Object -First 1
}

if ($target) {
	Set-WallpaperPath -Path $target.FullName
}
'@

Set-Content -Path $syncScriptPath -Value $syncScript -Encoding UTF8

$lightScriptPath = Join-Path $syncDir "set-light-mode.ps1"
$lightScript = @'
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize" -Name AppsUseLightTheme -Value 1 -Type DWord -ErrorAction SilentlyContinue
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize" -Name SystemUsesLightTheme -Value 1 -Type DWord -ErrorAction SilentlyContinue
& "$env:LOCALAPPDATA\BluefinCLI\theme-mode-sync.ps1"
'@
Set-Content -Path $lightScriptPath -Value $lightScript -Encoding UTF8

$darkScriptPath = Join-Path $syncDir "set-dark-mode.ps1"
$darkScript = @'
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize" -Name AppsUseLightTheme -Value 0 -Type DWord -ErrorAction SilentlyContinue
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize" -Name SystemUsesLightTheme -Value 0 -Type DWord -ErrorAction SilentlyContinue
& "$env:LOCALAPPDATA\BluefinCLI\theme-mode-sync.ps1"
'@
Set-Content -Path $darkScriptPath -Value $darkScript -Encoding UTF8

Start-Process -FilePath $themePath | Out-Null
`

	f, err := os.CreateTemp("", "bluefin-theme-*.ps1")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary PowerShell script: %w", err)
	}

	if _, err := f.WriteString(script); err != nil {
		f.Close()
		return "", fmt.Errorf("failed to write temporary PowerShell script: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary PowerShell script: %w", err)
	}

	return f.Name(), nil
}

func loadWindowsThemeState() (*windowsThemeState, error) {
	path, err := windowsThemeStatePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &windowsThemeState{}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read windows theme state: %w", err)
	}

	state := &windowsThemeState{}
	if err := json.Unmarshal(content, state); err != nil {
		return nil, fmt.Errorf("failed to parse windows theme state: %w", err)
	}

	return state, nil
}

func saveWindowsThemeState(state *windowsThemeState) error {
	path, err := windowsThemeStatePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create windows theme state directory: %w", err)
	}

	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal windows theme state: %w", err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write windows theme state: %w", err)
	}

	return nil
}

func windowsThemeStatePath() (string, error) {
	dir, err := env.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, windowsThemeStateFile), nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}

	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)

	return result
}
