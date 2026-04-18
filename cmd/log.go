package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
)

var (
	showAll    bool
	verboseOut bool
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show what needs your attention",
	Long: `Analyzes your code and shows:
- What needs to be implemented (planned)
- What's changing too much (unstable)
- What's not being used (unused/abandoned)

Focus on what matters. Active code is hidden by default.

Flags:
  -a, --all     Show all symbols including active ones
  -v, --verbose   Show detailed analysis info`,
	Run: func(cmd *cobra.Command, args []string) {
		runLog()
	},
}

// @ignore-unused
func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all symbols including active ones")
	logCmd.Flags().BoolVarP(&verboseOut, "verbose", "v", false, "Show detailed analysis info")
}

func runLog() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	if verboseOut {
		fmt.Println(ui.Muted("Analyzing project..."))
		fmt.Println()
	}

	// Always analyze with current project state
	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error analyzing files: %v\n"), err)
		os.Exit(1)
	}

	if verboseOut {
		fmt.Printf(ui.Muted("Extracted %d signals\n"), len(projectState.Symbols))
		fmt.Println()
	}

	// Save state using StateStore
	stateStore := store.NewStateStore(config.TrackDirPath(trackDir))
	if err := stateStore.SaveProjectState(projectState); err != nil {
		fmt.Fprintf(os.Stderr, ui.Warning("Warning: Could not save state: %v\n"), err)
	}

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

	// Get trackDir for intent loading
	trackDir, _ := commonPkg.FindProjectRoot()
	intentStore := store.NewIntentStore(config.TrackDirPath(trackDir))

	fmt.Println()

	// Quick summary
	fmt.Printf("%s %s\n", ui.Success("✓ Active:"), ui.Success(fmt.Sprintf("%d", len(active))))
	if len(planned) > 0 {
		fmt.Printf("  %s %s\n", ui.Warning("Planned:"), ui.Warning(fmt.Sprintf("%d", len(planned))))
	}
	if len(unused) > 0 {
		fmt.Printf("  %s %s\n", ui.Info("Unused:"), ui.Info(fmt.Sprintf("%d", len(unused))))
	}
	if len(unstable) > 0 {
		fmt.Printf("  %s %s\n", ui.Error("   Unstable:"), ui.Error(fmt.Sprintf("%d", len(unstable))))
	}
	if len(abandoned) > 0 {
		fmt.Printf("  %s %s\n", ui.Muted("   Abandoned:"), ui.Muted(fmt.Sprintf("%d", len(abandoned))))
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
	render.RenderSymbolGroup(render.RenderArgs{
		Title:      "PLANNED",
		Subtitle:   "needs implementation",
		Symbols:    planned,
		ShowAll:    showAll,
		Verbose:    verboseOut,
		MaxPerFile: 3,
		ShowChurn:  false,
	})

	// 2. UNSTABLE - High priority
	render.RenderSymbolGroup(render.RenderArgs{
		Title:      "UNSTABLE",
		Subtitle:   "changing too much",
		Symbols:    unstable,
		ShowAll:    showAll,
		Verbose:    verboseOut,
		MaxPerFile: 3,
		ShowChurn:  true,
	})

	// 3. UNUSED - Medium priority
	render.RenderSymbolGroup(render.RenderArgs{
		Title:      "UNUSED",
		Subtitle:   "not called anywhere",
		Symbols:    unused,
		ShowAll:    showAll,
		Verbose:    verboseOut,
		MaxPerFile: 3,
		ShowChurn:  false,
	})

	// 4. ABANDONED - Consider removal
	render.RenderSymbolGroup(render.RenderArgs{
		Title:      "ABANDONED",
		Subtitle:   "old, not used",
		Symbols:    abandoned,
		ShowAll:    showAll,
		Verbose:    verboseOut,
		MaxPerFile: 2,
		ShowChurn:  true,
	})

	activeTodos := ps.GetActiveTodos()

	// Load and render intents
	if intentState, err := intentStore.LoadIntentState(); err == nil {
		activeIntents := intentState.GetActiveIntents()
		render.RenderTodosAndIntents(activeTodos, activeIntents)
	} else {
		// Fallback to just todos for backward compatibility
		render.RenderTodos(activeTodos)
	}

	if showAll && len(active) > 0 {
		printActiveSymbols(active)
	}

	fmt.Println()
	fmt.Println(ui.Header("  💡 NEXT ACTION"))
	fmt.Println()
	suggestion := state.SuggestNextAction(ps)
	fmt.Printf("  %s\n", ui.Info("→ "+suggestion))
	fmt.Println()

	fmt.Printf(ui.Muted("  %d symbols · %d need attention · %d active\n"),
		summary.TotalSymbols, totalIssues, len(active))
	if !showAll && totalIssues > 10 {
		fmt.Printf("%s", ui.Muted("  Use 'dizz log --all' to see everything\n"))
	}
	fmt.Println()
}

func printActiveSymbols(active []state.Symbol) {
	fmt.Println(ui.Success("✓ ACTIVE") + ui.Muted(" (working well)"))
	for i, sym := range active {
		if i >= 10 {
			fmt.Printf(ui.Muted("     ... and %d more\n"), len(active)-10)
			break
		}
		fmt.Printf("  %s\n", ui.Muted(sym.Name))
	}
	fmt.Println()
}
