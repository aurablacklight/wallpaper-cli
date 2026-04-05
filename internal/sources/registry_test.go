package sources

import (
	"context"
	"testing"
)

type mockSource struct {
	name string
}

func (m *mockSource) Name() string { return m.name }
func (m *mockSource) Search(ctx context.Context, params *SearchParams) (*SearchResult, error) {
	return &SearchResult{}, nil
}
func (m *mockSource) Capabilities() *Capabilities {
	return &Capabilities{Name: m.name}
}

func TestRegisterAndGet(t *testing.T) {
	// Clear registry for test isolation
	registryMu.Lock()
	old := registry
	registry = make(map[string]SourceFactory)
	registryMu.Unlock()
	defer func() {
		registryMu.Lock()
		registry = old
		registryMu.Unlock()
	}()

	Register("test-source", func(cfg map[string]string) (Source, error) {
		return &mockSource{name: "test-source"}, nil
	})

	src, err := Get("test-source", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if src.Name() != "test-source" {
		t.Errorf("Name = %q, want test-source", src.Name())
	}
}

func TestGetUnknown(t *testing.T) {
	_, err := Get("nonexistent-source-xyz", nil)
	if err == nil {
		t.Error("expected error for unknown source")
	}
}

func TestList(t *testing.T) {
	registryMu.Lock()
	old := registry
	registry = make(map[string]SourceFactory)
	registryMu.Unlock()
	defer func() {
		registryMu.Lock()
		registry = old
		registryMu.Unlock()
	}()

	Register("alpha", func(cfg map[string]string) (Source, error) {
		return &mockSource{name: "alpha"}, nil
	})
	Register("beta", func(cfg map[string]string) (Source, error) {
		return &mockSource{name: "beta"}, nil
	})

	names := List()
	if len(names) != 2 {
		t.Fatalf("List() returned %d names, want 2", len(names))
	}

	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	if !found["alpha"] || !found["beta"] {
		t.Errorf("List() = %v, want alpha and beta", names)
	}
}

func TestIsRegistered(t *testing.T) {
	registryMu.Lock()
	old := registry
	registry = make(map[string]SourceFactory)
	registryMu.Unlock()
	defer func() {
		registryMu.Lock()
		registry = old
		registryMu.Unlock()
	}()

	Register("exists", func(cfg map[string]string) (Source, error) {
		return &mockSource{name: "exists"}, nil
	})

	if !IsRegistered("exists") {
		t.Error("IsRegistered(exists) = false, want true")
	}
	if IsRegistered("nope") {
		t.Error("IsRegistered(nope) = true, want false")
	}
}

func TestConfigPassthrough(t *testing.T) {
	registryMu.Lock()
	old := registry
	registry = make(map[string]SourceFactory)
	registryMu.Unlock()
	defer func() {
		registryMu.Lock()
		registry = old
		registryMu.Unlock()
	}()

	var receivedCfg map[string]string
	Register("cfg-test", func(cfg map[string]string) (Source, error) {
		receivedCfg = cfg
		return &mockSource{name: "cfg-test"}, nil
	})

	cfg := map[string]string{"api_key": "secret123"}
	_, err := Get("cfg-test", cfg)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if receivedCfg["api_key"] != "secret123" {
		t.Errorf("config not passed through: got %v", receivedCfg)
	}
}
