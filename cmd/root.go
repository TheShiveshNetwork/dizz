package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "dizz",
	Short: "Progress-aware dev CLI - know what to work on next",
	Long: `dizz analyzes your codebase to show what's used, unused, and planned.
	
No magic. Just facts from your code.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// initialize all the subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(whereamiCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(listCmd)
}

