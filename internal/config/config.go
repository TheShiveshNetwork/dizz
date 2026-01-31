package config

import "path/filepath"

const (
	AppName				= "dizz"

	TrackDirName	= ".dizz"
	ObjectsDirName= "objects"
	RefsDirName		= "refs"
	GitRefDirName	= "refs/git"

	ConfigFile		= "config.json"
	StateFile			= "state.json"
	HistoryDir		= "history"

	DefaultBranch	= "main"
)

type Config struct {
	ProjectName string   `json:"project_name"`
	RootPath    string   `json:"root_path"`
	Include     []string `json:"include"`
	Exclude     []string `json:"exclude"`
}

func TrackDirPath(root string) string {
	return filepath.Join(root, TrackDirName)
}

func ObjectsDirPath(root string) string {
	return filepath.Join(root, ObjectsDirName)
}

func RefsDirPath(root string) string {
	return filepath.Join(root, RefsDirName)
}

func ConfigFilePath(root string) string {
	return filepath.Join(root, ConfigFile)
}

func StateFilePath(root string) string {
	return filepath.Join(root, StateFile)
}

func HistoryDirPath(root string) string {
	return filepath.Join(root, HistoryDir)
}

// return a sensible default configuration usable in any project
func DefaultConfig(projectName string) *Config {
	return &Config{
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

