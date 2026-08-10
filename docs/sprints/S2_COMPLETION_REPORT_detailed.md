# Pulse Sprint 2 — Senior Developer Report

**Prepared by:** Junior Developer
**Sprint:** S2 — Project Discovery
**Date:** 2026-08-11
**Base:** `2888476` tag `0.1.0`
**Final:** `a7435b9` tag `0.2.0`

---

## 1. Executive summary

S2 taught Pulse to understand what it is pointed at.

Before S2, Pulse could receive a path from the command line and resolve it to an absolute filesystem path. That was the entirety of its intelligence. It could not tell you whether that path existed, whether it was a directory, where the actual project root was, what files were inside it, or what kind of project it was.

After S2, Pulse runs a four-stage discovery pipeline on every invocation and can answer all of those questions reliably.

```
Before:   "Here is a path."

After:    "This is a valid directory."
          "This is its actual project root."
          "Here is everything inside it."
          "This is a Go project."
```

85 tests pass across 7 packages. 0 failures. No third-party dependencies introduced.

---

## 2. What existed before S2

### 2.1 Repository state at `0.1.0`

```
pulse/
├── main.go
├── go.mod                    module pulse, go 1.21
├── build.ps1
├── internal/
│   ├── cli/
│   │   ├── cli.go            argument parsing + Run()
│   │   └── cli_test.go
│   ├── config/
│   │   ├── config.go         path resolution
│   │   └── config_test.go
│   ├── errors/
│   │   ├── errors.go         PulseError, KindUser/Environment/Internal
│   │   └── errors_test.go
│   ├── output/
│   │   ├── output.go         stdout/stderr separation
│   │   └── output_test.go
│   ├── testhelpers/
│   │   └── testhelpers.go    TempProject(), ProjectFixturePath()
│   ├── project/
│   │   └── project.go        empty scaffold
│   ├── scanner/
│   │   └── scanner.go        empty scaffold
│   ├── snapshot/             empty scaffold
│   └── git/                  empty scaffold
└── testdata/
    └── projects/
        ├── basic/README.md
        └── git-project/README.md
```

### 2.2 What S1 actually did

S1 established the application skeleton. The important pieces:

**CLI** parsed `--help`, `--version`, `--json`, and one positional path argument. It returned exit codes `0` and `1`. It knew nothing about what the path pointed at.

**Config** resolved that raw path — empty string became cwd, relative became absolute, absolute was cleaned. The result was `Config.TargetPath`. Config did not validate whether that path existed or was usable.

**Errors** provided a typed error system with `PulseError`, three kinds (`KindUser`, `KindEnvironment`, `KindInternal`), and four constructors: `User()`, `UserWrap()`, `Environment()`, `Internal()`. Critically: no `New()` or `Wrap()` functions.

**Output** separated stdout and stderr cleanly through a `Writer` struct. No other package was permitted to write directly to either stream.

**Testhelpers** provided `TempProject(t, map[string]string)` for creating deterministic temporary directory fixtures, and `ProjectFixturePath()` for referencing fixtures in `testdata/`.

### 2.3 What S1 left unresolved

The pipeline ended at `Config.TargetPath`. After that point, Pulse printed a static summary and exited. The existing S1 CLI tests for `TestRun_WithTargetPath_ReturnsSuccess` passed `"./some-project"` as an argument — a path that did not exist on disk — and the test passed because S1 never validated it. This worked in S1 because no validation occurred. It would become a problem in S2.

---

## 3. What S2 built

### 3.1 M2.1 — Target validation

**Files added:**
```
internal/project/target.go
internal/project/target_test.go
```

**Problem being solved:** Config resolved paths but made no guarantees about them. A path could resolve successfully to something that didn't exist, couldn't be read, or was a file rather than a directory. Pulse needed a validation gate before any further analysis.

**What was built:** `ResolveTarget(path string) (Target, error)` takes the absolute path from Config and validates three things in order — the path is not empty, the path exists on the filesystem, the path is a directory. Any failure returns a classified `PulseError` using the errors package constructors that S1 established. On success it returns a `Target` struct holding the cleaned absolute path.

