package defaults

import (
	"github.com/TheShiveshNetwork/dizz/internal/config"
)

// return a sensible default configuration usable in any project
func DefaultConfig(projectName string) *config.Config {
	return &config.Config{
		ProjectName: projectName,
		RootPath:    ".",
		// Include is nil by default — CodeFilesWithIncludes falls back to the
		// language registry's extension list, so only known source-code files
		// are discovered.  Users can override with explicit patterns in their
		// .dizz/config.json if they need to analyze non-registry languages.
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
	content := "# dizz Skill for " + projectName + "\n\n"
	content += "dizz is a state-aware project assistant. It tracks intents, code health, and project state.\n\n"
	content += "## Quick Start\n\n"
	content += "Run `dizz context` to get a token-optimized project summary.\n\n"
	content += "## Available Commands\n\n"
	content += "- `dizz context` - Token-optimized project context for agents\n"
	content += "- `dizz intent list` - View active intents (what needs doing)\n"
	content += "- `dizz status` - Project health overview\n"
	content += "- `dizz log` - Detailed symbol health and todos\n"
	content += "- `dizz snapshot --auto` - Record current state before making changes\n"
	content += "- `dizz intent add \"msg\" --type todo` - Add a new intent\n\n"
	content += "## Data Format\n\n"
	content += "All intent data is stored in `.dizz/intent.ton` (Token-Optimized Notation) - a line-oriented,\n"
	content += "pipe-delimited format readable by any agent without a parser.\n"
	return content
}

func GlobalSkillInstructions() string {
	content := "# dizz Global Skill\n\n"
	content += "dizz is a CLI tool that tracks project state, intents, and code health.\n"
	content += "It is available system-wide after installation.\n\n"
	content += "## Usage\n\n"
	content += "To use dizz in any project, run `dizz init` in the project root.\n"
	content += "This creates the `.dizz/` directory with intent tracking and analysis.\n\n"
	content += "## Primary Agent Command\n\n"
	content += "Run `dizz context` inside any initialized project to get a token-optimized summary of:\n"
	content += "- Active intents (goals, tasks, questions)\n"
	content += "- Symbol health (active, unstable, unused, abandoned)\n"
	content += "- TODOs found in source code\n"
	content += "- Git context (branch, commit)\n\n"
	content += "The output is in TON format - pipe-delimited, one record per line.\n"
	content += "No parser needed: split on `|` to read.\n"
	return content
}

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
