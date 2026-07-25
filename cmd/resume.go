package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
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

// @ignore-unused
func init() {
	rootCmd.AddCommand(resumeCmd)
}

func runResume() {
	_, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("✗")+" "+err.Error())
		os.Exit(1)
	}

	planned := projectState.GetSymbolsByState(state.Planned)
	unstable := projectState.GetSymbolsByState(state.Unstable)
	unused := projectState.GetSymbolsByState(state.Unused)
	summary := projectState.GetSummary()
	active := summary.ByState[state.Active]
	issues := len(planned) + len(unstable) + len(unused)

	branch := ""
	hasGit := integrations.IsRepo()
	if hasGit {
		branch, _ = integrations.GetCurrentBranch()
	}

	suggestion := state.SuggestNextAction(projectState)

	data := &render.ResumeData{
		Branch:       branch,
		HasGit:       hasGit,
		TimeSince:    time.Since(projectState.UpdatedAt),
		Planned:      planned,
		ActiveCount:  active,
		IssueCount:   issues,
		TotalSymbols: summary.TotalSymbols,
		Suggestion:   suggestion,
	}

	render.ResumeOutput(data)
}
