package graph

import (
	"fmt"
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-3
}

func TestSimilarity(t *testing.T) {
	cases := []struct {
		a, b string
		want float64
	}{
		{"refactor auth", "refactor auth", 1},
		{"Refactor auth", "refactor token", 0.5},
		{"fix login bug", "login form", 1 / math.Sqrt(6)},
		{"completely unrelated xyz", "token refactor", 0},
		{"", "anything", 0},
		{"auth refactor", "refactor auth", 1},
		{"caching cache", "cache", 1},
		{"tokens weighting", "token weight", 1},
		{"the tokens", "token", 1},
		{"cache invalidation", "caching invalidations", 1},
	}
	for _, c := range cases {
		if got := Similarity(c.a, c.b); !almostEqual(got, c.want) {
			t.Errorf("Similarity(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestTokenizeStemmingAndStopwords(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"Refactor auth", []string{"refactor", "auth"}},
		{"the tokens weighting", []string{"token", "weight"}},
		{"caching cache caches", []string{"cach", "cach", "cach"}},
		{"refactor refactoring", []string{"refactor", "refactor"}},
		{"pkg/auth.go", []string{"pkg", "auth", "go"}},
	}
	for _, c := range cases {
		got := tokenize(c.in)
		if len(got) != len(c.want) {
			t.Errorf("tokenize(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("tokenize(%q) = %v, want %v", c.in, got, c.want)
				break
			}
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

	// IDF-weighted cosine scores over the fixture corpus (message + tags only):
	//   intentA <-> intentB : 0.172  (single shared stem "refactor", below default 0.2)
	//   intentA <-> todo    : 0.270  (refactor)
	//   intentB <-> todo    : 0.638  (refactor + token)
	expected := []struct {
		a, b string
		want float64
	}{
		{intentA, todoID, 0.270},
		{todoID, intentA, 0.270},
		{intentB, todoID, 0.638},
		{todoID, intentB, 0.638},
	}
	for _, e := range expected {
		if !g.HasEdge(EdgeRelatedTo, e.a, e.b) {
			t.Errorf("missing RELATED_TO edge %s -> %s", e.a, e.b)
			continue
		}
		edge := g.Edge(EdgeKey(EdgeRelatedTo, e.a, e.b))
		if edge == nil || !almostEqual(edge.Weight, e.want) {
			t.Errorf("edge %s -> %s weight = %v, want %v", e.a, e.b, edge.Weight, e.want)
		}
		if edge != nil && edge.Attrs["similarity"] != fmt.Sprintf("%.3f", edge.Weight) {
			t.Errorf("edge %s -> %s similarity attr = %v", e.a, e.b, edge.Attrs["similarity"])
		}
	}

	if n := len(g.EdgesOfType(EdgeRelatedTo)); n != len(expected) {
		t.Errorf("RELATED_TO edge count = %d, want %d", n, len(expected))
	}

	// Single-stem overlap ("refactor" shared once) is below the default
	// threshold, so intentA and intentB must not be linked.
	if g.HasEdge(EdgeRelatedTo, intentA, intentB) {
		t.Errorf("unexpected RELATED_TO edge %s -> %s at default threshold", intentA, intentB)
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

func TestSimilarityTopKTruncation(t *testing.T) {
	projectRoot := writeFixtureProject(t)
	opts := DefaultBuildOptions(projectRoot)
	opts.SimilarityThreshold = 0.1
	opts.SimilarityTopK = 1
	g, err := Build(opts)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// With topK 1 each node links to at most one other, so the 3-candidate
	// corpus yields at most 3 directed edges (one best pick per node).
	if n := len(g.EdgesOfType(EdgeRelatedTo)); n > 3 {
		t.Errorf("topK=1 should cap RELATED_TO edges, got %d", n)
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
