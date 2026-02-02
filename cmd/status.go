package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Quick project health check",
	Long: `Shows a compact summary of your project state.
Use this for a quick health check without full analysis.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

func runStatus() {
	// Always analyze with current project state
	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("✗")+" "+err.Error())
		os.Exit(1)
	}

	summary := projectState.GetSummary()

	// Calculate health score
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

	// Project info
	cwd, _ := os.Getwd()
	trackDir := config.TrackDirPath(cwd)
	var cfg config.Config
	configPath := config.ConfigFilePath(trackDir)
	if err := store.Load(configPath, &cfg); err == nil {
		fmt.Printf("  %s %s\n", ui.Muted("Project:"), ui.Highlight(cfg.ProjectName))
	}

	// Git info
	if integrations.IsRepo() {
		if branch, err := integrations.GetCurrentBranch(); err == nil {
			fmt.Printf("  %s %s\n", ui.Muted("Branch:"), ui.Info(branch))
		}
		if projectState.GitCommit != "" {
			shortCommit := projectState.GitCommit
			if len(shortCommit) > 7 {
				shortCommit = shortCommit[:7]
			}
			fmt.Printf("  %s %s\n", ui.Muted("Commit:"), ui.Muted(shortCommit))
		}
	}

	// Last updated
	timeSince := time.Since(projectState.UpdatedAt)
	timeDisplay := utils.FormatTime(timeSince)
	fmt.Printf("  %s %s\n", ui.Muted("Updated:"), ui.Muted(timeDisplay.Text))
	fmt.Println()

	// Health indicator
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

	fmt.Printf("  %s %s %s\n",
		ui.Muted("Health:"),
		ui.Colorize(healthIcon, healthColor),
		ui.Colorize(fmt.Sprintf("%d%%", healthScore), healthColor))
	fmt.Println()

	// Symbol breakdown
	fmt.Println(ui.Muted("  Symbols:"))

	if activeCount > 0 {
		bar := createBar(activeCount, summary.TotalSymbols, ui.BrightGreen)
		fmt.Printf("    %s %-12s %s %s\n",
			ui.Success("✓"),
			"Active",
			ui.Success(fmt.Sprintf("%3d", activeCount)),
			bar)
	}

	if summary.ByState[state.Planned] > 0 {
		bar := createBar(summary.ByState[state.Planned], summary.TotalSymbols, ui.BrightYellow)
		fmt.Printf("    %s %-12s %s %s\n",
			ui.Warning("⚠"),
			"Planned",
			ui.Warning(fmt.Sprintf("%3d", summary.ByState[state.Planned])),
			bar)
	}

	if summary.ByState[state.Unused] > 0 {
		bar := createBar(summary.ByState[state.Unused], summary.TotalSymbols, ui.BrightCyan)
		fmt.Printf("    %s%-12s %s %s\n",
			ui.Info("⚪"),
			"Unused",
			ui.Info(fmt.Sprintf("%3d", summary.ByState[state.Unused])),
			bar)
	}

	if summary.ByState[state.Unstable] > 0 {
		bar := createBar(summary.ByState[state.Unstable], summary.TotalSymbols, ui.BrightRed)
		fmt.Printf("      %-12s %s %s\n",
			"Unstable",
			ui.Error(fmt.Sprintf("%3d", summary.ByState[state.Unstable])),
			bar)
	}

	if summary.ByState[state.Abandoned] > 0 {
		bar := createBar(summary.ByState[state.Abandoned], summary.TotalSymbols, ui.Gray)
		fmt.Printf("       %-12s %s %s\n",
			"Abandoned",
			ui.Muted(fmt.Sprintf("%3d", summary.ByState[state.Abandoned])),
			bar)
	}

	fmt.Printf(ui.Muted("    ──────────────────────\n"))
	fmt.Printf(ui.Muted("      Total        %3d\n"), summary.TotalSymbols)
	fmt.Println()

	if summary.ActiveTodos > 0 {
		fmt.Printf("  %s %s\n", ui.Info("📝 TODOs:"), ui.Info(fmt.Sprintf("%d", summary.ActiveTodos)))
		fmt.Println()
	}

	// Action items
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

// createBar creates a simple progress bar
func createBar(count, total int, color string) string {
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

// @ignore-unstable - this function is intentionally excluded from instability analysis
func anotherFunction() {
	// This function would be excluded from instability analysis
}
