package cli

import (
	"github.com/raghavendrashivam474/aayam/internal/capability"
	"github.com/raghavendrashivam474/aayam/internal/output"
)

type discoveryCapability struct {
	name        string
	description string
	writer      *output.Writer
}

func (c discoveryCapability) Name() string {
	return c.name
}

func (c discoveryCapability) Description() string {
	return c.description
}

func (c discoveryCapability) Run(ctx capability.Context) (capability.Result, error) {
	if ctx.JSON {
		if err := c.writer.PrintDiscoveryJSON(ctx.Snap); err != nil {
			return capability.Result{}, err
		}
	} else {
		c.writer.PrintDiscovery(ctx.Snap)
	}

	return capability.Result{
		CapabilityName: c.name,
	}, nil
}

func newCapabilityRegistry(w *output.Writer) *capability.Registry {
	reg := capability.NewRegistry()

	mustRegister := func(c capability.Capability) {
		if err := reg.Register(c); err != nil {
			panic(err)
		}
	}

	mustRegister(discoveryCapability{
		name:        "overview",
		description: "Show a project overview",
		writer:      w,
	})
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
