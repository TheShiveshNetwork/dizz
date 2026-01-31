package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Snapshot current state to history",
	Run: func(cmd *cobra.Command, args []string) {
		runCommit()
	},
}

func runCommit() {
	// Load current state
	var projectState state.ProjectState
	if err := store.Load(".dizz/state.json", &projectState); err != nil {
		fmt.Fprintln(os.Stderr, "Error: No state found. Run 'dizz whereami' first.")
		os.Exit(1)
	}

	// Create timestamped snapshot
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	historyPath := filepath.Join(".dizz", "history", fmt.Sprintf("state_%s.json", timestamp))

	if err := store.Save(historyPath, projectState); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving snapshot: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Saved snapshot: %s\n", historyPath)
	fmt.Println()
	fmt.Println("💡 Use this to track progress over time")
}
