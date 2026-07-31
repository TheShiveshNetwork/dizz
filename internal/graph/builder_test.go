package graph

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/signals"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

func writeFixtureProject(t *testing.T) string {
	t.Helper()
	projectRoot := t.TempDir()

	writeFile(t, filepath.Join(projectRoot, "go.mod"), "module example.com/test\n\ngo 1.22\n")
	writeFile(t, filepath.Join(projectRoot, "pkg/auth.go"), "")
	writeFile(t, filepath.Join(projectRoot, "pkg/validate.go"), "")
	writeFile(t, filepath.Join(projectRoot, "pkg/auth_test.go"), "")

	trackDir := config.TrackDirPath(projectRoot)
	stateStore := store.NewStateStore(trackDir)
	ps := state.NewProjectState()
	ps.Symbols = []state.Symbol{
		{Name: "Login", File: filepath.Join(projectRoot, "pkg/auth.go"), Line: 1, EndLine: 10, Type: "function", Language: "go", State: state.Active, Confidence: 0.9, SignalSource: "ast", IsDefined: true, IsCalled: true},
		{Name: "Token", File: filepath.Join(projectRoot, "pkg/auth.go"), Line: 12, EndLine: 20, Type: "function", Language: "go", State: state.Active, Confidence: 0.9, SignalSource: "ast", IsDefined: true, IsCalled: false},
		{Name: "Validate", File: filepath.Join(projectRoot, "pkg/validate.go"), Line: 1, EndLine: 10, Type: "function", Language: "go", State: state.Active, Confidence: 0.9, SignalSource: "ast", IsDefined: true, IsCalled: true},
	}
	ps.Files = []state.FileContext{
		{Path: filepath.Join(projectRoot, "pkg/auth.go"), Language: "go"},
		{Path: filepath.Join(projectRoot, "pkg/validate.go"), Language: "go"},
		{Path: filepath.Join(projectRoot, "pkg/auth_test.go"), Language: "go"},
	}
	ps.Todos = []state.Todo{
		{File: filepath.Join(projectRoot, "pkg/auth.go"), Line: 15, Text: "refactor token", Type: "TODO", Language: "go"},
	}
	if err := stateStore.SaveProjectState(ps); err != nil {
		t.Fatalf("save state: %v", err)
	}

	is := state.NewIntentState()
	is.Intents = append(is.Intents, state.Intent{
		ID: "int_001", Type: state.Refactor, Message: "Refactor auth", Scope: "pkg/**",
		CreatedAt: time.Now(), UpdatedAt: time.Now(), CreatedBy: "test", Severity: 2,
		Confidence: 1.0, Status: state.IntentActive,
	})
	is.Intents = append(is.Intents, state.Intent{
		ID: "int_002", Type: state.Refactor, Message: "Refactor token system", Scope: "pkg/auth.go",
		CreatedAt: time.Now(), UpdatedAt: time.Now(), CreatedBy: "test", Severity: 1,
		Confidence: 1.0, Status: state.IntentResolved,
	})
	if err := store.NewIntentStore(trackDir).SaveIntentState(is); err != nil {
		t.Fatalf("save intent: %v", err)
	}

	cacheDir := config.CacheDirPath(projectRoot)
	cache := store.NewSignalCache(projectRoot, cacheDir)
	now := time.Now()
	authSigs := []signals.Signal{
		{Type: signals.FunctionDefined, Name: "Login", File: "pkg/auth.go", Line: 1, EndLine: 10, Language: "go", Confidence: 0.9, Metadata: map[string]interface{}{"source_tier": "ast"}},
		{Type: signals.FunctionDefined, Name: "Token", File: "pkg/auth.go", Line: 12, EndLine: 20, Language: "go", Confidence: 0.9, Metadata: map[string]interface{}{"source_tier": "ast"}},
		{Type: signals.FunctionCalled, Name: "Validate", File: "pkg/auth.go", Line: 3, Confidence: 1.0, Metadata: map[string]interface{}{"source_tier": "ast"}},
		{Type: signals.ImportFound, Name: "example.com/test/pkg", File: "pkg/auth.go", Line: 1, Confidence: 1.0, Metadata: map[string]interface{}{"source_tier": "ast"}},
		{Type: signals.ImportFound, Name: "golang.org/x/crypto/bcrypt", File: "pkg/auth.go", Line: 2, Confidence: 1.0, Metadata: map[string]interface{}{"source_tier": "ast"}},
	}
	if err := cache.SetRel("pkg/auth.go", "h1", now, authSigs); err != nil {
		t.Fatalf("cache auth: %v", err)
	}
	valSigs := []signals.Signal{
		{Type: signals.FunctionDefined, Name: "Validate", File: "pkg/validate.go", Line: 1, EndLine: 10, Language: "go", Confidence: 0.9, Metadata: map[string]interface{}{"source_tier": "ast"}},
	}
	if err := cache.SetRel("pkg/validate.go", "h2", now, valSigs); err != nil {
		t.Fatalf("cache validate: %v", err)
	}
	testSigs := []signals.Signal{
		{Type: signals.FunctionDefined, Name: "TestLogin", File: "pkg/auth_test.go", Line: 1, EndLine: 5, Language: "go", Confidence: 0.9, Metadata: map[string]interface{}{"source_tier": "ast"}},
		{Type: signals.ImportFound, Name: "example.com/test/pkg", File: "pkg/auth_test.go", Line: 1, Confidence: 1.0, Metadata: map[string]interface{}{"source_tier": "ast"}},
	}
	if err := cache.SetRel("pkg/auth_test.go", "h3", now, testSigs); err != nil {
		t.Fatalf("cache test: %v", err)
	}
	if err := cache.SaveManifest(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cfg := config.Config{Version: config.ConfigVersion, ProjectName: "test"}
	cfg.Guardrails = []config.Guardrail{
		{ID: "gr1", Paths: []string{"pkg/auth.go"}, Action: config.ActionReadOnly, Reason: "never touch"},
	}
	if err := store.NewConfigStore(trackDir).SaveConfig(&cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	return projectRoot
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestBuildDerivesGraphFromPersistedState(t *testing.T) {
	projectRoot := writeFixtureProject(t)
	g, err := Build(DefaultBuildOptions(projectRoot))
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	for _, name := range []string{"Login", "Token", "Validate"} {
		if len(g.SymbolsNamed(name)) == 0 {
			t.Errorf("missing symbol %s", name)
		}
	}
	for _, rel := range []string{"pkg/auth.go", "pkg/validate.go", "pkg/auth_test.go"} {
		if g.FileNode(rel) == nil {
			t.Errorf("missing file %s", rel)
		}
	}

	login := g.SymbolsNamed("Login")[0].ID
	validate := g.SymbolsNamed("Validate")[0].ID
	if !g.HasEdge(EdgeCalls, login, validate) {
		t.Errorf("missing CALLS edge %s -> %s", login, validate)
	}

	authFile := FileID("pkg/auth.go")
	foundImport := false
	for _, e := range g.Outgoing(authFile) {
		if e.Type == EdgeImports {
			foundImport = true
		}
	}
	if !foundImport {
		t.Errorf("missing IMPORTS edge from %s", authFile)
	}

	depID := DepID("golang.org/x/crypto/bcrypt")
	if g.Node(depID) == nil {
		t.Errorf("missing dep node %s", depID)
	}
	if !g.HasEdge(EdgeDependsOn, authFile, depID) {
		t.Errorf("missing DEPENDS_ON edge %s -> %s", authFile, depID)
	}

	if g.Node(IntentID("int_001")) == nil {
		t.Errorf("missing intent node int_001")
	}

	testID := TestID("pkg/auth_test.go")
	if g.Node(testID) == nil {
		t.Fatalf("missing test node")
	}
	if !g.HasEdge(EdgeTests, testID, login) {
		t.Errorf("missing TESTS edge %s -> %s", testID, login)
	}

	todoID := TodoID("pkg/auth.go", 15)
	if !g.HasEdge(EdgeHasTodo, authFile, todoID) {
		t.Errorf("missing HAS_TODO edge %s -> %s", authFile, todoID)
	}

	if g.Node(ModuleID("example.com/test")) == nil {
		t.Errorf("missing module node")
	}

	if g.Node(GuardrailID("gr1")) == nil {
		t.Errorf("missing guardrail node")
	}
	if !g.HasEdge(EdgeProtects, GuardrailID("gr1"), authFile) {
		t.Errorf("missing PROTECTS edge")
	}

	if g.Node(IntentID("int_001")) == nil {
		t.Errorf("missing intent node")
	}
	matched := false
	for _, e := range g.Incoming(authFile) {
		if e.Type == EdgeScopeMatch {
			matched = true
		}
	}
	if !matched {
		t.Errorf("missing SCOPE_MATCH edge on %s", authFile)
	}
}

func TestBuildFailsWithoutState(t *testing.T) {
	projectRoot := t.TempDir()
	_, err := Build(DefaultBuildOptions(projectRoot))
	if err == nil {
		t.Fatalf("expected error for project without state, got nil")
	}
}

func TestBuildOffByDefault(t *testing.T) {
	projectRoot := writeFixtureProject(t)
	g, err := Build(DefaultBuildOptions(projectRoot))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	stats := g.ComputeStats()
	if stats.HasCoChange {
		t.Errorf("co-change should be off by default")
	}
}

func TestBuildCoChangeIncludedWhenRequested(t *testing.T) {
	if _, err := os.Stat(filepath.Join(".", ".git")); err != nil {
		t.Skip("not in a git repository")
	}
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	opts := DefaultBuildOptions(projectRoot)
	opts.IncludeCoChange = true
	g, err := Build(opts)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if !g.ComputeStats().HasCoChange {
		t.Skip("no co-change pairs in this repo")
	}
}
