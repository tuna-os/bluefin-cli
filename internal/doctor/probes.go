package doctor

import (
	"fmt"
	"path/filepath"
)

// CommandExists checks if the specified binary is executable on PATH.
func CommandExists(opts Options, bin string) bool {
	o := opts.withDefaults()
	_, err := o.LookPath(bin)
	return err == nil
}

// CheckBrew checks package manager availability and suitability.
func CheckBrew(opts Options) CheckResult {
	o := opts.withDefaults()
	if o.IsAlpine() {
		switch {
		case CommandExists(o, "coldbrew"):
			return CheckResult{OK: true, Name: "Package manager: coldbrew"}
		case CommandExists(o, "apk"):
			return CheckResult{
				Name: "Package manager: coldbrew not set up yet",
				Warn: true,
				Note: "run 'bluefin-cli doctor --fix' to install it (rootless, sandboxed); falls back to 'sudo apk add' per package until then",
			}
		default:
			return CheckResult{
				Name: "Package manager",
				Warn: true,
				Note: "neither coldbrew nor apk found — shell tool installs will be skipped",
			}
		}
	}
	if _, err := o.LookPath("brew"); err == nil {
		return CheckResult{OK: true, Name: "Homebrew on PATH"}
	}
	for _, p := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew", "/home/linuxbrew/.linuxbrew/bin/brew"} {
		if _, err := o.Stat(p); err == nil {
			return CheckResult{
				Name: "Homebrew on PATH",
				Warn: true,
				Note: fmt.Sprintf("brew is installed at %s but not on PATH — add: eval \"$(%s shellenv)\"", p, p),
			}
		}
	}
	return CheckResult{
		Name: "Homebrew on PATH",
		Warn: true,
		Note: "brew not found — Install Apps bundles need it: https://brew.sh",
	}
}

// CheckShellIntegration checks whether bluefin-cli init hook is enabled in current shell rc.
func CheckShellIntegration(opts Options) CheckResult {
	o := opts.withDefaults()
	sh := o.CurrentShell
	status := o.ShellStatus()
	if status != nil && status[sh] {
		return CheckResult{OK: true, Name: fmt.Sprintf("Shell integration enabled (%s)", sh)}
	}
	return CheckResult{
		Name: fmt.Sprintf("Shell integration enabled (%s)", sh),
		Warn: true,
		Note: fmt.Sprintf("run: bluefin-cli shell %s on", sh),
	}
}

// CheckTools checks whether eza, fzf, starship are installed.
func CheckTools(opts Options) CheckResult {
	o := opts.withDefaults()
	missing := []string{}
	for _, tool := range []string{"eza", "fzf", "starship"} {
		if _, err := o.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if len(missing) == 0 {
		return CheckResult{OK: true, Name: "Managed tools installed (eza, fzf, starship)"}
	}
	return CheckResult{
		Name: "Managed tools installed",
		Warn: true,
		Note: fmt.Sprintf("missing: %v — install via the Bluefin Shell menu", missing),
	}
}

// CheckNetwork checks whether api.github.com is reachable.
func CheckNetwork(opts Options) CheckResult {
	o := opts.withDefaults()
	if err := o.CheckNetwork(); err != nil {
		return CheckResult{
			Name: "GitHub reachable",
			Warn: true,
			Note: "updates and bundle installs need network access: " + err.Error(),
		}
	}
	return CheckResult{OK: true, Name: "GitHub reachable"}
}

// CheckSelfUpdate checks whether self-updating is viable on direct install.
func CheckSelfUpdate(opts Options) CheckResult {
	o := opts.withDefaults()
	method := o.DetectUpdate()
	if method != "direct" {
		return CheckResult{OK: true, Name: "Updates managed by " + string(method)}
	}
	exe, err := o.Executable()
	if err != nil {
		return CheckResult{Name: "Self-update viable", Warn: true, Note: err.Error()}
	}
	dir := filepath.Dir(exe)
	f, err := o.CreateTempFile(dir, ".bluefin-doctor-*")
	if err != nil {
		return CheckResult{
			Name: "Self-update viable",
			Warn: true,
			Note: dir + " is not writable — 'bluefin-cli update' would need elevated permissions",
		}
	}
	name := f.Name()
	_ = f.Close()
	_ = o.RemoveTempFile(name)
	return CheckResult{OK: true, Name: "Self-update viable (install dir writable)"}
}

// CheckVersion checks the running version against latest release.
func CheckVersion(opts Options) CheckResult {
	o := opts.withDefaults()
	v := o.Version
	if v == "dev" {
		return CheckResult{OK: true, Name: "Version: dev build"}
	}
	rel, err := o.LatestRelease()
	if err != nil {
		return CheckResult{
			OK:   true,
			Name: "Version " + v,
			Note: "could not check for updates: " + err.Error(),
		}
	}
	if o.IsNewer(v, rel.TagName) {
		hint := "bluefin-cli update"
		if h := o.UpdateHint(o.DetectUpdate()); h != "" {
			hint = h
		}
		return CheckResult{
			Name: "Version " + v,
			Warn: true,
			Note: fmt.Sprintf("%s available — run: %s", rel.TagName, hint),
		}
	}
	return CheckResult{OK: true, Name: "Version " + v + " (latest)"}
}
