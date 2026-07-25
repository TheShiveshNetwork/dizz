package benchmarks

import (
	"fmt"
	"testing"

	"github.com/TheShiveshNetwork/dizz/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

// createTestSymbols creates realistic test symbols for benchmarking
func createTestSymbols(count int) []state.Symbol {
	symbols := make([]state.Symbol, 0, count)

	files := []string{
		"internal/analyzer/ast/parser.go",
		"internal/analyzer/regex/scanner.go",
		"internal/common/state.go",
		"internal/state/scorer.go",
		"internal/state/model.go",
		"cmd/status.go",
		"cmd/log.go",
		"internal/store/store.go",
		"internal/config/config.go",
		"internal/ui/colors.go",
	}

	functionNames := []string{
		"parseFunction", "analyzeSymbols", "calculateScore",
		"saveProjectState", "loadConfig", "renderStatus",
		"processFile", "validateState", "updateCache",
		"filterResults", "aggregateData", "formatOutput",
	}

	for i := 0; i < count; i++ {
		symbols = append(symbols, state.Symbol{
			Name:    fmt.Sprintf("%s_%d", functionNames[i%len(functionNames)], i),
			File:    files[i%len(files)],
			Line:    10 + (i*5)%200,
			EndLine: 15 + (i*5)%200 + 5,
		})
	}

	return symbols
}

// BenchmarkIndividualGitCalls tests the old approach (individual git calls)
func BenchmarkIndividualGitCalls(b *testing.B) {
	if !integrations.IsRepo() {
		b.Skip("Not in a git repository")
	}

	// Test with different symbol counts
	testSizes := []int{10, 50, 100, 175} // 175 = current project size

	for _, size := range testSizes {
		b.Run(fmt.Sprintf("symbols_%d", size), func(b *testing.B) {
			symbols := createTestSymbols(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, symbol := range symbols {
					integrations.GetFunctionChurn(symbol.File, symbol.Name, symbol.Line, symbol.EndLine, 20)
					integrations.GetFileLastModified(symbol.File)
				}
			}
		})
	}
}

// BenchmarkBatchGitCalls tests the new approach (batch git calls)
func BenchmarkBatchGitCalls(b *testing.B) {
	if !integrations.IsRepo() {
		b.Skip("Not in a git repository")
	}

	// Test with different symbol counts
	testSizes := []int{10, 50, 100, 175} // 175 = current project size

	for _, size := range testSizes {
		b.Run(fmt.Sprintf("symbols_%d", size), func(b *testing.B) {
			symbols := createTestSymbols(size)

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

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Clear cache between runs for fair comparison
				integrations.InvalidateCache()
				_, err := integrations.BatchGitAnalysis(symbolData)
				if err != nil {
					b.Fatalf("BatchGitAnalysis failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCachedBatchGitCalls tests the cache hit performance
func BenchmarkCachedBatchGitCalls(b *testing.B) {
	if !integrations.IsRepo() {
		b.Skip("Not in a git repository")
	}

	// Test with different symbol counts
	testSizes := []int{10, 50, 100, 175} // 175 = current project size

	for _, size := range testSizes {
		b.Run(fmt.Sprintf("symbols_%d", size), func(b *testing.B) {
			symbols := createTestSymbols(size)

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

			// Prime the cache first
			_, err := integrations.BatchGitAnalysis(symbolData)
			if err != nil {
				b.Fatalf("Failed to prime cache: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := integrations.BatchGitAnalysis(symbolData)
				if err != nil {
					b.Fatalf("BatchGitAnalysis failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkGitCommandsComparison directly compares individual vs batch approaches
func BenchmarkGitCommandsComparison(b *testing.B) {
	if !integrations.IsRepo() {
		b.Skip("Not in a git repository")
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

	b.Run("Individual_Calls", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, symbol := range symbols {
				integrations.GetFunctionChurn(symbol.File, symbol.Name, symbol.Line, symbol.EndLine, 20)
				integrations.GetFileLastModified(symbol.File)
			}
		}
	})

	b.Run("Batch_Calls_CacheMiss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			integrations.InvalidateCache()
			_, err := integrations.BatchGitAnalysis(symbolData)
			if err != nil {
				b.Fatalf("BatchGitAnalysis failed: %v", err)
			}
		}
	})

	b.Run("Batch_Calls_CacheHit", func(b *testing.B) {
		// Prime cache
		integrations.BatchGitAnalysis(symbolData)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := integrations.BatchGitAnalysis(symbolData)
			if err != nil {
				b.Fatalf("BatchGitAnalysis failed: %v", err)
			}
		}
	})
}
