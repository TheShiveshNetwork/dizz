package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
)

var rootCmd = &cobra.Command{
	Use:   "dizz",
	Short: "Progress-aware dev CLI - know what to work on next",
	Long:  `dizz analyzes your codebase to show what's used, unused, and planned.`,
}

func Execute() {
	autoWireHook()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// autoWireHook checks if the current repo has .dizz/hooks/post-commit and
// configures the local core.hooksPath if needed. This ensures hooks work
// immediately after clone without requiring an explicit "dizz init".
func autoWireHook() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	trackDir := config.TrackDirPath(cwd)
	if _, err := os.Stat(trackDir); os.IsNotExist(err) {
		return
	}
	if !integrations.IsRepo() {
		return
	}
	integrations.EnsureLocalHooksConfigured(trackDir)
}

// @ignore-unused
func init() {
	// Root commands are registered via init() in their respective files
}
