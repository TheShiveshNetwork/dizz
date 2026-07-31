package graph

import "testing"

func TestImportResolver(t *testing.T) {
	files := []string{
		"internal/store/store.go",
		"internal/store/other.go",
		"internal/graph/graph.go",
		"cmd/root.go",
		"cmd/state.go",
	}
	r := NewImportResolver("/proj", files)

	cases := []struct {
		name   string
		imp    string
		from   string
		want   string
		wantOK bool
	}{
		{"relative dir", "../store", "internal/graph/graph.go", "internal/store/other.go", true},
		{"relative file", "../store/store", "internal/graph/graph.go", "internal/store/store.go", true},
		{"module-style dir", "internal/store", "cmd/root.go", "internal/store/other.go", true},
		{"module-style suffix", "github.com/x/dizz/internal/store", "cmd/root.go", "internal/store/other.go", true},
		{"bare stem", "graph", "cmd/root.go", "internal/graph/graph.go", true},
		{"bare basename", "state", "cmd/root.go", "cmd/state.go", true},
		{"external dep", "golang.org/x/crypto/bcrypt", "cmd/root.go", "", false},
		{"unresolvable relative", "./nope", "cmd/root.go", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := r.Resolve(tc.imp, tc.from)
			if ok != tc.wantOK {
				t.Fatalf("Resolve(%q, %q) ok=%v want %v", tc.imp, tc.from, ok, tc.wantOK)
			}
			if ok && got != tc.want {
				t.Errorf("Resolve(%q, %q) = %q want %q", tc.imp, tc.from, got, tc.want)
			}
		})
	}
}

func TestImportResolverPickFileInDirIsDeterministic(t *testing.T) {
	files := []string{"internal/a/z.go", "internal/a/a.go", "internal/a/m.go"}
	r := NewImportResolver("/proj", files)
	got, ok := r.Resolve("internal/a", "cmd/root.go")
	if !ok {
		t.Fatalf("expected resolution, got none")
	}
	if got != "internal/a/a.go" {
		t.Errorf("pickFileInDir = %q, want %q", got, "internal/a/a.go")
	}
}

func TestImportResolverEmpty(t *testing.T) {
	r := NewImportResolver("/proj", nil)
	if _, ok := r.Resolve("anything", "cmd/root.go"); ok {
		t.Errorf("expected no resolution from empty file set")
	}
}
