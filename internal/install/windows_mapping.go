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

// loadWindowsPackageAliases parses the embedded mapping.
//
// A parse failure here is a build-time defect, not a runtime condition: the
// bytes are compiled into the binary, so if they are malformed they were
// malformed for every user of that build. Degrading to an empty map made that
// defect invisible (tuna-os/bluefin-cli#232) -- windowsCandidates would return
// just the Homebrew name for every package, so `install` on Windows searched
// winget for "fd" and "bat" instead of "sharkdp.fd" and "sharkdp.bat", failing
// one package at a time with nothing anywhere reporting that the mapping had
// not loaded. Panicking instead makes any test that touches the loader fail
// loudly in CI, which is the only place the problem can still be fixed.
func loadWindowsPackageAliases() map[string][]string {
	aliases := make(map[string][]string)
	if err := json.Unmarshal(windowsMappingJSON, &aliases); err != nil {
		panic("install: embedded windows_mapping.json does not parse: " + err.Error())
	}
	if len(aliases) == 0 {
		panic("install: embedded windows_mapping.json parsed to an empty mapping")
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
