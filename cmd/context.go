package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/render"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/spf13/cobra"
)

var (
	contextIntentOnly bool
	contextSymbolOnly bool
	contextTodosOnly  bool
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Token-optimized project context for agents",
	Long: `Outputs a compact, token-efficient summary of the project state
in TON (Token-Optimized Notation) format, designed for AI agent consumption.

Includes active intents, symbol health, todos, and git context
in a pipe-delimited, single-line-per-record format.`,
	Run: func(cmd *cobra.Command, args []string) {
		runContext()
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.Flags().BoolVar(&contextIntentOnly, "intents", false, "Show intents only")
	contextCmd.Flags().BoolVar(&contextSymbolOnly, "symbols", false, "Show symbols only")
	contextCmd.Flags().BoolVar(&contextTodosOnly, "todos", false, "Show todos only")
}

func runContext() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprint(os.Stderr, ui.Error("No dizz project found. Run 'dizz init' first.\n"))
		os.Exit(1)
	}

	info := render.ContextInfo{
		ProjectName: filepath.Base(trackDir),
		HasGit:      integrations.IsRepo(),
	}

	configStore := store.NewConfigStore(config.TrackDirPath(trackDir))
	if cfg, err := configStore.LoadConfig(); err == nil {
		info.ProjectName = cfg.ProjectName
		info.Agentic = cfg.Agentic
		info.ConfigRoot = cfg.RootPath
		info.ConfigIncludeCount = len(cfg.Include)
		info.ConfigExcludeCount = len(cfg.Exclude)
	}

	if info.HasGit {
		if branch, err := integrations.GetCurrentBranch(); err == nil {
			info.Branch = branch
		}
		if commit, err := integrations.GetCurrentCommit(); err == nil {
			if len(commit) > 7 {
				commit = commit[:7]
			}
			info.Commit = commit
		}
	}

	intentStore := store.NewIntentStore(config.TrackDirPath(trackDir))
	intentState, err := intentStore.LoadIntentState()
	if err != nil {
		intentState = state.NewIntentState()
	}

	if contextIntentOnly {
		data, err := intentState.MarshalTON()
		if err != nil {
			fmt.Fprint(os.Stderr, ui.Error("Error marshaling intents\n"))
			os.Exit(1)
		}
		fmt.Print(string(data))
		return
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		projectState = state.NewProjectState()
	}

	if contextTodosOnly {
		var buf bytes.Buffer
		for _, todo := range projectState.GetActiveTodos() {
			fmt.Fprintf(&buf, "%s|%d|%s|%s\n", todo.File, todo.Line, todo.Type, todo.Text)
		}
		fmt.Print(buf.String())
		return
	}

	renderer := render.NewContextRenderer()
	output, err := renderer.Render(projectState, intentState, info, nil)
	if err != nil {
		fmt.Fprint(os.Stderr, ui.Error("Error rendering context\n"))
		os.Exit(1)
	}
	fmt.Print(output)
}
