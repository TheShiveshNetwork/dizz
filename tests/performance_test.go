package benchmarks

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheShiveshNetwork/dizz/integrations"
)

// TestGitPerformanceComparison is a test that measures actual performance differences
// and reports them in a human-readable format
func TestGitPerformanceComparison(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	symbols := createTestSymbols(175) // Realistic project size

	// Prepare symbol data for batch processing
	symbolData := make([]integrations.SymbolRange, len(symbols))
	for i, symbol := range symbols {
		symbolData[i] = integrations.SymbolRange{
			File:    symbol.File,
			Name:    symbol.Name,
			Line:    symbol.Line,
			EndLine: symbol.EndLine,
		}
	}

	// Test 1: Individual calls
	t.Run("IndividualCalls", func(t *testing.T) {
		start := time.Now()
		for _, symbol := range symbols {
			integrations.GetFunctionChurn(symbol.File, symbol.Name, symbol.Line, symbol.EndLine, 20)
			integrations.GetFileLastModified(symbol.File)
		}
		duration := time.Since(start)

		t.Logf("Individual git calls: %v (%d symbols, %d git operations)",
			duration, len(symbols), len(symbols)*2)
	})

	// Test 2: Batch calls (cache miss)
	t.Run("BatchCallsCacheMiss", func(t *testing.T) {
		integrations.InvalidateCache()
		start := time.Now()
		_, err := integrations.BatchGitAnalysis(symbolData)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("BatchGitAnalysis failed: %v", err)
		}

		t.Logf("Batch git calls (cache miss): %v (%d symbols)", duration, len(symbols))
	})

	// Test 3: Batch calls (cache hit)
	t.Run("BatchCallsCacheHit", func(t *testing.T) {
		// Prime cache
		integrations.BatchGitAnalysis(symbolData)

		start := time.Now()
		_, err := integrations.BatchGitAnalysis(symbolData)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("BatchGitAnalysis failed: %v", err)
		}

		t.Logf("Batch git calls (cache hit): %v (%d symbols)", duration, len(symbols))
	})
}

// BenchmarkMicroGitOperations benchmarks individual git operations to find the real bottlenecks
func BenchmarkMicroGitOperations(b *testing.B) {
	if !integrations.IsRepo() {
		b.Skip("Not in a git repository")
	}

	// Test specific git operations that might be slow
	b.Run("GetCurrentCommit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			integrations.GetCurrentCommit()
		}
	})

	b.Run("GetFileLastModified_Single", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			integrations.GetFileLastModified("internal/common/state.go")
		}
	})

	b.Run("GetFunctionChurn_Single", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			integrations.GetFunctionChurn("internal/common/state.go", "runCurrentAnalysisAtRoot", 100, 200, 20)
		}
	})

	b.Run("GetCurrentCommitWithMessage", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			integrations.GetCurrentCommitWithMessage()
		}
	})
}

// TestAnalysisBreakdown breaks down the analysis time by component
func TestAnalysisBreakdown(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	t.Log("=== Analysis Performance Breakdown ===")

	// Test individual components
	start := time.Now()
	headHash, err := integrations.GetCurrentCommit()
	if err != nil {
		t.Fatalf("GetCurrentCommit failed: %v", err)
	}
	getCommitTime := time.Since(start)
	t.Logf("GetCurrentCommit: %v", getCommitTime)

	start = time.Now()
	commit, err := integrations.GetCurrentCommitWithMessage()
	if err != nil {
		t.Fatalf("GetCurrentCommitWithMessage failed: %v", err)
	}
	getCommitWithMsgTime := time.Since(start)
	t.Logf("GetCurrentCommitWithMessage: %v", getCommitWithMsgTime)

	// Test file operations
	files := []string{
		"internal/common/state.go",
		"internal/state/model.go",
		"internal/state/scorer.go",
		"cmd/status.go",
		"internal/integrations/git.go",
	}

	start = time.Now()
	for _, file := range files {
		integrations.GetFileLastModified(file)
	}
	fileTime := time.Since(start)
	t.Logf("GetFileLastModified (%d files): %v", len(files), fileTime)

	// Test function operations
	start = time.Now()
	for i := 0; i < 10; i++ {
		file := files[i%len(files)]
		startLine := 10 + i*10
		endLine := startLine + 20
		integrations.GetFunctionChurn(file, fmt.Sprintf("func%d", i), startLine, endLine, 20)
	}
	funcTime := time.Since(start)
	t.Logf("GetFunctionChurn (10 functions): %v", funcTime)

	t.Logf("\nHEAD: %s", headHash[:7])
	t.Logf("Commit: %s", commit.Message)
}
