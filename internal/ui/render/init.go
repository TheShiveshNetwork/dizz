package render

import (
	"fmt"
	"os"

	"github.com/TheShiveshNetwork/dizz/config"
)

type InitData struct {
	ProjectName string
	TrackDir    string
	IsGitRepo   bool
}

func InitAlreadyInitialized(data *InitData) {
	fmt.Println("dizz already initialized")
	fmt.Printf("  Project: %s\n", data.ProjectName)
	fmt.Printf("  Path: %s\n", data.TrackDir)
}

func InitNotAGitRepo() {
	fmt.Println("Not a git repository")
	fmt.Println("   Run 'git init' first, or continue without git integration")
	fmt.Print("\nContinue anyway? [y/N]: ")
}

func InitErrorCreatingDir(dir string, err error) {
	fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
}

func InitErrorWritingConfig(err error) {
	fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
}

func InitErrorWritingGitignore(err error) {
	fmt.Fprintf(os.Stderr, "Error writing gitignore file: %v\n", err)
}

func InitGitHookError(err error) {
	fmt.Printf("Could not install post-commit hook: %v\n", err)
	fmt.Println("   You can still use dizz manually with 'dizz snapshot'")
}

func InitGitHookOK() {
	fmt.Println("Installed post-commit hook to .dizz/hooks/")
}

func InitGitHooksPathError(err error) {
	fmt.Printf("Could not set hooks path: %v\n", err)
}

func InitGitHooksPathOK() {
	fmt.Println("Configured git to use hooks from .dizz/hooks/")
}

func InitSummary(data *InitData) {
	fmt.Printf("Initialized %s\n", data.ProjectName)
	fmt.Printf("  Created %s/\n", data.TrackDir)
	if data.IsGitRepo {
		fmt.Println("  Git integration: enabled")
	} else {
		fmt.Println("  Git integration: disabled")
	}
	fmt.Printf("  Project: %s\n", data.ProjectName)
	fmt.Printf("  Agent skill: .agents/skills/dizz/\n")
	fmt.Printf("\nNext: Run '%s log' to see your project state\n", config.AppName)
	fmt.Printf("      Run '%s context' for agent-optimized project context\n", config.AppName)
}

func InitSkillWarning(msg string) {
	fmt.Printf("Warning: could not write SKILL.md: %s\n", msg)
}

func InitSkillDirWarning(err error) {
	fmt.Printf("Warning: could not create skill directory: %v\n", err)
}
