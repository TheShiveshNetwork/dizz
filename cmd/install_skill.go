package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/internal/defaults"
	"github.com/TheShiveshNetwork/dizz/internal/skill"
)

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install dizz skill for AI agents on this system",
	Long: `Detects installed AI agents (Claude Code, Cursor, Gemini CLI,
OpenCode, Codex CLI / Copilot) and installs the dizz SKILL.md
into each agent's skill directory.

Agents use the skill to discover and invoke dizz for project
state, context, and intent tracking.`,
	RunE: runInstallSkill,
}

func init() {
	rootCmd.AddCommand(installSkillCmd)
}

func runInstallSkill(cmd *cobra.Command, args []string) error {
	fmt.Println("Detecting AI agent skill directories...")

	dirs := skill.DetectAgentDirs()
	if len(dirs) == 0 {
		fmt.Println("  No supported AI agent directories found on this system.")
		fmt.Println()
		fmt.Println("Install an AI agent first, then run this command again.")
		fmt.Println("  - Claude Code: https://docs.anthropic.com/en/docs/claude-code/overview")
		fmt.Println("  - Cursor: https://cursor.com")
		fmt.Println("  - Gemini CLI: https://google-gemini.github.io/gemini-cli/")
		fmt.Println()
		fmt.Println("You can also manually install the skill:")
		fmt.Println("  mkdir -p ~/.agents/skills/dizz-global")
		fmt.Println("  See agent-skills/dizz-global/SKILL.md for content")
		return nil
	}

	for _, d := range dirs {
		fmt.Printf("  Found: %s (%s)\n", d.Name, d.Path)
	}
	fmt.Println()

	fmt.Println("Fetching canonical SKILL.md...")
	content, err := fetchGlobalSkillContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not fetch SKILL.md: %v\n", err)
		fmt.Println("  Using bundled template instead.")
		content = []byte(defaults.GlobalSkillInstructions())
	} else {
		fmt.Println("  Downloaded from GitHub.")
	}

	fmt.Println("Installing to agent directories...")
	results := skill.InstallToAll(content)

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
		return nil
	}

	fmt.Println()
	fmt.Printf("Installed dizz skill to %d agent director(ies).\n", successCount)
	fmt.Println()
	fmt.Println("Agents can now discover dizz automatically.")
	fmt.Println("Try: dizz context")
	fmt.Println()

	return nil
}

func fetchGlobalSkillContent() ([]byte, error) {
	url := skill.FetchGlobalSkillURL()
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty response body")
	}

	return data, nil
}
