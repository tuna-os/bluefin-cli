package install

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed windows_mapping.json
var windowsMappingJSON []byte

var (
	windowsPackageAliases     map[string][]string
	windowsMappingLoadOnce    sync.Once
	windowsPackageAliasesLoad = loadWindowsPackageAliases
)

func loadWindowsPackageAliases() map[string][]string {
	aliases := make(map[string][]string)
	if err := json.Unmarshal(windowsMappingJSON, &aliases); err != nil {
		return map[string][]string{}
	}
	return aliases
}

func getWindowsPackageAliases() map[string][]string {
	windowsMappingLoadOnce.Do(func() {
		windowsPackageAliases = windowsPackageAliasesLoad()
	})

	if windowsPackageAliases == nil {
		return map[string][]string{}
	}

	return windowsPackageAliases
}

func windowsCandidates(name string) []string {
	candidates := []string{name}
	aliasesMap := getWindowsPackageAliases()
	if aliases, ok := aliasesMap[name]; ok {
		seen := map[string]bool{name: true}
		for _, alias := range aliases {
			if !seen[alias] {
				candidates = append(candidates, alias)
				seen[alias] = true
			}
		}
	}
	return candidates
}
