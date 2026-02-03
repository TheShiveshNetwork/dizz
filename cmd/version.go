package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current verison of dizz",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dizz %s\n", version)
	},
}

// @ignore-unused
func init() {
	rootCmd.AddCommand(versionCmd)
}
