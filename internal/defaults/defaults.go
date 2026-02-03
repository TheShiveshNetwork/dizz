package defaults

import (
	"github.com/TheShiveshNetwork/dizz/internal/config"
)

// return a sensible default configuration usable in any project
func DefaultConfig(projectName string) *config.Config {
	return &config.Config{
		ProjectName: projectName,
		RootPath:    ".",
		Include:     []string{"**/*"},
		Exclude: []string{
			"vendor/**",
			"node_modules/**",
			".git/**",
			".dizz/**",
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
!.gitignore`
}

func GitPostCommitHookContent(appName string) string {
	return `#!/usr/bin/env sh

# Detect repo remote (best-effort, silent failure)
REMOTE_URL="$(git config --get remote.origin.url 2>/dev/null || true)"

USE_LOCAL=false
case "$REMOTE_URL" in
    *TheShiveshNetwork/dizz*)
        USE_LOCAL=true
        ;;
esac

if [ "$USE_LOCAL" = "true" ] && [ -x "$(pwd)/` + appName + `" ]; then
    # Use local dev binary ONLY for dizz repo
    DIZZ_BIN="$(pwd)/` + appName + `"
else
    # Fallback to system-installed dizz
    DIZZ_BIN="$(command -v ` + appName + ` 2>/dev/null || true)"
fi

if [ -n "$DIZZ_BIN" ] && [ -x "$DIZZ_BIN" ]; then
    "$DIZZ_BIN" snapshot --auto >/dev/null 2>&1 || true
fi
`
}

