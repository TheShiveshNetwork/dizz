package graph

import "sort"

// Graph is an in-memory directed graph with adjacency lists and indices for
// the query patterns that matter (lookup by file, lookup by symbol name).
type Graph struct {
	nodes    map[string]*Node
	edges    map[string]*Edge
	outgoing map[string][]*Edge
	incoming map[string][]*Edge

	fileNodes     map[string]string // relpath -> file node id
	symbolsByName map[string][]string
	byType        map[NodeType][]*Node
	symbolNodes   map[string]*Node
	symbolsByFile map[string][]*Node
	ProjectRoot   string
}

// NewGraph creates an empty graph.
func NewGraph() *Graph {
	return &Graph{
		nodes:         make(map[string]*Node),
		edges:         make(map[string]*Edge),
		outgoing:      make(map[string][]*Edge),
		incoming:      make(map[string][]*Edge),
		fileNodes:     make(map[string]string),
		symbolsByName: make(map[string][]string),
		byType:        make(map[NodeType][]*Node),
		symbolNodes:   make(map[string]*Node),
		symbolsByFile: make(map[string][]*Node),
	}
}

// AddNode inserts a node, returning the existing node when the ID is a dup.
func (g *Graph) AddNode(n *Node) *Node {
	if existing, ok := g.nodes[n.ID]; ok {
		return existing
	}
	g.nodes[n.ID] = n
	g.byType[n.Type] = append(g.byType[n.Type], n)
	switch n.Type {
	case NodeSymbol:
		g.symbolNodes[n.ID] = n
		g.symbolsByName[n.Label] = append(g.symbolsByName[n.Label], n.ID)
		if file := n.Attr("file"); file != "" {
			g.symbolsByFile[file] = append(g.symbolsByFile[file], n)
		}
	case NodeFile:
		if path := n.Attr("path"); path != "" {
			g.fileNodes[path] = n.ID
		}
	}
	return n
}

// Node returns the node with the given ID, or nil.
func (g *Graph) Node(id string) *Node {
	return g.nodes[id]
}

// Nodes returns all nodes, ordered by ID for determinism.
func (g *Graph) Nodes() []*Node {
	out := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// AddEdge inserts an edge. Duplicates (same type/from/to) are merged: the
// weight becomes the max and an occurrence counter is incremented. This keeps
// the graph canonical even when the same relationship is seen from several
// sources (e.g. multiple call sites).
func (g *Graph) AddEdge(e *Edge) *Edge {
	key := e.key()
	if existing, ok := g.edges[key]; ok {
		count, _ := parseAttrInt(existing.Attrs, "occurrences")
		if count == 0 {
			count = 1
		}
		if e.Weight > existing.Weight {
			existing.Weight = e.Weight
		}
		existing.Rationale.Confidence = max(existing.Rationale.Confidence, e.Rationale.Confidence)
		if existing.Rationale.Evidence == "" {
			existing.Rationale.Evidence = e.Rationale.Evidence
		}
		if existing.Attrs == nil {
			existing.Attrs = make(map[string]string)
		}
		existing.Attrs["occurrences"] = itoa(count + 1)
		return existing
	}
	e.ID = key
	g.edges[key] = e
	g.outgoing[e.From] = append(g.outgoing[e.From], e)
	g.incoming[e.To] = append(g.incoming[e.To], e)
	return e
}

// @dizz-ignore-abandoned
// Edge returns the edge with the given ID, or nil.
func (g *Graph) Edge(id string) *Edge {
	return g.edges[id]
}

// Edges returns all edges, ordered by key for determinism.
func (g *Graph) Edges() []*Edge {
	out := make([]*Edge, 0, len(g.edges))
	for _, e := range g.edges {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// @dizz-ignore-abandoned
// EdgesOfType returns all edges of a given type, ordered by key.
func (g *Graph) EdgesOfType(t EdgeType) []*Edge {
	out := make([]*Edge, 0, len(g.edges))
	for _, e := range g.edges {
		if e.Type == t {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Outgoing returns the edges leaving a node.
func (g *Graph) Outgoing(id string) []*Edge {
	return g.outgoing[id]
}

// Incoming returns the edges entering a node.
func (g *Graph) Incoming(id string) []*Edge {
	return g.incoming[id]
}

// FileNode returns the file node for a relative path, or nil.
func (g *Graph) FileNode(relPath string) *Node {
	if id, ok := g.fileNodes[relPath]; ok {
		return g.nodes[id]
	}
	return nil
}

// SymbolsNamed returns the symbol nodes with the given label (symbol name).
func (g *Graph) SymbolsNamed(name string) []*Node {
	ids := g.symbolsByName[name]
	out := make([]*Node, 0, len(ids))
	for _, id := range ids {
		if n, ok := g.symbolNodes[id]; ok {
			out = append(out, n)
		}
	}
	return out
}

// SymbolsInFile returns the symbol nodes defined in the given relative path.
func (g *Graph) SymbolsInFile(relPath string) []*Node {
	return g.symbolsByFile[relPath]
}

// HasEdge reports whether an edge with the given type/from/to exists.
func (g *Graph) HasEdge(t EdgeType, from, to string) bool {
	_, ok := g.edges[EdgeKey(t, from, to)]
	return ok
}

// Stats summarizes graph shape for stats queries.
type Stats struct {
	Nodes       int
	Edges       int
	ByNodeType  map[NodeType]int
	ByEdgeType  map[EdgeType]int
	UnusedDeps  int
	HasCoChange bool
}

// ComputeStats produces graph shape statistics.
func (g *Graph) ComputeStats() Stats {
	s := Stats{
		ByNodeType: make(map[NodeType]int),
		ByEdgeType: make(map[EdgeType]int),
	}
	for _, n := range g.nodes {
		s.Nodes++
		s.ByNodeType[n.Type]++
	}
	for _, e := range g.edges {
		s.Edges++
		s.ByEdgeType[e.Type]++
		if e.Type == EdgeCoChangesWith {
			s.HasCoChange = true
		}
	}
	// A dependency is "unused" when no file that declares DEPENDS_ON edges to
	// it is reachable from any CALLS edge. For simplicity, a dep is unused when
	// it has no incoming CALLS through its importing file — approximated here
	// by counting dep nodes with DEPENDS_ON incoming edges only.
	for _, n := range g.nodes {
		if n.Type == NodeDep && len(g.incoming[n.ID]) > 0 {
			s.UnusedDeps++
		}
	}
	return s
}

func parseAttrInt(attrs map[string]string, key string) (int, bool) {
	if attrs == nil {
		return 0, false
	}
	v, ok := attrs[key]
	if !ok {
		return 0, false
	}
	n := 0
	neg := false
	i := 0
	if i < len(v) && (v[i] == '-' || v[i] == '+') {
		neg = v[i] == '-'
		i++
	}
	for ; i < len(v); i++ {
		if v[i] < '0' || v[i] > '9' {
			return 0, false
		}
		n = n*10 + int(v[i]-'0')
	}
	if neg {
		n = -n
	}
	return n, true
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
