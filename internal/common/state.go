package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/analyzer"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/ast"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/regex"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/discover"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/signals"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

var (
	gitignoreCache      []string
	gitignoreCacheMtx   sync.Mutex
	gitignoreCacheMtime time.Time
)

func cachedGitignore(projectRoot string) ([]string, error) {
	path := filepath.Join(projectRoot, ".gitignore")
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	gitignoreCacheMtx.Lock()
	defer gitignoreCacheMtx.Unlock()

	if info.ModTime().Equal(gitignoreCacheMtime) && gitignoreCache != nil {
		return gitignoreCache, nil
	}

	patterns, err := discover.ParseGitignore(projectRoot)
	if err != nil {
		return nil, err
	}

	gitignoreCache = patterns
	gitignoreCacheMtime = info.ModTime()
	return patterns, nil
}

// AnalysisOptions controls what to include/exclude from analysis
type AnalysisOptions struct {
	IgnoreUnstable  bool
	IgnoreUnused    bool
	IgnoreAbandoned bool
	SkipGit         bool
	// Add more options as needed
}

// EnsureCurrentStateWithAnalysis ensures we have current project state by always analyzing
// This provides live, up-to-date data on every call
func EnsureCurrentStateWithAnalysis(options *AnalysisOptions) (*state.ProjectState, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}
	return runCurrentAnalysisAtRoot(projectRoot, options)
}

// @ignore-unused
// EnsureCurrentState ensures we have up-to-date project state (legacy for list command)
func EnsureCurrentState() (*state.ProjectState, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}

	// Load existing state (for list command that uses cached data)
	trackDir := config.TrackDirPath(projectRoot)
	stateStore := store.NewStateStore(trackDir)

	projectState, err := stateStore.LoadProjectState()
	if err != nil {
		// No state exists, run analysis to create it
		return runCurrentAnalysisAtRoot(projectRoot, &AnalysisOptions{})
	}

	return projectState, nil
}

// @ignore-unused
// runCurrentAnalysis performs a full project analysis (legacy for backward compatibility)
func runCurrentAnalysis() (*state.ProjectState, error) {
	return runCurrentAnalysisWithOptions(&AnalysisOptions{})
}

// runCurrentAnalysisWithOptions performs a full project analysis with filtering options
func runCurrentAnalysisWithOptions(options *AnalysisOptions) (*state.ProjectState, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}
	return runCurrentAnalysisAtRoot(projectRoot, options)
}