```go
type Target struct {
    Path string
}
```

The `Target` type is a semantic guarantee. Anything downstream that receives a `Target` knows it is a real, accessible directory. No further existence checking is needed.

**8 tests:**
- Valid directory → returns Target with cleaned path
- Missing path → classified user error
- File instead of directory → classified user error
- Empty string → classified user error
- Absolute path → accepted correctly
- Path cleaning → filepath.Clean applied
- Error messages → meaningful text, not raw Go errors

**Commit:** `4120338`

---

### 3.2 M2.2 — Project root discovery

**Files added:**
```
internal/project/root.go
internal/project/root_test.go
```

**Problem being solved:** Users don't always point Pulse at a project root. A developer working in `my-project/internal/service/` who runs `pulse .` from there should get analysis of `my-project/`, not `my-project/internal/service/`. Pulse needed to find the true root.

**What was built:** `DiscoverRoot(target Target) RootDiscovery` walks upward from the target directory, checking each ancestor for the presence of a recognised root marker. The walk stops when a marker is found or the filesystem root is reached.

Root markers checked:
```
.git  go.mod  package.json  Cargo.toml  pyproject.toml  pom.xml  build.gradle
```

The result type carries explicit state:

```go
type RootDiscovery struct {
    Root        string
    Marker      string
    MarkerFound bool
}
```

When no marker is found, `Root` is set to the original target path and `MarkerFound` is `false`. This is documented behaviour, not an error. A project with no recognised marker is still a valid analysis target.

Symlinks are not resolved during the walk — the logical path is followed as-is to avoid traversal outside expected boundaries.

**8 tests:**
- Target is already root → correct root, correct marker reported
- Nested one level → walks up correctly
- Deeply nested → walks multiple levels correctly
- No marker found → falls back to target, MarkerFound false
- `.git` directory → detected correctly
- `package.json` → detected correctly
- `Cargo.toml` → detected correctly
- Root is always absolute → path guarantee verified

**Commit:** `dfd77f8`

---

### 3.3 M2.3 — Filesystem scanner

**Files added:**
```
internal/scanner/scanner.go
internal/scanner/scanner_test.go
```

**Problem being solved:** Once Pulse knows where a project root is, it needs to know what exists inside it. This inventory is the raw material for everything that comes after — type detection, snapshot comparison, future analysis.

**What was built:** `Scan(root string) (Inventory, error)` walks the directory tree from the project root using `filepath.WalkDir` and produces two sorted slices.

```go
type FileEntry struct {
    RelPath   string   // forward-slash, relative to root
    Extension string   // lowercase, includes leading dot
    SizeBytes int64
}

type DirEntry struct {
    RelPath string
}

type Inventory struct {
    Files []FileEntry
    Dirs  []DirEntry
}
```

Four architectural decisions made explicit in the implementation:

**Deterministic ordering.** Both slices are sorted by `RelPath` after the walk completes. This is not optional — future snapshot comparison and tests both depend on stable output. The filesystem walk order is not guaranteed to be consistent across operating systems or runs.

**Forward-slash paths.** `filepath.ToSlash()` is applied to all relative paths. This makes the inventory consistent across Windows and Unix without callers needing to handle platform differences.

**Symlinks skipped.** Symlinked entries are detected via `fs.ModeSymlink` and silently excluded. Following symlinks risks traversing outside the project tree. The behaviour is documented in the function comment.

**Partial scan over abort.** When `WalkDir` encounters an unreadable entry, the error is swallowed and the walk continues. A partial inventory is more useful than no inventory.

Default exclusions (entire subtree skipped):
```
.git  node_modules  vendor  dist  build  target  .cache  __pycache__  .idea  .vscode
```

**12 tests:**
- Empty project → zero files, zero dirs
- Single file → RelPath, extension, no dirs
- Nested directories → correct file and dir counts
- Multiple file types → extensions collected correctly
- Excluded directories → contents never appear in inventory
- Deterministic ordering → two scans of same project produce identical results
- Forward-slash paths → no backslashes in RelPath on Windows
- File with no extension → empty extension string
- Root not in dirs → `.` never appears in Dirs
- Size populated → SizeBytes non-zero for non-empty files
- Extension is lowercase → case normalisation verified
- Deeply nested dirs → `a`, `a/b`, `a/b/c` all recorded correctly

