package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"dizz/internal/config"
)

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Initialize dizz in current directory",
	Run: func (cmd *cobra.Command, args []string) {
		runInit()
	},
}

type DizzConfig struct {
	ProjectName	string		`json:"project_name"`
	RootPath		string		`json:"root_path"`
	Include			[]string	`json:"include"`
	Exclude			[]string	`json:"exclude"`
}

func runInit() {
	// Create .dizz directory to store the states
	cwd, _ := os.Getwd()
	projectName := filepath.Base(cwd)
	// DEBUG: 
	trackDir := config.TrackDirPath(cwd)
	if err := os.MkdirAll(filepath.Join(trackDir, "history"), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .prog directory: %v\n", err)
		os.Exit(1)
	}

	// handle config for the project
	dizzConfig := DizzConfig{
		ProjectName : projectName,
		RootPath		: ".",
		Include			: []string{"**/*.go"},
		Exclude			: []string{"vendor/**",  "node_modules/**", ".git/**"},
	}

	dizzConfigPath := config.ConfigFilePath(trackDir)
	data, _ := json.MarshalIndent(dizzConfig, "", "  ")
	if err := os.WriteFile(dizzConfigPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	// finalization
	fmt.Printf("✓ Initialized %s\n", projectName)
	fmt.Printf("  Created %s/\n", trackDir)
	fmt.Printf("  Project: %s\n", projectName)
	fmt.Printf("\nNext: Run '%s whereami' to see your project state\n", config.AppName)
}

