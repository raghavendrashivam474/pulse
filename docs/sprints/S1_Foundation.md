# Pulse v1.0 — Sprint 1 Continuation
## Formal Engineering Report: M1.5 → M1.8 Implementation

**Report Type:** Sprint Implementation Post-Mortem
**Project:** Pulse v1.0.0
**Sprint:** S1 — Foundation (Continuation)
**Scope:** M1.5 → M1.8
**Author:** Implementation Engineer
**Reported To:** Senior Developer
**Status:** Sprint 1 Complete

---

## 1. Executive Summary

This report documents the implementation of Milestones 1.5 through 1.8 of Sprint 1 for the Pulse CLI project. The objective of this continuation sprint was to complete the engineering foundation established in M1.1–M1.4 by adding structured error handling, a runtime configuration model, a testing foundation, and a reproducible build baseline.

Implementation was completed successfully. All four milestones are closed. The final verification suite passes cleanly. Two infrastructure problems were encountered during execution — both were diagnosed, understood, and resolved without altering the architectural design or scope of any milestone.

The foundation is confirmed ready for S2 — Project Discovery.

---

## 2. Scope Delivered

The following packages and artifacts were created during this continuation sprint.

### 2.1 New Packages

| Package | Purpose |
|---|---|
| `internal/errors` | Application-level error types with kind classification |
| `internal/config` | Runtime configuration model and path resolution |
| `internal/testhelpers` | Shared test utilities for deterministic filesystem testing |

### 2.2 Updated Packages

| Package | Changes |
|---|---|
| `internal/output` | Added `PrintError()` writing to stderr. Added explicit `Writer` struct with injectable `Out` and `Err` fields for testability |
| `internal/cli` | Wired through `Config`. Added `ExitSuccess` / `ExitFailure` constants. Error path now uses `output.Writer.PrintError()` |

### 2.3 New Files

| File | Purpose |
|---|---|
| `build.ps1` | Cross-platform release build script |
| `testdata/projects/basic/` | Committed minimal project fixture |
| `testdata/projects/git-project/` | Committed minimal project fixture |
| `docs/sprints/S1_Foundation.md` | Sprint documentation and conventions reference |

### 2.4 Final Verification Results

```text
go build ./...     clean
go test ./...      all packages pass
go vet ./...       clean
gofmt -l .         clean

pulse              Pulse v1.0.0 / No project analysis available yet.
pulse --version    Pulse v1.0.0
pulse --json       { "version": "1.0.0", "status": "no_analysis" }
pulse --unknown    Error: unknown flag: --unknown  [exit 1]
```

---

## 3. Problems Encountered

Two distinct infrastructure problems were encountered. Neither was a design problem. Neither required any architectural change. Both were tooling and environment issues that are important to document so they are not repeated in S2.

---

### Problem 1: Module Name Mismatch

#### 3.1.1 Description

The first build attempt failed immediately with:

```text
main.go:3:8: package pulse/internal/cli is not in std
(C:\Program Files\Go\src\pulse\internal\cli)
```

The binary was never produced. All subsequent commands that attempted to run `pulse.exe` failed with:

```text
The term '.\pulse.exe' is not recognized as the name of a cmdlet,
function, script file, or operable program.
```

This was a cascading failure. The binary did not exist because the build had failed. Every downstream command failed for the same root cause.

#### 3.1.2 Root Cause

The `go.mod` file contained the module declaration:

```text
module github.com/raghavendrashivam474/pulse
```

All import paths in the source code used:

```go
import "pulse/internal/cli"
import "pulse/internal/config"
import pulseErrors "pulse/internal/errors"
```

Go requires that import paths begin with the module name declared in `go.mod`. Because the module name was `github.com/raghavendrashivam474/pulse` and the imports used `pulse`, Go could not resolve the packages as internal module packages. It attempted to find them in the Go standard library instead, which is why the error message referenced `C:\Program Files\Go\src\pulse\internal\cli`.

