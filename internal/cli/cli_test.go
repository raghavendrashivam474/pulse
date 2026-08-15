package cli_test

import (
	"os"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/capability"
	"github.com/raghavendrashivam474/aayam/internal/cli"
	"github.com/raghavendrashivam474/aayam/internal/testhelpers"
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

func testRegistry(t *testing.T) *capability.Registry {
	t.Helper()

	reg := capability.NewRegistry()
	for _, name := range []string{"overview", "project", "structure", "relationships"} {
		if err := reg.Register(&stubCapability{
			name:        name,
			description: name + " capability",
		}); err != nil {
			t.Fatalf("failed to register stub capability %q: %v", name, err)
		}
	}

	return reg
}

func TestParseArgs_Empty(t *testing.T) {
	args, err := cli.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if args.Help || args.Version || args.JSON {
		t.Error("expected all flags to be false")
	}

	if args.TargetPath != "" {
		t.Errorf("expected empty target path, got %q", args.TargetPath)
	}
}

func TestParseArgs_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		args, err := cli.ParseArgs([]string{flag})
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", flag, err)
		}
		if !args.Help {
			t.Errorf("expected Help=true for %q", flag)
		}
	}
}

func TestParseArgs_Version(t *testing.T) {
	for _, flag := range []string{"--version", "-v"} {
		args, err := cli.ParseArgs([]string{flag})
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", flag, err)
		}
		if !args.Version {
			t.Errorf("expected Version=true for %q", flag)
		}
	}
}

func TestParseArgs_JSON(t *testing.T) {
	args, err := cli.ParseArgs([]string{"--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !args.JSON {
		t.Error("expected JSON=true")
	}
}

func TestParseArgs_TargetPath(t *testing.T) {
	args, err := cli.ParseArgs([]string{"./my-project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if args.TargetPath != "./my-project" {
		t.Errorf("expected target path %q, got %q", "./my-project", args.TargetPath)
	}
}

func TestParseArgs_TargetPathWithFlags(t *testing.T) {
	args, err := cli.ParseArgs([]string{"--json", "./my-project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !args.JSON {
		t.Error("expected JSON=true")
	}

	if args.TargetPath != "./my-project" {
		t.Errorf("expected target path %q, got %q", "./my-project", args.TargetPath)
	}
}

func TestParseArgs_CapabilityAndTarget(t *testing.T) {
	args, err := cli.ParseArgs([]string{"project", "./my-project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(args.Positionals) != 2 {
		t.Fatalf("expected 2 positional args, got %d", len(args.Positionals))
	}

	if args.Positionals[0] != "project" || args.Positionals[1] != "./my-project" {
		t.Errorf("unexpected positionals: %#v", args.Positionals)
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	if _, err := cli.ParseArgs([]string{"--unknown"}); err == nil {
		t.Fatal("expected unknown flag error")
	}
}

func TestParseArgs_TooManyPositional(t *testing.T) {
	if _, err := cli.ParseArgs([]string{"project", "./a", "./b"}); err == nil {
		t.Fatal("expected too many arguments error")
	}
}

func TestResolveCommand_LegacyTargetPath(t *testing.T) {
	args, err := cli.ParseArgs([]string{"./my-project"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if err := cli.ResolveCommand(args, testRegistry(t)); err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	if args.CapabilityName != "" {
		t.Errorf("expected no capability, got %q", args.CapabilityName)
	}

	if args.TargetPath != "./my-project" {
		t.Errorf("expected target path %q, got %q", "./my-project", args.TargetPath)
	}
}

func TestResolveCommand_CapabilityAndTarget(t *testing.T) {
	args, err := cli.ParseArgs([]string{"project", "./my-project"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if err := cli.ResolveCommand(args, testRegistry(t)); err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	if args.CapabilityName != "project" {
		t.Errorf("expected capability %q, got %q", "project", args.CapabilityName)
	}

	if args.TargetPath != "./my-project" {
		t.Errorf("expected target path %q, got %q", "./my-project", args.TargetPath)
	}
}

func TestResolveCommand_CapabilityWithoutTargetDefaultsToDot(t *testing.T) {
	args, err := cli.ParseArgs([]string{"project"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if err := cli.ResolveCommand(args, testRegistry(t)); err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	if args.CapabilityName != "project" {
		t.Errorf("expected capability %q, got %q", "project", args.CapabilityName)
	}

	if args.TargetPath != "." {
		t.Errorf("expected target path %q, got %q", ".", args.TargetPath)
	}
}

func TestResolveCommand_UnknownCapability(t *testing.T) {
	args, err := cli.ParseArgs([]string{"unknown", "./my-project"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if err := cli.ResolveCommand(args, testRegistry(t)); err == nil {
		t.Fatal("expected unknown capability error")
	}
}

func TestRun_NoArgs_ReturnsSuccess(t *testing.T) {
	code := cli.Run([]string{})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", code)
	}
}

func TestRun_Version_ReturnsSuccess(t *testing.T) {
	code := cli.Run([]string{"--version"})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", code)
	}
}

func TestRun_Help_ReturnsSuccess(t *testing.T) {
	code := cli.Run([]string{"--help"})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", code)
	}
}

func TestRun_JSON_ReturnsSuccess(t *testing.T) {
	code := cli.Run([]string{"--json"})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", code)
	}
}

func TestRun_UnknownFlag_ReturnsFailure(t *testing.T) {
	code := cli.Run([]string{"--unknown"})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure, got %d", code)
	}
}

func TestRun_TooManyArgs_ReturnsFailure(t *testing.T) {
	code := cli.Run([]string{"project", "./a", "./b"})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure, got %d", code)
	}
}

func TestRun_LegacyTargetPathCompatibility_ReturnsSuccess(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test project\n",
	})

	code := cli.Run([]string{dir})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", code)
	}
}

func TestRun_WithJSONAndTargetPath_ReturnsSuccess(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test project\n",
	})

	code := cli.Run([]string{"--json", dir})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", code)
	}
}

func TestRun_WithCapabilitiesAndTarget_ReturnsSuccess(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test project\n",
	})

	for _, capabilityName := range []string{"overview", "project", "structure", "relationships"} {
		t.Run(capabilityName, func(t *testing.T) {
			code := cli.Run([]string{capabilityName, dir})
			if code != cli.ExitSuccess {
				t.Errorf("expected ExitSuccess for %q, got %d", capabilityName, code)
			}
		})
	}
}

func TestRun_UnknownCapability_ReturnsFailure(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test project\n",
	})

	code := cli.Run([]string{"unknown", dir})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure, got %d", code)
	}
}

func TestRun_NonexistentPath_ReturnsFailure(t *testing.T) {
	code := cli.Run([]string{"./nonexistent-project-dir"})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure, got %d", code)
	}
}

func TestRun_FileTarget_ReturnsFailure(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go": "package main\n",
	})

	filePath := dir + string(os.PathSeparator) + "main.go"
	code := cli.Run([]string{filePath})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure, got %d", code)
	}
}
