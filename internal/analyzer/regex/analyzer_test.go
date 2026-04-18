package regex

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

func TestAnalyze_TodoExtractionFromAnchoredComments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "example.js")
	content := `// TODO: line comment
const endpoint = "https://example.com/TODO: not-a-comment";
/* TODO: block comment start
 * TODO: block comment continuation
 */
const value = "TODO: still-not-a-comment";
`

	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	analyzer := NewAnalyzer()
	sigSet, err := analyzer.Analyze([]string{testFile})
	if err != nil {
		t.Fatalf("analyze file: %v", err)
	}

	todos := sigSet.ByType(signals.TodoFound)
	if len(todos) != 3 {
		t.Fatalf("expected 3 TODO signals from comment lines, got %d", len(todos))
	}
}
