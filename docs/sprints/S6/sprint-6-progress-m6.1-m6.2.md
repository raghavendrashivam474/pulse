# Sprint 6 Progress Report — M6.1 + M6.2 Complete

**To:** Senior Developer
**Sprint:** S6 — Capability Intelligence Foundation
**Date:** 2026-08-15
**Author:** Junior Developer
**Repository:** `aryntra-aayaam`
**Branch:** `main`

---

## Status

Two of five milestones complete and committed.

```text
M6.1  feat(cli): introduce capability command architecture   3ef8481  DONE
M6.2  feat(capability): add project overview                1f4014b  DONE
M6.3  feat(capability): add project capability                        NEXT
M6.4  feat(capability): add structure capability                      PENDING
M6.5  feat(capability): add relationship capability                   PENDING
```

---

## What was built

### M6.1 — Capability Command Architecture

**New package:** `internal/capability`

Introduced the core capability abstraction that the rest of S6 builds on.

```text
internal/capability/
  capability.go       — Capability interface, Context, Result, Registry
  capability_test.go  — Registration, lookup, duplicate, sorted names, nil guard
```

**`Capability` interface:**

```go
type Capability interface {
    Name()        string
    Description() string
    Run(ctx Context) (Result, error)
}
```

**`Registry`:**

- `Register` — adds a capability, returns error on duplicate or nil
- `Lookup` — retrieves by name, bool ok pattern
- `IsKnown` — fast existence check
- `Names` — returns sorted name slice for deterministic help output

**CLI changes:**

- `ParseArgs` extended to collect up to two positional arguments
- `ResolveCommand` added — maps positionals to either capability mode or legacy target mode
- `Run` routes through capability registry when a capability name is resolved
- `--help` now lists all registered capabilities with descriptions
- Unknown capability produces a clean user-facing error

**New output helper:** `internal/output/help.go` — `PrintHelpWithCapabilities`

**Routing logic:**

```text
aayam .                   → legacy target mode  → overview (after M6.2)
aayam project .           → capability mode     → project capability
aayam unknown .           → error: unknown capability
aayam --help              → help with capability list
```

**Backward compatibility:** `aayam .` and `aayam --json .` both continue working without modification to any domain package.

---

### M6.2 — Overview Capability

**New file:** `internal/output/overview.go`

Introduced `PrintOverview` and `PrintOverviewJSON` — a concise, summary-only rendering of project intelligence.

**Before M6.2 — `aayam .` produced:**

```text
Project / Languages / Structure / Packages / Dependencies /
Relationships / Git / History / Contributors / Health
```

**After M6.2 — `aayam .` produces:**

```text
Project
  Name: aryntra-aayaam
  Type: Go

Structure
  Files:       82
  Directories: 31

Languages
  CSS / Go / HTML / JavaScript / Markdown / PowerShell / Python / TypeScript

Relationships
  Nodes: 132
  Edges: 259
  Contains: 82  Belongs to: 52  Imports: 83  Depends on: 42

Git
  Branch:       main
  Working Tree: Dirty

Health
  Commits:      Yes
  Contributors: Yes
  Working Tree: Dirty
```

**What overview deliberately omits:**

```text
Root path
Packages list
Dependencies list
History detail
Contributors detail
HEAD state
```

These are not removed from the system — they are reserved for dedicated capabilities.

**JSON overview:**

```powershell
aayam --json .
```

```json
{
  "application": "Aryntra Aayam",
  "version": "1.0.0",
  "capability": "overview",
  "name": "aryntra-aayaam",
  "type": "Go",
  "file_count": 82,
  "directory_count": 31,
  "languages": [...],
  "relationships": {
    "nodes": 132,
    "edges": 259,
    "contains": 82,
    "belongs_to": 52,
    "imports": 83,
    "depends_on": 42
  },
  "git": {
    "is_repository": true,
    "branch": "main",
    "working_tree": "dirty",
    "has_commits": true,
    "has_contributors": true
  }
}
```

The JSON shape is capability-scoped — `"capability": "overview"` — which sets the pattern for M6.3–M6.5.

---

## Architecture state after M6.2

```text
                    CLI
                     │
             ParseArgs + ResolveCommand
                     │
              Capability Registry
                     │
       ┌─────────────┼─────────────┬────────────┐
       ▼             ▼             ▼            ▼
   Overview       Project      Structure  Relationships
   (own renderer) (full dump)  (full dump) (full dump)
       │             │             │            │
       └─────────────┴─────────────┴────────────┘
                     ▼
              ProjectSnapshot
              (single discovery, shared by all)
```

`project`, `structure`, and `relationships` currently route through `PrintDiscovery` — the full output. This is intentional. Each gets its own renderer in M6.3, M6.4, and M6.5 respectively.

---

## Package changes summary

| File | Change |
|---|---|
| `internal/capability/capability.go` | New — interface + registry |
| `internal/capability/capability_test.go` | New — full registry test coverage |
| `internal/cli/cli.go` | Extended — capability routing, ResolveCommand |
| `internal/cli/cli_test.go` | Extended — routing tests, capability compat tests |
| `internal/cli/capabilities.go` | New — registered capabilities wired to output |
| `internal/output/help.go` | New — PrintHelpWithCapabilities |
| `internal/output/overview.go` | New — PrintOverview, PrintOverviewJSON |
| `internal/output/output_test.go` | Extended — overview tests |

---

## Test status

```text
go build ./...     PASS
go test ./...      PASS  (all 11 packages)
go vet ./...       PASS
gofmt              PASS
```

No existing tests were broken. No domain packages were modified.

---

## Manual verification matrix

| Command | Result |
|---|---|
| `aayam --help` | Shows capabilities list |
| `aayam .` | Concise overview |
| `aayam overview .` | Identical to `aayam .` |
| `aayam --json .` | Scoped JSON with `capability: overview` |
| `aayam project .` | Full discovery output (temporary, M6.3 target) |
| `aayam structure .` | Full discovery output (temporary, M6.4 target) |
| `aayam relationships .` | Full discovery output (temporary, M6.5 target) |
| `aayam unknown .` | `Error: unknown capability: unknown` |

---

## What M6.3 will do

`aayam project .` will stop rendering the full discovery dump and instead render only:

```text
Aryntra Aayaam — Project

Name:          aryntra-aayaam
Type:          Go
Root:          C:\...\aryntra-aayaam
Files:         82
Directories:   31
Languages:     CSS / Go / HTML / ...
```

No packages. No dependencies. No git history. No graph detail.

The capability boundary for project identity will be clean and independently useful.

---

## No domain packages were touched

`scanner`, `project`, `snapshot`, `codebase`, `git` — all untouched.

All new intelligence rendering consumes the existing `ProjectSnapshot` without re-running discovery.