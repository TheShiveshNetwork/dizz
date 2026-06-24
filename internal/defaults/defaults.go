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
!intent.json
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
