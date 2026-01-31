package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"dizz/internal/state"
	"dizz/internal/store"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show quick status summary",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

func runStatus() {
	// Load state
	var projectState state.ProjectState
	if err := store.LoadJson(".dizz/state.json", &projectState); err != nil {
		fmt.Fprintln(os.Stderr, "Error: No state found. Run 'dizz whereami' first.")
		os.Exit(1)
	}

	// Count by state
	used := 0
	unused := 0
	planned := 0

	for _, fn := range projectState.Functions {
		switch fn.State {
		case state.Used:
			used++
		case state.Unused:
			unused++
		case state.Planned:
			planned++
		}
	}

	total := len(projectState.Functions)

	fmt.Println("Project Status")
	fmt.Println()
	fmt.Printf("Last updated: %s\n", projectState.UpdatedAt)
	fmt.Println()
	fmt.Printf("Functions:\n")
	fmt.Printf("  ✓ Used:    %d\n", used)
	fmt.Printf("  ⚠ Unused:  %d\n", unused)
	fmt.Printf("  ✗ Planned: %d\n", planned)
	fmt.Printf("  ─────────────\n")
	fmt.Printf("  Total:     %d\n", total)
	fmt.Println()
	fmt.Printf("TODOs: %d\n", len(projectState.Todos))

	if unused > 0 {
		fmt.Println()
		fmt.Println("💡 Tip: Run 'dizz whereami' for detailed analysis")
	}
}
