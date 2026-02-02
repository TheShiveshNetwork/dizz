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
