package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/spf13/cobra"
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
	cwd, _ := os.Getwd()
	trackDir := config.TrackDirPath(cwd)

	// Load state
	var projectState state.ProjectState
	statePath := config.StateFilePath(trackDir)
	if err := store.Load(statePath, &projectState); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("✗")+" No state found. Run 'dizz whereami' first.")
		os.Exit(1)
	}

	// Load config
	var cfg config.Config
	configPath := config.ConfigFilePath(trackDir)
	store.Load(configPath, &cfg)

	// Calculate time away
	timeSince := time.Since(projectState.UpdatedAt)

	// Header
	fmt.Println()
	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(ui.Header("  🔄 RESUMING: " + cfg.ProjectName))
	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Time context
	var timeStr string
	var timeColor string
	if timeSince < 24*time.Hour {
		timeStr = "Less than a day"
		timeColor = ui.BrightGreen
	} else if timeSince < 7*24*time.Hour {
		days := int(timeSince.Hours() / 24)
		timeStr = fmt.Sprintf("%d days", days)
		timeColor = ui.BrightYellow
	} else if timeSince < 30*24*time.Hour {
		weeks := int(timeSince.Hours() / 24 / 7)
		timeStr = fmt.Sprintf("%d weeks", weeks)
		timeColor = ui.BrightYellow
	} else {
		months := int(timeSince.Hours() / 24 / 30)
		timeStr = fmt.Sprintf("%d months", months)
		timeColor = ui.BrightRed
	}

	fmt.Printf("  %s %s\n", ui.Muted("Last worked on:"), ui.Colorize(timeStr+" ago", timeColor))

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
		fmt.Println(ui.Header("  📝 WHERE YOU LEFT OFF"))
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

		if len(unstable) > 0 {
			fmt.Printf("  %s Code with high churn:\n", ui.Error("🔥"))
			for i := 0; i < len(unstable) && i < 3; i++ {
				fmt.Printf("    • %s ", ui.Highlight(unstable[i].Name))
				fmt.Printf(ui.Error("(%d changes)\n"), unstable[i].ChurnCount)
			}
			fmt.Println()
		}
	}

	// Quick summary
	summary := projectState.GetSummary()
	active := summary.ByState[state.Active]
	issues := len(planned) + len(unstable) + len(unused)

	fmt.Println(ui.Header("  📊 QUICK SUMMARY"))
	fmt.Println()
	fmt.Printf("  %s %s symbols working well\n", ui.Success("✓"), ui.Success(fmt.Sprintf("%d", active)))
	if issues > 0 {
		fmt.Printf("  %s %s items need attention\n", ui.Warning("⚠"), ui.Warning(fmt.Sprintf("%d", issues)))
	}
	fmt.Println()

	// Suggested next action
	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(ui.Header("  💡 WHAT TO DO NOW"))
	fmt.Println(ui.Header("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if timeSince > 24*time.Hour {
		fmt.Printf("  %s Re-analyze to get current state\n", ui.Highlight("1."))
		fmt.Printf("     %s\n", ui.Muted("dizz whereami"))
		fmt.Println()
	}

	suggestion := state.SuggestNextAction(&projectState)
	fmt.Printf("  %s %s\n", ui.Highlight("→"), suggestion)
	fmt.Println()

	// Footer
	if timeSince > 24*time.Hour {
		fmt.Println(ui.Muted("  💡 Run 'dizz whereami' to refresh the analysis"))
	} else {
		fmt.Println(ui.Muted("  💡 Run 'dizz status' for a quick health check"))
	}
	fmt.Println()
}
