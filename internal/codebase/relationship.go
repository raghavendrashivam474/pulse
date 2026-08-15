package codebase

import "sort"

// NodeKind identifies the type of entity a Node represents in the code graph.
//
// The vocabulary is deliberately small. Only kinds that Aryntra Aayam can actually
// discover and prove are defined here.
type NodeKind string

const (
	// NodeProject represents the root project itself.
	NodeProject NodeKind = "project"

	// NodeDirectory represents a filesystem directory.
	NodeDirectory NodeKind = "directory"

	// NodeFile represents a single source file.
	NodeFile NodeKind = "file"

	// NodePackage represents a discovered package (Go package, etc).
	NodePackage NodeKind = "package"
)

// RelationshipKind identifies the type of relationship between two nodes.
//
// Every RelationshipKind must be backed by actual evidence from the
// discovery layer. Speculative or inferred relationships must not be
// introduced here.
type RelationshipKind string

const (
	// RelationshipContains means the From node is a structural parent
	// of the To node. Example: a directory contains a file.
	RelationshipContains RelationshipKind = "contains"

	// RelationshipBelongsTo is the inverse of contains.
	// Example: a file belongs_to a directory.
	RelationshipBelongsTo RelationshipKind = "belongs_to"

	// RelationshipImports means the From file or package imports
	// the To package. Backed by parsed import declarations.
	RelationshipImports RelationshipKind = "imports"

	// RelationshipDependsOn means the From package depends on
	// the To package. Backed by inter-package import analysis.
	RelationshipDependsOn RelationshipKind = "depends_on"
)

// Node is a single vertex in the code graph.
//
// Every entity that participates in a relationship must have a Node.
// Node IDs must be stable across runs given the same codebase.
// Node IDs must not contain machine-specific absolute paths.
//
// ID convention:
//
//	"project:<name>"
//	"dir:<relative/path>"
//	"file:<relative/path>"
//	"package:<name>"
type Node struct {
	// ID is the unique, stable identifier for this node within the graph.
	// Used as the source and target of edges.
	ID string `json:"id"`

	// Kind classifies the type of entity this node represents.
	Kind NodeKind `json:"kind"`

	// Name is the short human-readable name for this node.
	// For files: the filename. For packages: the package name.
	// For directories: the directory name. For projects: the project name.
	Name string `json:"name"`

	// Path is the forward-slash-separated path relative to the project root.
	// Empty for nodes that have no meaningful path (e.g. a project node).
	Path string `json:"path,omitempty"`
}

// Relationship is a directed edge in the code graph.
//
// It connects two nodes with a typed relationship.
// Both From and To are Node IDs.
//
// Example:
//
//	Relationship{
//	    From: "package:cli",
//	    To:   "package:config",
//	    Kind: RelationshipDependsOn,
//	}
type Relationship struct {
	// From is the ID of the source node.
	From string `json:"from"`

	// To is the ID of the target node.
	To string `json:"to"`

	// Kind classifies the type of relationship.
	Kind RelationshipKind `json:"kind"`
}

// CodeGraph is the complete relationship model of a codebase.
//
// It contains all discovered nodes and the directed edges between them.
// This is the primary S5 deliverable.
//
// Design constraints:
//   - All collections are deterministically ordered.
//   - No visualization state. This is a pure data model.
//   - Consumers must not mutate the graph after construction.
//   - Future capability engines, CLI queries, and visualization layers
//     all consume this model without modifying it.
//
// Architecture:
//
//	Codebase Model (S4)
//	       |
//	       v
//	  CodeGraph  <- this type
//	       |
//	  +---------+-----------+
//	  |         |           |
//	 CLI    Capability   Graph UI
//	        Engine       (future)
type CodeGraph struct {
	// Nodes contains all discovered entities in the codebase.
	// Sorted by ID for deterministic ordering.
	Nodes []Node `json:"nodes"`

	// Edges contains all discovered relationships between nodes.
	// Sorted by (From, To, Kind) for deterministic ordering.
	Edges []Relationship `json:"edges"`
}

// NewCodeGraph returns an empty CodeGraph ready for population.
func NewCodeGraph() CodeGraph {
	return CodeGraph{
		Nodes: []Node{},
		Edges: []Relationship{},
	}
}

// AddNode appends a node to the graph if a node with the same ID does
// not already exist. Duplicate IDs are silently ignored.
//
// Call Normalise after all nodes and edges have been added to
// restore deterministic ordering.
func (g *CodeGraph) AddNode(n Node) {
	for _, existing := range g.Nodes {
		if existing.ID == n.ID {
			return
		}
	}
	g.Nodes = append(g.Nodes, n)
}

// AddEdge appends a relationship to the graph if an identical edge
// (same From, To, and Kind) does not already exist.
// Duplicate edges are silently ignored.
//
// Call Normalise after all nodes and edges have been added to
// restore deterministic ordering.
func (g *CodeGraph) AddEdge(r Relationship) {
	for _, existing := range g.Edges {
		if existing.From == r.From && existing.To == r.To && existing.Kind == r.Kind {
			return
		}
	}
	g.Edges = append(g.Edges, r)
}

// Normalise sorts nodes by ID and edges by (From, To, Kind).
//
// This must be called after all nodes and edges have been added to
// guarantee deterministic output regardless of insertion order.
func (g *CodeGraph) Normalise() {
	sort.Slice(g.Nodes, func(i, j int) bool {
		return g.Nodes[i].ID < g.Nodes[j].ID
	})

	sort.Slice(g.Edges, func(i, j int) bool {
		ei, ej := g.Edges[i], g.Edges[j]
		if ei.From != ej.From {
			return ei.From < ej.From
		}
		if ei.To != ej.To {
			return ei.To < ej.To
		}
		return ei.Kind < ej.Kind
	})
}

// NodeCount returns the number of nodes in the graph.
func (g *CodeGraph) NodeCount() int {
	return len(g.Nodes)
}

// EdgeCount returns the number of edges in the graph.
func (g *CodeGraph) EdgeCount() int {
	return len(g.Edges)
}

// EdgesByKind returns a map from RelationshipKind to the count of
// edges of that kind. Useful for summary output.
func (g *CodeGraph) EdgesByKind() map[RelationshipKind]int {
	counts := make(map[RelationshipKind]int)
	for _, e := range g.Edges {
		counts[e.Kind]++
	}
	return counts
}
