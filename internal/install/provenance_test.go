package install

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The files under internal/install/resources/ are compiled into every binary
// and are owned upstream, not here: the Brewfiles drive package installation
// and the cask list decides which wallpapers the CLI offers. Until
// tuna-os/bluefin-cli#258 they were identified only by a mutable branch URL,
// so a reviewer could see the copied diff but not which upstream state
// produced it, and nothing detected a resource that had been hand-edited or
// truncated after the fact. PROVENANCE.json records that, and these tests are
// what make it load-bearing rather than decorative.

type provenanceEntry struct {
	SourceRepo   string `json:"source_repo"`
	SourceRef    string `json:"source_ref"`
	SourceCommit string `json:"source_commit"`
	SourcePath   string `json:"source_path"`
	Derivation   string `json:"derivation,omitempty"`
	SHA256       string `json:"sha256"`
}

type provenanceManifest struct {
	GeneratedAt string                     `json:"generated_at"`
	Resources   map[string]provenanceEntry `json:"resources"`
}

const (
	resourcesDir    = "resources"
	provenanceFile  = "resources/PROVENANCE.json"
	digestAlgPrefix = "sha256:"
)

func loadProvenance(t *testing.T) provenanceManifest {
	t.Helper()
	data, err := os.ReadFile(provenanceFile)
	if err != nil {
		t.Fatalf("reading %s: %v", provenanceFile, err)
	}
	var manifest provenanceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parsing %s: %v", provenanceFile, err)
	}
	if len(manifest.Resources) == 0 {
		t.Fatalf("%s lists no resources", provenanceFile)
	}
	return manifest
}

// embeddedResourceFiles walks the resources tree, skipping the manifest itself.
func embeddedResourceFiles(t *testing.T) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(resourcesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(resourcesDir, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if rel == "PROVENANCE.json" {
			return nil
		}
		found = append(found, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", resourcesDir, err)
	}
	sort.Strings(found)
	return found
}

// TestProvenanceDigestsMatchCommittedResources is the gate: every embedded
// resource must hash to the digest the manifest records. A truncated, corrupted
// or hand-edited resource fails here rather than reaching a release.
func TestProvenanceDigestsMatchCommittedResources(t *testing.T) {
	manifest := loadProvenance(t)

	for rel, entry := range manifest.Resources {
		data, err := os.ReadFile(filepath.Join(resourcesDir, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("%s is listed in PROVENANCE.json but could not be read: %v", rel, err)
			continue
		}
		sum := sha256.Sum256(data)
		got := digestAlgPrefix + hex.EncodeToString(sum[:])
		if got != entry.SHA256 {
			t.Errorf("%s has digest %s, but PROVENANCE.json records %s — "+
				"the resource was changed without `just update-resources`, or the manifest is stale",
				rel, got, entry.SHA256)
		}
	}
}

// TestProvenanceCoversEveryEmbeddedResource catches the other direction: a new
// resource added to the tree without a manifest entry would otherwise ship
// with no provenance at all, which is exactly the state #258 describes.
func TestProvenanceCoversEveryEmbeddedResource(t *testing.T) {
	manifest := loadProvenance(t)

	for _, rel := range embeddedResourceFiles(t) {
		if _, ok := manifest.Resources[rel]; !ok {
			t.Errorf("%s is embedded but has no PROVENANCE.json entry — add one via `just update-resources`", rel)
		}
	}

	for rel := range manifest.Resources {
		if _, err := os.Stat(filepath.Join(resourcesDir, filepath.FromSlash(rel))); os.IsNotExist(err) {
			t.Errorf("PROVENANCE.json lists %s, which is not in the resources tree", rel)
		}
	}
}

// TestProvenanceEntriesAreWellFormed a manifest entry that names no upstream
// is not provenance, so the fields that identify the source are required.
func TestProvenanceEntriesAreWellFormed(t *testing.T) {
	manifest := loadProvenance(t)

	for rel, entry := range manifest.Resources {
		if entry.SourceRepo == "" {
			t.Errorf("%s: source_repo is empty", rel)
		}
		if entry.SourcePath == "" {
			t.Errorf("%s: source_path is empty", rel)
		}
		if !strings.HasPrefix(entry.SHA256, digestAlgPrefix) || len(entry.SHA256) != len(digestAlgPrefix)+64 {
			t.Errorf("%s: sha256 %q is not a sha256: digest", rel, entry.SHA256)
		}
	}
}

// TestProvenanceBundleFilesAreCovered every bundle the CLI can install must
// resolve to an embedded Brewfile with recorded provenance; otherwise a bundle
// silently falls through to the network download path.
func TestProvenanceBundleFilesAreCovered(t *testing.T) {
	manifest := loadProvenance(t)

	for name, spec := range bundles {
		rel := "brewfiles/" + spec.File
		if _, ok := manifest.Resources[rel]; !ok {
			t.Errorf("bundle %q maps to %s, which has no PROVENANCE.json entry", name, rel)
		}
	}
}
