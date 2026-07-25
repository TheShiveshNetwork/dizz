package benchmarks

import (
	"testing"
	"time"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/integrations"
)

// TestSymbolExtraction tests if symbols are being extracted from files
func TestSymbolExtraction(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	t.Log("=== Symbol Extraction Test ===")

	// Get project root
	_, err := commonPkg.FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot failed: %v", err)
	}

	// Clear cache to ensure fresh analysis
	integrations.InvalidateCache()

	// Run analysis
	start := time.Now()
	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}
	duration := time.Since(start)

	t.Logf("Analysis duration: %v", duration)
	t.Logf("Symbols found: %d", len(projectState.Symbols))
	t.Logf("Git commit: %v", projectState.GitCommit != nil)

	// If we have symbols, show a few examples
	if len(projectState.Symbols) > 0 {
		t.Logf("Sample symbols:")
		maxToShow := 5
		for i, symbol := range projectState.Symbols {
			if i >= maxToShow {
				break
			}
			t.Logf("- %s (%s:%d)", symbol.Name, symbol.File, symbol.Line)
		}
	} else {
		t.Error("No symbols extracted - this is the real bottleneck!")

		// Debug: Check what files were actually analyzed
		t.Log("Debugging analysis pipeline...")

		// The issue might be in the analyzers themselves
		t.Log("Files discovered in analysis should be ~34-40, but symbols = 0")
		t.Log("This indicates symbol extraction is failing")
	}
}
