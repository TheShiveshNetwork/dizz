package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/defaults"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
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

	data := &render.InitData{
		ProjectName: projectName,
		TrackDir:    trackDir,
	}

	if _, err := os.Stat(dizzConfigPath); err == nil {
		render.InitAlreadyInitialized(data)
		return
	}

	data.IsGitRepo = integrations.IsRepo()
	if !data.IsGitRepo {
		render.InitNotAGitRepo()
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
			render.InitErrorCreatingDir(dir, err)
			os.Exit(1)
		}
	}

	dizzConfig := defaults.DefaultConfig(projectName)
	cfgData, _ := json.MarshalIndent(dizzConfig, "", "  ")
	if err := os.WriteFile(dizzConfigPath, cfgData, 0644); err != nil {
		render.InitErrorWritingConfig(err)
		os.Exit(1)
	}

	gitignorePath := filepath.Join(trackDir, ".gitignore")
	gitignoreContent := defaults.GitignoreContent()
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		render.InitErrorWritingGitignore(err)
		os.Exit(1)
	}

	createProjectSkill(cwd, projectName)
	render.InitSummary(data)

	if data.IsGitRepo {
		hooksDir := config.HooksDirPath(cwd)
		hookPath := filepath.Join(hooksDir, "post-commit")
		hookContent := defaults.LocalPostCommitHookContent(config.AppName)

		if err := integrations.InstallLocalPostCommitHook(hookPath, hookContent); err != nil {
			render.InitGitHookError(err)
		} else {
			render.InitGitHookOK()
		}

		if err := integrations.SetLocalHooksPath(); err != nil {
			render.InitGitHooksPathError(err)
		} else {
			render.InitGitHooksPathOK()
		}
	}
}

func createProjectSkill(cwd, projectName string) {
	skillDir := filepath.Join(cwd, ".agents", "skills", "dizz")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		render.InitSkillDirWarning(err)
		return
	}

	skillContent := defaults.SkillInstructions(projectName)
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		render.InitSkillWarning("could not write SKILL.md")
	}
}
