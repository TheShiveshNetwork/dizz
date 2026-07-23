package render

import (
	"fmt"
	"time"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
)

type StatusData struct {
	Config        *config.Config
	ProjectState  *state.ProjectState
	IntentSummary *state.IntentSummary
	TimeSince     time.Duration
	Branch        string
	HasGit        bool
	IsUntracked   bool
}

func StatusDashboard(data *StatusData) {
	summary := data.ProjectState.GetSummary()

	totalSymbols := float64(summary.TotalSymbols)
	if totalSymbols == 0 {
		fmt.Println(ui.Warning("⚠ No symbols found. Run 'dizz log' to analyze."))
		return
	}

	activeCount := summary.ByState[state.Active]
	issueCount := summary.ByState[state.Planned] +
		summary.ByState[state.Unstable] +
		summary.ByState[state.Unused] +
		summary.ByState[state.Abandoned]

	healthScore := int((float64(activeCount) / totalSymbols) * 100)

	fmt.Println()

	if data.Config != nil {
		fmt.Printf("  %s %s\n", ui.Muted("Project:"), ui.Highlight(data.Config.ProjectName))
	}

	timeDisplay := utils.FormatTime(data.TimeSince)
	fmt.Printf("  %s %s\n", ui.Muted("Code Updated:"), ui.Muted(timeDisplay.Text))

	if data.HasGit {
		if data.IsUntracked {
			fmt.Printf("  %s %s\n",
				ui.Warning("●"),
				ui.Warning("Changes not committed"),
			)
		}
		fmt.Println()
		if data.Branch != "" {
			fmt.Printf("  %s %s\n", ui.Muted("Branch:"), ui.Info(data.Branch))
		}
		if data.ProjectState.GitCommit != nil {
			commitMsg := data.ProjectState.GitCommit.Message
			shortCommit := data.ProjectState.GitCommit.Hash
			if len(shortCommit) > 7 {
				shortCommit = shortCommit[:7]
			}
			fmt.Printf("  %s %s %s\n", ui.Muted("Last Commit:"), commitMsg, ui.Muted("("+shortCommit+")"))
		}
	}

	var healthColor string
	var healthIcon string
	if healthScore >= 80 {
		healthColor = ui.BrightGreen
		healthIcon = "●"
	} else if healthScore >= 60 {
		healthColor = ui.BrightYellow
		healthIcon = "◐"
	} else {
		healthColor = ui.BrightRed
		healthIcon = "○"
	}

	fmt.Println()
	fmt.Printf("  %s %s %s\n",
		ui.Muted("Code Score:"),
		ui.Colorize(healthIcon, healthColor),
		ui.Colorize(fmt.Sprintf("%d%%", healthScore), healthColor))
	fmt.Println()

	fmt.Println(ui.Muted("  Symbols:"))

	if activeCount > 0 {
		bar := ProgressBar(activeCount, summary.TotalSymbols, ui.BrightGreen)
		fmt.Printf("    %s %-12s %s %s\n",
			ui.Success("✓"),
			"Active",
			ui.Success(fmt.Sprintf("%3d", activeCount)),
			bar)
	}

	if summary.ByState[state.Planned] > 0 {
		bar := ProgressBar(summary.ByState[state.Planned], summary.TotalSymbols, ui.BrightYellow)
		fmt.Printf("      %-12s %s %s\n",
			"Planned",
			ui.Warning(fmt.Sprintf("%3d", summary.ByState[state.Planned])),
			bar)
	}

	if summary.ByState[state.Unused] > 0 {
		bar := ProgressBar(summary.ByState[state.Unused], summary.TotalSymbols, ui.BrightCyan)
		fmt.Printf("      %-12s %s %s\n",
			"Unused",
			ui.Info(fmt.Sprintf("%3d", summary.ByState[state.Unused])),
			bar)
	}

	if summary.ByState[state.Unstable] > 0 {
		bar := ProgressBar(summary.ByState[state.Unstable], summary.TotalSymbols, ui.BrightRed)
		fmt.Printf("      %-12s %s %s\n",
			"Unstable",
			ui.Error(fmt.Sprintf("%3d", summary.ByState[state.Unstable])),
			bar)
	}

	if summary.ByState[state.Abandoned] > 0 {
		bar := ProgressBar(summary.ByState[state.Abandoned], summary.TotalSymbols, ui.Gray)
		fmt.Printf("      %-12s %s %s\n",
			"Abandoned",
			ui.Muted(fmt.Sprintf("%3d", summary.ByState[state.Abandoned])),
			bar)
	}

	fmt.Printf("%s", ui.Muted("    ──────────────────────\n"))
	fmt.Printf(ui.Muted("      Total        %3d\n"), summary.TotalSymbols)
	fmt.Println()

	if summary.ActiveTodos > 0 {
		fmt.Printf("  %s %s\n", ui.Info("📝 TODOs:"), ui.Info(fmt.Sprintf("%d", summary.ActiveTodos)))
		fmt.Println()
	}

	if data.IntentSummary != nil && data.IntentSummary.ActiveIntents > 0 {
		fmt.Printf("  %s %s", ui.Info("🎯 Intents:"), ui.Info(fmt.Sprintf("%d", data.IntentSummary.ActiveIntents)))
		if data.IntentSummary.HighSeverity > 0 {
			fmt.Printf(" %s", ui.Error(fmt.Sprintf("(%d high)", data.IntentSummary.HighSeverity)))
		}
		fmt.Println()
		fmt.Println()
	}

	if issueCount > 0 {
		fmt.Println()
		fmt.Printf("  %s items need attention\n", ui.Warning(fmt.Sprintf("%d", issueCount)))
		fmt.Println(ui.Muted("  Run 'dizz log' for details"))
		fmt.Println()
	} else {
		fmt.Println(ui.Success("  🎉 Everything looks good!"))
		fmt.Println()
	}
}

func ProgressBar(count, total int, color string) string {
	if total == 0 {
		return ""
	}

	barWidth := 20
	filled := int(float64(count) / float64(total) * float64(barWidth))

	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += ui.Colorize("█", color)
		} else {
			bar += ui.Muted("░")
		}
	}

	return bar
}
