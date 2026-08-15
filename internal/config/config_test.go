package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/raghavendrashivam474/aayam/internal/config"
)

func TestNew_EmptyPath_UsesCWD(t *testing.T) {
	cfg, err := config.New("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	expected := filepath.Clean(cwd)
	if cfg.TargetPath != expected {
		t.Errorf("expected %q, got %q", expected, cfg.TargetPath)
	}
}

func TestNew_AbsolutePath(t *testing.T) {
	abs := filepath.Clean(os.TempDir())
	cfg, err := config.New(abs, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TargetPath != abs {
		t.Errorf("expected %q, got %q", abs, cfg.TargetPath)
	}
}

func TestNew_RelativePath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	cfg, err := config.New("some-subdir", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Clean(filepath.Join(cwd, "some-subdir"))
	if cfg.TargetPath != expected {
		t.Errorf("expected %q, got %q", expected, cfg.TargetPath)
	}
}

func TestNew_DotPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	cfg, err := config.New(".", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Clean(cwd)
	if cfg.TargetPath != expected {
		t.Errorf("expected %q, got %q", expected, cfg.TargetPath)
	}
}

func TestNew_PathIsCleaned(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	cfg, err := config.New("./sub/../sub", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Clean(filepath.Join(cwd, "sub"))
	if cfg.TargetPath != expected {
		t.Errorf("expected cleaned path %q, got %q", expected, cfg.TargetPath)
	}
}

func TestNew_PathIsAbsolute(t *testing.T) {
	cfg, err := config.New("./any-path", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(cfg.TargetPath) {
		t.Errorf("expected absolute path, got %q", cfg.TargetPath)
	}
}

func TestNew_JSONMode(t *testing.T) {
	cfg, err := config.New("", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.JSON {
		t.Error("expected JSON=true")
	}
}

func TestNew_JSONFalse(t *testing.T) {
	cfg, err := config.New("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JSON {
		t.Error("expected JSON=false")
	}
}

func TestNew_TargetPath_NeverEmpty(t *testing.T) {
	cfg, err := config.New("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(cfg.TargetPath) == "" {
		t.Error("expected non-empty TargetPath")
	}
}
