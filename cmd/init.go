package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/defaults"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dizz in current directory",
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

// @ignore-unused
func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() {
	cwd, _ := os.Getwd()
	projectName := filepath.Base(cwd)
	trackDir := config.TrackDirPath(cwd)
	dizzConfigPath := config.ConfigFilePath(trackDir)

	// return if dizz config already exists
	if _, err := os.Stat(dizzConfigPath); err == nil {
		fmt.Println("✓ dizz already initialized")
		fmt.Printf("  Project: %s\n", projectName)
		fmt.Printf("  Path: %s\n", trackDir)
		return
	}

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

	objectsDir := config.ObjectsDirPath(cwd)
	refsDir := config.RefsDirPath(cwd)

	// create all the necessary dirs
	dirs := []string{
		trackDir,
		objectsDir,
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
	dizzConfig := defaults.DefaultConfig(projectName)
	data, _ := json.MarshalIndent(dizzConfig, "", "  ")
	if err := os.WriteFile(dizzConfigPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	// add not needed files to gitignore (which can be regenerated)
	gitignorePath := filepath.Join(trackDir, ".gitignore")
	gitignoreContent := defaults.GitignoreContent()
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing gitignore file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Initialized %s\n", projectName)
	// install local post-commit hook (tracked in .dizz/hooks/)
	if isGitRepo {
		hooksDir := config.HooksDirPath(cwd)
		hookPath := filepath.Join(hooksDir, "post-commit")
		hookContent := defaults.LocalPostCommitHookContent(config.AppName)

		if err := integrations.InstallLocalPostCommitHook(hookPath, hookContent); err != nil {
			fmt.Printf("⚠️  Could not install post-commit hook: %v\n", err)
			fmt.Println("   You can still use dizz manually with 'dizz snapshot'")
		} else {
			fmt.Println("✓ Installed post-commit hook to .dizz/hooks/")
		}

		if err := integrations.SetLocalHooksPath(); err != nil {
			fmt.Printf("⚠️  Could not set hooks path: %v\n", err)
		} else {
			fmt.Println("✓ Configured git to use hooks from .dizz/hooks/")
		}
	}
	fmt.Printf("  Created %s/\n", trackDir)
	if isGitRepo {
		fmt.Println("  Git integration: enabled")
	} else {
		fmt.Println("  Git integration: disabled")
	}
	fmt.Printf("  Project: %s\n", projectName)
	fmt.Printf("\nNext: Run '%s log' to see your project state\n", config.AppName)
}
