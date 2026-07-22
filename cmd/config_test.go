package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

func TestRunConfigAddConvention(t *testing.T) {
	projectRoot := setupDizzProject(t)
	restore := chdir(t, projectRoot)
	defer restore()

	configAddConventionRule = "Always run go test before merge"
	configAddConventionScope = "internal/**"
	defer resetConfigAddConventionFlags()

	if err := runConfigAddConvention(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg := mustLoadConfig(t, projectRoot)
	if len(cfg.Conventions) != 1 || cfg.Conventions[0].Rule != "Always run go test before merge" || cfg.Conventions[0].Scope != "internal/**" {
		t.Fatalf("convention not persisted: %#v", cfg.Conventions)
	}

	if err := runConfigAddConvention(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg = mustLoadConfig(t, projectRoot)
	if len(cfg.Conventions) != 1 {
		t.Fatalf("duplicate convention should not be added: %#v", cfg.Conventions)
	}
}

func TestRunConfigAddGuardrail(t *testing.T) {
	projectRoot := setupDizzProject(t)
	restore := chdir(t, projectRoot)
	defer restore()

	configAddGuardrailPath = "internal/generated/**"
	configAddGuardrailAction = "read_only"
	configAddGuardrailReason = "auto-generated"
	defer resetConfigAddGuardrailFlags()

	if err := runConfigAddGuardrail(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg := mustLoadConfig(t, projectRoot)
	if len(cfg.Guardrails) != 1 || cfg.Guardrails[0].Path != "internal/generated/**" || cfg.Guardrails[0].Action != "read_only" || cfg.Guardrails[0].Reason != "auto-generated" {
		t.Fatalf("guardrail not persisted: %#v", cfg.Guardrails)
	}

	if err := runConfigAddGuardrail(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg = mustLoadConfig(t, projectRoot)
	if len(cfg.Guardrails) != 1 {
		t.Fatalf("duplicate guardrail should not be added: %#v", cfg.Guardrails)
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
	if cfg.Description != "Persistent instructions for agents" {
		t.Fatalf("description mismatch: %q", cfg.Description)
	}
}

func setupDizzProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	trackDir := filepath.Join(root, ".dizz")
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal config with only the required fields
	cfg := &config.Config{
		Version:       1,
		ProjectName:   "test-project",
		Description:   "",
		Include:       []string{"**/*"},
		Exclude:       []string{"**/*_test.go", "vendor/**", "node_modules/**", ".git/**", ".dizz/**"},
	}
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

func resetConfigAddConventionFlags() {
	configAddConventionRule = ""
	configAddConventionScope = ""
}

func resetConfigAddGuardrailFlags() {
	configAddGuardrailPath = ""
	configAddGuardrailAction = ""
	configAddGuardrailReason = ""
}
