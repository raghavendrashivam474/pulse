package cli

import (
	"github.com/raghavendrashivam474/aayam/internal/capability"
	"github.com/raghavendrashivam474/aayam/internal/output"
)

// overviewCapability renders a concise project overview.
type overviewCapability struct {
	writer *output.Writer
}

func (c overviewCapability) Name() string        { return "overview" }
func (c overviewCapability) Description() string { return "Show a concise project overview" }
func (c overviewCapability) Run(ctx capability.Context) (capability.Result, error) {
	if ctx.JSON {
		if err := c.writer.PrintOverviewJSON(ctx.Snap); err != nil {
			return capability.Result{}, err
		}
	} else {
		c.writer.PrintOverview(ctx.Snap)
	}
	return capability.Result{CapabilityName: "overview"}, nil
}

// discoveryCapability routes to the full discovery output.
// Used by project, structure, relationships until M6.3-M6.5 give them their own renderers.
type discoveryCapability struct {
	name        string
	description string
	writer      *output.Writer
}

func (c discoveryCapability) Name() string        { return c.name }
func (c discoveryCapability) Description() string { return c.description }
func (c discoveryCapability) Run(ctx capability.Context) (capability.Result, error) {
	if ctx.JSON {
		if err := c.writer.PrintDiscoveryJSON(ctx.Snap); err != nil {
			return capability.Result{}, err
		}
	} else {
		c.writer.PrintDiscovery(ctx.Snap)
	}
	return capability.Result{CapabilityName: c.name}, nil
}

func newCapabilityRegistry(w *output.Writer) *capability.Registry {
	reg := capability.NewRegistry()

	mustRegister := func(c capability.Capability) {
		if err := reg.Register(c); err != nil {
			panic(err)
		}
	}

	mustRegister(overviewCapability{writer: w})
	mustRegister(discoveryCapability{
		name:        "project",
		description: "Show project identity and metadata",
		writer:      w,
	})
	mustRegister(discoveryCapability{
		name:        "structure",
		description: "Show filesystem and language structure",
		writer:      w,
	})
	mustRegister(discoveryCapability{
		name:        "relationships",
		description: "Show relationship graph intelligence",
		writer:      w,
	})

	return reg
}

func capabilityHelpEntries(reg *capability.Registry) []output.CapabilityInfo {
	names := reg.Names()
	entries := make([]output.CapabilityInfo, 0, len(names))
	for _, name := range names {
		c, ok := reg.Lookup(name)
		if !ok {
			continue
		}
		entries = append(entries, output.CapabilityInfo{
			Name:        c.Name(),
			Description: c.Description(),
		})
	}
	return entries
}
