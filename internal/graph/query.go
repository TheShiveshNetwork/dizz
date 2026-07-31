package graph

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// QueryEngine executes traversals over a derived graph.
type QueryEngine struct {
	g *Graph
}

// NewQueryEngine wraps a graph for querying.
func NewQueryEngine(g *Graph) *QueryEngine {
	return &QueryEngine{g: g}
}

// ResolveEntity resolves a user-supplied entity specifier to a node ID.
// Accepted forms:
//
//	symbol:<name>            unique symbol by name
//	symbol:<name>@<file>     symbol by name in a file
//	symbol:<file>::<name>    symbol by file and name
//	file:<path>              file by relative path
//	<name>                   unique symbol by name, else file by path
func (qe *QueryEngine) ResolveEntity(term string) (*Node, error) {
	term = strings.TrimSpace(term)
	if term == "" {
		return nil, fmt.Errorf("empty entity")
	}

	switch {
	case strings.HasPrefix(term, "symbol:"):
		body := strings.TrimPrefix(term, "symbol:")
		if at := strings.LastIndex(body, "@"); at >= 0 {
			name, file := body[:at], body[at+1:]
			return qe.resolveSymbol(name, file)
		}
		if idx := strings.Index(body, "::"); idx >= 0 {
			return qe.resolveSymbol(body[idx+2:], body[:idx])
		}
		return qe.resolveSymbol(body, "")

	case strings.HasPrefix(term, "file:"):
		rel := strings.TrimPrefix(term, "file:")
		if n := qe.g.FileNode(rel); n != nil {
			return n, nil
		}
		return nil, fmt.Errorf("file not found: %s", rel)

	case strings.HasPrefix(term, "intent:"), strings.HasPrefix(term, "test:"),
		strings.HasPrefix(term, "dep:"), strings.HasPrefix(term, "snapshot:"),
		strings.HasPrefix(term, "commit:"), strings.HasPrefix(term, "guardrail:"),
		strings.HasPrefix(term, "module:"):
		if n := qe.g.Node(term); n != nil {
			return n, nil
		}
		return nil, fmt.Errorf("node not found: %s", term)
	}

	// Bare term: try unique symbol by name, then file by path.
	if syms := qe.g.SymbolsNamed(term); len(syms) == 1 {
		return syms[0], nil
	} else if len(syms) > 1 {
		return nil, fmt.Errorf("symbol %q is ambiguous (%d definitions); use symbol:<name>@<file>", term, len(syms))
	}
	if n := qe.g.FileNode(term); n != nil {
		return n, nil
	}
	return nil, fmt.Errorf("no entity matches %q (try symbol:<name>, file:<path>, or intent:<id>)", term)
}

func (qe *QueryEngine) resolveSymbol(name, file string) (*Node, error) {
	if file == "" {
		syms := qe.g.SymbolsNamed(name)
		if len(syms) == 0 {
			return nil, fmt.Errorf("symbol not found: %s", name)
		}
		if len(syms) > 1 {
			return nil, fmt.Errorf("symbol %q is ambiguous (%d definitions); use symbol:<name>@<file>", name, len(syms))
		}
		return syms[0], nil
	}
	file = strings.TrimPrefix(file, "./")
	id := SymbolID(file, name)
	if n := qe.g.Node(id); n != nil {
		return n, nil
	}
	return nil, fmt.Errorf("symbol not found: %s in %s", name, file)
}

// Affected describes an entity within a blast radius, with accumulated score
// and the edge types that connect it to the origin.
type Affected struct {
	NodeID string
	Node   *Node
	Score  float64
	Depth  int
	Via    []EdgeType
}

// blastStep is a candidate edge in a blast-radius traversal with the resolved
// target node ID (edges may be followed in reverse for dependency types).
type blastStep struct {
	e      *Edge
	target string
}

