package output

import (
    "encoding/json"
    "fmt"
    "sort"

    "github.com/raghavendrashivam474/aayam/internal/snapshot"
)

// PrintDependencies renders the package dependency summary to stdout.
//
// This is the dedicated renderer for: aayam dependencies .
// It answers: "What does this project depend on?"
// It does NOT include project identity, structure counts, or git.
func (w *Writer) PrintDependencies(snap snapshot.ProjectSnapshot) {
    fmt.Fprintf(w.Out, "Aryntra Aayam - Dependencies\n\n")

    deps := snap.Codebase.Dependencies
    if len(deps) == 0 {
        fmt.Fprintf(w.Out, "  (no internal dependencies detected)\n")
        return
    }

    // Group: From -> []To. Dependencies are already sorted by (From, To).
    type group struct {
        from string
        tos  []string
    }

    var groups []group
    index := make(map[string]int)

    for _, d := range deps {
        if i, ok := index[d.From]; ok {
            groups[i].tos = append(groups[i].tos, d.To)
        } else {
            index[d.From] = len(groups)
            groups = append(groups, group{from: d.From, tos: []string{d.To}})
        }
    }

    sort.Slice(groups, func(i, j int) bool {
        return groups[i].from < groups[j].from
    })

    fmt.Fprintf(w.Out, "Packages\n\n")
    for _, g := range groups {
        fmt.Fprintf(w.Out, "  %s\n", g.from)
        for _, to := range g.tos {
            fmt.Fprintf(w.Out, "    -> %s\n", to)
        }
        fmt.Fprintf(w.Out, "\n")
    }
}

// JSONDependenciesResult is the top-level structure for machine-readable
// dependencies output.
type JSONDependenciesResult struct {
    Application  string                  `json:"application"`
    Version      string                  `json:"version"`
    Capability   string                  `json:"capability"`
    Dependencies JSONDependenciesPayload `json:"dependencies"`
}

// JSONDependenciesPayload holds the dependencies-capability-specific fields.
type JSONDependenciesPayload struct {
    Packages []JSONPackageDependency `json:"packages"`
}

// JSONPackageDependency represents one package and what it depends on.
type JSONPackageDependency struct {
    Name      string   `json:"name"`
    DependsOn []string `json:"depends_on"`
}

// PrintDependenciesJSON renders the dependency summary as JSON to stdout.
func (w *Writer) PrintDependenciesJSON(snap snapshot.ProjectSnapshot) error {
    deps := snap.Codebase.Dependencies

    var packages []JSONPackageDependency
    index := make(map[string]int)

    for _, d := range deps {
        if i, ok := index[d.From]; ok {
            packages[i].DependsOn = append(packages[i].DependsOn, d.To)
        } else {
            index[d.From] = len(packages)
            packages = append(packages, JSONPackageDependency{
                Name:      d.From,
                DependsOn: []string{d.To},
            })
        }
    }

    sort.Slice(packages, func(i, j int) bool {
        return packages[i].Name < packages[j].Name
    })

    if packages == nil {
        packages = []JSONPackageDependency{}
    }

    result := JSONDependenciesResult{
        Application: "Aryntra Aayam",
        Version:     version,
        Capability:  "dependencies",
        Dependencies: JSONDependenciesPayload{
            Packages: packages,
        },
    }

    enc := json.NewEncoder(w.Out)
    enc.SetIndent("", "  ")
    return enc.Encode(result)
}
