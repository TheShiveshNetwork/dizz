package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
)

var todoCmd = &cobra.Command{
	Use:     "todo",
	Aliases: []string{"todos"},
	Short:   "View todos and fixmes found in code",
	Long: `Lists all TODO, FIXME, and other markers found in your codebase during analysis.
This command shows human-authored markers embedded in your source code.`,
}

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all todos in code",
	Run: func(cmd *cobra.Command, args []string) {
		runTodoList()
	},
}

// @ignore-unused
func init() {
	rootCmd.AddCommand(todoCmd)
	todoCmd.AddCommand(todoListCmd)
}

func runTodoList() {
	// Always analyze to get current project state
	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error analyzing files: %v\n"), err)
		os.Exit(1)
	}

	todos := projectState.GetActiveTodos()
	render.RenderTodoList(todos)
}
