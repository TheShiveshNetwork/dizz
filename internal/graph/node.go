// Package graph implements a derived knowledge graph over the state dizz
// already produces. The graph is never materialized: every query builds it
// in-memory from the persisted state files (state.ton.gz, intent.ton), the
// per-file signal cache and, on demand, git history. Because it is a pure
// function of on-disk state, it is always live and never goes stale.
package graph

import (
	"strconv"
)

// NodeType enumerates the kinds of entities that can appear in the graph.
type NodeType string

const (
	NodeSymbol    NodeType = "symbol"
	NodeFile      NodeType = "file"
	NodeModule    NodeType = "module"
	NodeIntent    NodeType = "intent"
	NodeCommit    NodeType = "commit"
	NodeTest      NodeType = "test"
	NodeDep       NodeType = "dep"
	NodeSnapshot  NodeType = "snapshot"
	NodeGuardrail NodeType = "guardrail"
	NodeTodo      NodeType = "todo"
)

// Rationale carries the evidence behind a node or edge, not just the verdict.
// Every claim in the graph can be traced to a source with a confidence so
// consumers can weight it accordingly.
type Rationale struct {
	SourceTier string  `json:"source_tier,omitempty"` // ast, lexical, regex
	Confidence float64 `json:"confidence"`            // 0.0-1.0
	Evidence   string  `json:"evidence,omitempty"`    // "12 commits in 30 days"
	LineRange  string  `json:"line_range,omitempty"`  // "40-80"
	Timestamp  int64   `json:"timestamp,omitempty"`
	SourceType string  `json:"source_type,omitempty"` // git, static_analysis, ...
}

// Node is a single entity in the knowledge graph.
type Node struct {
	ID        string                 `json:"id"`
	Type      NodeType               `json:"type"`
	Label     string                 `json:"label"`
	Attrs     map[string]interface{} `json:"attrs,omitempty"`
	Rationale Rationale              `json:"rationale,omitempty"`
}

// Attr returns the value of a node attribute, or "" when absent.
func (n *Node) Attr(key string) string {
	if n == nil || n.Attrs == nil {
		return ""
	}
	if v, ok := n.Attrs[key].(string); ok {
		return v
	}
	return ""
}

// Node ID constructors. IDs are deterministic so identical derivations
// produce identical graphs.

// SymbolID builds the ID of a symbol node from its relative file path.
func SymbolID(relFile, name string) string {
	return "symbol:" + relFile + "::" + name
}

// FileID builds the ID of a file node from its relative path.
func FileID(relFile string) string {
	return "file:" + relFile
}

// IntentID builds the ID of an intent node.
func IntentID(id string) string {
	return "intent:" + id
}

// TestID builds the ID of a test node from its relative path.
func TestID(relFile string) string {
	return "test:" + relFile
}

// DepID builds the ID of a dependency node from its import path.
func DepID(name string) string {
	return "dep:" + name
}

// SnapshotID builds the ID of a snapshot node from its hash.
func SnapshotID(hash string) string {
	return "snapshot:" + hash
}

// GuardrailID builds the ID of a guardrail node.
func GuardrailID(id string) string {
	return "guardrail:" + id
}

// TodoID builds the ID of a todo node from its relative path and line.
func TodoID(relFile string, line int) string {
	return "todo:" + relFile + ":" + strconv.Itoa(line)
}

// ModuleID builds the ID of a module node from its module path.
func ModuleID(path string) string {
	return "module:" + path
}
