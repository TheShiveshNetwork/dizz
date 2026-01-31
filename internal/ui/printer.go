package ui

import (
	"fmt"
	"sort"

	"dizz/internal/state"
)

func PrintProjectState(ps state.ProjectState) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("WHERE YOU ARE")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Organize functions by state
	usedFuncs := []state.Function{}
	unusedFuncs := []state.Function{}
	plannedFuncs := []state.Function{}

	for _, fn := range ps.Functions {
		switch fn.State {
		case state.Used:
			usedFuncs = append(usedFuncs, fn)
		case state.Unused:
			unusedFuncs = append(unusedFuncs, fn)
		case state.Planned:
			plannedFuncs = append(plannedFuncs, fn)
		}
	}

	// Sort by name for consistent output
	sort.Slice(usedFuncs, func(i, j int) bool { return usedFuncs[i].Name < usedFuncs[j].Name })
	sort.Slice(unusedFuncs, func(i, j int) bool { return unusedFuncs[i].Name < unusedFuncs[j].Name })
	sort.Slice(plannedFuncs, func(i, j int) bool { return plannedFuncs[i].Name < plannedFuncs[j].Name })

	// Print sections
	if len(usedFuncs) > 0 {
		fmt.Println("✓ USED FUNCTIONS")
		for _, fn := range usedFuncs {
			fmt.Printf("  %s\n", fn.Name)
			fmt.Printf("    %s\n", fn.File)
		}
		fmt.Println()
	}

	if len(unusedFuncs) > 0 {
		fmt.Println("⚠  UNUSED FUNCTIONS (declared but not called)")
		for _, fn := range unusedFuncs {
			fmt.Printf("  %s\n", fn.Name)
			fmt.Printf("    %s\n", fn.File)
		}
		fmt.Println()
	}

	if len(plannedFuncs) > 0 {
		fmt.Println("✗ PLANNED (TODOs found)")
		for _, fn := range plannedFuncs {
			fmt.Printf("  %s\n", fn.Name)
			fmt.Printf("    %s\n", fn.File)
		}
		fmt.Println()
	}

	// Print TODO summary
	if len(ps.Todos) > 0 {
		fmt.Printf("📝 TODOs: %d found\n", len(ps.Todos))
		if len(ps.Todos) <= 5 {
			for _, todo := range ps.Todos {
				fmt.Printf("  • %s\n", todo)
			}
		} else {
			for i := 0; i < 5; i++ {
				fmt.Printf("  • %s\n", ps.Todos[i])
			}
			fmt.Printf("  ... and %d more\n", len(ps.Todos)-5)
		}
		fmt.Println()
	}

	// Suggest next action
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("💡 NEXT SUGGESTED ACTION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	suggestion := state.SuggestNextAction(ps.Functions)
	fmt.Printf("→ %s\n", suggestion)
	fmt.Println()

	// Print summary
	fmt.Printf("Summary: %d functions analyzed (%d used, %d unused, %d planned)\n",
		len(ps.Functions), len(usedFuncs), len(unusedFuncs), len(plannedFuncs))
}

func PrintCompact(ps state.ProjectState) {
	used := 0
	unused := 0
	planned := 0

	for _, fn := range ps.Functions {
		switch fn.State {
		case state.Used:
			used++
		case state.Unused:
			unused++
		case state.Planned:
			planned++
		}
	}

	fmt.Printf("✓ %d used | ⚠ %d unused | ✗ %d planned | 📝 %d todos\n",
		used, unused, planned, len(ps.Todos))
}

