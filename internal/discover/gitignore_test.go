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
