package profile

import (
	"errors"
	"testing"
)

type fakeSyncClient struct {
	created  []byte
	updated  []byte
	updateID string
	fetched  []byte
	err      error
}

type fakeSyncIDStore struct{ id string }

func (f *fakeSyncIDStore) Get() string         { return f.id }
func (f *fakeSyncIDStore) Set(id string) error { f.id = id; return nil }

func (f *fakeSyncClient) Create(data []byte) (string, error) {
	f.created = data
	return "new-id", f.err
}

func (f *fakeSyncClient) Update(id string, data []byte) error {
	f.updateID, f.updated = id, data
	return f.err
}

func (f *fakeSyncClient) Fetch(string) ([]byte, error) { return f.fetched, f.err }

func TestSyncPushCreatesOrUpdates(t *testing.T) {
	p := &Profile{Version: 1, EnabledShells: []string{"bash"}}
	fake := &fakeSyncClient{}
	store := &fakeSyncIDStore{}
	sync := NewSync(fake, store)

	id, err := sync.Push(p)
	if err != nil || id != "new-id" || store.id != "new-id" || len(fake.created) == 0 {
		t.Fatalf("create = (%q, %v), data=%q", id, err, fake.created)
	}
	store.id = "existing"
	id, err = sync.Push(p)
	if err != nil || id != "existing" || fake.updateID != "existing" || len(fake.updated) == 0 {
		t.Fatalf("update = (%q, %v), target=%q data=%q", id, err, fake.updateID, fake.updated)
	}
}

func TestSyncFetchValidatesBeforeApply(t *testing.T) {
	fake := &fakeSyncClient{fetched: []byte(`{"version":1,"enabled_shells":["fish"]}`)}
	store := &fakeSyncIDStore{}
	p, err := NewSync(fake, store).Fetch("id")
	if err != nil || len(p.EnabledShells) != 1 || p.EnabledShells[0] != "fish" {
		t.Fatalf("Fetch() = (%+v, %v)", p, err)
	}
	if store.id != "id" {
		t.Fatalf("fetched id was not persisted: %q", store.id)
	}

	fake.fetched = []byte(`{"version":99}`)
	if _, err := NewSync(fake, store).Fetch("id"); err == nil {
		t.Fatal("expected unsupported version to fail")
	}
	fake.err = errors.New("offline")
	if _, err := NewSync(fake, store).Fetch("id"); !errors.Is(err, fake.err) {
		t.Fatalf("expected transport error, got %v", err)
	}
}