// blastNeighbors expands a node in the "what breaks if this changes" sense:
// everything it points at, plus everything that depends on it (callers,
// covering tests, importers, matching intents, co-change partners).
func (qe *QueryEngine) blastNeighbors(id string) []blastStep {
	var out []blastStep
	for _, e := range qe.g.Outgoing(id) {
		out = append(out, blastStep{e: e, target: e.To})
	}
	for _, e := range qe.g.Incoming(id) {
		switch e.Type {
		case EdgeCalls, EdgeTests, EdgeImports, EdgeScopeMatch, EdgeCoChangesWith, EdgeDependsOn:
			out = append(out, blastStep{e: e, target: e.From})
		}
	}
	return out
}

// BlastRadius computes the transitive set of nodes affected by changing the
// entity, within the given depth, scoring each by the product of edge weights
// along the path. Both outgoing dependencies and incoming dependents (callers,
// covering tests, importers) are followed. Returns results sorted by
// descending score.
func (qe *QueryEngine) BlastRadius(id string, depth int) []Affected {
	start := qe.g.Node(id)
	if start == nil {
		return nil
	}
	results := make(map[string]*Affected)
	type visit struct {
		id    string
		score float64
		depth int
		via   []EdgeType
	}
	queue := []visit{{id: id, score: 1.0, depth: 0}}
	seen := map[string]bool{id: true}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.depth >= depth {
			continue
		}
		for _, step := range qe.blastNeighbors(cur.id) {
			e := step.e
			if seen[step.target] {
				// Still accumulate score along alternate routes.
				if a, ok := results[step.target]; ok {
					a.Score += cur.score * e.Weight
					a.Via = appendUniqueEdgeType(a.Via, e.Type)
				}
				continue
			}
			seen[step.target] = true
			score := cur.score * e.Weight
			via := append([]EdgeType{}, cur.via...)
			via = append(via, e.Type)
			a := &Affected{
				NodeID: step.target,
				Node:   qe.g.Node(step.target),
				Score:  score,
				Depth:  cur.depth + 1,
				Via:    via,
			}
			results[step.target] = a
			queue = append(queue, visit{id: step.target, score: score, depth: cur.depth + 1, via: via})
		}
	}

	out := make([]Affected, 0, len(results))
	for _, a := range results {
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

func appendUniqueEdgeType(list []EdgeType, t EdgeType) []EdgeType {
	for _, existing := range list {
		if existing == t {
			return list
		}
	}
	return append(list, t)
}

// EdgeRef is a referenced edge in a trace result.
type EdgeRef struct {
	Type      EdgeType
	From      string
	To        string
	Weight    float64
	Attrs     map[string]string
	Rationale Rationale
}

// TraceResult is the full bidirectional neighborhood of a node.
type TraceResult struct {
	Node *Node

	Callers    []EdgeRef // incoming CALLS
	Callees    []EdgeRef // outgoing CALLS
	Imports    []EdgeRef // outgoing IMPORTS/DEPENDS_ON
	ImportedBy []EdgeRef // incoming IMPORTS
	Tests      []EdgeRef // incoming TESTS (tests covering this symbol/file)
	Tested     []EdgeRef // outgoing TESTS
	Intents    []EdgeRef // incoming SCOPE_MATCH
	Todo       []EdgeRef // outgoing HAS_TODO
	CoChanges  []EdgeRef // outgoing+incoming CO_CHANGES_WITH
	Contains   []EdgeRef // outgoing CONTAINS
	DefinedIn  *EdgeRef  // outgoing DEFINED_IN
	Protects   []EdgeRef // incoming PROTECTS
	Protected  []EdgeRef // outgoing PROTECTS
}

// Trace gathers every relationship of a node.
func (qe *QueryEngine) Trace(id string) *TraceResult {
	node := qe.g.Node(id)
	if node == nil {
		return nil
	}
	tr := &TraceResult{Node: node}
	for _, e := range qe.g.Incoming(id) {
		switch e.Type {
		case EdgeCalls:
			tr.Callers = append(tr.Callers, toRef(e))
		case EdgeTests:
			tr.Tests = append(tr.Tests, toRef(e))
		case EdgeImports:
			tr.ImportedBy = append(tr.ImportedBy, toRef(e))
		case EdgeScopeMatch:
			tr.Intents = append(tr.Intents, toRef(e))
		case EdgeCoChangesWith:
			tr.CoChanges = append(tr.CoChanges, toRef(e))
		case EdgeProtects:
			tr.Protects = append(tr.Protects, toRef(e))
		}
	}
	for _, e := range qe.g.Outgoing(id) {
		switch e.Type {
		case EdgeCalls:
			tr.Callees = append(tr.Callees, toRef(e))
		case EdgeImports:
			tr.Imports = append(tr.Imports, toRef(e))
		case EdgeDependsOn:
			tr.Imports = append(tr.Imports, toRef(e))
		case EdgeTests:
			tr.Tested = append(tr.Tested, toRef(e))
		case EdgeScopeMatch:
			tr.Intents = append(tr.Intents, toRef(e))
		case EdgeCoChangesWith:
			tr.CoChanges = append(tr.CoChanges, toRef(e))
		case EdgeHasTodo:
			tr.Todo = append(tr.Todo, toRef(e))
		case EdgeContains:
			tr.Contains = append(tr.Contains, toRef(e))
		case EdgeDefinedIn:
			ref := toRef(e)
			tr.DefinedIn = &ref
		case EdgeProtects:
			tr.Protected = append(tr.Protected, toRef(e))
		}
	}
	return tr
}

func toRef(e *Edge) EdgeRef {
	return EdgeRef{Type: e.Type, From: e.From, To: e.To, Weight: e.Weight, Attrs: e.Attrs, Rationale: e.Rationale}
}

// TestCoverage returns the TESTS edges covering a node.
func (qe *QueryEngine) TestCoverage(id string) []EdgeRef {
	var out []EdgeRef
	for _, e := range qe.g.Incoming(id) {
		if e.Type == EdgeTests {
			out = append(out, toRef(e))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].From < out[j].From })
	return out
}

