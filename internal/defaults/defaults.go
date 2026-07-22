package defaults

import (
	"strings"

	"github.com/TheShiveshNetwork/dizz/config"
)

// return a sensible default configuration usable in any project
func DefaultConfig(projectName string) *config.Config {
	return &config.Config{
		Version:       1,
		ProjectName:   projectName,
		Description:   "",
		RootPath:      ".",
		Include:       []string{"**/*"},
		Exclude: []string{
			// Directories
			"vendor/**",
			"node_modules/**",
			".git/**",
			".dizz/**",

			// Config, data, and env files (no analysis value)
			"**/*.json",
			"**/*.yaml",
			"**/*.yml",
			"**/*.toml",
			"**/*.xml",
			"**/*.config",
			"**/*.conf",
			"**/*.cfg",
			"**/*.ini",
			"**/*.env",
			"**/*.env.*",

			// Documentation and text
			"**/*.md",
			"**/*.txt",
			"**/*.rst",

			// Build and lock files
			"**/*.lock",
			"**/Makefile",
			"**/Dockerfile",

			// Logs and temp
			"**/*.log",
			"**/*.tmp",
			"**/*.swp",

			// OS artifacts
			"**/.DS_Store",
			"**/Thumbs.db",
		},
	}
}

func GitignoreContent() string {
	return `# Ignore everything by default
*

# Keep config
!config.json
!.gitignore

# Keep hooks (tracked for post-commit snapshot)
!hooks/
!hooks/**`
}

// LocalPostCommitHookContent returns the content for a local post-commit hook.
func LocalPostCommitHookContent(appName string) string {
	return `#!/usr/bin/env sh

DIZZ_BIN="$(command -v ` + appName + ` 2>/dev/null || true)"

if [ -n "$DIZZ_BIN" ] && [ -x "$DIZZ_BIN" ]; then
    "$DIZZ_BIN" snapshot --auto >/dev/null 2>&1 || true
fi
`
}

// GitPostCommitHookContent is a legacy alias kept for backward compatibility.
// New code should call LocalPostCommitHookContent instead.
var GitPostCommitHookContent = LocalPostCommitHookContent

func SkillMetadata(projectName string) map[string]interface{} {
	return map[string]interface{}{
		"name":        "dizz",
		"version":     "1.0.0",
		"description": "State-aware project assistant for agents",
		"project":     projectName,
		"global":      false,
		"commands": []map[string]string{
			{"name": "context", "description": "Token-optimized project context", "command": "dizz context"},
			{"name": "intents", "description": "List active project intents", "command": "dizz intent list"},
			{"name": "intent-add", "description": "Add a new intent", "command": `dizz intent add "message" --type todo --severity 2`},
			{"name": "status", "description": "Project health overview", "command": "dizz status"},
			{"name": "snapshot", "description": "Record current project state", "command": "dizz snapshot --auto"},
			{"name": "log", "description": "Symbol health details", "command": "dizz log"},
		},
	}
}

// GlobalSkillMetadata returns the metadata for the global dizz skill.
func GlobalSkillMetadata() map[string]interface{} {
	return map[string]interface{}{
		"name":        "dizz",
		"version":     "1.0.0",
		"description": "State-aware project assistant for AI agents",
		"global":      true,
		"commands": []map[string]string{
			{"name": "context", "description": "Token-optimized project context dump", "command": "dizz context"},
			{"name": "intents", "description": "View active project intents", "command": "dizz intent list"},
			{"name": "status", "description": "Project health overview", "command": "dizz status"},
			{"name": "snapshot", "description": "Record current project state", "command": "dizz snapshot --auto"},
			{"name": "log", "description": "Symbol health and todos", "command": "dizz log"},
		},
	}
}

