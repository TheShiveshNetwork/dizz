package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/config"
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
		fmt.Printf("  mkdir -p ~/.agents/skills/dizz\n")
		fmt.Printf("  %s context > ~/.agents/skills/dizz/SKILL.md\n", config.AppName)
		return nil
	}

	for _, d := range dirs {
		fmt.Printf("  Found: %s (%s)\n", d.Name, d.Path)
	}
	fmt.Println()

	fmt.Println("Fetching canonical SKILL.md...")
	content, err := fetchSkillContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not fetch SKILL.md: %v\n", err)
		fmt.Println("  Using bundled template instead.")
		content = []byte(bundledSkillContent())
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

func fetchSkillContent() ([]byte, error) {
	url := skill.FetchSkillURL()
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

func bundledSkillContent() string {
	projectName := "project"

	cwd, err := os.Getwd()
	if err == nil {
		trackDir := config.TrackDirPath(cwd)
		cfgPath := config.ConfigFilePath(trackDir)
		if data, err := os.ReadFile(cfgPath); err == nil {
			s := string(data)
			prefix := `"project_name":"`
			if start := strings.Index(s, prefix); start >= 0 {
				start += len(prefix)
				if end := strings.Index(s[start:], `"`); end >= 0 {
					projectName = s[start : start+end]
				}
			}
		}
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: dizz\n")
	b.WriteString("description: State-aware project assistant. Tracks intents, code health, and symbol states. Use to understand project state, find work items, or detect dead code.\n")
	b.WriteString("license: MIT\n")
	b.WriteString("metadata:\n")
	b.WriteString("  version: \"1.0.0\"\n")
	b.WriteString("---\n")
	b.WriteString("\n")
	b.WriteString("# dizz Skill for " + projectName + "\n")
	b.WriteString("\n")
	b.WriteString("dizz analyzes your project to show what is used, unused, and planned.\n")
	b.WriteString("\n")
	b.WriteString("## Commands\n")
	b.WriteString("\n")
	b.WriteString("- dizz context - Token-optimized project context\n")
	b.WriteString("- dizz intent list - View active intents\n")
	b.WriteString("- dizz status - Project health overview\n")
	b.WriteString("- dizz log - Symbol health and todos\n")
	b.WriteString("- dizz snapshot --auto - Record current state\n")
	b.WriteString("- dizz intent add \"msg\" --type todo - Add intent\n")
	b.WriteString("\n")
	b.WriteString("## First use\n")
	b.WriteString("\n")
	b.WriteString("Run dizz context to get started.\n")
	return b.String()
}
