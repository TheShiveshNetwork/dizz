package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertGitignorePattern(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"*.log", "**/*.log"},
		{"build/", "**/build/**"},
		{"/dist", "**/dist"},
		{"node_modules", "**/node_modules"},
		{"src/testdata/", "src/testdata/**"},
		{"build/*.o", "build/*.o"},
		{"dizz", "**/dizz"},
		{"vendor/**", "vendor/**"},
		{"node_modules/", "**/node_modules/**"},
		{"", ""},  // not reachable from real .gitignore
		{"/", ""}, // lone / not valid in practice
	}

	for _, tt := range tests {
		result := convertGitignorePattern(tt.input)
		if result != tt.expected {
			t.Errorf("convertGitignorePattern(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseGitignore(t *testing.T) {
	dir := t.TempDir()
	content := `# comment
*.log
build/
/dist
!keep-this
`
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	patterns, err := ParseGitignore(dir)
	if err != nil {
		t.Fatalf("ParseGitignore failed: %v", err)
	}

	expected := []string{"**/*.log", "**/build/**", "**/dist"}
	if len(patterns) != len(expected) {
		t.Fatalf("got %d patterns, want %d: %v", len(patterns), len(expected), patterns)
	}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("pattern[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestParseGitignoreNoFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ParseGitignore(dir)
	if err == nil {
		t.Error("expected error for missing .gitignore")
	}
}

func TestMatchGitignoreEndToEnd(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		expect  bool
	}{
		// Basic patterns
		{"main.js", "**/main.js", true},
		{"src/main.js", "**/main.js", true},
		{"main.js", "**/*.js", true},
		{"src/main.js", "**/*.js", true},

		// Single * wildcard (matches suffix)
		{"main.js", "*.js", true},
		{"src/main.js", "*.js", true},

		// **/dirname/** pattern (from gitignore dir patterns)
		{"sub/node_modules/test.js", "**/node_modules/**", true},
		{"node_modules/pkg/index.js", "**/node_modules/**", true},
		{"src/sub/node_modules/pkg/index.js", "**/node_modules/**", true},
		{"mynode_modules_stuff/test.js", "**/node_modules/**", false}, // not exact match

		// Directory patterns (no **)
		{"vendor/pkg/file.go", "vendor/**", true},
		{"src/vendor/pkg/file.go", "vendor/**", false}, // anchored

		// Extension patterns
		{"file.log", "**/*.log", true},
		{"sub/file.log", "**/*.log", true},
	}

	for _, tt := range tests {
		result := matchPattern(tt.path, tt.pattern)
		if result != tt.expect {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.expect)
		}
	}
}

func TestConvertNestedGitignorePattern(t *testing.T) {
	tests := []struct {
		dirRel   string
		input    string
		expected string
	}{
		{".opencode", "node_modules/", ".opencode/**/node_modules/**"},
		{".opencode", "node_modules", ".opencode/**/node_modules"},
		{".opencode", "/skill", ".opencode/skill"},
		{".opencode", "cache/", ".opencode/**/cache/**"},
		{".opencode", "dist", ".opencode/**/dist"},
		{".opencode", "foo/bar/", ".opencode/foo/bar/**"},
		{".opencode", "build/*.o", ".opencode/build/*.o"},
		{"sub", "node_modules/", "sub/**/node_modules/**"},
		{"", "", ""},
	}

	for _, tt := range tests {
		result := convertNestedGitignorePattern(tt.dirRel, tt.input)
		if result != tt.expected {
			t.Errorf("convertNestedGitignorePattern(%q, %q) = %q, want %q", tt.dirRel, tt.input, result, tt.expected)
		}
	}
}

func TestMatchNestedGitignoreEndToEnd(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		expect  bool
	}{
		// Nested basename patterns anchored below their directory
		{".opencode/node_modules/pkg/index.js", ".opencode/**/node_modules/**", true},
		{".opencode/sub/node_modules/pkg/index.js", ".opencode/**/node_modules/**", true},
		{"node_modules/pkg/index.js", ".opencode/**/node_modules/**", false},
		{"other/node_modules/x.js", ".opencode/**/node_modules/**", false},

		// Nested anchored patterns
		{".opencode/skill/SKILL.md", ".opencode/skill/**", true},
		{".opencode/skill", ".opencode/skill/**", true},
		{"skill/SKILL.md", ".opencode/skill/**", false},

		// ** between segments (from nested basename dir patterns)
		{".opencode/cache/x/y", ".opencode/**/cache/**", true},
		{".opencode/a/b/cache", ".opencode/**/cache/**", true},
	}

	for _, tt := range tests {
		result := matchPattern(tt.path, tt.pattern)
		if result != tt.expect {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.expect)
		}
	}
}

func TestParseGitignoreTree(t *testing.T) {
	dir := t.TempDir()

	// Root .gitignore ignores node_modules everywhere and .opencode directly.
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules/\n.opencode/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Nested .gitignore under .opencode ignores its own node_modules (the
	// case opencode creates when running inside a project).
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".opencode", ".gitignore"), []byte("node_modules/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Files that must be excluded.
	if err := os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "node_modules", "zod"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "pkg", "x.js"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".opencode", "node_modules", "zod", "y.js"), []byte("y"), 0644); err != nil {
		t.Fatal(err)
	}

	patterns, err := ParseGitignoreTree(dir)
	if err != nil {
		t.Fatalf("ParseGitignoreTree failed: %v", err)
	}

	// Excluded files must not be discovered.
	files, err := CodeFiles(dir, patterns)
	if err != nil {
		t.Fatalf("CodeFiles failed: %v", err)
	}
	for _, f := range files {
		rel, _ := filepath.Rel(dir, f)
		if rel == "node_modules/pkg/x.js" {
			t.Error("root node_modules file was not excluded")
		}
		if rel == ".opencode/node_modules/zod/y.js" {
			t.Error("nested .opencode/node_modules file was not excluded")
		}
	}
	if len(files) != 1 || files[0] != filepath.Join(dir, "main.go") {
		t.Errorf("unexpected files discovered: %v", files)
	}
}

func TestParseGitignoreTreeNestedOnly(t *testing.T) {
	dir := t.TempDir()

	// No root .gitignore; only a nested one. This is the opencode scenario:
	// .opencode/.gitignore exists, root has none.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".opencode", ".gitignore"), []byte("node_modules/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "node_modules", "zod"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".opencode", "node_modules", "zod", "y.js"), []byte("y"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patterns, err := ParseGitignoreTree(dir)
	if err != nil {
		t.Fatalf("ParseGitignoreTree failed: %v", err)
	}

	files, err := CodeFiles(dir, patterns)
	if err != nil {
		t.Fatalf("CodeFiles failed: %v", err)
	}
	for _, f := range files {
		rel, _ := filepath.Rel(dir, f)
		if rel == ".opencode/node_modules/zod/y.js" {
			t.Error("nested node_modules file was not excluded without a root .gitignore")
		}
	}
	if len(files) != 1 || files[0] != filepath.Join(dir, "main.go") {
		t.Errorf("unexpected files discovered: %v", files)
	}
}
