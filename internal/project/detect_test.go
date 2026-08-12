package project_test

import (
	"testing"

	"pulse/internal/project"
	"pulse/internal/scanner"
	"pulse/internal/testhelpers"
)

// scanDir scans a temp project and returns the inventory.
// Fails the test immediately on scan error.
func scanDir(t *testing.T, files map[string]string) scanner.Inventory {
	t.Helper()
	dir := testhelpers.TempProject(t, files)
	inv, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	return inv
}

func TestDetectType_Go(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"go.mod":  "module example.com/test\n\ngo 1.21\n",
		"main.go": "package main\n",
	})
	d := project.DetectType(inv)
	if d.Primary != project.TypeGo {
		t.Errorf("expected %q, got %q", project.TypeGo, d.Primary)
	}
}

func TestDetectType_Node(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"package.json": "{\"name\":\"app\",\"version\":\"1.0.0\"}",
		"index.js":     "console.log('hello');\n",
	})
	d := project.DetectType(inv)
	if d.Primary != project.TypeNode {
		t.Errorf("expected %q, got %q", project.TypeNode, d.Primary)
	}
}

func TestDetectType_Rust(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"Cargo.toml":  "[package]\nname = \"test\"\n",
		"src/main.rs": "fn main() {}\n",
	})
	d := project.DetectType(inv)
	if d.Primary != project.TypeRust {
		t.Errorf("expected %q, got %q", project.TypeRust, d.Primary)
	}
}

func TestDetectType_Python_Pyproject(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"pyproject.toml": "[project]\nname = \"myapp\"\n",
		"main.py":        "print('hello')\n",
	})
	d := project.DetectType(inv)
	if d.Primary != project.TypePython {
		t.Errorf("expected %q, got %q", project.TypePython, d.Primary)
	}
}

func TestDetectType_Python_Requirements(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"requirements.txt": "flask==2.0.0\n",
		"app.py":           "from flask import Flask\n",
	})
	d := project.DetectType(inv)
	if d.Primary != project.TypePython {
		t.Errorf("expected %q, got %q", project.TypePython, d.Primary)
	}
}

func TestDetectType_Java_Maven(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"pom.xml": "<project></project>\n",
	})
	d := project.DetectType(inv)
	if d.Primary != project.TypeJava {
		t.Errorf("expected %q, got %q", project.TypeJava, d.Primary)
	}
}

func TestDetectType_Java_Gradle(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"build.gradle": "apply plugin: 'java'\n",
	})
	d := project.DetectType(inv)
	if d.Primary != project.TypeJava {
		t.Errorf("expected %q, got %q", project.TypeJava, d.Primary)
	}
}

func TestDetectType_Unknown(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"README.md": "# My Project\n",
		"notes.txt": "some notes\n",
	})
	d := project.DetectType(inv)

	if d.Primary != project.TypeUnknown {
		t.Errorf("expected %q, got %q", project.TypeUnknown, d.Primary)
	}
	// Unknown is valid intelligence, not an error state
	if len(d.AllDetected) == 0 {
		t.Error("AllDetected must not be empty even for Unknown")
	}
	if d.AllDetected[0] != project.TypeUnknown {
		t.Errorf("AllDetected[0] expected %q, got %q", project.TypeUnknown, d.AllDetected[0])
	}
}

func TestDetectType_MultipleMarkers_GoWinsOverNode(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"go.mod":       "module example.com/test\n",
		"package.json": "{\"name\":\"app\"}",
	})
	d := project.DetectType(inv)

	if d.Primary != project.TypeGo {
		t.Errorf("expected Go to win priority over Node, got %q", d.Primary)
	}
	if len(d.AllDetected) < 2 {
		t.Errorf("expected at least 2 detected types, got %d: %v", len(d.AllDetected), d.AllDetected)
	}
}

func TestDetectType_Deterministic(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"go.mod":       "module example.com/test\n",
		"package.json": "{\"name\":\"app\"}",
		"Cargo.toml":   "[package]\nname=\"test\"\n",
	})

	d1 := project.DetectType(inv)
	d2 := project.DetectType(inv)

	if d1.Primary != d2.Primary {
		t.Errorf("non-deterministic Primary: %q vs %q", d1.Primary, d2.Primary)
	}
	if len(d1.AllDetected) != len(d2.AllDetected) {
		t.Errorf("non-deterministic AllDetected length: %d vs %d",
			len(d1.AllDetected), len(d2.AllDetected))
	}
	for i := range d1.AllDetected {
		if d1.AllDetected[i] != d2.AllDetected[i] {
			t.Errorf("AllDetected[%d] non-deterministic: %q vs %q",
				i, d1.AllDetected[i], d2.AllDetected[i])
		}
	}
}

func TestDetectType_EmptyInventory(t *testing.T) {
	var inv scanner.Inventory
	d := project.DetectType(inv)

	if d.Primary != project.TypeUnknown {
		t.Errorf("expected Unknown for empty inventory, got %q", d.Primary)
	}
}

func TestDetectType_MarkerInSubdirectory(t *testing.T) {
	// go.mod nested inside a subdir should still be detected
	inv := scanDir(t, map[string]string{
		"README.md":           "# Mono repo\n",
		"services/api/go.mod": "module example.com/api\n",
	})
	d := project.DetectType(inv)

	if d.Primary != project.TypeGo {
		t.Errorf("expected Go from nested go.mod, got %q", d.Primary)
	}
}

func TestDetectType_MarkersPopulated(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"go.mod": "module example.com/test\n",
	})
	d := project.DetectType(inv)

	if len(d.Markers) == 0 {
		t.Error("expected Markers to be populated")
	}
	found := false
	for _, m := range d.Markers {
		if m == "go.mod" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected go.mod in Markers, got %v", d.Markers)
	}
}

func TestDetectType_GoSourceWithoutGoMod(t *testing.T) {
	inv := scanDir(t, map[string]string{
		"main.go":         "package main\n",
		"internal/app.go": "package internal\n",
		"README.md":       "# fixture\n",
	})

	d := project.DetectType(inv)

	if d.Primary != project.TypeGo {
		t.Errorf("expected %q from Go source fallback, got %q", project.TypeGo, d.Primary)
	}
	if len(d.AllDetected) == 0 {
		t.Fatal("expected AllDetected to be populated")
	}
	if d.AllDetected[0] != project.TypeGo {
		t.Errorf("expected AllDetected[0] = %q, got %q", project.TypeGo, d.AllDetected[0])
	}
}
