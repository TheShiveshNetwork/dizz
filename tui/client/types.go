package client

import (
	"encoding/json"
)

type Symbol struct {
	Name       string  `json:"name"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	EndLine    int     `json:"end_line,omitempty"`
	Type       string  `json:"type"`
	Language   string  `json:"language"`
	State      string  `json:"state"`
	Confidence float64 `json:"confidence"`
	ChurnCount int     `json:"churn_count"`
	HasTodo    bool    `json:"has_todo"`
}

type Todo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Text     string `json:"text"`
	Type     string `json:"type"`
	Language string `json:"language"`
	Resolved bool   `json:"resolved"`
}

type Intent struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Message  string   `json:"message"`
	Scope    string   `json:"scope"`
	Severity int      `json:"severity"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags,omitempty"`
}

type SnapshotInfo struct {
	Hash      string
	Timestamp string
	Kind      string
	Size      string
}

type Summary struct {
	TotalSymbols int
	Active       int
	Planned      int
	Unstable     int
	Unused       int
	Abandoned    int
	ActiveTodos  int
	Intents      int
	TotalFiles   int
	ProjectName  string
	Branch       string
	Commit       string
}

type ConfigInstruction struct {
	Rule  string `json:"rule"`
	Scope string `json:"scope,omitempty"`
}

type ConfigGuardrail struct {
	ID         string   `json:"id,omitempty"`
	Paths      []string `json:"paths,omitempty"`
	RequireAll bool     `json:"require_all,omitempty"`
	Action     string   `json:"action"`
	Reason     string   `json:"reason"`
}

type ConfigAgentDefaults struct {
	DefaultLens string `json:"default_lens,omitempty"`
	MinSeverity int    `json:"min_severity,omitempty"`
}

type ProjectConfig struct {
	Version       string               `json:"version"`
	ProjectName   string               `json:"project_name"`
	Description   string               `json:"description"`
	Commands      map[string]string    `json:"commands,omitempty"`
	Instructions  json.RawMessage      `json:"instructions,omitempty"`
	Guardrails    []ConfigGuardrail    `json:"guardrails,omitempty"`
	SeverityScale map[string]string    `json:"severity_scale,omitempty"`
	AgentDefaults *ConfigAgentDefaults `json:"agent_defaults,omitempty"`
	Links         map[string]string    `json:"links,omitempty"`
	Include       []string             `json:"include"`
	Exclude       []string             `json:"exclude"`
}

func (c *ProjectConfig) ParseInstructions() []ConfigInstruction {
	if len(c.Instructions) == 0 {
		return nil
	}

	var result []ConfigInstruction

	// Try as string array first
	var strings []string
	if err := json.Unmarshal(c.Instructions, &strings); err == nil {
		for _, s := range strings {
			result = append(result, ConfigInstruction{Rule: s})
		}
		return result
	}

	// Try as object array
	var objects []ConfigInstruction
	if err := json.Unmarshal(c.Instructions, &objects); err == nil {
		return objects
	}

	return result
}
