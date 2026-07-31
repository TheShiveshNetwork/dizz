package graph

// DumpNode is a node in the JSON dump consumed by visualizers (TUI, web).
type DumpNode struct {
	ID     string   `json:"id"`
	Label  string   `json:"label"`
	Type   NodeType `json:"type"`
	State  string   `json:"state,omitempty"`
	File   string   `json:"file,omitempty"`
	Degree int      `json:"degree"`
}

// DumpLink is an edge in the JSON dump consumed by visualizers.
type DumpLink struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Type   EdgeType `json:"type"`
}

// Dump is the full JSON-serializable view of the graph.
type Dump struct {
	ProjectRoot string     `json:"project_root"`
	Nodes       []DumpNode `json:"nodes"`
	Links       []DumpLink `json:"links"`
}

// DumpJSON renders the graph for visualization consumers. Node and link order
// follow Graph.Nodes and Graph.Edges, so output is deterministic.
func (g *Graph) DumpJSON() Dump {
	nodes := g.Nodes()
	links := g.Edges()
	d := Dump{
		ProjectRoot: g.ProjectRoot,
		Nodes:       make([]DumpNode, 0, len(nodes)),
		Links:       make([]DumpLink, 0, len(links)),
	}
	for _, n := range nodes {
		dn := DumpNode{
			ID:    n.ID,
			Label: n.Label,
			Type:  n.Type,
			State: n.Attr("state"),
			File:  n.Attr("file"),
		}
		if dn.File == "" {
			dn.File = n.Attr("path")
		}
		dn.Degree = g.Degree(n.ID)
		d.Nodes = append(d.Nodes, dn)
	}
	for _, e := range links {
		d.Links = append(d.Links, DumpLink{Source: e.From, Target: e.To, Type: e.Type})
	}
	return d
}

// Degree returns the total number of incident edges for a node.
func (g *Graph) Degree(id string) int {
	return len(g.outgoing[id]) + len(g.incoming[id])
}