The mismatched module name was an artifact of how the repository was originally initialised. The GitHub-qualified name is conventional for public repositories but was inconsistent with the import paths that had been written throughout the codebase.

#### 3.1.3 Resolution Attempted — First

The `go.mod` file was rewritten to:

```text
module pulse

go 1.21
```

This was the correct fix in principle. However, the rewrite was performed using PowerShell's `Set-Content -Encoding UTF8` command, which introduced a secondary problem documented in Problem 2 below. The module name fix was correct but was obscured by the second failure.

#### 3.1.4 Resolution — Final

The module name `pulse` was confirmed as the correct value. All import paths in all source files already used `pulse/internal/...` consistently. Once the file encoding issue was separately resolved, the module name fix took effect and the build succeeded.

#### 3.1.5 Lesson

When initialising a Go repository, the module name in `go.mod` must be decided before any source files are written with import paths. Once import paths are established throughout the codebase, the module name in `go.mod` must match them exactly. A GitHub-qualified module name is appropriate for public distribution but must be reflected consistently in every import statement if used.

---

### Problem 2: UTF-8 BOM in go.mod

#### 3.2.1 Description

After correcting the module name, the build still failed with a new error:

```text
go: errors parsing go.mod:
go.mod:1: unexpected input character '\ufeff'
```

The `\ufeff` character is the Unicode Byte Order Mark (BOM). It is an invisible character that some editors and tools prepend to UTF-8 encoded files. It appeared at the very start of `go.mod`, before the `m` in `module`.

#### 3.2.2 Root Cause

PowerShell 5 — the version shipped with Windows — writes a UTF-8 BOM by default when using:

```powershell
Set-Content -Encoding UTF8
```

This is a known and documented behaviour of PowerShell 5. The BOM (`EF BB BF` in hexadecimal, `\ufeff` in Unicode) is invisible in most text editors and terminals. It does not cause problems in many contexts. However, the Go toolchain's `go.mod` parser is strict and rejects any file that begins with a BOM character.

This was confirmed by inspecting the raw bytes of the file:

```powershell
$bytes = [System.IO.File]::ReadAllBytes("go.mod")
Write-Host "First 3 bytes: $($bytes[0]) $($bytes[1]) $($bytes[2])"
```

A clean file begins with bytes `109 111 100` — the ASCII codes for `m`, `o`, `d`. A BOM-prefixed file begins with `239 187 191` before the actual content.

#### 3.2.3 Impact

The BOM affected every file written during the session using `Set-Content -Encoding UTF8`. This included:

```text
go.mod
main.go
internal/cli/cli.go
internal/cli/cli_test.go
internal/config/config.go
internal/config/config_test.go
internal/errors/errors.go
internal/errors/errors_test.go
internal/output/output.go
internal/output/output_test.go
internal/testhelpers/testhelpers.go
internal/testhelpers/testhelpers_test.go
```

The `go.mod` BOM caused an immediate hard failure. The BOM in Go source files did not prevent compilation in this case but caused `gofmt` to report all source files as needing formatting correction, which would have caused the `gofmt -l .` check to fail.

#### 3.2.4 Resolution

All files were rewritten using `[System.IO.File]::WriteAllText()` with an explicit `UTF8Encoding` object constructed with BOM suppressed:

```powershell
$encoding = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText($fullPath, $content, $encoding)
```

The `$false` parameter instructs the encoder to omit the BOM. This is the correct method for writing BOM-free UTF-8 files from PowerShell 5.

A reusable helper function was defined for the session:

```powershell
function Write-NoBOM {
    param([string]$Path, [string]$Content)
    $encoding = New-Object System.Text.UTF8Encoding $false
    $fullPath = Join-Path (Get-Location) $Path
    $dir = Split-Path $fullPath
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
    [System.IO.File]::WriteAllText($fullPath, $Content, $encoding)
}
```

