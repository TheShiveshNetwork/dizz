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

	ConfigFile     = "config.json"
	StateFile      = "state.json.gz"
	StateTONFile   = "state.ton.gz"
	ContextTONFile = "context.ton"
	IntentFile     = "intent.json"
	DeltaFileExt   = ".delta"

	DefaultBranch = "main"
)

// Config represents the dizz configuration structure.
type Config struct {
	Version         int    `json:"version"`
	ProjectName     string `json:"project_name"`
	Description     string `json:"description"`
	Include         []string `json:"include"`
	Exclude         []string `json:"exclude"`
	Commands        map[string]string `json:"commands,omitempty"`
	EntryPoints     map[string]string `json:"entry_points,omitempty"`
	Conventions     []Convention `json:"conventions,omitempty"`
	Guardrails      []Guardrail `json:"guardrails,omitempty"`
	SeverityScale   map[string]string `json:"severity_scale,omitempty"`
	AgentDefaults   AgentDefaults `json:"agent_defaults,omitempty"`
	Links           Links         `json:"links,omitempty"`
}

// Convention represents a coding convention.
type Convention struct {
	Rule   string `json:"rule"`
	Scope  string `json:"scope"`
}

// Guardrail represents a guardrail rule.
type Guardrail struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// AgentDefaults represents default settings for agents.
type AgentDefaults struct {
	DefaultLens string `json:"default_lens"`
	MinSeverity int    `json:"min_severity"`
}

// Links represents external links.
type Links struct {
	Contributing string `json:"contributing"`
	Docs         string `json:"docs"`
}

// TrackDirPath returns the path to the .dizz directory.
func TrackDirPath(root string) string {
	return filepath.Join(root, TrackDirName)
}

// ObjectsDirPath returns the path to the objects directory.
func ObjectsDirPath(root string) string {
	return filepath.Join(root, TrackDirName, ObjectsDirName)
}

// RefsDirPath returns the path to the refs directory.
func RefsDirPath(root string) string {
	return filepath.Join(root, TrackDirName, RefsDirName)
}

// ConfigFilePath returns the path to the config.json file.
func ConfigFilePath(root string) string {
	return filepath.Join(root, ConfigFile)
}

// StateFilePath returns the path to the state.json.gz file.
func StateFilePath(root string) string {
	return filepath.Join(root, StateFile)
}

// StateTONFilePath returns the path to the state.ton.gz file.
func StateTONFilePath(root string) string {
	return filepath.Join(root, StateTONFile)
}

// ContextTONFilePath returns the path to the context.ton file.
func ContextTONFilePath(root string) string {
	return filepath.Join(root, ContextTONFile)
}

// IntentFilePath returns the path to the intent.json file.
func IntentFilePath(root string) string {
	return filepath.Join(root, IntentFile)
}

// HooksDirPath returns the path to the hooks directory.
func HooksDirPath(root string) string {
	return filepath.Join(root, TrackDirName, HooksDirName)
}

// CacheDirPath returns the path to the cache directory.
func CacheDirPath(root string) string {
	return filepath.Join(root, TrackDirName, CacheDirName)
}

// CacheManifestPath returns the path to the cache manifest.json file.
func CacheManifestPath(root string) string {
	return filepath.Join(root, TrackDirName, CacheDirName, "manifest.json")
}

// CacheSignalsDirPath returns the path to the cache signals directory.
func CacheSignalsDirPath(root string) string {
	return filepath.Join(root, TrackDirName, CacheDirName, "signals")
}

// DeltaFilePath returns the path to a delta file.
func DeltaFilePath(root string, hash string) string {
	return filepath.Join(root, TrackDirName, ObjectsDirName, hash[:2], hash[2:]+DeltaFileExt)
}
