package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/internal/defaults"
	"github.com/TheShiveshNetwork/dizz/internal/skill"
)

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install dizz skill for AI agents on this system",
	Long: `Detects installed AI agents and installs the dizz SKILL.md
into each agent's skill directory.

Available providers:
  agents          ~/.agents/skills/ (tool-agnostic standard)
  claude-code     ~/.claude/skills/
  claude-desktop  ~/.claude/skills/
  copilot         ~/.copilot/skills/ (VS Code)
  cursor          ~/.cursor/skills/
  gemini-cli      ~/.gemini/skills/
  gemini-config   ~/.gemini/config/skills/
  antigravity     ~/.gemini/config/skills/ (Google Antigravity)
  opencode        ~/.config/opencode/skills/

Use --provider to install to a specific agent only.
Without --provider, installs to all detected agents.`,
	RunE: runInstallSkill,
}

func init() {
	rootCmd.AddCommand(installSkillCmd)
	installSkillCmd.Flags().StringP("provider", "p", "", "install to a specific provider only (e.g. opencode, cursor)")
}

func runInstallSkill(cmd *cobra.Command, args []string) error {
	provider, _ := cmd.Flags().GetString("provider")

	fmt.Println("Fetching canonical SKILL.md...")
	content, err := fetchGlobalSkillContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not fetch SKILL.md: %v\n", err)
		fmt.Println("  Using bundled template instead.")
		content = []byte(defaults.GlobalSkillInstructions())
	} else {
		fmt.Println("  Downloaded from GitHub.")
	}

	if provider != "" {
		return installToSingleProvider(content, provider)
	}

	return installToAllProviders(content)
}

func installToSingleProvider(content []byte, provider string) error {
	provider = strings.ToLower(provider)
	fmt.Printf("Installing to provider: %s\n", provider)

	results, err := skill.InstallToProvider(content, provider)
	if err != nil {
		return err
	}

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
		return nil
	}

	fmt.Println()
	fmt.Printf("Installed dizz skill to %s.\n", provider)
	fmt.Println("Agents can now discover dizz automatically.")
	fmt.Println("Try: dizz context")
	return nil
}

func installToAllProviders(content []byte) error {
	fmt.Println("Detecting AI agent skill directories...")

	dirs := skill.DetectAgentDirs()
	if len(dirs) == 0 {
		fmt.Println("  No supported AI agent directories found on this system.")
		fmt.Println()
		fmt.Println("Install an AI agent first, then run this command again.")
		fmt.Println("Or use --provider to install to a specific agent:")
		fmt.Println("  dizz install-skill --provider opencode")
		fmt.Println("  dizz install-skill --provider cursor")
		fmt.Println()
		fmt.Println("Run 'dizz install-skill --help' to see all available providers.")
		return nil
	}

	for _, d := range dirs {
		fmt.Printf("  Found: %s (%s)\n", d.Name, d.Path)
	}
	fmt.Println()

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
