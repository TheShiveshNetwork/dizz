package benchmarks

import (
	"testing"
	"time"

	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
)

// TestFullAnalysisBreakdown breaks down the complete analysis to find remaining bottlenecks
func TestFullAnalysisBreakdown(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	t.Log("=== Full Analysis Performance Breakdown ===")

	// Step 1: Find project root
	start := time.Now()
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot failed: %v", err)
	}
	findRootTime := time.Since(start)
	t.Logf("FindProjectRoot: %v", findRootTime)

	// Step 2: Load existing state (if any)
	start = time.Now()
	// This would normally load from disk, but let's measure the structure
	loadStateTime := time.Since(start)
	t.Logf("LoadState: %v", loadStateTime)

	// Step 3: Discover files
	start = time.Now()
	// This is part of the analysis - let's measure the actual analysis
	analysisTime := time.Since(start)
	t.Logf("FileDiscovery: %v", analysisTime)

	// Step 4: Run actual analysis
	start = time.Now()
	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}
	fullAnalysisTime := time.Since(start)
	t.Logf("FullAnalysis: %v", fullAnalysisTime)

	// Step 5: Break down what happened during analysis
	t.Logf("\nAnalysis Results:")
	t.Logf("- Symbols found: %d", len(projectState.Symbols))
	t.Logf("- Git commit: %v", projectState.GitCommit != nil)
	t.Logf("- Updated at: %v", projectState.UpdatedAt)

	// Test git operations separately
	start = time.Now()
	headHash, err := integrations.GetCurrentCommit()
	if err != nil {
		t.Fatalf("GetCurrentCommit failed: %v", err)
	}
	gitOpsTime := time.Since(start)
	t.Logf("GitOperations: %v", gitOpsTime)

	// Calculate what percentage git operations take up
	if fullAnalysisTime > 0 {
		gitPercentage := float64(gitOpsTime) / float64(fullAnalysisTime) * 100
		t.Logf("\nGit operations are %.1f%% of total analysis time", gitPercentage)
	}

	t.Logf("\nProject root: %s", trackDir)
	t.Logf("Current HEAD: %s", headHash[:7])
}

// TestRealWorldUsage tests the actual CLI usage patterns
func TestRealWorldUsage(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	t.Log("=== Real World Usage Performance ===")

	// Test 1: Cold start (first run)
	t.Run("ColdStart", func(t *testing.T) {
		start := time.Now()
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}
		duration := time.Since(start)

		t.Logf("Cold start analysis: %v", duration)
		t.Logf("Symbols analyzed: %d", len(projectState.Symbols))
	})

	// Test 2: Warm start (subsequent run)
	t.Run("WarmStart", func(t *testing.T) {
		start := time.Now()
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}
		duration := time.Since(start)

		t.Logf("Warm start analysis: %v", duration)
		t.Logf("Symbols analyzed: %d", len(projectState.Symbols))
	})

	// Test 3: Multiple rapid calls (like CLI might do)
	t.Run("MultipleCalls", func(t *testing.T) {
		times := make([]time.Duration, 3)
		for i := 0; i < 3; i++ {
			start := time.Now()
			_, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
			if err != nil {
				t.Fatalf("Analysis %d failed: %v", i, err)
			}
			times[i] = time.Since(start)
		}

		t.Logf("Multiple calls: %v, %v, %v", times[0], times[1], times[2])

		avg := times[0] + times[1] + times[2]
		avg /= 3
		t.Logf("Average: %v", avg)
	})
}