func SkillInstructions(projectName string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: dizz\n")
	b.WriteString("description: State-aware project assistant. Tracks intents, code health, and symbol states. Use to understand project state, find work items, or detect dead code.\n")
	b.WriteString("license: MIT\n")
	b.WriteString("metadata:\n")
	b.WriteString("  project: " + projectName + "\n")
	b.WriteString("  version: \"1.0.0\"\n")
	b.WriteString("---\n")
	b.WriteString("\n")
	b.WriteString("# dizz Skill for " + projectName + "\n")
	b.WriteString("\n")
	b.WriteString("dizz tracks project intents, code health, and symbol states.\n")
	b.WriteString("\n")
	b.WriteString("## Quick Start\n")
	b.WriteString("\n")
	b.WriteString("Run dizz context to get a token-optimized project summary.\n")
	b.WriteString("\n")
	b.WriteString("## Commands\n")
	b.WriteString("\n")
	b.WriteString("- dizz context - Token-optimized project context for agents\n")
	b.WriteString("- dizz intent list - View active intents (what needs doing)\n")
	b.WriteString("- dizz status - Project health overview (unstable, unused, abandoned symbols)\n")
	b.WriteString("- dizz log - Detailed symbol health and todos\n")
	b.WriteString("- dizz snapshot --auto - Record current state before making changes\n")
	b.WriteString("- dizz intent add \"msg\" --type todo - Add a new intent\n")
	b.WriteString("\n")
	b.WriteString("## Data Format\n")
	b.WriteString("\n")
	b.WriteString("All intent data is stored in .dizz/intent.ton (Token-Optimized Notation) -\n")
	b.WriteString("a line-oriented, pipe-delimited format readable by any agent without a parser.\n")
	b.WriteString("Split on | to read. No JSON parser needed.\n")
	return b.String()
}

// GlobalSkillInstructions returns the instructions for the global dizz skill.
func GlobalSkillInstructions() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: dizz\n")
	b.WriteString("description: State-aware project assistant for AI agents. Tracks project state, intents, code health, and symbol states. Use when you need to understand project state, find work items, or detect dead code.\n")
	b.WriteString("license: MIT\n")
	b.WriteString("compatibility: Designed for AI agents (Claude Code, OpenClaw, Cursor, Gemini CLI, OpenCode, Codex CLI / Copilot)\n")
	b.WriteString("metadata:\n")
	b.WriteString("  version: \"1.0.0\"\n")
	b.WriteString("---\n")
	b.WriteString("\n")
	b.WriteString("# dizz Global Skill\n")
	b.WriteString("\n")
	b.WriteString("dizz tracks project state, intents, and code health.\n")
	b.WriteString("Available system-wide after installation.\n")
	b.WriteString("\n")
	b.WriteString("## Usage\n")
	b.WriteString("\n")
	b.WriteString("To use dizz in any project, run dizz init in the project root.\n")
	b.WriteString("This creates the .dizz/ directory with intent tracking and analysis.\n")
	b.WriteString("\n")
	b.WriteString("## Primary Agent Command\n")
	b.WriteString("\n")
	b.WriteString("Run dizz context inside any initialized project to get a token-optimized summary of:\n")
	b.WriteString("- Active intents (goals, tasks, questions)\n")
	b.WriteString("- Symbol health (active, unstable, unused, abandoned)\n")
	b.WriteString("- TODOs found in source code\n")
	b.WriteString("- Git context (branch, commit)\n")
	b.WriteString("\n")
	b.WriteString("The output is in TON format - pipe-delimited, one record per line.\n")
	b.WriteString("No parser needed: split on | to read.\n")
	return b.String()
}

// GlobalRouterHookContent returns the content for the global router hook.
func GlobalRouterHookContent() string {
	return `#!/usr/bin/env sh

# dizz global router hook
# This hook runs on every commit in every repo on this machine.
# It checks if the current repo has dizz hooks and delegates to them.

DIZZ_HOOKS=".dizz/hooks/post-commit"

if [ -f "$DIZZ_HOOKS" ] && [ -x "$DIZZ_HOOKS" ]; then
    # Configure local hooks path so future commits bypass the router
    git config core.hooksPath ".dizz/hooks" 2>/dev/null || true
    exec "$DIZZ_HOOKS"
	fi
`
}