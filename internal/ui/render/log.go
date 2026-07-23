package render

import (
	"fmt"

	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
)

func LogSummary(ps *state.ProjectState) {
	planned := ps.GetSymbolsByState(state.Planned)
	unstable := ps.GetSymbolsByState(state.Unstable)
	unused := ps.GetSymbolsByState(state.Unused)
	abandoned := ps.GetSymbolsByState(state.Abandoned)
	active := ps.GetSymbolsByState(state.Active)

	fmt.Println()
	fmt.Printf("%s %s\n", ui.Success("✓ Active:"), ui.Success(fmt.Sprintf("%d", len(active))))
	if len(planned) > 0 {
		fmt.Printf("  %s %s\n", ui.Warning("Planned:"), ui.Warning(fmt.Sprintf("%d", len(planned))))
	}
	if len(unused) > 0 {
		fmt.Printf("  %s %s\n", ui.Info("Unused:"), ui.Info(fmt.Sprintf("%d", len(unused))))
	}
	if len(unstable) > 0 {
		fmt.Printf("  %s %s\n", ui.Error("Unstable:"), ui.Error(fmt.Sprintf("%d", len(unstable))))
	}
	if len(abandoned) > 0 {
		fmt.Printf("  %s %s\n", ui.Muted("Abandoned:"), ui.Muted(fmt.Sprintf("%d", len(abandoned))))
	}
	fmt.Println()
}

func LogNoIssues() {
	fmt.Println(ui.Success("  🎉 All clear! Everything is active and stable."))
	fmt.Println()
}

func LogActiveSymbols(active []state.Symbol) {
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

func LogNextAction(suggestion string, totalSymbols, issueCount, activeCount int, showAll bool) {
	fmt.Println()
	fmt.Println(ui.Header("  💡 NEXT ACTION"))
	fmt.Println()
	fmt.Printf("  %s\n", ui.Info("→ "+suggestion))
	fmt.Println()

	fmt.Printf(ui.Muted("  %d symbols · %d need attention · %d active\n"),
		totalSymbols, issueCount, activeCount)
	if !showAll && issueCount > 10 {
		fmt.Printf("%s", ui.Muted("  Use 'dizz log --all' to see everything\n"))
	}
	fmt.Println()
}

func LogAnalyzing() {
	fmt.Println(ui.Muted("Analyzing project..."))
	fmt.Println()
}

func LogExtracted(count int) {
	fmt.Printf(ui.Muted("Extracted %d signals\n"), count)
	fmt.Println()
}