// runCurrentAnalysisAtRoot performs analysis at a specific project root.
// Phases 2+3: per-file signal cache + git data carry-forward for unchanged files.
func runCurrentAnalysisAtRoot(projectRoot string, options *AnalysisOptions) (*state.ProjectState, error) {
	trackDir := config.TrackDirPath(projectRoot)

	// Load config
	configStore := store.NewConfigStore(trackDir)
	cfg, err := configStore.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Resolve analysis root
	analysisRoot := cfg.RootPath
	if !filepath.IsAbs(analysisRoot) {
		analysisRoot = filepath.Join(projectRoot, analysisRoot)
	}
	analysisRoot = filepath.Clean(analysisRoot)

	// Merge .gitignore patterns into exclude list (copy first to avoid mutating cfg.Exclude)
	excludePatterns := make([]string, len(cfg.Exclude))
	copy(excludePatterns, cfg.Exclude)
	if integrations.IsRepo() {
		if gitignorePatterns, err := cachedGitignore(projectRoot); err == nil {
			excludePatterns = append(excludePatterns, gitignorePatterns...)
		}
	}

	// Discover files
	files, err := discover.CodeFilesWithIncludes(analysisRoot, cfg.Include, excludePatterns)
	if err != nil {
		return nil, err
	}

	// Build analyzer registry
	registry := analyzer.NewRegistry()
	registry.Register(&ast.Analyzer{})
	registry.Register(regex.NewAnalyzer())

	// ── Phase 2: Per-file signal cache ──
	cacheDir := config.CacheDirPath(projectRoot)
	signalCache := store.NewSignalCache(projectRoot, cacheDir)
	signalCache.LoadManifest()

	var allSignals []signals.Signal
	var changedFiles []string

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		mtime := info.ModTime()

		// Fast path: mtime matches cache → no content read needed
		if sigs, ok := signalCache.GetByMTime(file, mtime); ok {
			allSignals = append(allSignals, sigs...)
			continue
		}

		// Slow path: mtime differs — read + hash + (re-)analyze
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		h := sha256.Sum256(content)
		contentHash := hex.EncodeToString(h[:])

		if sigs, ok := signalCache.Get(file, contentHash, mtime); ok {
			// MTime changed but content is the same (e.g. git checkout)
			allSignals = append(allSignals, sigs...)
		} else {
			sigs, err := registry.AnalyzeSingleFile(file, content)
			if err != nil {
				continue
			}
			allSignals = append(allSignals, sigs...)
			signalCache.Set(file, contentHash, mtime, sigs)
			changedFiles = append(changedFiles, file)
		}
	}

	// Evict cache entries for deleted files
	discoveredSet := make(map[string]struct{})
	for _, f := range files {
		discoveredSet[f] = struct{}{}
	}
	signalCache.EvictStale(discoveredSet)
	signalCache.SaveManifest()

	// ── Phase 4: Load previous state early for signal-set comparison ──
	prevStateStore := store.NewStateStore(trackDir)
	prevState, _ := prevStateStore.LoadProjectState()

	// Merge all signals into a SignalSet
	mergedSigSet := &signals.SignalSet{Signals: allSignals}

	// Compute hash of the merged signal set for identity detection
	signalJSON, _ := json.Marshal(mergedSigSet)
	signalHashBytes := sha256.Sum256(signalJSON)
	signalHashStr := hex.EncodeToString(signalHashBytes[:])

	// If the signal set is identical to the previous run, skip scoring entirely.
	// This closes the loop for truly zero-work runs: no file changes, no HEAD
	// change, no analysis, no scoring, no git — just return the cached state.
	if prevState != nil {
		if prevHash, ok := prevState.Metadata["signal_set_hash"].(string); ok && prevHash == signalHashStr {
			prevState.UpdatedAt = time.Now()
			if options != nil {
				var filteredSymbols []state.Symbol
				for _, symbol := range prevState.Symbols {
					shouldInclude := true
					if options.IgnoreUnstable && symbol.State == state.Unstable {
						shouldInclude = false
					}
					if options.IgnoreUnused && symbol.State == state.Unused {
						shouldInclude = false
					}
					if options.IgnoreAbandoned && symbol.State == state.Abandoned {
						shouldInclude = false
					}
					if shouldInclude {
						filteredSymbols = append(filteredSymbols, symbol)
					}
				}
				prevState.Symbols = filteredSymbols
			}
			return prevState, nil
		}
	}

	// Normal flow: run the scorer
	intentStore := store.NewIntentStore(trackDir)
	intentState, _ := intentStore.LoadIntentState()

	scorer := state.NewScorer()
	projectState := scorer.InterpretSignalsWithIntent(mergedSigSet, intentState, prevState)

	// Store signal hash in metadata for future comparison
	if projectState.Metadata == nil {
		projectState.Metadata = make(map[string]interface{})
	}
	projectState.Metadata["signal_set_hash"] = signalHashStr

	// Filter symbols based on ignore options
	if options != nil {
		var filteredSymbols []state.Symbol
		for _, symbol := range projectState.Symbols {
			shouldInclude := true
			if options.IgnoreUnstable && symbol.State == state.Unstable {
				shouldInclude = false
			}
			if options.IgnoreUnused && symbol.State == state.Unused {
				shouldInclude = false
			}
			if options.IgnoreAbandoned && symbol.State == state.Abandoned {
				shouldInclude = false
			}
			if shouldInclude {
				filteredSymbols = append(filteredSymbols, symbol)
			}
		}
		projectState.Symbols = filteredSymbols
	}

	// ── Phase 3: Git data carry-forward + analysis ──
	if (options == nil || !options.SkipGit) && integrations.IsRepo() {
		if commit, err := integrations.GetCurrentCommitWithMessage(); err == nil {
			projectState.GitCommit = &commit
		}

		// Build previous symbol index for git data carry-forward
		prevIndex := make(map[string]int)
		if prevState != nil {
			for i, sym := range prevState.Symbols {
				prevIndex[sym.File+"::"+sym.Name] = i
			}
		}

		// Build changed-files set for O(1) lookup
		changedSet := make(map[string]struct{})
		for _, f := range changedFiles {
			changedSet[f] = struct{}{}
		}

		// Carry forward git data for unchanged symbols;
		// collect changed symbols for batch git analysis.
		var changedSymbolIdx []int
		for i := range projectState.Symbols {
			sym := &projectState.Symbols[i]
			if _, isChanged := changedSet[sym.File]; !isChanged {
				if prevIdx, ok := prevIndex[sym.File+"::"+sym.Name]; ok {
					sym.ChurnCount = prevState.Symbols[prevIdx].ChurnCount
					sym.LastTouched = prevState.Symbols[prevIdx].LastTouched
				}
			} else {
				changedSymbolIdx = append(changedSymbolIdx, i)
			}
		}

		// Run batch git analysis only for changed/new symbols
		if len(changedSymbolIdx) > 0 {
			symbolData := make([]interface{}, len(changedSymbolIdx))
			for j, idx := range changedSymbolIdx {
				sym := projectState.Symbols[idx]
				symbolData[j] = struct {
					File    string
					Name    string
					Line    int
					EndLine int
				}{
					File:    sym.File,
					Name:    sym.Name,
					Line:    sym.Line,
					EndLine: sym.EndLine,
				}
			}

			if gitResult, err := integrations.BatchGitAnalysis(symbolData); err == nil {
				for _, idx := range changedSymbolIdx {
					sym := &projectState.Symbols[idx]
					if lastMod, exists := gitResult.FileLastModified[sym.File]; exists {
						sym.LastTouched = &lastMod
					}
					rangeKey := fmt.Sprintf("%s:%d:%d", sym.File, sym.Line, sym.EndLine)
					if churn, exists := gitResult.FunctionChurn[rangeKey]; exists {
						sym.ChurnCount = churn
					}
				}
			} else {
				for _, idx := range changedSymbolIdx {
					sym := &projectState.Symbols[idx]
					if churn, err := integrations.GetFunctionChurn(sym.File, sym.Name, sym.Line, sym.EndLine, 20); err == nil {
						sym.ChurnCount = churn
					}
					if lastMod, err := integrations.GetFileLastModified(sym.File); err == nil {
						sym.LastTouched = &lastMod
					}
				}
			}
		}
	}

	// Save state
	stateStore := store.NewStateStore(trackDir)
	if err := stateStore.SaveProjectState(projectState); err != nil {
		// Continue even if saving fails
	}

	return projectState, nil
}

// FindProjectRoot searches up directory tree for .dizz directory
func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		trackDir := config.TrackDirPath(dir)
		if _, err := os.Stat(trackDir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", errors.New("Not a dizz project. Run 'dizz init' first.")
		}
		dir = parent
	}
}
