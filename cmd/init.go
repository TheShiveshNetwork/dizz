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

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() {
	cwd, _ := os.Getwd()
	projectName := filepath.Base(cwd)
	trackDir := config.TrackDirPath(cwd)
	dizzConfigPath := config.ConfigFilePath(trackDir)

	if _, err := os.Stat(dizzConfigPath); err == nil {
		fmt.Println("dizz already initialized")
		fmt.Printf("  Project: %s\n", projectName)
		fmt.Printf("  Path: %s\n", trackDir)
		return
	}

	isGitRepo := integrations.IsRepo()
	if !isGitRepo {
		fmt.Println("Not a git repository")
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

	dizzConfig := defaults.DefaultConfig(projectName)
	data, _ := json.MarshalIndent(dizzConfig, "", "  ")
	if err := os.WriteFile(dizzConfigPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	gitignorePath := filepath.Join(trackDir, ".gitignore")
	gitignoreContent := defaults.GitignoreContent()
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing gitignore file: %v\n", err)
		os.Exit(1)
	}

	createProjectSkill(cwd, projectName)

	fmt.Printf("Initialized %s\n", projectName)

	if isGitRepo {
		hooksDir := config.HooksDirPath(cwd)
		hookPath := filepath.Join(hooksDir, "post-commit")
		hookContent := defaults.LocalPostCommitHookContent(config.AppName)

		if err := integrations.InstallLocalPostCommitHook(hookPath, hookContent); err != nil {
			fmt.Printf("Could not install post-commit hook: %v\n", err)
			fmt.Println("   You can still use dizz manually with 'dizz snapshot'")
		} else {
			fmt.Println("Installed post-commit hook to .dizz/hooks/")
		}

		if err := integrations.SetLocalHooksPath(); err != nil {
			fmt.Printf("Could not set hooks path: %v\n", err)
		} else {
			fmt.Println("Configured git to use hooks from .dizz/hooks/")
		}
	}
	fmt.Printf("  Created %s/\n", trackDir)
	if isGitRepo {
		fmt.Println("  Git integration: enabled")
	} else {
		fmt.Println("  Git integration: disabled")
	}
	fmt.Printf("  Project: %s\n", projectName)
	fmt.Printf("  Agent skill: .agents/skills/dizz/\n")
	fmt.Printf("\nNext: Run '%s log' to see your project state\n", config.AppName)
	fmt.Printf("      Run '%s context' for agent-optimized project context\n", config.AppName)
}

func createProjectSkill(cwd, projectName string) {
	skillDir := filepath.Join(cwd, ".agents", "skills", "dizz")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		fmt.Printf("Warning: could not create skill directory: %v\n", err)
		return
	}

	skillMeta := defaults.SkillMetadata(projectName)
	metaJSON, _ := json.MarshalIndent(skillMeta, "", "  ")
	metaPath := filepath.Join(skillDir, "skill.json")
	if err := os.WriteFile(metaPath, metaJSON, 0644); err != nil {
		fmt.Printf("Warning: could not write skill.json: %v\n", err)
	}

	skillContent := defaults.SkillInstructions(projectName)
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		fmt.Printf("Warning: could not write SKILL.md: %v\n", err)
	}
}
