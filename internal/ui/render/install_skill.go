package render

import (
	"fmt"

	"github.com/TheShiveshNetwork/dizz/internal/skill"
)

func InstallSkillFetching() {
	fmt.Println("Fetching canonical SKILL.md...")
}

func InstallSkillFetchFallback() {
	fmt.Println("  Using bundled template instead.")
}

func InstallSkillDownloaded() {
	fmt.Println("  Downloaded from GitHub.")
}

func InstallSkillToProvider(provider string) {
	fmt.Printf("Installing to provider: %s\n", provider)
}

func InstallSkillDetecting() {
	fmt.Println("Detecting AI agent skill directories...")
}

func InstallSkillDetectedDirs(dirs []skill.AgentDir) {
	for _, d := range dirs {
		fmt.Printf("  Found: %s (%s)\n", d.Name, d.Path)
	}
	fmt.Println()
}

func InstallSkillInstalling() {
	fmt.Println("Installing to agent directories...")
}

func InstallSkillResults(results []skill.InstallResult) {
	successCount := 0
	for _, r := range results {
		if r.Err == nil && !r.Skipped {
			successCount++
		}
	}

	fmt.Print(skill.FormatInstallResults(results))

	if successCount == 0 {
		fmt.Println()
		fmt.Println("No skills were installed. See errors above.")
		return
	}

	fmt.Println()
	fmt.Printf("Installed dizz skill to %d agent director(ies).\n", successCount)
	fmt.Println()
	fmt.Println("Agents can now discover dizz automatically.")
	fmt.Println("Try: dizz context")
}

func InstallSkillSingleResults(results []skill.InstallResult, provider string) {
	successCount := 0
	for _, r := range results {
		if r.Err == nil && !r.Skipped {
			successCount++
		}
	}

	fmt.Print(skill.FormatInstallResults(results))

	if successCount == 0 {
		fmt.Println()
		fmt.Println("Skill was not installed. See errors above.")
		return
	}

	fmt.Println()
	fmt.Printf("Installed dizz skill to %s.\n", provider)
	fmt.Println("Agents can now discover dizz automatically.")
	fmt.Println("Try: dizz context")
}

func InstallSkillNoDirs() {
	fmt.Println("  No supported AI agent directories found on this system.")
	fmt.Println()
	fmt.Println("Install an AI agent first, then run this command again.")
	fmt.Println("Or use --provider to install to a specific agent:")
	fmt.Println("  dizz install-skill --provider opencode")
	fmt.Println("  dizz install-skill --provider cursor")
	fmt.Println()
	fmt.Println("Run 'dizz install-skill --help' to see all available providers.")
}
