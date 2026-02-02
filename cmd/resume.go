package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Quick context after time away",
	Long: `Shows what you need to know after being away from the project.
Optimized for the "I haven't touched this in weeks" scenario.`,
	Run: func(cmd *cobra.Command, args []string) {
		runResume()
	},
}

func runResume() {
	_, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	// Ensure we have current project state
	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("✗")+" "+err.Error())
		os.Exit(1)
	}

	// Calculate time away
	timeSince := time.Since(projectState.UpdatedAt)

	fmt.Println()

	timeDisplay := utils.FormatTime(timeSince)
	fmt.Printf("  %s %s\n", ui.Muted("Last worked on:"), ui.Colorize(timeDisplay.Text, timeDisplay.Color))

	// Git context
	if integrations.IsRepo() {
		if branch, err := integrations.GetCurrentBranch(); err == nil {
			fmt.Printf("  %s %s\n", ui.Muted("Current branch:"), ui.Info(branch))
		}
	}
	fmt.Println()

	// Get priority items
	planned := projectState.GetSymbolsByState(state.Planned)
	unstable := projectState.GetSymbolsByState(state.Unstable)
	unused := projectState.GetSymbolsByState(state.Unused)

	// What you were working on
	if len(planned) > 0 || len(unstable) > 0 {
		fmt.Println()

		if len(planned) > 0 {
			fmt.Printf("  %s You had %s planned work:\n", ui.Warning("⚠"), ui.Warning(fmt.Sprintf("%d", len(planned))))
			limit := 3
			if len(planned) < limit {
				limit = len(planned)
			}
			for i := 0; i < limit; i++ {
				fmt.Printf("    • %s\n", ui.Highlight(planned[i].Name))
				fmt.Printf("      %s\n", ui.Muted(planned[i].File))
			}
			if len(planned) > 3 {
				fmt.Printf(ui.Muted("    ... and %d more\n"), len(planned)-3)
			}
			fmt.Println()
		}
	}

	// Quick summary
	summary := projectState.GetSummary()
	active := summary.ByState[state.Active]
	issues := len(planned) + len(unstable) + len(unused)

	fmt.Println(ui.Header("  QUICK SUMMARY"))
	fmt.Println()
	fmt.Printf("  %s %s symbols working well\n", ui.Success("✓"), ui.Success(fmt.Sprintf("%d", active)))
	if issues > 0 {
		fmt.Printf("  %s %s items need attention\n", ui.Warning("⚠"), ui.Warning(fmt.Sprintf("%d", issues)))
	}
	fmt.Println()
	fmt.Println()
	fmt.Println(ui.Header("  💡 WHAT TO DO NOW"))
	fmt.Println()

	if timeSince > 24*time.Hour {
		fmt.Printf("  %s Re-analyze to get current state\n", ui.Highlight("1."))
		fmt.Printf("     %s\n", ui.Muted("dizz log"))
		fmt.Println()
	}

	suggestion := state.SuggestNextAction(projectState)
	fmt.Printf("  %s %s\n", ui.Highlight("→"), suggestion)
	fmt.Println()

	// Footer
	if timeSince > 24*time.Hour {
		fmt.Println(ui.Muted("  💡 Run 'dizz log' to refresh the analysis"))
	} else {
		fmt.Println(ui.Muted("  💡 Run 'dizz status' for a quick health check"))
	}
	fmt.Println()
}
