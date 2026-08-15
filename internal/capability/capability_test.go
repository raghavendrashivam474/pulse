package capability_test

import (
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/capability"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

type stubCapability struct {
	name        string
	description string
}

func (s *stubCapability) Name() string {
	return s.name
}

func (s *stubCapability) Description() string {
	return s.description
}

func (s *stubCapability) Run(_ capability.Context) (capability.Result, error) {
	return capability.Result{CapabilityName: s.name}, nil
}

func newStub(name, description string) capability.Capability {
	return &stubCapability{
		name:        name,
		description: description,
	}
}

func TestRegistry_RegisterAndLookup(t *testing.T) {
	r := capability.NewRegistry()

	if err := r.Register(newStub("project", "Project intelligence")); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	c, ok := r.Lookup("project")
	if !ok {
		t.Fatal("expected registered capability to be found")
	}

	if c.Name() != "project" {
		t.Errorf("expected capability name %q, got %q", "project", c.Name())
	}
}

func TestRegistry_LookupUnknown(t *testing.T) {
	r := capability.NewRegistry()

	if _, ok := r.Lookup("unknown"); ok {
		t.Fatal("expected unknown capability lookup to fail")
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	r := capability.NewRegistry()

	if err := r.Register(newStub("project", "Project intelligence")); err != nil {
		t.Fatalf("unexpected first register error: %v", err)
	}

	if err := r.Register(newStub("project", "Duplicate")); err == nil {
		t.Fatal("expected duplicate registration error")
	}
}

func TestRegistry_IsKnown(t *testing.T) {
	r := capability.NewRegistry()
	if err := r.Register(newStub("structure", "Structure intelligence")); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if !r.IsKnown("structure") {
		t.Error("expected IsKnown to return true for registered capability")
	}

	if r.IsKnown("relationships") {
		t.Error("expected IsKnown to return false for unregistered capability")
	}
}

func TestRegistry_NamesSorted(t *testing.T) {
	r := capability.NewRegistry()
	_ = r.Register(newStub("structure", ""))
	_ = r.Register(newStub("project", ""))
	_ = r.Register(newStub("relationships", ""))

	names := r.Names()
	expected := []string{"project", "relationships", "structure"}

	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}

	for i := range expected {
		if names[i] != expected[i] {
			t.Errorf("names[%d]: expected %q, got %q", i, expected[i], names[i])
		}
	}
}

func TestRegistry_ZeroValueRegisterWorks(t *testing.T) {
	var r capability.Registry

	if err := r.Register(newStub("overview", "Overview intelligence")); err != nil {
		t.Fatalf("expected zero-value registry to work, got error: %v", err)
	}

	if !r.IsKnown("overview") {
		t.Fatal("expected overview capability to be registered")
	}
}

func TestRegistry_RegisterNil(t *testing.T) {
	var r capability.Registry

	if err := r.Register(nil); err == nil {
		t.Fatal("expected nil capability registration to fail")
	}
}

func TestCapability_RunReturnsName(t *testing.T) {
	c := newStub("project", "Project intelligence")

	result, err := c.Run(capability.Context{
		Snap: snapshot.ProjectSnapshot{},
		JSON: false,
	})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	if result.CapabilityName != "project" {
		t.Errorf("expected CapabilityName %q, got %q", "project", result.CapabilityName)
	}
}
