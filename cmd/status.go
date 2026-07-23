package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
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

// @dizz-ignore-unused
func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("✗")+" "+err.Error())
		os.Exit(1)
	}

	configStore := store.NewConfigStore(trackDir)
	cfg, _ := configStore.LoadConfig()

	var intentSummary *state.IntentSummary
	intentStore := store.NewIntentStore(config.TrackDirPath(trackDir))
	if intentState, err := intentStore.LoadIntentState(); err == nil {
		s := intentState.GetIntentSummary()
		intentSummary = &s
	}

	branch := ""
	if integrations.IsRepo() {
		branch, _ = integrations.GetCurrentBranch()
	}

	data := &render.StatusData{
		Config:        cfg,
		ProjectState:  projectState,
		IntentSummary: intentSummary,
		TimeSince:     time.Since(projectState.UpdatedAt),
		Branch:        branch,
		HasGit:        integrations.IsRepo(),
		IsUntracked:   integrations.HasUntrackedOrModifiedChanges(),
	}

	render.StatusDashboard(data)
}

// @dizz-ignore-unstable - this function is intentionally excluded from instability analysis
func anotherFunction() {
	// This function would be excluded from instability analysis
}
