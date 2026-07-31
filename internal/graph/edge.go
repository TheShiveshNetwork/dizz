package graph

import "fmt"

// EdgeType enumerates the relationships stored in the graph.
type EdgeType string

const (
	EdgeDefinedIn     EdgeType = "DEFINED_IN"
	EdgeContains      EdgeType = "CONTAINS"
	EdgeCalls         EdgeType = "CALLS"
	EdgeImports       EdgeType = "IMPORTS"
	EdgeDependsOn     EdgeType = "DEPENDS_ON"
	EdgeTests         EdgeType = "TESTS"
	EdgeCoChangesWith EdgeType = "CO_CHANGES_WITH"
	EdgeResolves      EdgeType = "RESOLVES"
	EdgeScopeMatch    EdgeType = "SCOPE_MATCH"
	EdgeProtects      EdgeType = "PROTECTS"
	EdgeModifies      EdgeType = "MODIFIED_BY"
	EdgeChangedIn     EdgeType = "CHANGED_IN"
	EdgeHasTodo       EdgeType = "HAS_TODO"
)

// Edge is a directed relationship between two nodes. It always carries a
// rationale so consumers can trace how the relationship was derived.
type Edge struct {
	ID        string
	Type      EdgeType
	From      string
	To        string
	Weight    float64
	Attrs     map[string]string
	Rationale Rationale
}

// Attr returns the value of an edge attribute, or "" when absent.
func (e *Edge) Attr(key string) string {
	if e == nil || e.Attrs == nil {
		return ""
	}
	return e.Attrs[key]
}

// EdgeKey returns the deduplication key for an edge: two edges with the same
// key describe the same relationship and are merged on insert.
func EdgeKey(t EdgeType, from, to string) string {
	return string(t) + "|" + from + "|" + to
}

func (e *Edge) key() string {
	return EdgeKey(e.Type, e.From, e.To)
}

// Describe renders a compact human/agent readable description of the edge.
func (e *Edge) Describe() string {
	return fmt.Sprintf("%s %s -> %s %s", e.Type, e.From, e.To, formatRationale(e.Rationale))
}
