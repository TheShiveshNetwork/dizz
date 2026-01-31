package config

import "path/filepath"

const (
	AppName				= "dizz"

	TrackDirName	= ".dizz"

	ConfigFile		= "config.json"
	StateFile			= "state.json"
	HistoryDir		= "history"

	DefaultBranch	= "main"
)

func TrackDirPath(root string) string {
	return filepath.Join(root, TrackDirName)
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

