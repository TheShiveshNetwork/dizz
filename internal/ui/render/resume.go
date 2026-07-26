package render

import (
	"fmt"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
)

type ResumeData struct {
	Branch       string
	HasGit       bool
	TimeSince    time.Duration
	Planned      []state.Symbol
	ActiveCount  int
	IssueCount   int
	TotalSymbols int
	Suggestion   string
}

func ResumeOutput(data *ResumeData) {
	fmt.Println()

	timeDisplay := utils.FormatTime(data.TimeSince)
	fmt.Printf("  %s %s\n", ui.Muted("Last worked on:"), ui.Colorize(timeDisplay.Text, timeDisplay.Color))

	if data.HasGit && data.Branch != "" {
		fmt.Printf("  %s %s\n", ui.Muted("Current branch:"), ui.Info(data.Branch))
	}
	fmt.Println()

	if len(data.Planned) > 0 {
		fmt.Println()
		fmt.Printf("  %s You had %s planned work:\n", ui.Warning("⚠"), ui.Warning(fmt.Sprintf("%d", len(data.Planned))))
		limit := 3
		if len(data.Planned) < limit {
			limit = len(data.Planned)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("    • %s\n", ui.Highlight(data.Planned[i].Name))
			fmt.Printf("      %s\n", ui.Muted(utils.RelPath(data.Planned[i].File)))
		}
		if len(data.Planned) > 3 {
			fmt.Printf(ui.Muted("    ... and %d more\n"), len(data.Planned)-3)
		}
		fmt.Println()
	}

	fmt.Println(ui.Header("  QUICK SUMMARY"))
	fmt.Println()
	fmt.Printf("  %s %s symbols working well\n", ui.Success("✓"), ui.Success(fmt.Sprintf("%d", data.ActiveCount)))
	if data.IssueCount > 0 {
		fmt.Printf("  %s %s items need attention\n", ui.Warning("⚠"), ui.Warning(fmt.Sprintf("%d", data.IssueCount)))
	}
	fmt.Println()
	fmt.Println()
	fmt.Println(ui.Header("  💡 WHAT TO DO NOW"))
	fmt.Println()

	if data.TimeSince > 24*time.Hour {
		fmt.Printf("  %s Re-analyze to get current state\n", ui.Highlight("1."))
		fmt.Printf("     %s\n", ui.Muted("dizz log"))
		fmt.Println()
	}

	fmt.Printf("  %s %s\n", ui.Highlight("→"), data.Suggestion)
	fmt.Println()

	if data.TimeSince > 24*time.Hour {
		fmt.Println(ui.Muted("  💡 Run 'dizz log' to refresh the analysis"))
	} else {
		fmt.Println(ui.Muted("  💡 Run 'dizz status' for a quick health check"))
	}
	fmt.Println()
}
