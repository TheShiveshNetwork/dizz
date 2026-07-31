package graph

import (
	"testing"
)

func TestSimilarity(t *testing.T) {
	cases := []struct {
		a, b string
		want float64
	}{
		{"refactor auth", "refactor auth", 1},
		{"Refactor auth", "refactor token", 0.5},
		{"fix login bug", "login form", 0.5},
		{"completely unrelated xyz", "token refactor", 0},
		{"", "anything", 0},
		{"auth refactor", "refactor auth", 1},
	}
	for _, c := range cases {
		if got := Similarity(c.a, c.b); got != c.want {
			t.Errorf("Similarity(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestBuildIntentSimilarityEdges(t *testing.T) {
	projectRoot := writeFixtureProject(t)
	g, err := Build(DefaultBuildOptions(projectRoot))
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	intentA := IntentID("int_001")
	intentB := IntentID("int_002")
	todoID := TodoID("pkg/auth.go", 15)

	// Intent <-> intent: "Refactor auth" vs "Refactor token system" share
	// one of two tokens -> 0.5. Resolved intents are included too.
	for _, pair := range [][2]string{{intentA, intentB}, {intentB, intentA}} {
		if !g.HasEdge(EdgeRelatedTo, pair[0], pair[1]) {
			t.Errorf("missing RELATED_TO edge %s -> %s", pair[0], pair[1])
		}
	}
	if e := g.Edge(EdgeKey(EdgeRelatedTo, intentA, intentB)); e == nil || e.Weight != 0.5 {
		t.Errorf("intent-intent weight = %v, want 0.5", e.Weight)
	}

	// Intent <-> todo: "Refactor auth" vs "refactor token" -> 0.5.
	if !g.HasEdge(EdgeRelatedTo, intentA, todoID) {
		t.Errorf("missing RELATED_TO intent -> todo")
	}
	if !g.HasEdge(EdgeRelatedTo, todoID, intentA) {
		t.Errorf("missing RELATED_TO todo -> intent")
	}
	if e := g.Edge(EdgeKey(EdgeRelatedTo, intentA, todoID)); e == nil || e.Weight != 0.5 {
		t.Errorf("intent-todo weight = %v, want 0.5", e.Weight)
	}
	if e := g.Edge(EdgeKey(EdgeRelatedTo, intentA, todoID)); e == nil || e.Attrs["similarity"] != "0.500" {
		t.Errorf("similarity attr = %v, want 0.500", e.Attrs["similarity"])
	}

	// No RELATED_TO edges to symbols or files.
	for _, e := range g.EdgesOfType(EdgeRelatedTo) {
		n := g.Node(e.To)
		if n != nil && (n.Type == NodeSymbol || n.Type == NodeFile) {
			t.Errorf("RELATED_TO must not target symbols or files: %s -> %s", e.From, e.To)
		}
	}
}

func TestSimilarityThresholdAndTopK(t *testing.T) {
	projectRoot := writeFixtureProject(t)
	opts := DefaultBuildOptions(projectRoot)
	opts.SimilarityThreshold = 1.01
	g, err := Build(opts)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if n := len(g.EdgesOfType(EdgeRelatedTo)); n != 0 {
		t.Errorf("threshold above 1.0 should produce no RELATED_TO edges, got %d", n)
	}
}

func TestDumpJSONCarriesAttrsAndWeights(t *testing.T) {
	projectRoot := writeFixtureProject(t)
	g, err := Build(DefaultBuildOptions(projectRoot))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	d := g.DumpJSON()

	var intentA *DumpNode
	for i := range d.Nodes {
		if d.Nodes[i].Type == NodeIntent {
			intentA = &d.Nodes[i]
			break
		}
	}
	if intentA == nil {
		t.Fatal("no intent node in dump")
	}
	if intentA.Attrs == nil || intentA.Attrs["severity"] == "" {
		t.Errorf("intent dump missing attrs: %+v", intentA.Attrs)
	}

	var relLink *DumpLink
	for i := range d.Links {
		if d.Links[i].Type == EdgeRelatedTo {
			relLink = &d.Links[i]
			break
		}
	}
	if relLink == nil {
		t.Fatal("no RELATED_TO link in dump")
	}
	if relLink.Weight <= 0 {
		t.Errorf("RELATED_TO weight not carried in dump: %+v", relLink)
	}
	if relLink.Attrs == nil || relLink.Attrs["similarity"] == "" {
		t.Errorf("RELATED_TO attrs not carried in dump: %+v", relLink.Attrs)
	}
}
