package capability

import (
	"fmt"

	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// Context holds the shared intelligence passed to a capability.
type Context struct {
	Snap snapshot.ProjectSnapshot
	JSON bool
}

// Result is the minimal structured result returned by a capability.
type Result struct {
	CapabilityName string
}

// Capability defines an independently invokable intelligence unit.
type Capability interface {
	Name() string
	Description() string
	Run(ctx Context) (Result, error)
}

// Registry stores registered capabilities by name.
// The zero value is ready to use.
type Registry struct {
	entries map[string]Capability
}

// NewRegistry returns an initialized registry.
func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]Capability),
	}
}

// Register adds a capability to the registry.
func (r *Registry) Register(c Capability) error {
	if c == nil {
		return fmt.Errorf("capability cannot be nil")
	}

	if r.entries == nil {
		r.entries = make(map[string]Capability)
	}

	name := c.Name()
	if name == "" {
		return fmt.Errorf("capability name cannot be empty")
	}

	if _, exists := r.entries[name]; exists {
		return fmt.Errorf("capability %q is already registered", name)
	}

	r.entries[name] = c
	return nil
}

// Lookup returns the registered capability for a name.
func (r *Registry) Lookup(name string) (Capability, bool) {
	c, ok := r.entries[name]
	return c, ok
}

// IsKnown reports whether a capability name is registered.
func (r *Registry) IsKnown(name string) bool {
	_, ok := r.Lookup(name)
	return ok
}

// Names returns the sorted registered capability names.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.entries))
	for name := range r.entries {
		names = append(names, name)
	}

	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}

	return names
}
