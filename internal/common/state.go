package cmd

import (
	"errors"
	"os"

	"github.com/TheShiveshNetwork/dizz/internal/analyzer"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/ast"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/regex"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/discover"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

// AnalysisOptions controls what to include/exclude from analysis
type AnalysisOptions struct {
	IgnoreUnstable  bool
	IgnoreUnused    bool
	IgnoreAbandoned bool
	// Add more options as needed
}

// EnsureCurrentStateWithAnalysis ensures we have current project state by always analyzing
// This provides live, up-to-date data on every call
func EnsureCurrentStateWithAnalysis(options *AnalysisOptions) (*state.ProjectState, error) {
	cwd, _ := os.Getwd()
	trackDir := config.TrackDirPath(cwd)

	// Check if project is initialized
	if _, err := os.Stat(trackDir); os.IsNotExist(err) {
		return nil, errors.New("Not a dizz project. Run 'dizz init' first.")
	}

	// Always run full analysis for live data
	return runCurrentAnalysisWithOptions(options)
}

// EnsureCurrentState ensures we have up-to-date project state (legacy for list command)
func EnsureCurrentState() (*state.ProjectState, error) {
	cwd, _ := os.Getwd()
	trackDir := config.TrackDirPath(cwd)

	// Check if project is initialized
	if _, err := os.Stat(trackDir); os.IsNotExist(err) {
		return nil, errors.New("Not a dizz project. Run 'dizz init' first.")
	}

	// Load existing state (for list command that uses cached data)
	var projectState state.ProjectState
	statePath := config.StateFilePath(trackDir)

	if err := store.Load(statePath, &projectState); err != nil {
		// No state exists, run analysis to create it
		return runCurrentAnalysis()
	}

	return &projectState, nil
}

// runCurrentAnalysis performs a full project analysis (legacy for backward compatibility)
func runCurrentAnalysis() (*state.ProjectState, error) {
	return runCurrentAnalysisWithOptions(&AnalysisOptions{})
}

// runCurrentAnalysisWithOptions performs a full project analysis with filtering options
func runCurrentAnalysisWithOptions(options *AnalysisOptions) (*state.ProjectState, error) {
	cwd, _ := os.Getwd()
	trackDir := config.TrackDirPath(cwd)

	// Load config
	var cfg config.Config
	configPath := config.ConfigFilePath(trackDir)
	if err := store.Load(configPath, &cfg); err != nil {
		return nil, err
	}

	// Step 1: Discover files
	files, err := discover.CodeFiles(cfg.RootPath, cfg.Exclude)
	if err != nil {
		return nil, err
	}

	// Step 2: Build analyzer registry
	registry := analyzer.NewRegistry()
	registry.Register(&ast.Analyzer{})
	registry.Register(regex.NewAnalyzer())

	// Step 3: Analyze files to extract signals
	sigSet, err := registry.AnalyzeFiles(files)
	if err != nil {
		return nil, err
	}

	// Step 4: Interpret signals into state
	scorer := state.NewScorer()
	projectState := scorer.InterpretSignals(sigSet)

	// Step 5: Filter symbols based on ignore options
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

	// Step 6: Enrich with git context if available
	if integrations.IsRepo() {
		if commit, err := integrations.GetCurrentCommit(); err == nil {
			projectState.GitCommit = commit
		}

		// Add churn data to symbols
		for i := range projectState.Symbols {
			symbol := &projectState.Symbols[i]
			if churn, err := integrations.GetFunctionChurn(symbol.File, symbol.Name, symbol.Line, symbol.EndLine, 20); err == nil {
				symbol.ChurnCount = churn
			}
			if lastMod, err := integrations.GetFileLastModified(symbol.File); err == nil {
				symbol.LastTouched = &lastMod
			}
		}
	}

	// Step 7: Save state
	statePath := config.StateFilePath(trackDir)
	if err := store.Save(statePath, projectState); err != nil {
		// Continue even if saving fails
	}

	return projectState, nil
}
