package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
)

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Initialize dizz in current directory",
	Run: func (cmd *cobra.Command, args []string) {
		runInit()
	},
}

func runInit() {
	// check git status
	isGitRepo := integrations.IsRepo()
	if !isGitRepo {
		fmt.Println("⚠️  Not a git repository")
		fmt.Println("   Run 'git init' first, or continue without git integration")
		fmt.Print("\nContinue anyway? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
	}

	// Create .dizz directory to store the states
	cwd, _ := os.Getwd()
	projectName := filepath.Base(cwd)
	trackDir := config.TrackDirPath(cwd)
	objectsDir := config.ObjectsDirPath(trackDir)
	refsDir := config.RefsDirPath(trackDir)
	
	// create all the necessary dirs
	dirs := []string{
		trackDir,
		objectsDir,
		filepath.Join(objectsDir, "00"),
		filepath.Join(objectsDir, "01"),
		refsDir,
		filepath.Join(refsDir, "git"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// create dizz config
	dizzConfig := config.DefaultConfig(projectName)
	dizzConfigPath := config.ConfigFilePath(trackDir)
	data, _ := json.MarshalIndent(dizzConfig, "", "  ")
	if err := os.WriteFile(dizzConfigPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	// finalization
	fmt.Printf("✓ Initialized %s\n", projectName)
	// install post-commit hook
	if isGitRepo {
		hookContent := integrations.GetHookContent(config.AppName)
		if err := integrations.InstallPostCommitHook(hookContent); err != nil {
			fmt.Printf("⚠️  Could not install git hook: %v\n", err)
			fmt.Println("   You can still use dizz manually with 'dizz snapshot'")
		} else {
			fmt.Println("✓ Installed git post-commit hook")
		}
	}
	fmt.Printf("  Created %s/\n", trackDir)
	if isGitRepo {
		fmt.Println("  Git integration: enabled")
	} else {
		fmt.Println("  Git integration: disabled")
	}
	fmt.Printf("  Project: %s\n", projectName)
	fmt.Printf("\nNext: Run '%s whereami' to see your project state\n", config.AppName)
}

