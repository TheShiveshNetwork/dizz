package cmd

import (
	"fmt"
	"os"
	"strings"

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
	dumpFull   bool
	logFilters []string
)

var validLogFilters = map[string]bool{
	"active": true, "planned": true, "unused": true, "unstable": true, "abandoned": true,
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show what needs your attention",
	Long: `Analyzes your code and shows what needs attention.

Symbols needing action (planned, unstable, unused, abandoned) are shown by default.
Active code is hidden unless --dump is used.

Flags:
  -a, --all      Show all items needing action (no per-file limit)
  -v, --verbose  Show detailed analysis info
  -d, --dump     Dump every symbol including active with full details
  --filter       Filter by state (repeatable: --filter=unused --filter=planned)
                 Valid states: active, planned, unused, unstable, abandoned`,
	Run: func(cmd *cobra.Command, args []string) {
		runLog()
	},
}

// @ignore-unused
func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all items needing action (no per-file limit)")
	logCmd.Flags().BoolVarP(&verboseOut, "verbose", "v", false, "Show detailed analysis info")
	logCmd.Flags().BoolVarP(&dumpFull, "dump", "d", false, "Dump every symbol including active with full details")
	logCmd.Flags().StringArrayVar(&logFilters, "filter", nil, "Filter by state (repeatable: --filter=unused --filter=planned)")
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

func parseLogFilters() []state.SymbolState {
	if len(logFilters) == 0 {
		return nil
	}
	var states []state.SymbolState
	for _, f := range logFilters {
		lower := strings.ToLower(f)
		if !validLogFilters[lower] {
			valid := make([]string, 0, len(validLogFilters))
			for k := range validLogFilters {
				valid = append(valid, k)
			}
			fmt.Fprintf(os.Stderr, ui.Error("Error: invalid filter %q. Valid states: %s\n"), f, strings.Join(valid, ", "))
			os.Exit(1)
		}
		states = append(states, state.SymbolState(lower))
	}
	return states
}

func hasFilter(filters []state.SymbolState, target state.SymbolState) bool {
	for _, f := range filters {
		if f == target {
			return true
		}
	}
	return false
}

func printFocusedState(ps *state.ProjectState) {
	planned := ps.GetSymbolsByState(state.Planned)
	unstable := ps.GetSymbolsByState(state.Unstable)
	unused := ps.GetSymbolsByState(state.Unused)
	abandoned := ps.GetSymbolsByState(state.Abandoned)
	active := ps.GetSymbolsByState(state.Active)

	filters := parseLogFilters()
	filtered := filters != nil

	if filtered {
		if !hasFilter(filters, state.Planned) {
			planned = nil
		}
		if !hasFilter(filters, state.Unstable) {
			unstable = nil
		}
		if !hasFilter(filters, state.Unused) {
			unused = nil
		}
		if !hasFilter(filters, state.Abandoned) {
			abandoned = nil
		}
		if !hasFilter(filters, state.Active) {
			active = nil
		}
	}

	summary := ps.GetSummary()
	totalIssues := len(planned) + len(unstable) + len(unused) + len(abandoned)

	trackDir, _ := commonPkg.FindProjectRoot()
	intentStore := store.NewIntentStore(config.TrackDirPath(trackDir))

	render.LogSummary(ps)

	if totalIssues == 0 && (!filtered || !hasFilter(filters, state.Active)) {
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

	if !filtered || hasFilter(filters, state.Active) {
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
	}

	suggestion := state.SuggestNextAction(ps)
	render.LogNextAction(suggestion, summary.TotalSymbols, totalIssues, len(active), showAll)
}
