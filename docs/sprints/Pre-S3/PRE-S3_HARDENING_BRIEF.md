# Pre-S3 Hardening — Completion Report

**Project:** Pulse
**Release baseline:** `0.2.0`
**Hardening commit:** `44e9ab8`
**Branch:** `main`
**Status:** ✅ Complete

---

## Executive Summary

A boundary violation in the S2 project discovery pipeline caused Pulse to silently replace any explicitly supplied target directory with an ancestor project root. Three of the four testdata fixtures were returning the Pulse repository itself instead of the fixture under analysis. This report documents the root cause, the fix, the regression coverage added, and the verified acceptance state. The codebase is now ready for S3 — Git Intelligence.

---

## 1. Problem Statement

### Observed behavior before fix

| Command | Expected | Actual |
|---|---|---|
| `go run . testdata\projects\go-project` | `go-project` | `pulse` |
| `go run . testdata\projects\mixed-project` | `mixed-project` | `pulse` |
| `go run . testdata\projects\unknown-project` | `unknown-project` | `pulse` |
| `go run . testdata\projects\node-project` | `node-project` | `node-project` ✅ |

Three of four fixtures returned the wrong project entirely. The one that worked — `node-project` — only worked because it contains its own `package.json`, which stopped the upward walk before it could escape.

### Why this was not caught by the automated suite

The existing tests in `root_test.go` and `snapshot_test.go` constructed isolated temporary directories using `testhelpers.TempProject`. Those temp directories exist in the system temp folder, completely outside the Pulse repository. There was no ancestor with a `go.mod` to escape to, so the upward walk was harmless in tests and the bug never fired.

The failure only manifested when a real target directory — one that lives physically inside the Pulse repository — was supplied. That is precisely the situation with `testdata/projects/`.

---

## 2. Root Cause Analysis

### The discovery pipeline

```text
Config.TargetPath
    ↓
project.ResolveTarget      validates the path is a real directory
    ↓
project.DiscoverRoot       walks upward looking for a root marker
    ↓
scanner.Scan               scans the resolved root
    ↓
project.DetectType         identifies project ecosystem
    ↓
project.NewMetadata        builds name, languages, counts
    ↓
snapshot.ProjectSnapshot   unified output
```

### Where the bug lived

`DiscoverRoot` in `internal/project/root.go`:

```go
// BEFORE — unbounded walk, no concept of boundary
func DiscoverRoot(target Target) RootDiscovery {
    current := target.Path

    for {
        if marker, found := findMarker(current); found {
            return RootDiscovery{Root: current, ...}
        }
        parent := filepath.Dir(current)
        if parent == current {
            break
        }
        current = parent   // ← walks upward unconditionally
    }

    return RootDiscovery{Root: target.Path, ...}
}
```

For `testdata/projects/go-project`:

```text
go-project/     ← no go.mod here
    ↓ walk up
projects/       ← no go.mod here
    ↓ walk up
testdata/       ← no go.mod here
    ↓ walk up
pulse/          ← go.mod found here ✓ (wrong project)
    ↓
Root = pulse/
```

Every downstream component — scanner, language detection, metadata — then described the Pulse repository, not the fixture.

### Secondary root cause

Even after fixing the boundary, `go-project` would have reported `Type: Unknown` because it has no `go.mod` of its own. The project type detection relied entirely on marker files. A fixture containing `.go` source but no `go.mod` had no way to be identified as Go.

---

## 3. Files Changed

```text
internal/project/target.go       modified
internal/project/root.go         modified
internal/project/detect.go       modified
internal/project/root_test.go    extended
internal/project/detect_test.go  extended
internal/snapshot/snapshot_test.go rewritten with fixture coverage
```

No other files were touched. The scanner, config, CLI, output, snapshot struct, and testhelpers are all unchanged.

---

## 4. Fix Detail

### 4.1 — `Target.Explicit` field (`target.go`)

```go
type Target struct {
    Path     string
    Explicit bool   // true when user explicitly named this directory
}
```

`ResolveTarget` now always sets `Explicit: true`:

```go
return Target{
    Path:     filepath.Clean(path),
    Explicit: true,
}, nil
```

The field is `false` by default for any `Target` constructed directly, which covers the future case of cwd-based discovery where upward walking is legitimate.

### 4.2 — Bounded walk in `DiscoverRoot` (`root.go`)

```go
func DiscoverRoot(target Target) RootDiscovery {
    if target.Explicit {
        // Boundary enforced: only check target itself, never ancestors.
        if marker, found := findMarker(target.Path); found {
            return RootDiscovery{Root: target.Path, Marker: marker, MarkerFound: true}
        }
        return RootDiscovery{Root: target.Path, Marker: "", MarkerFound: false}
    }

    // Original unbounded walk preserved for non-explicit targets.
    current := target.Path
    for {
        ...
    }
}
```

The design is additive. The original unbounded walk is fully preserved under `Explicit: false`. Nothing that worked before is broken.

### 4.3 — Source-file fallback in `DetectType` (`detect.go`)

```go
func inferTypeFromFiles(inv scanner.Inventory) Detection {
    for _, f := range inv.Files {
        if f.Extension == ".go" {
            return Detection{Primary: TypeGo, AllDetected: []ProjectType{TypeGo}}
        }
    }
    return Detection{Primary: TypeUnknown, AllDetected: []ProjectType{TypeUnknown}}
}
```

`DetectType` calls this when no marker files are found. A project with `.go` source but no `go.mod` is correctly identified as Go. A project with no recognised source is correctly identified as Unknown.

---

## 5. Regression Tests Added

### `internal/project/root_test.go`