All affected files were rewritten through this function. After rewriting, byte inspection confirmed the BOM was absent from `go.mod`:

```text
First 3 bytes: 109 111 100  (= 'm' 'o' 'd')
```

#### 3.2.5 The gofmt Formatting Problem

After the BOM was resolved, `gofmt -l .` still reported all source files as needing formatting:

```text
Needs formatting: internal\cli\cli.go internal\cli\cli_test.go
internal\config\config.go internal\config\config_test.go
internal\errors\errors.go internal\errors\errors_test.go
internal\output\output.go internal\output\output_test.go
internal\testhelpers\testhelpers.go internal\testhelpers\testhelpers_test.go
main.go
```

This was caused by PowerShell here-strings using spaces for indentation in the written source files, whereas Go requires tabs. The `gofmt` tool enforces tab indentation as part of Go's canonical formatting standard.

Resolution was a single command:

```powershell
gofmt -w .
```

The `-w` flag rewrites all `.go` files in place with correct formatting. After this command, `gofmt -l .` returned no output, confirming all files were clean. The build and test suite continued to pass after the rewrite.

#### 3.2.6 Lesson

PowerShell 5's `Set-Content -Encoding UTF8` must never be used to write files consumed by Go tooling. The correct pattern for all future file writes from PowerShell in this project is:

```powershell
$encoding = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText($fullPath, $content, $encoding)
```

Additionally, `gofmt -w .` should be run immediately after any session that writes Go source files through external tooling or scripts. It is a safe, idempotent operation and costs nothing.

---

## 4. What Was Not a Problem

It is worth recording what did not cause difficulty, because it validates the architectural decisions made in M1.1–M1.4.

### 4.1 Package Architecture

The internal package boundaries established in M1.4 required no restructuring. `internal/errors`, `internal/config`, and `internal/testhelpers` fit naturally into the existing layout. The stub packages (`git`, `project`, `scanner`, `snapshot`) required no modification.

### 4.2 Error Design

The decision to keep the error system lightweight — a simple `PulseError` struct with a `Kind` field rather than a framework — proved appropriate. The design was implemented in one file, tested fully, and integrated into the CLI without friction.

### 4.3 Path Resolution

The `config.New()` path resolution logic worked correctly across all test cases on the first implementation. The Go standard library's `os.Getwd()`, `filepath.IsAbs()`, `filepath.Join()`, and `filepath.Clean()` compose cleanly for this purpose.

### 4.4 stdout / stderr Separation

Making `output.Writer` hold explicit `io.Writer` fields for both `Out` and `Err` made testing the stdout/stderr separation straightforward. Tests could inject `bytes.Buffer` instances and assert independently that errors appeared on stderr and that stdout remained clean. This pattern will extend naturally into S2.

### 4.5 Test Determinism

All tests pass without depending on personal machine state, local project structure, or network access. The `testhelpers.TempProject()` function creates isolated temporary directories that are automatically cleaned up. The `testhelpers.ProjectFixturePath()` function locates committed fixtures by walking up to `go.mod` rather than using hardcoded absolute paths.

---

## 5. Decisions Made During Implementation

### 5.1 Module Name: `pulse` not `github.com/raghavendrashivam474/pulse`

The module name was set to `pulse` to match the import paths already established in the codebase. This is appropriate for the current phase of the project. If Pulse is later published as a public Go module, the module name in `go.mod` and all import paths will need to be updated together. This is a known and manageable future task, not a defect.

### 5.2 Go Version: 1.21

The `go.mod` Go version directive was set to `1.21`. The original file declared `go 1.26.5`, which does not exist as a released Go version at the time of this sprint. `go 1.21` is a stable, widely available release that supports all language features used in this sprint.

### 5.3 No Environment Variables Introduced

M1.6 intentionally did not introduce environment variable configuration. No S1 requirement existed for it. Configuration infrastructure was added because Pulse needs it, not because mature CLIs conventionally have it.

