package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/render"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
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

// @dizz-ignore-unused
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
		info.Description = cfg.Description
		info.Instructions = cfg.Instructions
		info.Guardrails = cfg.Guardrails
		info.Commands = cfg.Commands
		info.AgentDefaults = cfg.AgentDefaults
	}

	if info.HasGit {
		if branch, err := integrations.GetCurrentBranch(); err == nil {
			info.Branch = branch
		}
		if commitInfo, err := integrations.GetCurrentCommitWithMessage(); err == nil {
			hash := commitInfo.Hash
			if len(hash) > 7 {
				hash = hash[:7]
			}
			info.Commit = hash
			info.CommitMessage = commitInfo.Message
		}
		info.Dirty = integrations.HasUntrackedOrModifiedChanges()
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

	if contextSymbolOnly {
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			projectState = state.NewProjectState()
		}
		var buf bytes.Buffer
		for _, symbol := range projectState.Symbols {
			fmt.Fprintf(&buf, "%s|%s|%s|%d\n", utils.RelPath(symbol.File), symbol.Name, symbol.Type, symbol.ChurnCount)
		}
		fmt.Print(buf.String())
		return
	}

	if contextTodosOnly {
		projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
		if err != nil {
			projectState = state.NewProjectState()
		}
		var buf bytes.Buffer
		for _, todo := range projectState.GetActiveTodos() {
			fmt.Fprintf(&buf, "%s|%d|%s|%s\n", utils.RelPath(todo.File), todo.Line, todo.Type, todo.Text)
		}
		fmt.Print(buf.String())
		return
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		projectState = state.NewProjectState()
	}

	renderer := render.NewContextRenderer()
	output, err := renderer.Render(projectState, intentState, info, nil)
	if err != nil {
		fmt.Fprint(os.Stderr, ui.Error("Error rendering context\n"))
		os.Exit(1)
	}
	fmt.Print(output)
}
