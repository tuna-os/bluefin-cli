package profile

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
)

const gistFilename = "bluefin-profile.json"

// SyncClient is the transport boundary used by profile synchronization.
// Implementations exchange profile documents without deciding when they are
// applied or where the remote identifier is persisted.
type SyncClient interface {
	Create(data []byte) (string, error)
	Update(id string, data []byte) error
	Fetch(id string) ([]byte, error)
}

// SyncIDStore persists the remote profile identifier independently from the
// transport and command layer.
type SyncIDStore interface {
	Get() string
	Set(string) error
}

// Sync serializes profiles and validates fetched documents independently from
// applying them to the local machine.
type Sync struct {
	client SyncClient
	store  SyncIDStore
}

func NewSync(client SyncClient, store SyncIDStore) *Sync {
	return &Sync{client: client, store: store}
}

func (s *Sync) Push(p *Profile) (string, error) {
	data, err := p.Marshal()
	if err != nil {
		return "", err
	}
	id := s.store.Get()
	if id != "" {
		if err := s.client.Update(id, data); err != nil {
			return "", err
		}
		return id, nil
	}
	id, err = s.client.Create(data)
	if err != nil {
		return "", err
	}
	if err := s.store.Set(id); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Sync) Fetch(id string) (*Profile, error) {
	configured := s.store.Get()
	if id == "" {
		id = configured
	}
	if id == "" {
		return nil, fmt.Errorf("no gist configured — run 'profile push' first or pass a gist id")
	}
	data, err := s.client.Fetch(id)
	if err != nil {
		return nil, err
	}
	p, err := Parse(data)
	if err != nil {
		return nil, err
	}
	if configured == "" {
		if err := s.store.Set(id); err != nil {
			return nil, err
		}
	}
	return p, nil
}

type ConfigSyncIDStore struct{}

func (ConfigSyncIDStore) Get() string { return viper.GetString("profile.gist_id") }

func (ConfigSyncIDStore) Set(id string) error {
	viper.Set("profile.gist_id", id)
	return config.Save()
}

// GitHubCLIClient adapts the gh gist commands to SyncClient. The gh CLI only
// accepts files for create and edit, so temporary-file handling stays inside
// this transport adapter rather than leaking into Cobra command handlers.
type GitHubCLIClient struct {
	command func(string, ...string) *exec.Cmd
}

func NewGitHubCLIClient() (*GitHubCLIClient, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("profile sync uses the GitHub CLI — install gh and run 'gh auth login'")
	}
	return &GitHubCLIClient{command: exec.Command}, nil
}

func (c *GitHubCLIClient) Create(data []byte) (string, error) {
	filename, cleanup, err := writeGistFile(data)
	if err != nil {
		return "", err
	}
	defer cleanup()
	out, err := c.command("gh", "gist", "create", "--desc", "bluefin-cli profile", filename).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("creating gist: %v\n%s", err, out)
	}
	return path.Base(strings.TrimSpace(string(out))), nil
}

func (c *GitHubCLIClient) Update(id string, data []byte) error {
	filename, cleanup, err := writeGistFile(data)
	if err != nil {
		return err
	}
	defer cleanup()
	out, err := c.command("gh", "gist", "edit", id, "--filename", gistFilename, filename).CombinedOutput()
	if err != nil {
		return fmt.Errorf("updating gist %s: %v\n%s", id, err, out)
	}
	return nil
}

func (c *GitHubCLIClient) Fetch(id string) ([]byte, error) {
	out, err := c.command("gh", "gist", "view", id, "--filename", gistFilename, "--raw").Output()
	if err != nil {
		return nil, fmt.Errorf("fetching gist %s: %w", id, err)
	}
	return out, nil
}

func writeGistFile(data []byte) (string, func(), error) {
	dir, err := os.MkdirTemp("", "bluefin-profile-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	filename := filepath.Join(dir, gistFilename)
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		cleanup()
		return "", nil, err
	}
	return filename, cleanup, nil
}
