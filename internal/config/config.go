package config

import "path/filepath"

const (
	AppName = "dizz"

	TrackDirName   = ".dizz"
	ObjectsDirName = "objects"
	RefsDirName    = "refs"
	GitRefDirName  = "refs/git"
	HooksDirName   = "hooks"
	CacheDirName   = "cache"

	ConfigFile = "config.json"
	StateFile  = "state.json.gz"
	IntentFile = "intent.json"

	DefaultBranch = "main"
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
	return filepath.Join(root, TrackDirName, ObjectsDirName)
}

func RefsDirPath(root string) string {
	return filepath.Join(root, TrackDirName, RefsDirName)
}

func ConfigFilePath(root string) string {
	return filepath.Join(root, ConfigFile)
}

func StateFilePath(root string) string {
	return filepath.Join(root, StateFile)
}

func IntentFilePath(root string) string {
	return filepath.Join(root, IntentFile)
}

func HooksDirPath(root string) string {
	return filepath.Join(root, TrackDirName, HooksDirName)
}

func CacheDirPath(root string) string {
	return filepath.Join(root, TrackDirName, CacheDirName)
}

func CacheManifestPath(root string) string {
	return filepath.Join(root, TrackDirName, CacheDirName, "manifest.json")
}

func CacheSignalsDirPath(root string) string {
	return filepath.Join(root, TrackDirName, CacheDirName, "signals")
}
