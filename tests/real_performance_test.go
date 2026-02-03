package benchmarks

import (
	"os"
	"testing"
	"time"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
)

// TestRealAnalysisPerformance tests the actual analysis performance in the same way CLI runs
func TestRealAnalysisPerformance(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	// Save current working directory
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origWD)

	// Change to project root (like CLI would)
	projectRoot, err := commonPkg.FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot failed: %v", err)
	}

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	t.Log("=== Real Analysis Performance Test ===")
	t.Logf("Working directory: %s", projectRoot)

	// Test 1: Cold start (clear all caches)
	t.Run("ColdStart", func(t *testing.T) {
		integrations.InvalidateCache()

		start := time.Now()
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}
		duration := time.Since(start)

		t.Logf("Cold start: %v", duration)
		t.Logf("Symbols found: %d", len(projectState.Symbols))

		if len(projectState.Symbols) == 0 {
			t.Error("No symbols found - this is NOT expected")
		} else {
			t.Logf("✓ Symbol extraction working: %d symbols", len(projectState.Symbols))
		}
	})

	// Test 2: Warm start (using cached results)
	t.Run("WarmStart", func(t *testing.T) {
		start := time.Now()
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}
		duration := time.Since(start)

		t.Logf("Warm start: %v", duration)
		t.Logf("Symbols found: %d", len(projectState.Symbols))
	})

	// Test 3: Measure git operations specifically
	t.Run("GitOperationsOnly", func(t *testing.T) {
		integrations.InvalidateCache()

		// First run to cache non-git parts
		_, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			t.Fatalf("First analysis failed: %v", err)
		}

		// Clear git cache only
		integrations.InvalidateCache()

		start := time.Now()
		_, err = commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			t.Fatalf("Second analysis failed: %v", err)
		}
		duration := time.Since(start)

		t.Logf("Git operations only: %v", duration)
	})

	// Test 4: Performance breakdown
	t.Run("PerformanceBreakdown", func(t *testing.T) {
		integrations.InvalidateCache()

		totalStart := time.Now()
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		totalDuration := time.Since(totalStart)

		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		// Now measure git operations
		gitStart := time.Now()
		integrations.InvalidateCache()

		// Simulate just git operations by running batch analysis
		symbolData := make([]interface{}, len(projectState.Symbols))
		for i, symbol := range projectState.Symbols {
			symbolData[i] = struct {
				File    string
				Name    string
				Line    int
				EndLine int
			}{
				File:    symbol.File,
				Name:    symbol.Name,
				Line:    symbol.Line,
				EndLine: symbol.EndLine,
			}
		}

		_, err = integrations.BatchGitAnalysis(symbolData)
		if err != nil {
			t.Fatalf("Git analysis failed: %v", err)
		}
		gitDuration := time.Since(gitStart)

		t.Logf("=== Performance Breakdown ===")
		t.Logf("Total analysis: %v", totalDuration)
		t.Logf("Git operations: %v", gitDuration)

		if totalDuration > 0 {
			gitPercentage := float64(gitDuration) / float64(totalDuration) * 100
			t.Logf("Git operations are %.1f%% of total time", gitPercentage)
		}

		t.Logf("Symbols processed: %d", len(projectState.Symbols))
	})
}
