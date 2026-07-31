package graph

import (
	"strings"
	"testing"
)

func buildQueryFixture(t *testing.T) (*Graph, *QueryEngine) {
	t.Helper()
	g := NewGraph()
	g.ProjectRoot = "/proj"

	addFile := func(rel string) string {
		id := FileID(rel)
		g.AddNode(&Node{ID: id, Type: NodeFile, Label: rel, Attrs: map[string]interface{}{"path": rel}})
		return id
	}
	addSymbol := func(file, name string) string {
		id := SymbolID(file, name)
		g.AddNode(&Node{ID: id, Type: NodeSymbol, Label: name, Attrs: map[string]interface{}{"file": file}})
		return id
	}
	addDep := func(name string) string {
		id := DepID(name)
		g.AddNode(&Node{ID: id, Type: NodeDep, Label: name, Attrs: map[string]interface{}{"import_path": name}})
		return id
	}

	handlerFile := addFile("handler.go")
	serviceFile := addFile("service.go")
	addFile("handler_test.go")
	login := addSymbol("handler.go", "Login")
	parseToken := addSymbol("handler.go", "ParseToken")
	validate := addSymbol("service.go", "Validate")
	addDep("golang.org/x/crypto/bcrypt")

	g.AddEdge(&Edge{Type: EdgeDefinedIn, From: login, To: handlerFile, Weight: 1})
	g.AddEdge(&Edge{Type: EdgeDefinedIn, From: parseToken, To: handlerFile, Weight: 1})
	g.AddEdge(&Edge{Type: EdgeDefinedIn, From: validate, To: serviceFile, Weight: 1})
	g.AddEdge(&Edge{Type: EdgeContains, From: handlerFile, To: login, Weight: 1})
	g.AddEdge(&Edge{Type: EdgeContains, From: handlerFile, To: parseToken, Weight: 1})
	g.AddEdge(&Edge{Type: EdgeContains, From: serviceFile, To: validate, Weight: 1})

	g.AddEdge(&Edge{Type: EdgeCalls, From: login, To: parseToken, Weight: 1})
	g.AddEdge(&Edge{Type: EdgeCalls, From: login, To: validate, Weight: 1})

	testID := TestID("handler_test.go")
	g.AddNode(&Node{ID: testID, Type: NodeTest, Label: "handler_test.go", Attrs: map[string]interface{}{"path": "handler_test.go"}})
	g.AddEdge(&Edge{Type: EdgeTests, From: testID, To: login, Weight: 0.9})

	g.AddEdge(&Edge{Type: EdgeCoChangesWith, From: handlerFile, To: serviceFile, Weight: 0.6})
	g.AddEdge(&Edge{Type: EdgeCoChangesWith, From: serviceFile, To: handlerFile, Weight: 0.6})

	return g, NewQueryEngine(g)
}

func TestResolveEntity(t *testing.T) {
	_, qe := buildQueryFixture(t)

	if n, err := qe.ResolveEntity("Login"); err != nil || n.ID != SymbolID("handler.go", "Login") {
		t.Errorf("bare symbol: %v %v", n, err)
	}
	if n, err := qe.ResolveEntity("symbol:Login@handler.go"); err != nil || n.ID != SymbolID("handler.go", "Login") {
		t.Errorf("symbol@file: %v %v", n, err)
	}
	if n, err := qe.ResolveEntity("symbol:handler.go::Login"); err != nil || n.ID != SymbolID("handler.go", "Login") {
		t.Errorf("symbol::file: %v %v", n, err)
	}
	if n, err := qe.ResolveEntity("file:handler.go"); err != nil || n.ID != FileID("handler.go") {
		t.Errorf("file: %v %v", n, err)
	}
	if _, err := qe.ResolveEntity("no_such_symbol_xyz"); err == nil {
		t.Errorf("expected error for unknown entity")
	}
}

func TestResolveEntityAmbiguous(t *testing.T) {
	g := NewGraph()
	for _, f := range []string{"a.go", "b.go"} {
		g.AddNode(&Node{ID: SymbolID(f, "Dup"), Type: NodeSymbol, Label: "Dup", Attrs: map[string]interface{}{"file": f}})
	}
	qe := NewQueryEngine(g)
	if _, err := qe.ResolveEntity("Dup"); err == nil {
		t.Errorf("expected ambiguity error")
	}
	if n, err := qe.ResolveEntity("symbol:Dup@b.go"); err != nil || n.ID != SymbolID("b.go", "Dup") {
		t.Errorf("disambiguated resolve failed: %v %v", n, err)
	}
}

