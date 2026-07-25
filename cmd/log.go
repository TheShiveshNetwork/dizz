package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/config"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
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
		render.LogAnalyzing()
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error analyzing files: %v\n"), err)
		os.Exit(1)
	}

	if verboseOut {
		render.LogExtracted(len(projectState.Symbols))
	}

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

	trackDir, _ := commonPkg.FindProjectRoot()
	intentStore := store.NewIntentStore(config.TrackDirPath(trackDir))

	render.LogSummary(ps)

	if totalIssues == 0 {
		render.LogNoIssues()
		if showAll && len(active) > 0 {
			render.LogActiveSymbols(active)
		}
		return
	}

	render.RenderSymbolGroup(render.RenderArgs{
		Title:      "PLANNED",
		Subtitle:   "needs implementation",
		Symbols:    planned,
		ShowAll:    showAll,
		Verbose:    verboseOut,
		MaxPerFile: 3,
		ShowChurn:  false,
	})

	render.RenderSymbolGroup(render.RenderArgs{
		Title:      "UNSTABLE",
		Subtitle:   "changing too much",
		Symbols:    unstable,
		ShowAll:    showAll,
		Verbose:    verboseOut,
		MaxPerFile: 3,
		ShowChurn:  true,
	})

	render.RenderSymbolGroup(render.RenderArgs{
		Title:      "UNUSED",
		Subtitle:   "not called anywhere",
		Symbols:    unused,
		ShowAll:    showAll,
		Verbose:    verboseOut,
		MaxPerFile: 3,
		ShowChurn:  false,
	})

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

	if intentState, err := intentStore.LoadIntentState(); err == nil {
		activeIntents := intentState.GetActiveIntents()
		render.RenderTodosAndIntents(activeTodos, activeIntents)
	} else {
		render.RenderTodos(activeTodos)
	}

	if showAll && len(active) > 0 {
		render.LogActiveSymbols(active)
	}

	suggestion := state.SuggestNextAction(ps)
	render.LogNextAction(suggestion, summary.TotalSymbols, totalIssues, len(active), showAll)
}
