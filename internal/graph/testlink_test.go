package graph

import (
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

func TestIsTestFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"pkg/auth_test.go", true},
		{"pkg/test_auth.py", true},
		{"pkg/spec_helper.rb", true},
		{"pkg/parser_test.rs", true},
		{"pkg/parser_spec.rs", true},
		{"src/app.test.ts", true},
		{"src/app.spec.jsx", true},
		{"src/app.go", false},
		{"src/auth.go", false},
		{"src/README.md", false},
		{"pkg/testdata/fixture.go", false},
	}
	for _, tc := range cases {
		if got := IsTestFile(tc.path); got != tc.want {
			t.Errorf("IsTestFile(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestTestCandidatesFromNames(t *testing.T) {
	got := testCandidates("pkg/auth_test.go", []string{"TestLogin", "TestTokenExpiry"})
	for _, want := range []string{"auth", "Login", "TokenExpiry"} {
		found := false
		for _, c := range got {
			if c == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("testCandidates missing %q (got %v)", want, got)
		}
	}
}

func TestLinkTests(t *testing.T) {
	g := NewGraph()
	g.ProjectRoot = "/proj"
	for _, sym := range []struct{ file, name string }{
		{"pkg/auth.go", "Login"},
		{"pkg/auth.go", "Token"},
		{"pkg/validate.go", "Validate"},
	} {
		id := SymbolID(sym.file, sym.name)
		g.AddNode(&Node{ID: id, Type: NodeSymbol, Label: sym.name, Attrs: map[string]interface{}{"file": sym.file}})
	}
	for _, rel := range []string{"pkg/auth.go", "pkg/validate.go", "pkg/auth_test.go"} {
		g.AddNode(&Node{ID: FileID(rel), Type: NodeFile, Label: rel, Attrs: map[string]interface{}{"path": rel}})
	}

	files := []string{"pkg/auth.go", "pkg/validate.go", "pkg/auth_test.go"}
	resolver := NewImportResolver("/proj", files)
	cached := map[string][]signals.Signal{
		"pkg/auth_test.go": {
			{Type: signals.FunctionDefined, Name: "TestLogin", File: "pkg/auth_test.go", Line: 1, EndLine: 5},
			{Type: signals.ImportFound, Name: "example.com/test/pkg", File: "pkg/auth_test.go", Line: 1},
		},
	}
	g.LinkTests(resolver, cached)

	testID := TestID("pkg/auth_test.go")
	if g.Node(testID) == nil {
		t.Fatalf("missing test node")
	}
	login := g.SymbolsNamed("Login")[0].ID
	if !g.HasEdge(EdgeTests, testID, login) {
		t.Errorf("missing TESTS edge %s -> %s (naming method)", testID, login)
	}
	// Import proximity links the test to every symbol in the resolved package.
	// example.com/test/pkg resolves to the pkg/ directory, so Token is covered too.
	if !g.HasEdge(EdgeTests, testID, g.SymbolsNamed("Token")[0].ID) {
		t.Errorf("missing TESTS edge %s -> Token (imports method)", testID)
	}
}
