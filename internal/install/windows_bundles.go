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

func getWindowsBundleManifest() map[string]WindowsBundle {
	windowsBundleManifestOnce.Do(func() {
		manifest := make(map[string]WindowsBundle)
		if err := json.Unmarshal(windowsBundlesJSON, &manifest); err != nil {
			windowsBundleManifest = map[string]WindowsBundle{}
			return
		}
		windowsBundleManifest = manifest
	})

	if windowsBundleManifest == nil {
		return map[string]WindowsBundle{}
	}

	return windowsBundleManifest
}
