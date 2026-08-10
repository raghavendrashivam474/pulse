# Pulse — Sprint 2 Completion Report

**Sprint:** S2 — Project Discovery
**Version:** `0.2.0`
**Base commit:** `2888476` (tag: `0.1.0`)
**Final commit:** `fd42253` (tag: `0.2.0`)
**Status:** COMPLETE

---

## Milestones delivered

### M2.1 — Target Resolution & Validation
**Commit:** `4120338`
**File:** `internal/project/target.go`

Introduced `Target` and `ResolveTarget()`.

Given `Config.TargetPath`, Pulse now validates:
- path is not empty
- path exists on the filesystem
- path is a directory, not a file

Failures are classified using the existing `errors` package:
- missing path → `KindUser`
- inaccessible path → `KindEnvironment`
- file instead of directory → `KindUser`

**Tests:** 8 cases covering valid directory, missing path, file target,
empty path, absolute path, path cleaning, and meaningful error messages.

---

### M2.2 — Project Root Discovery
**Commit:** `dfd77f8`
**File:** `internal/project/root.go`

Introduced `RootDiscovery` and `DiscoverRoot()`.

Walks upward from the validated target looking for recognised root markers:
.git go.mod package.json Cargo.toml pyproject.toml pom.xml build.gradle

text


Behaviour when no marker is found: target itself becomes root.
`RootDiscovery.MarkerFound` is false — this is documented, not an error.

Walk stops at filesystem root to prevent runaway traversal.
Symlinks in the path are not resolved.

**Tests:** 8 cases covering target-is-root, nested one level, deeply nested,
no marker fallback, .git, package.json, Cargo.toml, and absolute path guarantee.

---

### M2.3 — Filesystem Scanner
**Commit:** `8cad990`
**File:** `internal/scanner/scanner.go`

Introduced `FileEntry`, `DirEntry`, `Inventory`, and `Scan()`.

Given a project root, produces a deterministic filesystem inventory:
- Files: relative path, lowercase extension, size in bytes
- Directories: relative path

Architectural guarantees:
- Results sorted by RelPath — deterministic across all runs
- Paths use forward slashes on all platforms
- Symlinks skipped — not followed, silently excluded
- Root directory not included in Dirs
- Unreadable entries skipped rather than aborting the scan

Default exclusions:
.git node_modules vendor dist build target .cache pycache .idea .vscode

text


**Tests:** 12 cases covering empty project, single file, nested directories,
multiple file types, excluded directories, deterministic ordering, forward-slash
paths, files with no extension, root not in dirs, size populated, lowercase
extensions, and deeply nested directory paths.

---

### M2.4 — Project Type Detection
**Commit:** `fd42253`
**File:** `internal/project/detect.go`

Introduced `ProjectType`, `Detection`, and `DetectType()`.

Analyses the filesystem inventory and returns a classified project type.

Supported types and their markers:

| Type    | Markers                          |
|---------|----------------------------------|
| Go      | go.mod                           |
| Rust    | Cargo.toml                       |
| Python  | pyproject.toml, requirements.txt |
| Java    | pom.xml, build.gradle            |
| Node.js | package.json                     |
| Unknown | (no recognised marker)           |

Detection priority when multiple markers are present:
Go > Rust > Python > Java > Node.js

text


`TypeUnknown` is valid intelligence, not an error condition:
"unable to analyse" != "unknown project type"

text


Detection is deterministic: identical inventory always produces identical output.

**Tests:** 13 cases covering all six project types, unknown, multiple markers
with priority, deterministic output, empty inventory, nested markers, and
markers slice population.

Also in this commit:
- CLI wired to run the full M2.1–M2.4 pipeline on every invocation
- Scanner test corrected: `go.mod` has extension `.mod`, not empty string
- CLI tests corrected: replaced non-existent `./some-project` with real
  temp directories via `testhelpers.TempProject`

---

## Pipeline established
os.Args
|
ParseArgs
|
config.New <- resolves and cleans path
|
project.ResolveTarget <- M2.1: validates real directory
|
project.DiscoverRoot <- M2.2: finds true project root
|
scanner.Scan <- M2.3: inventories the filesystem
|
project.DetectType <- M2.4: identifies the ecosystem
|
output

text


---

## Test summary

| Package                  | Tests |
|--------------------------|-------|
| pulse/internal/cli       | 18    |
| pulse/internal/config    | 9     |
| pulse/internal/errors    | 5     |
| pulse/internal/output    | 6     |
| pulse/internal/project   | 29    |
| pulse/internal/scanner   | 12    |
| pulse/internal/testhelpers | 6   |
| **Total**                | **85** |

All 85 tests pass. No failures.

---

## Architecture boundaries respected

S2 did not touch:
git/ <- still waiting for S3
snapshot/ <- still waiting for M2.7
build.ps1
.gitignore

text


No third-party dependencies were introduced.
Standard library only throughout.

Package responsibilities remain clean:
project/ <- identity and reasoning
scanner/ <- filesystem inspection only
config/ <- runtime input
errors/ <- failure classification
output/ <- presentation
cli/ <- application flow

text


---

## Observable behaviour at 0.2.0
$ pulse .
Pulse - Project Discovery

Project
Root: C:...\pulse
Type: Go

Filesystem
Files: 34
Directories: 16

$ pulse .\README.md
Error: target must be a directory, not a file: ...\README.md

$ pulse .\does-not-exist
Error: target path does not exist: ...\does-not-exist

$ pulse --json .
{
"root": "C:\...\pulse",
"type": "Go",
"files": 34,
"directories": 16
}

text


---

## What S2 does not do

Intentionally excluded from scope:

- Git history analysis
- Dependency graph analysis
- Code quality or health scoring
- Framework deep detection
- Polished output formatting (M2.8)
- Snapshot comparison (M2.7)
- AI or LLM analysis
- External dependencies

---

## Ready for

- M2.5 onwards as defined in the S2 brief
- S3 Git integration building on the established root discovery