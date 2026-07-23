package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

func TestRunConfigAddInstruction(t *testing.T) {
	projectRoot := setupDizzProject(t)
	restore := chdir(t, projectRoot)
	defer restore()

	configAddInstructionRule = "Always run go test before merge"
	configAddInstructionScope = "internal/**"
	defer resetConfigAddInstructionFlags()

	if err := runConfigAddInstruction(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg := mustLoadConfig(t, projectRoot)
	if len(cfg.Instructions) != 1 || cfg.Instructions[0].Rule != "Always run go test before merge" || cfg.Instructions[0].Scope != "internal/**" {
		t.Fatalf("instruction not persisted: %#v", cfg.Instructions)
	}

	if err := runConfigAddInstruction(nil, nil); err != nil {
		t.Fatal(err)
	}

	cfg = mustLoadConfig(t, projectRoot)
	if len(cfg.Instructions) != 1 {
		t.Fatalf("duplicate instruction should not be added: %#v", cfg.Instructions)
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

	cfg := &config.Config{
		Version:     config.ConfigVersion,
		ProjectName: "test-project",
		Description: "",
		Include:     []string{"**/*"},
		Exclude:     []string{"**/*_test.go", "vendor/**", "node_modules/**", ".git/**", ".dizz/**"},
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

func resetConfigAddInstructionFlags() {
	configAddInstructionRule = ""
	configAddInstructionScope = ""
}

func resetConfigAddGuardrailFlags() {
	configAddGuardrailPath = ""
	configAddGuardrailAction = ""
	configAddGuardrailReason = ""
}

func TestRunConfigShow(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowJSON(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowJSON = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyInstructions(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowInstructionsOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyGuardrails(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowGuardrailsOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyCommands(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowCommandsOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyDescription(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowDescriptionOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyName(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowNameOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyVersion(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowVersionOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyInclude(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowIncludeOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyExclude(t *testing.T) {
	projectRoot := setupDizzProjectWithInstructions(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowExcludeOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlySeverityScale(t *testing.T) {
	projectRoot := setupDizzProjectWithFullConfig(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowSeverityScaleOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyAgentDefaults(t *testing.T) {
	projectRoot := setupDizzProjectWithFullConfig(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowAgentDefaultsOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowOnlyLinks(t *testing.T) {
	projectRoot := setupDizzProjectWithFullConfig(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowLinksOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfigShowMultipleFilters(t *testing.T) {
	projectRoot := setupDizzProjectWithFullConfig(t)
	restore := chdir(t, projectRoot)
	defer restore()
	defer resetAllShowFlags()

	configShowNameOnly = true
	configShowInstructionsOnly = true
	if err := runConfigShow(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func setupDizzProjectWithInstructions(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	trackDir := filepath.Join(root, ".dizz")
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Version:     config.ConfigVersion,
		ProjectName: "test-project",
		Description: "Test description",
		Include:     []string{"**/*"},
		Exclude:     []string{"**/*_test.go"},
		Commands: map[string]string{
			"build": "go build",
			"test":  "go test ./...",
		},
		Instructions: []config.Instruction{
			{Rule: "Run tests before merge", Scope: "internal/**"},
		},
		Guardrails: []config.Guardrail{
			{Path: "generated/**", Action: "read_only", Reason: "auto-generated"},
		},
	}
	configStore := store.NewConfigStore(trackDir)
	if err := configStore.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	return root
}

func setupDizzProjectWithFullConfig(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	trackDir := filepath.Join(root, ".dizz")
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Version:     config.ConfigVersion,
		ProjectName: "test-project",
		Description: "Test description",
		Include:     []string{"**/*"},
		Exclude:     []string{"**/*_test.go"},
		Commands: map[string]string{
			"build": "go build",
			"test":  "go test ./...",
		},
		Instructions: []config.Instruction{
			{Rule: "Run tests before merge", Scope: "internal/**"},
		},
		Guardrails: []config.Guardrail{
			{Path: "generated/**", Action: "read_only", Reason: "auto-generated"},
		},
		SeverityScale: map[string]string{
			"0": "exploratory",
			"1": "minor",
			"2": "important",
			"3": "critical",
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
	configStore := store.NewConfigStore(trackDir)
	if err := configStore.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	return root
}
