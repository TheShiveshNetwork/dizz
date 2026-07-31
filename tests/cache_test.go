package benchmarks

import (
	"testing"
	"time"

	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
)

// TestCacheInvalidation tests how cache affects analysis performance
func TestCacheInvalidation(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	t.Log("=== Cache Invalidation Analysis ===")

	// Test 1: Clear all caches and run fresh analysis
	t.Run("FreshAnalysis", func(t *testing.T) {
		// Clear git cache
		integrations.InvalidateCache()

		start := time.Now()
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		t.Logf("Fresh analysis: %v", duration)
		t.Logf("Symbols found: %d", len(projectState.Symbols))
	})

	// Test 2: Run again (should use cache)
	t.Run("CachedAnalysis", func(t *testing.T) {
		start := time.Now()
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		t.Logf("Cached analysis: %v", duration)
		t.Logf("Symbols found: %d", len(projectState.Symbols))
	})

	// Test 3: Clear cache again and measure git operations only
	t.Run("GitOperationsOnly", func(t *testing.T) {
		// First get symbols without git (this caches the non-git part)
		projectState1, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			t.Fatalf("First analysis failed: %v", err)
		}

		// Clear git cache only
		integrations.InvalidateCache()

		// Now measure just git operations
		start := time.Now()
		projectState2, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Second analysis failed: %v", err)
		}

		t.Logf("Git operations only: %v", duration)
		t.Logf("Symbols found: %d (first: %d)", len(projectState2.Symbols), len(projectState1.Symbols))
	})
}
