package codebase

// PackageNodeID returns the stable graph node ID for a package name.
//
// Convention: "package:<name>"
//
// This function is the single source of truth for package node ID
// formatting. All graph construction code must use this function
// rather than constructing IDs inline.
func PackageNodeID(name string) string {
	return "package:" + name
}

// BuildPackageGraph constructs a CodeGraph from the package and dependency
// information already present in a Codebase.
//
// It does not re-scan the filesystem or re-parse any source files.
// It reads only:
//
//	Codebase.Packages     -> Node per package
//	Codebase.Dependencies -> Edge per inter-package dependency
//
// Deduplication:
//   - Duplicate nodes are ignored (AddNode is idempotent on ID).
//   - Duplicate edges are ignored (AddEdge is idempotent on From+To+Kind).
//
// Determinism:
//
//	Normalise() is called before returning. Given the same Codebase
//	the result is always identical.
func BuildPackageGraph(cb Codebase) CodeGraph {
	g := NewCodeGraph()

	// One node per discovered package.
	for _, pkg := range cb.Packages {
		g.AddNode(Node{
			ID:   PackageNodeID(pkg.Name),
			Kind: NodePackage,
			Name: pkg.Name,
			Path: pkg.Path,
		})
	}

	// One depends_on edge per inter-package dependency.
	for _, dep := range cb.Dependencies {
		// Ensure both endpoints have nodes even if the package list
		// did not include them (defensive — should not happen in practice).
		g.AddNode(Node{
			ID:   PackageNodeID(dep.From),
			Kind: NodePackage,
			Name: dep.From,
		})
		g.AddNode(Node{
			ID:   PackageNodeID(dep.To),
			Kind: NodePackage,
			Name: dep.To,
		})

		g.AddEdge(Relationship{
			From: PackageNodeID(dep.From),
			To:   PackageNodeID(dep.To),
			Kind: RelationshipDependsOn,
		})
	}

	g.Normalise()
	return g
}
