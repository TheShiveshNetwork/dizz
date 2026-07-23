package store

import (
	"testing"

	"github.com/TheShiveshNetwork/dizz/config"
)

func TestConfigStore_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := NewConfigStore(dir)

	cfg := &config.Config{
		Version:     config.ConfigVersion,
		ProjectName: "demo",
		Description: "Project guidance",
		Include:     []string{"**/*.go"},
		Exclude:     []string{"vendor/**"},
		Commands: map[string]string{
			"build": "make build",
			"test":  "go test ./...",
			"lint":  "golangci-lint run",
		},
		Instructions: []config.Instruction{
			{Rule: "Errors are wrapped with context, never silently swallowed", Scope: "internal/**"},
		},
		Guardrails: []config.Guardrail{
			{Path: "internal/generated/**", Action: "read_only", Reason: "auto-generated"},
		},
		SeverityScale: map[string]string{
			"0": "exploratory",
			"1": "minor",
			"2": "important",
			"3": "critical, blocks release",
		},
		AgentDefaults: config.AgentDefaults{
			DefaultLens: "priority",
			MinSeverity: 1,
		},
		Links: config.Links{
			Contributing: "CONTRIBUTING.md",
			Docs:         "https://dizz.shitworks.co/docs",
		},
	}

	if err := s.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Version != cfg.Version {
		t.Fatalf("version mismatch: got %q want %q", loaded.Version, cfg.Version)
	}
	if loaded.ProjectName != cfg.ProjectName {
		t.Fatalf("project name mismatch: got %q want %q", loaded.ProjectName, cfg.ProjectName)
	}
	if loaded.Description != cfg.Description {
		t.Fatalf("description mismatch: got %q want %q", loaded.Description, cfg.Description)
	}
	if len(loaded.Include) != len(cfg.Include) || loaded.Include[0] != cfg.Include[0] {
		t.Fatalf("include mismatch: got %#v want %#v", loaded.Include, cfg.Include)
	}
	if len(loaded.Exclude) != len(cfg.Exclude) || loaded.Exclude[0] != cfg.Exclude[0] {
		t.Fatalf("exclude mismatch: got %#v want %#v", loaded.Exclude, cfg.Exclude)
	}
	// Check commands
	if len(loaded.Commands) != len(cfg.Commands) {
		t.Fatalf("commands length mismatch: got %d want %d", len(loaded.Commands), len(cfg.Commands))
	}
	for k, v := range cfg.Commands {
		if loaded.Commands[k] != v {
			t.Fatalf("command mismatch for key %s: got %q want %q", k, loaded.Commands[k], v)
		}
	}
	// Check instructions
	if len(loaded.Instructions) != len(cfg.Instructions) {
		t.Fatalf("instructions length mismatch: got %d want %d", len(loaded.Instructions), len(cfg.Instructions))
	}
	for i, v := range cfg.Instructions {
		if loaded.Instructions[i].Rule != v.Rule || loaded.Instructions[i].Scope != v.Scope {
			t.Fatalf("instruction mismatch at index %d: got %#v want %#v", i, loaded.Instructions[i], v)
		}
	}
	// Check guardrails
	if len(loaded.Guardrails) != len(cfg.Guardrails) {
		t.Fatalf("guardrails length mismatch: got %d want %d", len(loaded.Guardrails), len(cfg.Guardrails))
	}
	for i, v := range cfg.Guardrails {
		if loaded.Guardrails[i].Path != v.Path || loaded.Guardrails[i].Action != v.Action || loaded.Guardrails[i].Reason != v.Reason {
			t.Fatalf("guardrail mismatch at index %d: got %#v want %#v", i, loaded.Guardrails[i], v)
		}
	}
	// Check severity scale
	if len(loaded.SeverityScale) != len(cfg.SeverityScale) {
		t.Fatalf("severity scale length mismatch: got %d want %d", len(loaded.SeverityScale), len(cfg.SeverityScale))
	}
	for k, v := range cfg.SeverityScale {
		if loaded.SeverityScale[k] != v {
			t.Fatalf("severity scale mismatch for key %s: got %q want %q", k, loaded.SeverityScale[k], v)
		}
	}
	// Check agent defaults
	if loaded.AgentDefaults.DefaultLens != cfg.AgentDefaults.DefaultLens {
		t.Fatalf("agent defaults lens mismatch: got %q want %q", loaded.AgentDefaults.DefaultLens, cfg.AgentDefaults.DefaultLens)
	}
	if loaded.AgentDefaults.MinSeverity != cfg.AgentDefaults.MinSeverity {
		t.Fatalf("agent defaults min severity mismatch: got %d want %d", loaded.AgentDefaults.MinSeverity, cfg.AgentDefaults.MinSeverity)
	}
	// Check links
	if loaded.Links.Contributing != cfg.Links.Contributing {
		t.Fatalf("links contributing mismatch: got %q want %q", loaded.Links.Contributing, cfg.Links.Contributing)
	}
	if loaded.Links.Docs != cfg.Links.Docs {
		t.Fatalf("link docs mismatch: got %q want %q", loaded.Links.Docs, cfg.Links.Docs)
	}
}
