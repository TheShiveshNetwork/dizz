package analyzer

import (
	"os"
	"strings"
	"sync"

	"github.com/TheShiveshNetwork/dizz/internal/language"
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

	// AnalyzeFile extracts signals from a single file.
	// Returns per-file signals for incremental caching.
	AnalyzeFile(file string) ([]signals.Signal, error)
}

// AnalyzerWithContent is an optional interface that analyzers can implement to
// accept pre-read file content, avoiding redundant I/O when the content is
// already available (e.g. from the cache slow path's SHA-256 hashing).
type AnalyzerWithContent interface {
	Analyzer
	AnalyzeFileContent(file string, content []byte) ([]signals.Signal, error)
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

// AnalyzeFile extracts signals from a single file using the best matching analyzer.
func (r *Registry) AnalyzeFile(file string) ([]signals.Signal, error) {
	analyzer := r.FindAnalyzer(file)
	if analyzer == nil {
		return nil, nil
	}
	return analyzer.AnalyzeFile(file)
}

// AnalyzeSingleFile runs the full per-file pipeline: main analysis + ignore markers.
// Used by the incremental cache path.  When content is non-nil it is used instead
// of re-reading the file from disk.
func (r *Registry) AnalyzeSingleFile(file string, content []byte) ([]signals.Signal, error) {
	var allSignals []signals.Signal

	analyzer := r.FindAnalyzer(file)
	if analyzer != nil {
		var sigs []signals.Signal
		var err error
		if content != nil {
			if awc, ok := analyzer.(AnalyzerWithContent); ok {
				sigs, err = awc.AnalyzeFileContent(file, content)
			} else {
				sigs, err = analyzer.AnalyzeFile(file)
			}
		} else {
			sigs, err = analyzer.AnalyzeFile(file)
		}
		if err != nil {
			return nil, err
		}
		allSignals = append(allSignals, sigs...)
	}

	// Analyze ignore markers for this file
	if content != nil {
		ignoreSigs := AnalyzeIgnoreMarkersFromSource(string(content), file)
		allSignals = append(allSignals, ignoreSigs...)
	} else {
		ignoreSigs := AnalyzeIgnoreMarkers(file)
		allSignals = append(allSignals, ignoreSigs...)
	}

	return allSignals, nil
}

// AnalyzeFiles runs appropriate analyzers on all files
func (r *Registry) AnalyzeFiles(files []string) (*signals.SignalSet, error) {
	if len(files) == 0 {
		return &signals.SignalSet{}, nil
	}

	// Optimization: For very small projects, skip parallelization to avoid overhead
	// Launching goroutines and managing channels is slower than just reading ~40 files.
	if len(files) < 100 {
		return r.analyzeSequentially(files)
	}

	// Group files by analyzer
	filesByAnalyzer := make(map[Analyzer][]string)
	for _, file := range files {
		analyzer := r.FindAnalyzer(file)
		if analyzer != nil {
			filesByAnalyzer[analyzer] = append(filesByAnalyzer[analyzer], file)
		}
	}

	allSignals := &signals.SignalSet{}
	var wg sync.WaitGroup

	// Use channels to collect results instead of a Mutex to avoid contention
	// Each analyzer gets one slot, plus one for ignore markers
	signalChan := make(chan []signals.Signal, len(filesByAnalyzer)+1)
	errChan := make(chan error, len(filesByAnalyzer))

	// 1. Run analyzers in parallel
	for analyzer, analyzerFiles := range filesByAnalyzer {
		wg.Add(1)
		go func(a Analyzer, af []string) {
			defer wg.Done()
			sigSet, err := a.Analyze(af)
			if err != nil {
				errChan <- err
				return
			}
			if sigSet != nil {
				signalChan <- sigSet.Signals
			}
		}(analyzer, analyzerFiles)
	}

	// 2. Analyze ignore markers in parallel (mostly I/O bound)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var ignoreSigs []signals.Signal
		// Limit concurrency for I/O to avoid system resource limits
		const workerCount = 8
		jobs := make(chan string)
		var innerWg sync.WaitGroup
		var mu sync.Mutex

		for i := 0; i < workerCount; i++ {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				for f := range jobs {
					sigs := AnalyzeIgnoreMarkers(f)
					mu.Lock()
					ignoreSigs = append(ignoreSigs, sigs...)
					mu.Unlock()
				}
			}()
		}

		for _, file := range files {
			jobs <- file
		}
		close(jobs)
		innerWg.Wait()
		signalChan <- ignoreSigs
	}()

	// Close signal channel once all workers are done
	go func() {
		wg.Wait()
		close(signalChan)
	}()

	// Collect all signals
	for sigs := range signalChan {
		for _, sig := range sigs {
			allSignals.Add(sig)
		}
	}

	// Return first error if any occurred
	select {
	case err := <-errChan:
		return nil, err
	default:
		return allSignals, nil
	}
}

// analyzeSequentially handles small projects without parallel overhead
func (r *Registry) analyzeSequentially(files []string) (*signals.SignalSet, error) {
	allSignals := &signals.SignalSet{}

	filesByAnalyzer := make(map[Analyzer][]string)
	for _, file := range files {
		analyzer := r.FindAnalyzer(file)
		if analyzer != nil {
			filesByAnalyzer[analyzer] = append(filesByAnalyzer[analyzer], file)
		}
	}

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

	if len(files) < 20 {
		for _, file := range files {
			sigs := AnalyzeIgnoreMarkers(file)
			for _, sig := range sigs {
				allSignals.Add(sig)
			}
		}
	} else {
		const workerCount = 8
		jobs := make(chan string, len(files))
		var innerWg sync.WaitGroup
		var mu sync.Mutex

		for i := 0; i < workerCount; i++ {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				for f := range jobs {
					sigs := AnalyzeIgnoreMarkers(f)
					mu.Lock()
					for _, sig := range sigs {
						allSignals.Add(sig)
					}
					mu.Unlock()
				}
			}()
		}

		for _, file := range files {
			jobs <- file
		}
		close(jobs)
		innerWg.Wait()
	}

	return allSignals, nil
}

// AnalyzeIgnoreMarkers analyzes a file for intent ignore markers.
// It reads the file from disk and delegates to AnalyzeIgnoreMarkersFromSource.
func AnalyzeIgnoreMarkers(filePath string) []signals.Signal {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []signals.Signal{}
	}
	return AnalyzeIgnoreMarkersFromSource(string(content), filePath)
}

// AnalyzeIgnoreMarkersFromSource analyzes pre-read file content for intent ignore markers.
// This avoids an extra file read when the content is already available.
func AnalyzeIgnoreMarkersFromSource(source string, filePath string) []signals.Signal {
	// Use the language registry for accurate language detection.
	langID := "unknown"
	if lc, ok := language.Detect(filePath); ok {
		langID = lc.ID
	}

	// Use the signals package to extract ignore markers
	ignoreSignals := signals.ExtractIgnoreMarkers(source, filePath, langID)

	var result []signals.Signal
	for _, ignoreSig := range ignoreSignals {
		ignoreType := extractIgnoreTypeFromSignal(ignoreSig, source)

		if symbolName, ok := ignoreSig.Metadata["symbol_name"].(string); ok {
			signal := signals.NewSignal(signals.IgnoreFlag, filePath).
				WithName(symbolName).
				WithRange(ignoreSig.Line, ignoreSig.Column, ignoreSig.EndLine, ignoreSig.EndColumn).
				WithLanguage(langID).
				WithMeta("ignore_type", ignoreType).
				WithMeta("symbol_name", symbolName)

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