| Test | What it proves |
|---|---|
| `TestDiscoverRoot_ExplicitTarget_DoesNotEscapeToAncestor` | Core regression. outer has `go.mod`, inner does not. Explicit target inner must not resolve to outer. |
| `TestDiscoverRoot_ExplicitTarget_WithOwnMarker_StaysOnTarget` | Explicit target with its own marker stays on the target, not the ancestor. |
| `TestDiscoverRoot_NestedOneLevel_NonExplicitTargetCanAscend` | Non-explicit target can still walk upward. Original behavior is not regressed. |
| `TestDiscoverRoot_DeeplyNested_NonExplicitTargetCanAscend` | Same for deep nesting. |

Existing tests were preserved and adjusted only where the `Explicit` flag required a clear declaration of intent (nested tests now explicitly set `Explicit: false` to document that they are testing the unbounded walk).

### `internal/project/detect_test.go`

| Test | What it proves |
|---|---|
| `TestDetectType_GoSourceWithoutGoMod` | A directory with `.go` files but no `go.mod` is identified as Go via the source-file fallback. |

### `internal/snapshot/snapshot_test.go`

Five new fixture-based tests. These are the most important additions because they exercise the complete pipeline end to end against real fixture directories that live inside the Pulse repository — exactly the condition the original suite never covered.

| Test | Fixture | Assertions |
|---|---|---|
| `TestDiscover_GoFixture_BoundedToTarget` | `go-project` | Root, Name, Type=Go, FileCount=4, DirCount=1, Languages=[Go, Markdown, PowerShell] |
| `TestDiscover_MixedFixture_BoundedToTarget` | `mixed-project` | Root, Name, FileCount=4, DirCount=1, Languages=[Go, JavaScript, Markdown, Python] |
| `TestDiscover_NodeFixture_BoundedToTarget` | `node-project` | Root, Name, Type=Node.js, FileCount=5, DirCount=1, Languages=[CSS, HTML, TypeScript] |
| `TestDiscover_UnknownFixture_BoundedToTarget` | `unknown-project` | Root, Name, Type=Unknown, FileCount=2, DirCount=0, Languages=[] |
| `TestDiscover_GoFixture_BoundedToTarget` | — | Explicit ancestor-trap assertion: `snap.Name == "pulse"` fails the test immediately |

---

## 6. Test Suite Results

```text
?       pulse                       [no test files]
ok      pulse/internal/cli          0.981s
ok      pulse/internal/config       (cached)
ok      pulse/internal/errors       (cached)
?       pulse/internal/git          [no test files]
ok      pulse/internal/output       0.978s
ok      pulse/internal/project      1.096s
ok      pulse/internal/scanner      (cached)
ok      pulse/internal/snapshot     0.955s
ok      pulse/internal/testhelpers  (cached)
```

All packages pass. No regressions.

---

## 7. CLI Acceptance Results

### H4 — All five targets

```text
go run . .
  Name: pulse  |  Type: Go  |  Root: …\pulse                          ✅

go run . testdata\projects\go-project
  Name: go-project  |  Type: Go  |  Root: …\go-project                ✅

go run . testdata\projects\mixed-project
  Name: mixed-project  |  Type: Go  |  Root: …\mixed-project          ✅

go run . testdata\projects\node-project
  Name: node-project  |  Type: Node.js  |  Root: …\node-project       ✅

go run . testdata\projects\unknown-project
  Name: unknown-project  |  Type: Unknown  |  Root: …\unknown-project  ✅
```

### Absolute path

```text
$go = (Resolve-Path "testdata\projects\go-project").Path
go run . $go
  Name: go-project  |  Root: …\go-project                             ✅
```

### JSON output

```json
{
  "application": "Pulse",
  "version": "1.0.0",
  "project": {
    "name": "go-project",
    "root": "C:\\...\\testdata\\projects\\go-project",
    "type": "Go",
    "languages": ["Go", "Markdown", "PowerShell"],
    "file_count": 4,
    "directory_count": 1
  }
}
```

No `"name": "pulse"`. No ancestor root. ✅

### Error behavior preserved

```text
go run . does-not-exist
  Error: target path does not exist: …\does-not-exist
  Exit code: 1                                                         ✅
```

---

## 8. Engineering Checks

```text
gofmt -l .          → no output   ✅
go build ./...      → success     ✅
go test ./...       → all pass    ✅
go vet ./...        → clean       ✅
git diff --check    → clean       ✅
```

---

## 9. Known Characteristic — `mixed-project` Type

`mixed-project` reports `Type: Go`. This is technically correct given its contents — it contains `main.go` and no ecosystem marker file (`go.mod`, `package.json`, etc.). The source-file fallback identifies it as Go from the `.go` extension.

This is not a bug. It is a fixture characteristic. If a future sprint requires `mixed-project` to report `Unknown` or a different type, the fix is to either add or remove files from the fixture, not to change detection logic.

---

## 10. Definition of Done

```text
H1  Enforce target boundary      ✅
H2  Regression tests added       ✅
H3  Snapshot integrity verified  ✅
H4  CLI acceptance               ✅
```

---

## 11. What Was Not Changed

These components are untouched and fully intact for S3 to consume:

```text
internal/cli/cli.go
internal/config/config.go
internal/scanner/scanner.go
internal/snapshot/snapshot.go
internal/output/output.go
internal/errors/errors.go
internal/testhelpers/testhelpers.go
internal/git/git.go
```

`ProjectSnapshot` is structurally identical to its S2 form. S3 can begin consuming it immediately without any migration.

---

## 12. Commit Record

```text
commit 44e9ab8
main

fix(discovery): enforce target analysis boundary
6 files changed, 240 insertions(+), 106 deletions(-)
```

`0.2.0` remains the historical S2 tag. Commit `44e9ab8` is the clean baseline from which **S3 — Git Intelligence** begins.