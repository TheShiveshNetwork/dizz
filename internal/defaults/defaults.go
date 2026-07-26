package defaults

//go:generate mkdir -p agent-skills
//go:generate cp -r ../../agent-skills/dizz agent-skills/dizz
//go:generate cp -r ../../agent-skills/dizz-global agent-skills/dizz-global

import (
	_ "embed"

	"github.com/TheShiveshNetwork/dizz/config"
)

//go:embed agent-skills/dizz/SKILL.md
var projectSkillContent []byte

//go:embed agent-skills/dizz-global/SKILL.md
var globalSkillContent []byte

// DefaultConfig returns a sensible default configuration usable in any project.
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

// SkillInstructions returns the content for the project-level dizz skill.
func SkillInstructions(_ string) []byte {
	return projectSkillContent
}

// GlobalSkillInstructions returns the content for the global dizz skill.
func GlobalSkillInstructions() []byte {
	return globalSkillContent
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
