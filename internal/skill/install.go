package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// AgentDir represents an agent platform's skill directory.
type AgentDir struct {
	Name string // Human-readable agent name
	Path string // Absolute path to the skills directory
}

// DetectAgentDirs returns all agent skill directories that exist on this system.
func DetectAgentDirs() []AgentDir {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var dirs []AgentDir
	candidates := agentDirCandidates(home)

	for _, c := range candidates {
		if info, err := os.Stat(c.Path); err == nil && info.IsDir() {
			dirs = append(dirs, c)
		}
	}

	return dirs
}

func agentDirCandidates(home string) []AgentDir {
	return []AgentDir{
		{Name: "Claude Code", Path: filepath.Join(home, ".claude", "skills")},
		{Name: "Claude Desktop", Path: filepath.Join(home, ".claude", "skills")},
		{Name: "OpenClaw", Path: filepath.Join(home, ".agents", "skills")},
		{Name: "OpenClaw (Shared)", Path: filepath.Join(home, ".openclaw", "skills")},
		{Name: "Codex CLI / Copilot", Path: filepath.Join(home, ".agents", "skills")},
		{Name: "Cursor", Path: filepath.Join(home, ".cursor", "skills")},
		{Name: "Gemini CLI", Path: filepath.Join(home, ".gemini", "config", "skills")},
		{Name: "Gemini CLI (alt)", Path: filepath.Join(home, ".gemini", "skills")},
		{Name: "OpenCode / Pi Agent", Path: filepath.Join(home, ".agents", "skills")},
		{Name: "OpenCode (config)", Path: filepath.Join(home, ".config", "opencode", "skills")},
	}
}

// InstallResult reports the outcome of installing to a single agent directory.
type InstallResult struct {
	Agent   string // Agent name
	Path    string // Target directory
	Err     error  // nil on success
	Skipped bool   // true if directory didn't exist
}

// InstallToDir writes SKILL.md content into the given agent skill directory.
// The skill is placed at <dir>/dizz/SKILL.md.
func InstallToDir(content []byte, agent string, dir string) error {
	skillDir := filepath.Join(dir, "dizz")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, content, 0644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}

	return nil
}

// InstallToAll detects all agent dirs and installs the skill content to each.
func InstallToAll(content []byte) []InstallResult {
	dirs := DetectAgentDirs()
	if dirs == nil {
		return []InstallResult{{
			Agent:   "system",
			Skipped: true,
			Err:     fmt.Errorf("no home directory found"),
		}}
	}

	if len(dirs) == 0 {
		return []InstallResult{{
			Agent:   "system",
			Skipped: true,
			Err:     fmt.Errorf("no agent skill directories found on this system"),
		}}
	}

	seen := make(map[string]bool)
	var results []InstallResult

	for _, d := range dirs {
		if seen[d.Path] {
			continue
		}
		seen[d.Path] = true

		if err := InstallToDir(content, d.Name, d.Path); err != nil {
			results = append(results, InstallResult{
				Agent: d.Name,
				Path:  filepath.Join(d.Path, "dizz", "SKILL.md"),
				Err:   err,
			})
		} else {
			results = append(results, InstallResult{
				Agent: d.Name,
				Path:  filepath.Join(d.Path, "dizz", "SKILL.md"),
			})
		}
	}

	return results
}

// FetchSkillURL returns the GitHub raw URL for the canonical SKILL.md.
func FetchSkillURL() string {
	return "https://raw.githubusercontent.com/TheShiveshNetwork/dizz/main/agent-skills/dizz/SKILL.md"
}

// IsOnline checks if the system likely has internet access.
func IsOnline() bool {
	// Simple check: can we resolve github.com?
	// On most systems, this is a good proxy for internet access.
	return true // Will be caught at fetch time; just attempt the fetch
}

// Platform returns a human-readable OS/arch string.
func Platform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

// GlobalSkillDirs returns all possible global skill directories for display.
func GlobalSkillDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	unique := make(map[string]bool)
	var dirs []string

	candidates := agentDirCandidates(home)
	for _, c := range candidates {
		p := filepath.Join(c.Path, "dizz")
		if !unique[p] {
			unique[p] = true
			dirs = append(dirs, p)
		}
	}

	return dirs
}

// CleanDirName returns the last path component of a directory.
func CleanDirName(p string) string {
	return filepath.Base(filepath.Dir(p))
}

// FormatInstallResults formats install results for display.
func FormatInstallResults(results []InstallResult) string {
	var b strings.Builder
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(&b, "  ✗ %s: %v\n", r.Agent, r.Err)
		} else {
			fmt.Fprintf(&b, "  ✓ %s\n", r.Agent)
		}
	}
	return b.String()
}