func TestBlastRadius(t *testing.T) {
	g, qe := buildQueryFixture(t)
	login := g.SymbolsNamed("Login")[0].ID

	affected := qe.BlastRadius(login, 2)
	ids := map[string]bool{}
	for _, a := range affected {
		ids[a.NodeID] = true
	}
	for _, want := range []string{
		SymbolID("handler.go", "ParseToken"),
		SymbolID("service.go", "Validate"),
		TestID("handler_test.go"),
		FileID("service.go"),
		FileID("handler.go"),
	} {
		if !ids[want] {
			t.Errorf("blast radius missing %s (got %v)", want, ids)
		}
	}
	// Login itself is the seed and not reported.
	if ids[login] {
		t.Errorf("blast radius should not include the seed node itself")
	}
	// Deterministic ordering by descending score.
	for i := 1; i < len(affected); i++ {
		if affected[i].Score > affected[i-1].Score {
			t.Fatalf("blast radius not sorted by descending score: %+v", affected)
		}
	}
}

func TestTrace(t *testing.T) {
	g, qe := buildQueryFixture(t)
	login := g.SymbolsNamed("Login")[0].ID

	tr := qe.Trace(login)
	if tr == nil {
		t.Fatalf("trace returned nil")
	}
	if len(tr.Callees) != 2 {
		t.Errorf("callees = %d, want 2", len(tr.Callees))
	}
	if len(tr.Tests) != 1 {
		t.Errorf("tests = %d, want 1", len(tr.Tests))
	}
	if tr.DefinedIn == nil || tr.DefinedIn.To != FileID("handler.go") {
		t.Errorf("defined_in = %+v", tr.DefinedIn)
	}
}

func TestTestCoverage(t *testing.T) {
	g, qe := buildQueryFixture(t)
	login := g.SymbolsNamed("Login")[0].ID

	refs := qe.TestCoverage(login)
	if len(refs) != 1 {
		t.Fatalf("test coverage = %d, want 1", len(refs))
	}
	if refs[0].From != TestID("handler_test.go") {
		t.Errorf("covering test = %s", refs[0].From)
	}
}

func TestCoChanges(t *testing.T) {
	_, qe := buildQueryFixture(t)
	handlerFile := FileID("handler.go")

	refs := qe.CoChanges(handlerFile, 0.5)
	if len(refs) == 0 {
		t.Fatalf("expected co-change edges")
	}
	for _, r := range refs {
		if r.From != handlerFile && r.To != handlerFile {
			t.Errorf("co-change edge unrelated to handler.go: %+v", r)
		}
	}
}

func TestShortestPath(t *testing.T) {
	_, qe := buildQueryFixture(t)
	login := SymbolID("handler.go", "Login")
	serviceFile := FileID("service.go")

	path := qe.ShortestPath(login, serviceFile)
	if len(path) != 3 {
		t.Fatalf("path length = %d, want 3 (%v)", len(path), path)
	}
	if path[0] != login || path[len(path)-1] != serviceFile {
		t.Errorf("path endpoints wrong: %v", path)
	}

	if got := qe.ShortestPath(login, login); len(got) != 1 || got[0] != login {
		t.Errorf("self path = %v", got)
	}
}

func TestImpactScope(t *testing.T) {
	_, qe := buildQueryFixture(t)
	nodes := qe.ImpactScope("handler*.go")
	if len(nodes) != 2 {
		t.Fatalf("impact scope = %d, want 2", len(nodes))
	}
	for _, n := range nodes {
		if !strings.HasPrefix(n.ID, "file:handler") {
			t.Errorf("unexpected node in scope: %s", n.ID)
		}
	}
}

func TestMarshalTON(t *testing.T) {
	_, qe := buildQueryFixture(t)
	g := qe.g
	data, err := g.MarshalTON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(data)
	for _, want := range []string{"# graph", "# nodes", "# edges", "CALLS|symbol:handler.go::Login|symbol:service.go::Validate"} {
		if !strings.Contains(out, want) {
			t.Errorf("TON output missing %q", want)
		}
	}
}
