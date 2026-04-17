package benchmarks

import (
	"testing"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
)

// BenchmarkExtractionOnly benchmarks only the symbol extraction part (no git)
func BenchmarkExtractionOnly(b *testing.B) {
	b.Run("Symbol_Extraction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Manually run discovery and extraction without git enrichment
			_, err := commonPkg.EnsureCurrentStateWithAnalysis(&commonPkg.AnalysisOptions{SkipGit: true})
			if err != nil {
				b.Fatalf("Analysis failed: %v", err)
			}
		}
	})
}

// BenchmarkFullAnalysis benchmarks the entire analysis pipeline
func BenchmarkFullAnalysis(b *testing.B) {
	if !integrations.IsRepo() {
		b.Skip("Not in a git repository")
	}

	b.Run("Complete_Analysis", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
			if err != nil {
				b.Fatalf("Analysis failed: %v", err)
			}
		}
	})

	b.Run("Complete_Analysis_WithCache", func(b *testing.B) {
		// Prime the analysis once
		_, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			b.Fatalf("Initial analysis failed: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
			if err != nil {
				b.Fatalf("Cached analysis failed: %v", err)
			}
		}
	})
}
