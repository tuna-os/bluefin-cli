package install

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed windows_bundles.json
var windowsBundlesJSON []byte

type WindowsPackage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Aliases     []string `json:"aliases,omitempty"`
}

type WindowsBundle struct {
	Description string           `json:"description"`
	Packages    []WindowsPackage `json:"packages"`
}

var (
	windowsBundleManifest     map[string]WindowsBundle
	windowsBundleManifestOnce sync.Once
)

// getWindowsBundleManifest parses the embedded bundle manifest. Like the alias
// mapping next door, a parse failure is a build-time defect rather than a
// runtime condition, so it is surfaced rather than swallowed
// (tuna-os/bluefin-cli#232). Both embedded data contracts in this package now
// fail the same way, so the gate is a rule of the package rather than
// something one of them happens to have.
func getWindowsBundleManifest() map[string]WindowsBundle {
	windowsBundleManifestOnce.Do(func() {
		manifest := make(map[string]WindowsBundle)
		if err := json.Unmarshal(windowsBundlesJSON, &manifest); err != nil {
			panic("install: embedded windows_bundles.json does not parse: " + err.Error())
		}
		if len(manifest) == 0 {
			panic("install: embedded windows_bundles.json parsed to an empty manifest")
		}
		windowsBundleManifest = manifest
	})

	if windowsBundleManifest == nil {
		return map[string]WindowsBundle{}
	}

	return windowsBundleManifest
}
