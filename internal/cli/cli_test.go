package cli_test

import (
	"os"
	"testing"

	"pulse/internal/cli"
	"pulse/internal/testhelpers"
)

func TestParseArgs_Empty(t *testing.T) {
	args, err := cli.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.Help || args.Version || args.JSON {
		t.Error("expected all flags false for empty args")
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
			t.Errorf("expected Help=true for flag %q", flag)
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
			t.Errorf("expected Version=true for flag %q", flag)
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
		t.Errorf("expected ./my-project, got %q", args.TargetPath)
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
		t.Errorf("expected ./my-project, got %q", args.TargetPath)
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	_, err := cli.ParseArgs([]string{"--unknown"})
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestParseArgs_TooManyPositional(t *testing.T) {
	_, err := cli.ParseArgs([]string{"./project-a", "./project-b"})
	if err == nil {
		t.Fatal("expected error for too many positional arguments, got nil")
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
		t.Errorf("expected ExitFailure for unknown flag, got %d", code)
	}
}

func TestRun_TooManyArgs_ReturnsFailure(t *testing.T) {
	code := cli.Run([]string{"./a", "./b"})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure for too many args, got %d", code)
	}
}

func TestRun_WithTargetPath_ReturnsSuccess(t *testing.T) {
	// Create a real temporary directory so target validation passes.
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test project\n",
	})

	// cli.Run uses config.New which resolves relative to cwd.
	// Pass the absolute temp dir path directly.
	code := cli.Run([]string{dir})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess with target path, got %d", code)
	}
}

func TestRun_WithJSONAndTargetPath_ReturnsSuccess(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"README.md": "# Test project\n",
	})

	code := cli.Run([]string{"--json", dir})
	if code != cli.ExitSuccess {
		t.Errorf("expected ExitSuccess with --json and target path, got %d", code)
	}
}

func TestRun_NonexistentPath_ReturnsFailure(t *testing.T) {
	code := cli.Run([]string{"./nonexistent-project-dir"})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure for nonexistent path, got %d", code)
	}
}

func TestRun_FileTarget_ReturnsFailure(t *testing.T) {
	dir := testhelpers.TempProject(t, map[string]string{
		"main.go": "package main\n",
	})

	filePath := dir + string(os.PathSeparator) + "main.go"
	code := cli.Run([]string{filePath})
	if code != cli.ExitFailure {
		t.Errorf("expected ExitFailure for file target, got %d", code)
	}
}
