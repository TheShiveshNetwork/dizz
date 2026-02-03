package benchmarks

import (
	"testing"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/discover"
)

// TestFileDiscovery tests if file discovery is working correctly
func TestFileDiscovery(t *testing.T) {
	t.Log("=== File Discovery Test ===")

	// Get actual project root like analysis does
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot failed: %v", err)
	}

	// Test finding files in project root with same exclude patterns as analysis
	files, err := discover.CodeFiles(trackDir, []string{
		".git/**",
		".dizz/**",
		"vendor/**",
		"node_modules/**",
		"tests/**", // We're excluding tests from analysis!
	})

	if err != nil {
		t.Fatalf("File discovery failed: %v", err)
	}

	t.Logf("Files discovered: %d", len(files))

	// Show first few files
	maxToShow := 10
	for i, file := range files {
		if i >= maxToShow {
			break
		}
		t.Logf("- %s", file)
	}

	if len(files) == 0 {
		t.Error("No files discovered - this explains 0 symbols!")
	}

	// Test including tests directory
	filesWithTests, err := discover.CodeFiles(trackDir, []string{
		".git/**",
		".dizz/**",
		"vendor/**",
		"node_modules/**",
		// Note: NOT excluding tests
	})

	if err != nil {
		t.Fatalf("File discovery with tests failed: %v", err)
	}

	t.Logf("Files discovered (including tests): %d", len(filesWithTests))
}