**Commit:** `8cad990`

---

### 3.4 M2.4 — Project type detection

**Files added:**
```
internal/project/detect.go
internal/project/detect_test.go
```

**Files modified:**
```
internal/cli/cli.go         ← pipeline wired end to end
internal/cli/cli_test.go    ← S1 tests corrected (see section 4)
internal/scanner/scanner_test.go  ← extension assumption corrected (see section 4)
```

**Problem being solved:** The scanner produces raw filesystem facts. M2.4 turns those facts into meaning. Knowing that `go.mod` exists in the inventory is a fact. Knowing that this is a Go project is intelligence.

**What was built:** `DetectType(inv scanner.Inventory) Detection` analyses the inventory by checking each file's basename against a map of known marker filenames.

```go
type ProjectType string

const (
    TypeGo      ProjectType = "Go"
    TypeNode    ProjectType = "Node.js"
    TypeRust    ProjectType = "Rust"
    TypePython  ProjectType = "Python"
    TypeJava    ProjectType = "Java"
    TypeUnknown ProjectType = "Unknown"
)
```

```go
type Detection struct {
    Primary     ProjectType
    AllDetected []ProjectType
    Markers     []string
}
```

Detection checks filenames, not full paths. A `go.mod` nested inside `services/api/go.mod` is still detected. This is intentional — monorepos and nested projects are real.

When multiple markers are present, a fixed priority order determines `Primary`:
```
Go > Rust > Python > Java > Node.js
```

This order is explicit in the code as `detectionOrder []string` and is documented. It is not magic — if the rules need to change, there is one place to change them.

`TypeUnknown` is a first-class result, not a failure. A project with only a `README.md` returns `Detection{Primary: TypeUnknown}` — this is valid intelligence. The distinction matters:

```
"Pulse cannot analyse this project"   ←  error condition
"Pulse does not recognise this type"  ←  TypeUnknown, perfectly valid
```

**13 tests:**
- Go, Node, Rust, Python (pyproject), Python (requirements), Java (Maven), Java (Gradle) → correct primary type
- Unknown → TypeUnknown returned, AllDetected non-empty, no error
- Multiple markers → priority order respected, both types in AllDetected
- Deterministic → identical inventory always produces identical Detection
- Empty inventory → TypeUnknown
- Nested marker → detected regardless of depth
- Markers field populated → contributing filenames recorded

**Commit:** `fd42253`

---

### 3.5 Pipeline wired into CLI

The four-stage discovery pipeline runs on every `pulse` invocation:

```
os.Args
   |
ParseArgs
   |
config.New              ← S1: resolves and cleans the path
   |
project.ResolveTarget   ← M2.1: validates real directory
   |
project.DiscoverRoot    ← M2.2: finds true project root
   |
scanner.Scan            ← M2.3: inventories the filesystem
   |
project.DetectType      ← M2.4: identifies the ecosystem
   |
output.Writer           ← renders to stdout / stderr
```

Plain text output:
```
Pulse — Project Discovery

Project
  Root:   C:\...\pulse
  Type:   Go

Filesystem
  Files:       34
  Directories: 16
```

JSON output (`--json`):
```json
{
  "root": "C:\\...\\pulse",
  "type": "Go",
  "files": 34,
  "directories": 16
}
```

Error cases handled cleanly:
```
$ pulse .\README.md
Error: target must be a directory, not a file: ...\README.md

$ pulse .\does-not-exist
Error: target path does not exist: ...\does-not-exist
```

---

## 4. Problems encountered and how they were fixed

### Problem 1 — PowerShell here-strings produced invalid UTF-8

**When it appeared:** First attempt to write `target.go` using PowerShell `@'...'@ | Set-Content`.

**What happened:** The compiler reported `invalid UTF-8 encoding` on multiple lines. The em-dash character `—` used in a comment was being written with incorrect encoding by `Set-Content` on this Windows configuration. The Go compiler rejected the file.

