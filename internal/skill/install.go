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
		// Tool-agnostic standard location
		{Name: "Agents (standard)", Path: filepath.Join(home, ".agents", "skills")},

		// Claude
		{Name: "Claude Code", Path: filepath.Join(home, ".claude", "skills")},
		{Name: "Claude Desktop", Path: filepath.Join(home, ".claude", "skills")},

		// GitHub Copilot (VS Code)
		{Name: "Copilot (VS Code)", Path: filepath.Join(home, ".copilot", "skills")},

		// Cursor
		{Name: "Cursor", Path: filepath.Join(home, ".cursor", "skills")},

		// Gemini
		{Name: "Gemini CLI", Path: filepath.Join(home, ".gemini", "skills")},
		{Name: "Gemini CLI (config)", Path: filepath.Join(home, ".gemini", "config", "skills")},
		{Name: "Google Antigravity", Path: filepath.Join(home, ".gemini", "config", "skills")},

		// OpenCode
		{Name: "OpenCode", Path: filepath.Join(home, ".config", "opencode", "skills")},
		{Name: "OpenCode (agents)", Path: filepath.Join(home, ".agents", "skills")},
	}
}

// ProviderID returns a lowercase, hyphen-free identifier for a provider name.
func ProviderID(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

// InstallToProvider installs the skill to a specific provider by name.
// Returns an error if the provider is not found or the directory doesn't exist.
func InstallToProvider(content []byte, provider string) ([]InstallResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("no home directory found")
	}

	targetID := strings.ToLower(strings.ReplaceAll(provider, " ", "-"))
	var matched []AgentDir

	for _, c := range agentDirCandidates(home) {
		if ProviderID(c.Name) == targetID {
			matched = append(matched, c)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("unknown provider %q - run with --help to see available providers", provider)
	}

	var results []InstallResult
	for _, d := range matched {
		if info, err := os.Stat(d.Path); err != nil || !info.IsDir() {
			results = append(results, InstallResult{
				Agent:   d.Name,
				Path:    filepath.Join(d.Path, "dizz", "SKILL.md"),
				Skipped: true,
			})
			continue
		}

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

	return results, nil
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

// FetchGlobalSkillURL returns the GitHub raw URL for the global SKILL.md.
func FetchGlobalSkillURL() string {
	return "https://raw.githubusercontent.com/TheShiveshNetwork/dizz/main/agent-skills/dizz-global/SKILL.md"
}

// @dizz-ignore-unused
// Platform returns a human-readable OS/arch string.
func Platform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

// @dizz-ignore-unused
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

// @dizz-ignore-unused
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
