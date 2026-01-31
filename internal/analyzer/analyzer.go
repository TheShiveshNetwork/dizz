package analyzer

import "github.com/TheShiveshNetwork/dizz/internal/signals"

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
	
	return allSignals, nil
}

// DefaultRegistry returns a registry with all built-in analyzers
func DefaultRegistry() *Registry {
	reg := NewRegistry()
	// Analyzers will be registered here as they're implemented
	return reg
}