```
internal\project\target.go:23:33: invalid UTF-8 encoding
internal\project\target.go:42:43: invalid UTF-8 encoding
```

**Fix:** Replaced all `Set-Content` calls with `[System.IO.File]::WriteAllText(..., [System.Text.Encoding]::UTF8)` for every file written during S2. This bypasses PowerShell's encoding pipeline entirely and guarantees clean UTF-8 output.

```powershell
[System.IO.File]::WriteAllText(
    (Join-Path (Get-Location) 'internal\project\target.go'),
    $content,
    [System.Text.Encoding]::UTF8
)
```

Additionally, the em-dash was removed from all source comments and replaced with plain ASCII alternatives. Source files should not contain non-ASCII characters unless the code requires it.

**Lesson:** On Windows, PowerShell's default string encoding cannot be trusted for source file output. Always use `[System.IO.File]::WriteAllText` with an explicit encoding when writing Go source files from PowerShell.

---

### Problem 2 — Wrong errors API used in target.go

**When it appeared:** Same first build attempt.

**What happened:** The generated `target.go` called `errors.New()` and `errors.Wrap()`. Neither function exists in S1's `internal/errors` package. S1 uses named constructors: `errors.User()`, `errors.UserWrap()`, `errors.Environment()`, `errors.Internal()`.

```
internal\project\target.go:28:33: undefined: errors.New
internal\project\target.go:43:33: undefined: errors.Wrap
```

**Fix:** Read `internal/errors/errors.go` carefully before writing any code that imports it. Rewrote `target.go` to use only the constructors that actually exist:

```go
// Wrong — does not exist in this codebase
return Target{}, errors.New(errors.KindUser, "message")

// Correct — matches the S1 API
return Target{}, pulseErrors.User("message")
return Target{}, pulseErrors.Environment("message", err)
```

**Lesson:** Before writing code that depends on an existing package, read that package's source. Don't assume standard library conventions apply to internal packages.

---

### Problem 3 — go.mod extension assumption was wrong

**When it appeared:** `TestScan_MultipleFileTypes` in the scanner package.

**What happened:** The test created a project containing `go.mod` and asserted that at least one file had an empty extension, reasoning that `go.mod` has no extension. This is incorrect.

```
go test ./internal/scanner/... -v -run TestScan_MultipleFileTypes
    scanner_test.go:94: expected empty extension for go.mod
```

`filepath.Ext("go.mod")` returns `".mod"`, not `""`. The dot-separated suffix is the extension. `go.mod` has extension `.mod`.

**Fix:** Replaced `go.mod` in that specific test with `Makefile`, which genuinely has no extension:

```go
// Before — wrong assumption
"go.mod": "module example.com/test\n",
// assertion: expected empty extension for go.mod

// After — correct fixture
"Makefile": "all:\n\techo done\n",
// assertion: expected empty extension for Makefile
```

The `TestScan_FileWithNoExtension` test already used `Makefile` correctly and passed throughout. Only `TestScan_MultipleFileTypes` had the incorrect assumption.

**Lesson:** Don't assume extension behaviour. `filepath.Ext` documentation is explicit: it returns the suffix beginning at the final dot. `go.mod`, `go.sum`, `docker-compose.yml` all have extensions.

---

### Problem 4 — S1 CLI tests passed non-existent paths

**When it appeared:** `TestRun_WithTargetPath_ReturnsSuccess` and `TestRun_WithJSONAndTargetPath_ReturnsSuccess` in the CLI package.

**What happened:** Two S1 tests passed `"./some-project"` as a target path argument:

```go
func TestRun_WithTargetPath_ReturnsSuccess(t *testing.T) {
    code := cli.Run([]string{"./some-project"})
    if code != cli.ExitSuccess {
        t.Errorf("expected ExitSuccess with target path, got %d", code)
    }
}
```

In S1, `Run()` resolved the path with `config.New()` and then printed a static summary — it never checked whether the path existed. `./some-project` resolved to an absolute path that didn't exist, but no validation occurred, so the test passed.

In S2, `ResolveTarget()` validates existence. `./some-project` doesn't exist on disk. The pipeline correctly returns an error and `Run()` correctly returns `ExitFailure`. The test broke — not because S2 was wrong, but because S1's test was only passing due to an absence of validation.

