package sources

import (
	"fmt"
	"sync"
)

// SourceFactory creates a Source with the given config.
type SourceFactory func(cfg map[string]string) (Source, error)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]SourceFactory)
)

// Register adds a source factory to the registry.
func Register(name string, factory SourceFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = factory
}

// Get creates a source by name using its registered factory.
func Get(name string, cfg map[string]string) (Source, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown source: %s", name)
	}
	return factory(cfg)
}

// List returns the names of all registered sources.
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// IsRegistered checks whether a source name has been registered.
func IsRegistered(name string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := registry[name]
	return ok
}
