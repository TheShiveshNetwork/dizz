package store

import (
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/config"
)

func TestConfigStore_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := NewConfigStore(dir)

	cfg := &config.Config{
		ProjectName: "demo",
		RootPath:    ".",
		Include:     []string{"**/*.go"},
		Exclude:     []string{"vendor/**"},
		Agentic: config.AgenticConfig{
			Description:  "Project guidance",
			Rules:        []string{"Never rewrite git history"},
			Standards:    []string{"Use gofmt"},
			Instructions: []string{"Run targeted tests first"},
		},
	}

	if err := s.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Agentic.Description != cfg.Agentic.Description {
		t.Fatalf("description mismatch: got %q want %q", loaded.Agentic.Description, cfg.Agentic.Description)
	}
	if len(loaded.Agentic.Rules) != 1 || loaded.Agentic.Rules[0] != cfg.Agentic.Rules[0] {
		t.Fatalf("rules mismatch: got %#v", loaded.Agentic.Rules)
	}
	if len(loaded.Agentic.Standards) != 1 || loaded.Agentic.Standards[0] != cfg.Agentic.Standards[0] {
		t.Fatalf("standards mismatch: got %#v", loaded.Agentic.Standards)
	}
	if len(loaded.Agentic.Instructions) != 1 || loaded.Agentic.Instructions[0] != cfg.Agentic.Instructions[0] {
		t.Fatalf("instructions mismatch: got %#v", loaded.Agentic.Instructions)
	}
}
