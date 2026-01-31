package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"dizz/internal/analyzer"
	"dizz/internal/state"
	"dizz/internal/store"
	"dizz/internal/ui"
)

var whereamiCmd = &cobra.Command{
	Use:   "whereami",
	Short: "Show current project state and what to work on next",
	Long: `Analyzes your code to show:
- Which functions are used
- Which are declared but unused
- What's planned (TODOs)

This is your north star for "what should I work on next?"`,
	Run: func(cmd *cobra.Command, args []string) {
		runWhereami()
	},
}

func runWhereami() {
	// Check if initialized
	if _, err := os.Stat(".dizz"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Error: Not a dizz project. Run 'dizz init' first.")
		os.Exit(1)
	}

	fmt.Println("🔍 Analyzing project...")
	fmt.Println()

	defined, called := analyzer.AnalyzeFunctions(".")
	todos := analyzer.FindAllTodos(".")
	
	var functions []state.Function
	
	for name, file := range defined {
		isCalled := called[name]
		hasTodo := false
		
		// Check if this function has a TODO tag
		for _, todo := range todos {
			if todo.File == file && containsFunction(todo.Line, name) {
				hasTodo = true
				break
			}
		}
		
		funcState, confidence := state.ScoreFunction(isCalled, hasTodo)
		
		functions = append(functions, state.Function{
			Name:       name,
			File:       file,
			State:      funcState,
			Confidence: confidence,
		})
	}
	
	projectState := state.ProjectState{
		UpdatedAt: time.Now().Format(time.RFC3339),
		Functions: functions,
		Todos:     formatTodos(todos),
	}
	
	// Step 4: Save state
	if err := store.SaveJson(".dizz/state.json", projectState); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not save state: %v\n", err)
	}
	
	// Step 5: Display results
	ui.PrintProjectState(projectState)
}

func containsFunction(line, funcName string) bool {
	// Simple heuristic: check if function name is in the line
	// This could be improved with better parsing
	return false // Simplified for now
}

func formatTodos(todos []analyzer.Todo) []string {
	var result []string
	for _, todo := range todos {
		result = append(result, fmt.Sprintf("%s:%d: %s", todo.File, todo.LineNum, todo.Line))
	}
	return result
}
