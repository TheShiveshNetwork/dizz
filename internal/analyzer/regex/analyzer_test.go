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
 ** TODO: block comment continuation 2
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
	if len(todos) != 4 {
		t.Fatalf("expected 4 TODO signals from comment lines, got %d", len(todos))
	}

	got := map[int]string{}
	for _, todo := range todos {
		got[todo.Line] = todo.Metadata["text"].(string)
	}

	want := map[int]string{
		1: "line comment",
		3: "block comment start",
		4: "block comment continuation",
		5: "block comment continuation 2",
	}

	for line, expectedText := range want {
		if got[line] != expectedText {
			t.Fatalf("expected TODO at line %d with text %q, got %q", line, expectedText, got[line])
		}
	}
}
