package config

import (
	"encoding/json"
	"path/filepath"
)

const (
	AppName       = "dizz"
	ConfigVersion = "1.0.0"

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

	DefaultBranch = "main"
)

// Config represents the dizz configuration structure.
type Config struct {
	Version       string            `json:"version"`
	ProjectName   string            `json:"project_name"`
	Description   string            `json:"description"`
	Commands      map[string]string `json:"commands,omitempty"`
	Instructions  []Instruction     `json:"instructions,omitempty"`
	Guardrails    []Guardrail       `json:"guardrails,omitempty"`
	SeverityScale map[string]string `json:"severity_scale,omitempty"`
	AgentDefaults AgentDefaults     `json:"agent_defaults,omitempty"`
	Links         Links             `json:"links,omitempty"`
	Include       []string          `json:"include"`
	Exclude       []string          `json:"exclude"`
}

// Instruction is a coding rule the agent should follow.
// In config JSON, an entry is either a plain string (applies globally)
// or an object with a glob scope, e.g. "*.ts" or "internal/**/*.go".
type Instruction struct {
	Rule  string `json:"rule"`
	Scope string `json:"scope,omitempty"` // omitted = global
}

// @dizz-ignore-unused
func (i *Instruction) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		i.Rule, i.Scope = s, ""
		return nil
	}
	type alias Instruction
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*i = Instruction(a)
	return nil
}

// @dizz-ignore-unused
func (i Instruction) MarshalJSON() ([]byte, error) {
	if i.Scope == "" {
		return json.Marshal(i.Rule)
	}
	type alias Instruction
	return json.Marshal(alias(i))
}

// Action is what the agent must do when a guardrail matches.
type Action string

const (
	ActionReadOnly      Action = "read_only"      // never modify
	ActionRequireReview Action = "require_review" // needs explicit user approval
	ActionWarn          Action = "warn"           // proceed, but flag it
	ActionSkip          Action = "skip"           // ignore this path in analysis entirely
	ActionForbid        Action = "forbid"         // hard block, e.g. deletions
)

var validActions = map[Action]bool{
	ActionReadOnly: true, ActionRequireReview: true, ActionWarn: true,
	ActionSkip: true, ActionForbid: true,
}

// Valid returns true if the action is a recognized guardrail action.
func (a Action) Valid() bool { return validActions[a] }

// Guardrail is a rule the agent must respect before touching files.
type Guardrail struct {
	ID         string   `json:"id,omitempty"`
	Paths      []string `json:"paths,omitempty"`       // glob(s); omitted or empty = applies globally
	RequireAll bool     `json:"require_all,omitempty"` // true = fires only when ALL paths are touched together
	Action     Action   `json:"action"`
	Reason     string   `json:"reason"`
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

// HooksDirPath returns the path to the hooks directory.
func HooksDirPath(root string) string {
	return filepath.Join(root, TrackDirName, HooksDirName)
}

// CacheDirPath returns the path to the cache directory.
func CacheDirPath(root string) string {
	return filepath.Join(root, TrackDirName, CacheDirName)
}
