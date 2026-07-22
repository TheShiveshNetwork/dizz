package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/defaults"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

func TestRunConfigAdd(t *testing.T) {
	projectRoot := setupDizzProject(t)
	restore := chdir(t, projectRoot)
	defer restore()

	configAddRule = "Always run go test before merge"
	configAddStandard = ""
	configAddInstruction = ""
	defer resetConfigAddFlags()

	if err := runConfigAdd(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg := mustLoadConfig(t, projectRoot)
	if len(cfg.Agentic.Rules) != 1 || cfg.Agentic.Rules[0] != "Always run go test before merge" {
		t.Fatalf("rule not persisted: %#v", cfg.Agentic.Rules)
	}

	if err := runConfigAdd(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg = mustLoadConfig(t, projectRoot)
	if len(cfg.Agentic.Rules) != 1 {
		t.Fatalf("duplicate rule should not be added: %#v", cfg.Agentic.Rules)
	}
}

func TestRunConfigSetDescription(t *testing.T) {
	projectRoot := setupDizzProject(t)
	restore := chdir(t, projectRoot)
	defer restore()

	if err := runConfigSetDescription(nil, []string{"Persistent instructions for agents"}); err != nil {
		t.Fatal(err)
	}

	cfg := mustLoadConfig(t, projectRoot)
	if cfg.Agentic.Description != "Persistent instructions for agents" {
		t.Fatalf("description mismatch: %q", cfg.Agentic.Description)
	}
}

func setupDizzProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	trackDir := filepath.Join(root, ".dizz")
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := defaults.DefaultConfig("test-project")
	configStore := store.NewConfigStore(trackDir)
	if err := configStore.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	return root
}

func mustLoadConfig(t *testing.T, root string) *config.Config {
	t.Helper()
	cfg, err := store.NewConfigStore(filepath.Join(root, ".dizz")).LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() {
		_ = os.Chdir(old)
	}
}

func resetConfigAddFlags() {
	configAddRule = ""
	configAddStandard = ""
	configAddInstruction = ""
}
