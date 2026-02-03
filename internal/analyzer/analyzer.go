package analyzer

import (
	"os"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// Analyzer is the interface that all language analyzers must implement
type Analyzer interface {
	// Language returns the language this analyzer supports
	Language() string

	// Supports checks if this analyzer can handle the given file
	Supports(file string) bool

	// Analyze extracts signals from the given files
	Analyze(files []string) (*signals.SignalSet, error)
}

// Registry manages all available analyzers
type Registry struct {
	analyzers []Analyzer
}

// NewRegistry creates a new analyzer registry
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make([]Analyzer, 0),
	}
}

// Register adds an analyzer to the registry
func (r *Registry) Register(analyzer Analyzer) {
	r.analyzers = append(r.analyzers, analyzer)
}

// FindAnalyzer returns the best analyzer for a given file
func (r *Registry) FindAnalyzer(file string) Analyzer {
	for _, analyzer := range r.analyzers {
		if analyzer.Supports(file) {
			return analyzer
		}
	}
	return nil
}

// AnalyzeFiles runs appropriate analyzers on all files
func (r *Registry) AnalyzeFiles(files []string) (*signals.SignalSet, error) {
	// Group files by analyzer
	filesByAnalyzer := make(map[Analyzer][]string)

	for _, file := range files {
		analyzer := r.FindAnalyzer(file)
		if analyzer != nil {
			filesByAnalyzer[analyzer] = append(filesByAnalyzer[analyzer], file)
		}
	}

	// Run each analyzer on its files
	allSignals := &signals.SignalSet{}

	for analyzer, analyzerFiles := range filesByAnalyzer {
		sigSet, err := analyzer.Analyze(analyzerFiles)
		if err != nil {
			return nil, err
		}

		if sigSet != nil {
			for _, sig := range sigSet.Signals {
				allSignals.Add(sig)
			}
		}
	}

	// Also analyze files for intent ignore markers
	for _, file := range files {
		ignoreSignals := analyzeIgnoreMarkers(file)
		for _, sig := range ignoreSignals {
			allSignals.Add(sig)
		}
	}

	return allSignals, nil
}

// analyzeIgnoreMarkers analyzes a file for intent ignore markers
func analyzeIgnoreMarkers(filePath string) []signals.Signal {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []signals.Signal{}
	}

	source := string(content)
	// Determine language from file extension
	language := "unknown"
	if strings.HasSuffix(filePath, ".go") {
		language = "go"
	} else if strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".ts") {
		language = "javascript"
	}

	// Use the signals package to extract ignore markers
	ignoreSignals := signals.ExtractIgnoreMarkers(source, filePath, language)

	// Convert IgnoreSignal to Signal
	var result []signals.Signal
	for _, ignoreSig := range ignoreSignals {
		// Extract ignore type from comment
		ignoreType := extractIgnoreTypeFromSignal(ignoreSig, source)

		if symbolName, ok := ignoreSig.Metadata["symbol_name"].(string); ok {
			signal := signals.NewSignal(signals.IgnoreFlag, filePath).
				WithName(symbolName).
				WithRange(ignoreSig.Line, ignoreSig.Column, ignoreSig.EndLine, ignoreSig.EndColumn).
				WithLanguage(language).
				WithMeta("ignore_type", ignoreType).
				WithMeta("symbol_name", symbolName) // Also copy symbol_name to new signal

			result = append(result, *signal)
		}
	}

	return result
}

// extractIgnoreTypeFromSignal extracts the ignore type from the original comment
func extractIgnoreTypeFromSignal(ignoreSig signals.IgnoreSignal, source string) string {
	lines := strings.Split(source, "\n")
	if ignoreSig.Line > 0 && ignoreSig.Line <= len(lines) {
		line := lines[ignoreSig.Line-1]
		if strings.Contains(line, "@ignore-unused") {
			return "unused"
		} else if strings.Contains(line, "@ignore-unstable") {
			return "unstable"
		} else if strings.Contains(line, "@ignore-abandoned") {
			return "abandoned"
		}
	}
	return "unknown"
}
