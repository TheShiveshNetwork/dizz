package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/internal/analyzer"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/regex"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/ast"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/discover"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
)

var (
	showAll    bool
	verboseOut bool
)

var whereamiCmd = &cobra.Command{
	Use:   "whereami",
	Short: "Show what needs your attention",
	Long: `Analyzes your code and shows:
- What needs to be implemented (planned)
- What's changing too much (unstable)
- What's not being used (unused/abandoned)

Focus on what matters. Active code is hidden by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		runWhereami()
	},
}

func init() {
	whereamiCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all symbols including active ones")
	whereamiCmd.Flags().BoolVarP(&verboseOut, "verbose", "v", false, "Show detailed analysis info")
}

func runWhereami() {
	// Check if initialized
	cwd, _ := os.Getwd()
	trackDir := config.TrackDirPath(cwd)
	if _, err := os.Stat(trackDir); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, ui.Error("✗")+" Not a dizz project. Run 'dizz init' first.")
		os.Exit(1)
	}

	if verboseOut {
		fmt.Println(ui.Muted("🔍 Analyzing project..."))
		fmt.Println()
	}

	// Load config
	var cfg config.Config
	configPath := config.ConfigFilePath(trackDir)
	if err := store.Load(configPath, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error loading config: %v\n"), err)
		os.Exit(1)
	}

	// Step 1: Discover files
	files, err := discover.CodeFiles(cfg.RootPath, cfg.Exclude)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error discovering files: %v\n"), err)
		os.Exit(1)
	}

	if verboseOut {
		fmt.Printf(ui.Muted("Found %d code files\n"), len(files))
	}

	// Step 2: Build analyzer registry
	registry := analyzer.NewRegistry()
	registry.Register(&ast.Analyzer{})
	registry.Register(regex.NewAnalyzer())

	// Step 3: Analyze files to extract signals
	sigSet, err := registry.AnalyzeFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error analyzing files: %v\n"), err)
		os.Exit(1)
	}

	if verboseOut {
		fmt.Printf(ui.Muted("Extracted %d signals\n"), len(sigSet.Signals))
		fmt.Println()
	}

	// Step 4: Interpret signals into state
	scorer := state.NewScorer()
	projectState := scorer.InterpretSignals(sigSet)

	// Step 5: Enrich with git context if available
	if integrations.IsRepo() {
		if commit, err := integrations.GetCurrentCommit(); err == nil {
			projectState.GitCommit = commit
		}

		// Add churn data to symbols
		for i := range projectState.Symbols {
			symbol := &projectState.Symbols[i]
			if churn, err := integrations.GetFileChurn(symbol.File, 20); err == nil {
				symbol.ChurnCount = churn
			}
			if lastMod, err := integrations.GetFileLastModified(symbol.File); err == nil {
				symbol.LastTouched = &lastMod
			}
		}

		// Re-score with churn data
		for i := range projectState.Symbols {
			scorer.Score(&projectState.Symbols[i])
		}
	}

	// Step 6: Save state
	statePath := config.StateFilePath(trackDir)
	if err := store.Save(statePath, projectState); err != nil {
		fmt.Fprintf(os.Stderr, ui.Warning("Warning: Could not save state: %v\n"), err)
	}

	// Step 7: Display results
	printFocusedState(projectState)
}

func printFocusedState(ps *state.ProjectState) {
	planned := ps.GetSymbolsByState(state.Planned)
	unstable := ps.GetSymbolsByState(state.Unstable)
	unused := ps.GetSymbolsByState(state.Unused)
	abandoned := ps.GetSymbolsByState(state.Abandoned)
	active := ps.GetSymbolsByState(state.Active)

	summary := ps.GetSummary()
	totalIssues := len(planned) + len(unstable) + len(unused) + len(abandoned)

	// Header
	fmt.Println()
	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(ui.Header("  📍 WHERE YOU ARE"))
	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Quick summary with colors
	fmt.Printf("  %s %s\n", ui.Success("✓ Active:"), ui.Success(fmt.Sprintf("%d", len(active))))
	if len(planned) > 0 {
		fmt.Printf("  %s %s\n", ui.Warning("⚠ Planned:"), ui.Warning(fmt.Sprintf("%d", len(planned))))
	}
	if len(unstable) > 0 {
		fmt.Printf("  %s %s\n", ui.Error("🔥 Unstable:"), ui.Error(fmt.Sprintf("%d", len(unstable))))
	}
	if len(unused) > 0 {
		fmt.Printf("  %s %s\n", ui.Info("⚪ Unused:"), ui.Info(fmt.Sprintf("%d", len(unused))))
	}
	if len(abandoned) > 0 {
		fmt.Printf("  %s %s\n", ui.Muted("❌ Abandoned:"), ui.Muted(fmt.Sprintf("%d", len(abandoned))))
	}
	fmt.Println()

	if totalIssues == 0 {
		fmt.Println(ui.Success("  🎉 All clear! Everything is active and stable."))
		fmt.Println()
		if showAll && len(active) > 0 {
			printActiveSymbols(active)
		}
		return
	}

	// Print sections that need attention (in priority order)
	
	// 1. PLANNED - Highest priority
	if len(planned) > 0 {
		fmt.Println(ui.Warning("━━ ⚠ PLANNED") + ui.Muted(" (needs implementation)"))
		for i, sym := range planned {
			if i >= 5 && !showAll {
				fmt.Printf(ui.Muted("     ... and %d more (use --all to show)\n"), len(planned)-5)
				break
			}
			fmt.Printf("  %s\n", ui.Highlight(sym.Name))
			fmt.Printf("     %s\n", ui.Muted(sym.File))
		}
		fmt.Println()
	}

	// 2. UNSTABLE - High priority
	if len(unstable) > 0 {
		fmt.Println(ui.Error("━━ 🔥 UNSTABLE") + ui.Muted(" (changing too much)"))
		for i, sym := range unstable {
			if i >= 5 && !showAll {
				fmt.Printf(ui.Muted("     ... and %d more (use --all to show)\n"), len(unstable)-5)
				break
			}
			fmt.Printf("  %s ", ui.Highlight(sym.Name))
			fmt.Printf(ui.Error("(churn: %d)\n"), sym.ChurnCount)
			fmt.Printf("     %s\n", ui.Muted(sym.File))
		}
		fmt.Println()
	}

	// 3. UNUSED - Medium priority
	if len(unused) > 0 {
		fmt.Println(ui.Info("━━ ⚪ UNUSED") + ui.Muted(" (not called anywhere)"))
		for i, sym := range unused {
			if i >= 5 && !showAll {
				fmt.Printf(ui.Muted("     ... and %d more (use --all to show)\n"), len(unused)-5)
				break
			}
			fmt.Printf("  %s\n", ui.Highlight(sym.Name))
			fmt.Printf("     %s\n", ui.Muted(sym.File))
		}
		fmt.Println()
	}

	// 4. ABANDONED - Consider removal
	if len(abandoned) > 0 {
		fmt.Println(ui.Muted("━━ ❌ ABANDONED") + ui.Muted(" (old, not used)"))
		for i, sym := range abandoned {
			if i >= 3 && !showAll {
				fmt.Printf(ui.Muted("     ... and %d more (use --all to show)\n"), len(abandoned)-3)
				break
			}
			fmt.Printf("  %s ", ui.Muted(sym.Name))
			fmt.Printf(ui.Muted("(churn: %d)\n"), sym.ChurnCount)
			fmt.Printf("     %s\n", ui.Muted(sym.File))
		}
		fmt.Println()
	}

	activeTodos := ps.GetActiveTodos()
	if len(activeTodos) > 0 {
		fmt.Println(ui.Info("━━ 📝 TODOS"))
		limit := 3
		if len(activeTodos) < limit {
			limit = len(activeTodos)
		}
		for i := 0; i < limit; i++ {
			todo := activeTodos[i]
			fmt.Printf("  %s\n", ui.Muted(fmt.Sprintf("%s:%d", todo.File, todo.Line)))
			fmt.Printf("     %s\n", todo.Text)
		}
		if len(activeTodos) > 3 {
			fmt.Printf(ui.Muted("     ... and %d more\n"), len(activeTodos)-3)
		}
		fmt.Println()
	}

	// Show active if requested
	if showAll && len(active) > 0 {
		printActiveSymbols(active)
	}

	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(ui.Header("  💡 NEXT ACTION"))
	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	suggestion := state.SuggestNextAction(ps)
	fmt.Printf("  %s\n", ui.Highlight("→ "+suggestion))
	fmt.Println()

	fmt.Printf(ui.Muted("  %d symbols · %d need attention · %d active\n"), 
		summary.TotalSymbols, totalIssues, len(active))
	if !showAll && totalIssues > 10 {
		fmt.Printf(ui.Muted("  Use 'dizz whereami --all' to see everything\n"))
	}
	fmt.Println()
}

func printActiveSymbols(active []state.Symbol) {
	fmt.Println(ui.Success("━━ ✓ ACTIVE") + ui.Muted(" (working well)"))
	for i, sym := range active {
		if i >= 10 {
			fmt.Printf(ui.Muted("     ... and %d more\n"), len(active)-10)
			break
		}
		fmt.Printf("  %s\n", ui.Muted(sym.Name))
	}
	fmt.Println()
}

