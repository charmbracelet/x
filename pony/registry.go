package pony

import (
	"sync"
)

// ComponentFactory is a function that creates an Element from props and children.
// Custom components should implement this signature.
type ComponentFactory func(props Props, children []Element) Element

var (
	registry   = make(map[string]ComponentFactory)
	registryMu sync.RWMutex
)

// Register registers a custom component with the given name.
// The factory function will be called to create instances of the component.
//
// Example:
//
//	pony.Register("badge", func(props Props, children []Element) Element {
//	    return &Badge{
//	        text:  props.Get("text"),
//	        color: props.Get("color"),
//	    }
//	})
func Register(name string, factory ComponentFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = factory
}

// Unregister removes a registered component.
func Unregister(name string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	delete(registry, name)
}

// GetComponent retrieves a component factory by name.
// Returns nil if the component is not registered.
func GetComponent(name string) (ComponentFactory, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	factory, ok := registry[name]
	return factory, ok
}

// RegisteredComponents returns a list of all registered component names.
func RegisteredComponents() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// ClearRegistry removes all registered components.
// Useful for testing.
func ClearRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]ComponentFactory)
}