// CoChanges returns the co-change edges for a node, either direction.
func (qe *QueryEngine) CoChanges(id string, minJaccard float64) []EdgeRef {
	var out []EdgeRef
	for _, e := range qe.g.Incoming(id) {
		if e.Type == EdgeCoChangesWith && e.Weight >= minJaccard {
			out = append(out, toRef(e))
		}
	}
	for _, e := range qe.g.Outgoing(id) {
		if e.Type == EdgeCoChangesWith && e.Weight >= minJaccard {
			out = append(out, toRef(e))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Weight > out[j].Weight })
	return out
}

// ShortestPath finds the shortest path between two nodes using BFS over all
// outgoing edges. Returns the node IDs along the path, or nil if disconnected.
func (qe *QueryEngine) ShortestPath(from, to string) []string {
	if from == to {
		return []string{from}
	}
	if qe.g.Node(from) == nil || qe.g.Node(to) == nil {
		return nil
	}
	prev := map[string]string{from: ""}
	queue := []string{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, e := range qe.g.Outgoing(cur) {
			if _, ok := prev[e.To]; ok {
				continue
			}
			prev[e.To] = cur
			if e.To == to {
				return backchain(prev, from, to)
			}
			queue = append(queue, e.To)
		}
	}
	return nil
}

func backchain(prev map[string]string, from, to string) []string {
	var rev []string
	for cur := to; cur != ""; cur = prev[cur] {
		rev = append(rev, cur)
	}
	out := make([]string, len(rev))
	for i := range rev {
		out[len(rev)-1-i] = rev[i]
	}
	return out
}

// ImpactScope returns the file nodes matching a glob pattern over relative
// paths (e.g. "internal/auth/**").
func (qe *QueryEngine) ImpactScope(pattern string) []*Node {
	pattern = strings.TrimSpace(pattern)
	pattern = strings.TrimPrefix(pattern, "file:")
	var out []*Node
	for _, rel := range qe.g.fileRelPaths() {
		ok := false
		if pattern == rel {
			ok = true
		} else if ok2, _ := path.Match(pattern, rel); ok2 {
			ok = true
		} else if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				ok = true
			}
		}
		if ok {
			if n := qe.g.FileNode(rel); n != nil {
				out = append(out, n)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
