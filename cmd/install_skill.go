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
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
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

	render.InstallSkillFetching()
	content, err := fetchGlobalSkillContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not fetch SKILL.md: %v\n", err)
		render.InstallSkillFetchFallback()
		content = []byte(defaults.GlobalSkillInstructions())
	} else {
		render.InstallSkillDownloaded()
	}

	if provider != "" {
		return installToSingleProvider(content, provider)
	}

	return installToAllProviders(content)
}

func installToSingleProvider(content []byte, provider string) error {
	provider = strings.ToLower(provider)
	render.InstallSkillToProvider(provider)

	results, err := skill.InstallToProvider(content, provider)
	if err != nil {
		return err
	}

	render.InstallSkillSingleResults(results, provider)
	return nil
}

func installToAllProviders(content []byte) error {
	render.InstallSkillDetecting()

	dirs := skill.DetectAgentDirs()
	if len(dirs) == 0 {
		render.InstallSkillNoDirs()
		return nil
	}

	render.InstallSkillDetectedDirs(dirs)

	render.InstallSkillInstalling()
	results := skill.InstallToAll(content)

	render.InstallSkillResults(results)
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
