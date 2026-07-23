package defaults

import (
	"strings"

	"github.com/TheShiveshNetwork/dizz/config"
)

// return a sensible default configuration usable in any project
func DefaultConfig(projectName string) *config.Config {
	return &config.Config{
		Version:     config.ConfigVersion,
		ProjectName: projectName,
		Description: "",
		Include:     []string{"**/*"},
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

func SkillInstructions(projectName string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: dizz\n")
	b.WriteString("description: Project-level dizz skill. Tracks intents, code health, and symbol states for this specific project. Use to understand project state, find work items, or detect dead code.\n")
	b.WriteString("license: MIT\n")
	b.WriteString("metadata:\n")
	b.WriteString("  scope: project\n")
	b.WriteString("  project: " + projectName + "\n")
	b.WriteString("  version: \"1.0.0\"\n")
	b.WriteString("---\n")
	b.WriteString("\n")
	b.WriteString("# dizz Project Skill\n")
	b.WriteString("\n")
	b.WriteString("dizz is this project's memory layer. It tracks intents, code health, and symbol states -\n")
	b.WriteString("information that survives across sessions and would otherwise be lost between conversations.\n")
	b.WriteString("It is read-only and works in any repo.\n")
	b.WriteString("\n")
	b.WriteString("## Quick Start\n")
	b.WriteString("\n")
	b.WriteString("Run `dizz context` to get a token-optimized project summary. This is the primary entry point for agents.\n")
	b.WriteString("\n")
	b.WriteString("## Project Config\n")
	b.WriteString("\n")
	b.WriteString("This project's agent rules live in `.dizz/config.json`. Run `dizz config show` to read them.\n")
	b.WriteString("\n")
	b.WriteString("Key fields:\n")
	b.WriteString("\n")
	b.WriteString("- `description` - project summary\n")
	b.WriteString("- `instructions` - agent rules scoped to file patterns (e.g., `rule@scope`)\n")
	b.WriteString("- `guardrails` - protected paths with actions (e.g., `path|action|reason`)\n")
	b.WriteString("- `commands` - project-specific commands (build, test, lint, etc.)\n")
	b.WriteString("\n")
	b.WriteString("To add rules:\n")
	b.WriteString("\n")
	b.WriteString("```\n")
	b.WriteString("dizz config add-instruction --rule \"Run tests before merge\" --scope \"internal/**\"\n")
	b.WriteString("dizz config add-guardrail --path \"generated/**\" --action \"read_only\" --reason \"auto-generated\"\n")
	b.WriteString("dizz config set-description \"Project description here\"\n")
	b.WriteString("```\n")
	b.WriteString("\n")
	b.WriteString("## Commands\n")
	b.WriteString("\n")
	b.WriteString("- `dizz context` - Token-optimized project context for agents\n")
	b.WriteString("- `dizz config show` - Read project agent config\n")
	b.WriteString("- `dizz intent list` - View active intents (what needs doing)\n")
	b.WriteString("- `dizz status` - Project health overview (unstable, unused, abandoned symbols)\n")
	b.WriteString("- `dizz log` - Detailed symbol health and todos\n")
	b.WriteString("- `dizz snapshot --auto` - Record current state before making changes\n")
	b.WriteString("- `dizz intent add \"msg\" --type todo` - Add a new intent\n")
	b.WriteString("\n")
	b.WriteString("## Data Format\n")
	b.WriteString("\n")
	b.WriteString("All intent data is stored in `.dizz/intent.ton` (Token-Optimized Notation) -\n")
	b.WriteString("a line-oriented, pipe-delimited format readable by any agent without a parser.\n")
	b.WriteString("Split on `|` to read. No JSON parser needed.\n")
	return b.String()
}

// GlobalSkillInstructions returns the content for the global dizz skill.
// This is installed system-wide by `dizz install-skill`.
func GlobalSkillInstructions() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: dizz-global\n")
	b.WriteString("description: Global dizz skill for AI agents. Detects dizz projects, reads config-backed agent guidance, and uses dizz commands for project state. Install system-wide via `dizz install-skill`.\n")
	b.WriteString("license: MIT\n")
	b.WriteString("metadata:\n")
	b.WriteString("  scope: global\n")
	b.WriteString("  version: \"1.0.0\"\n")
	b.WriteString("---\n")
	b.WriteString("\n")
	b.WriteString("# dizz Global Skill\n")
	b.WriteString("\n")
	b.WriteString("dizz is a state-aware assistant for AI agents. It keeps track of project intents, code health, and symbol states -\n")
	b.WriteString("information that survives across sessions and would otherwise be lost between conversations.\n")
	b.WriteString("It is read-only and works in any repo.\n")
	b.WriteString("\n")
	b.WriteString("This skill applies system-wide across every project. Each time you enter a project, determine whether dizz is relevant there before acting.\n")
	b.WriteString("\n")
	b.WriteString("## Step 1: Check applicability\n")
	b.WriteString("\n")
	b.WriteString("- If dizz is not installed on this machine, note that and move on - do not block the current task over it.\n")
	b.WriteString("- If dizz is installed but this project has no `.dizz/` directory yet, ask the user before initializing it.\n")
	b.WriteString("- If `.dizz/` already exists, proceed directly - no need to ask.\n")
	b.WriteString("\n")
	b.WriteString("## Step 2: Read project agent rules from config\n")
	b.WriteString("\n")
	b.WriteString("Project-level agent standards live in `.dizz/config.json`. Treat this as the source of truth for agent runs in this repository.\n")
	b.WriteString("\n")
	b.WriteString("Run `dizz config show` to read the current config. Key fields:\n")
	b.WriteString("\n")
	b.WriteString("- `project_name` - the project identifier\n")
	b.WriteString("- `description` - human-readable project summary\n")
	b.WriteString("- `instructions` - agent rules scoped to file patterns\n")
	b.WriteString("- `guardrails` - protected paths with actions and reasons\n")
	b.WriteString("- `commands` - project-specific commands (build, test, lint, etc.)\n")
	b.WriteString("- `agent_defaults` - default analysis lens and severity thresholds\n")
	b.WriteString("\n")
	b.WriteString("To update config:\n")
	b.WriteString("\n")
	b.WriteString("```\n")
	b.WriteString("dizz config add-instruction --rule \"...\" --scope \"...\"\n")
	b.WriteString("dizz config add-guardrail --path \"...\" --action \"...\" --reason \"...\"\n")
	b.WriteString("dizz config set-description \"...\"\n")
	b.WriteString("```\n")
	b.WriteString("\n")
	b.WriteString("## Step 3: Use dizz commands, not this file\n")
	b.WriteString("\n")
	b.WriteString("This file does not list dizz commands or flags. The CLI changes after this was written, and a stale list leads to confident wrong invocations.\n")
	b.WriteString("\n")
	b.WriteString("- Run `dizz --help` to see everything available right now.\n")
	b.WriteString("- Run `dizz <command> --help` for specific flags.\n")
	b.WriteString("\n")
	b.WriteString("Both are instant and local. Check rather than guess.\n")
	b.WriteString("\n")
	b.WriteString("## When to reach for dizz\n")
	b.WriteString("\n")
	b.WriteString("Once this project has dizz set up (Step 1), use dizz whenever:\n")
	b.WriteString("\n")
	b.WriteString("- **Starting a session** - get current state before reading files or asking the user to recap.\n")
	b.WriteString("- **Before and after making changes** - checkpoint state so nothing gets lost.\n")
	b.WriteString("- **The user states a long-term goal or decision** - record it immediately so it outlives this conversation.\n")
	b.WriteString("- **The user asks what to work on or what changed** - answer from dizz output, not a guess.\n")
	b.WriteString("- **Returning after a gap** - recover context from dizz first.\n")
	b.WriteString("\n")
	b.WriteString("## Failure modes\n")
	b.WriteString("\n")
	b.WriteString("| Symptom | Cause | Fix |\n")
	b.WriteString("|---|---|---|\n")
	b.WriteString("| `command not found: dizz` | Not installed on this machine | Mention it, offer the install link, continue without it |\n")
	b.WriteString("| No dizz project detected | This project has not been initialized yet | Ask the user before running `dizz init` |\n")
	b.WriteString("| Output looks stale | Local state out of date | Take a fresh checkpoint with `dizz snapshot --auto` |\n")
	b.WriteString("\n")
	b.WriteString("## Constraints\n")
	b.WriteString("\n")
	b.WriteString("- **Read-only** - dizz never edits source code. Any actual fix or resolution is the agent's job.\n")
	b.WriteString("- **Zero-config** by default, once a project has been initialized.\n")
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