### 5.4 No Config File Introduced

No `pulse.toml`, `.pulse/`, or similar file was introduced. The runtime configuration model covers only what is needed for S1: a resolved target path and an output mode flag.

### 5.5 Version Constant Kept in output Package

The version string `1.0.0` lives in `internal/output`. This is a pragmatic S1 decision. In S2 or later, the version should be injected via `ldflags` at build time and propagated cleanly through the application. The `build.ps1` script already demonstrates the ldflags pattern. This is documented as a known future improvement, not left as an undocumented shortcut.

---

## 6. Permanent Conventions Established

The following conventions are now established for the lifetime of the project.

### 6.1 File Writing from PowerShell

```powershell
# Always — never use Set-Content -Encoding UTF8 for Go files
$encoding = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText($fullPath, $content, $encoding)
```

### 6.2 Post-Write Formatting

```powershell
# Run immediately after writing any Go source files externally
gofmt -w .
```

### 6.3 Full Verification Suite

```powershell
go build ./...
go test ./...
go vet ./...
gofmt -l .
```

All four must pass before any milestone is considered closed.

### 6.4 Error Flow

```text
internal package          →    returns error
         ↓
    cli / application     →    wraps or passes error
         ↓
    output.Writer         →    PrintError() to stderr
         ↓
    process exit          →    ExitFailure (non-zero)
```

No internal package writes directly to stdout or stderr.

### 6.5 Output Separation

```text
stdout   →   normal Pulse output (summary, JSON, version, help)
stderr   →   errors and diagnostics only
```

### 6.6 Test Isolation

```text
testhelpers.TempProject()          →   isolated temp directory per test
testhelpers.ProjectFixturePath()   →   committed fixture, located via go.mod
```

No test depends on the developer's machine state.

---

## 7. Sprint 1 — Final Status

### Milestone Completion

| Milestone | Description | Status |
|---|---|---|
| M1.1 | Repository Bootstrap | ✅ Complete |
| M1.2 | Minimal CLI Entrypoint | ✅ Complete |
| M1.3 | CLI Argument Foundation | ✅ Complete |
| M1.4 | Internal Package Structure | ✅ Complete |
| M1.5 | Error Handling Foundation | ✅ Complete |
| M1.6 | Configuration & Environment | ✅ Complete |
| M1.7 | Testing Foundation | ✅ Complete |
| M1.8 | Build & Release Baseline | ✅ Complete |

### Final Suite

```text
go build ./...   ✅
go test ./...    ✅
go vet ./...     ✅
gofmt -l .       ✅
```

### Package Coverage

```text
pulse/internal/cli           tested
pulse/internal/config        tested
pulse/internal/errors        tested
pulse/internal/output        tested
pulse/internal/testhelpers   tested
pulse/internal/git           stub — S2
pulse/internal/project       stub — S2
pulse/internal/scanner       stub — S2
pulse/internal/snapshot      stub — S2
```

---

## 8. Handoff to S2

Sprint 1 is complete. The foundation is stable. The next engineer inheriting this repository can answer every structural question without reading implementation code:

| Question | Answer |
|---|---|
| How does Pulse start? | `main.go` → `cli.Main()` |
| How does Pulse receive input? | `cli.ParseArgs()` → `config.New()` |
| Where do failures go? | `output.Writer.PrintError()` → stderr → exit 1 |
| Where will project identity live? | `internal/project` |
| Where will filesystem scanning live? | `internal/scanner` |
| Where will Git logic live? | `internal/git` |
| Where will unified state live? | `internal/snapshot` |
| How is Pulse tested? | `go test ./...` from repository root |
| How is Pulse built for release? | `.\build.ps1` → `dist/` |

**S2 — Project Discovery is ready to begin.**

---

*Report prepared at conclusion of Sprint 1 implementation session.*
*Pulse v1.0.0 — Foundation complete.*