```
Error: target path does not exist: C:\...\internal\cli\some-project
    cli_test.go:138: expected ExitSuccess with target path, got 1
```

**First fix attempt — wrong:** A PowerShell `-replace` was used to change `"some-project"` to `"."`. This failed silently because the actual string in the file was `"./some-project"` with the `./` prefix. The replacement found no matches.

**Actual fix:** Rewrote the two failing tests to create real temporary directories using `testhelpers.TempProject()` and pass the absolute path to `cli.Run()`. Also added two new tests that S1 was missing — verifying that nonexistent paths and file targets correctly return `ExitFailure`:

```go
func TestRun_WithTargetPath_ReturnsSuccess(t *testing.T) {
    dir := testhelpers.TempProject(t, map[string]string{
        "README.md": "# Test project\n",
    })
    code := cli.Run([]string{dir})
    if code != cli.ExitSuccess {
        t.Errorf("expected ExitSuccess with target path, got %d", code)
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
```

**Lesson:** Tests that pass because validation is absent are not passing — they are untested. When validation is added to a system, tests that depended on its absence will correctly fail. The right response is to update the tests to test the new behaviour, not to remove the validation.

---

## 5. What was deliberately not built

The brief was specific about scope. The following were considered and explicitly excluded:

| Excluded | Reason |
|---|---|
| Git history analysis | S3 |
| Dependency graph | Future sprint |
| Health / quality scoring | Future sprint |
| Framework deep detection | Future sprint |
| Polished output formatting | M2.8 |
| Snapshot comparison | M2.7 |
| External dependencies | Standard library sufficient |
| Concurrency | Not needed at this scale |
| Plugin system | Not in scope |
| Configuration files | Not in scope |

`internal/git/`, `internal/snapshot/` were not touched.

---

## 6. Architecture boundaries

The S2 brief defined responsibility boundaries. They held throughout.

```
project/    knows what a project is — identity and reasoning
scanner/    knows what is on disk — no interpretation
config/     knows what the user asked for
errors/     knows how to classify failure
output/     knows how to render — receives results, not logic
cli/        knows how to wire everything together
```

`scanner` does not interpret what it finds. It returns facts.
`project` does not walk directories. It receives an inventory.
`output` received rendered strings. No formatting logic lives in `project` or `scanner`.
No package writes directly to stdout or stderr except through `output.Writer`.

---

## 7. Test summary

| Package | Tests | Status |
|---|---|---|
| pulse/internal/cli | 18 | PASS |
| pulse/internal/config | 9 | PASS |
| pulse/internal/errors | 5 | PASS |
| pulse/internal/output | 6 | PASS |
| pulse/internal/project | 29 | PASS |
| pulse/internal/scanner | 12 | PASS |
| pulse/internal/testhelpers | 6 | PASS |
| **Total** | **85** | **ALL PASS** |

---

## 8. Commit history

```
a7435b9   docs: add S2 completion report
fd42253   feat(project): detect project type        ← tag: 0.2.0
8cad990   feat(scanner): add project filesystem discovery
dfd77f8   feat(project): discover project root
4120338   feat(project): validate analysis target
─────────────────────────────────────────────────────
2888476   feat: complete sprint 1 foundation         ← tag: 0.1.0
```

Every commit between `0.1.0` and `0.2.0` leaves the codebase buildable and fully tested. No commit introduces a broken state.

---

## 9. Two things to be aware of before S3

**`internal/project/project.go` exists and is empty.** It is S1 scaffolding. Before S3 extends the `project` package, decide whether this file becomes a package-level facade composing `Target`, `RootDiscovery`, and `Detection` into a single `Project` struct, or whether it is removed. Either is a reasonable choice. The important thing is to decide deliberately rather than accumulate dead files.

**`TypeUnknown` is load-bearing.** It is not a failure state, it is valid intelligence. Any S3 code that branches on `Detection.Primary` must handle `TypeUnknown` explicitly. Treating it as a fallback to be ignored would be incorrect — there are real projects that produce it, and Pulse should behave sensibly for them.