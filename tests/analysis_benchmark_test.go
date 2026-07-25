package benchmarks

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/analyzer"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/ast"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/regex"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/discover"
	"github.com/TheShiveshNetwork/dizz/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

// BenchmarkExtractionOnly benchmarks only the symbol extraction part (no git)
func BenchmarkExtractionOnly(b *testing.B) {
	b.Run("Discovery_And_Extraction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runExtractionOnly()
			if err != nil {
				b.Fatalf("Extraction failed: %v", err)
			}
		}
	})
}

func runExtractionOnly() error {
	projectRoot, err := commonPkg.FindProjectRoot()
	if err != nil {
		return err
	}

	trackDir := config.TrackDirPath(projectRoot)
	configStore := store.NewConfigStore(trackDir)
	cfg, err := configStore.LoadConfig()
	if err != nil {
		return err
	}

	analysisRoot := projectRoot
	if !filepath.IsAbs(analysisRoot) {
		analysisRoot = filepath.Join(projectRoot, analysisRoot)
	}
	analysisRoot = filepath.Clean(analysisRoot)

	files, err := discover.CodeFiles(analysisRoot, cfg.Exclude)
	if err != nil {
		return err
	}

	registry := analyzer.NewRegistry()
	registry.Register(&ast.Analyzer{})
	registry.Register(regex.NewAnalyzer())
	sigSet, err := registry.AnalyzeFiles(files)
	if err != nil {
		return err
	}
	if len(sigSet.Signals) == 0 {
		return errors.New("analysis produced no signals")
	}

	return nil
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
