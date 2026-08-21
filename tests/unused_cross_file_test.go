package benchmarks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/analyzer"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/ast"
	"github.com/TheShiveshNetwork/dizz/internal/analyzer/regex"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

func TestUnusedDetectionAcrossGoFiles(t *testing.T) {
	dir := t.TempDir()
	defFile := filepath.Join(dir, "defs.go")
	useFile := filepath.Join(dir, "use.go")

	writeFile(t, defFile, "package sample\n\nfunc Exported() {}\n\nfunc UnusedLocal() {}\n")
	writeFile(t, useFile, "package sample\n\nfunc Use() {\n\tExported()\n}\n")

	ps := analyzeFilesToProjectState(t, []string{defFile, useFile})
	symbolsByName := mapSymbolsByName(ps.Symbols)

	exported, ok := symbolsByName["Exported"]
	if !ok {
		t.Fatalf("expected Exported symbol in analysis output")
	}
	if !exported.IsCalled {
		t.Fatalf("expected Exported to be called from another file")
	}
	if exported.State != state.Active {
		t.Fatalf("expected Exported state %q, got %q", state.Active, exported.State)
	}

	unusedLocal, ok := symbolsByName["UnusedLocal"]
	if !ok {
		t.Fatalf("expected UnusedLocal symbol in analysis output")
	}
	if unusedLocal.IsCalled {
		t.Fatalf("expected UnusedLocal to remain uncalled")
	}
	if unusedLocal.State != state.Unused {
		t.Fatalf("expected UnusedLocal state %q, got %q", state.Unused, unusedLocal.State)
	}
}

func TestUnusedDetectionAcrossPythonFiles(t *testing.T) {
	dir := t.TempDir()
	defFile := filepath.Join(dir, "defs.py")
	useFile := filepath.Join(dir, "use.py")

	writeFile(t, defFile, "def helper():\n    return 1\n\ndef local_only():\n    return 0\n")
	writeFile(t, useFile, "from defs import helper\n\nvalue = helper()\n")

	ps := analyzeFilesToProjectState(t, []string{defFile, useFile})
	symbolsByName := mapSymbolsByName(ps.Symbols)

	helper, ok := symbolsByName["helper"]
	if !ok {
		t.Fatalf("expected helper symbol in analysis output")
	}
	if !helper.IsCalled {
		t.Fatalf("expected helper to be called from another file")
	}
	if helper.State != state.Active {
		t.Fatalf("expected helper state %q, got %q", state.Active, helper.State)
	}

	localOnly, ok := symbolsByName["local_only"]
	if !ok {
		t.Fatalf("expected local_only symbol in analysis output")
	}
	if localOnly.IsCalled {
		t.Fatalf("expected local_only to remain uncalled")
	}
	if localOnly.State != state.Unused {
		t.Fatalf("expected local_only state %q, got %q", state.Unused, localOnly.State)
	}
}

func analyzeFilesToProjectState(t *testing.T, files []string) *state.ProjectState {
	t.Helper()

	registry := analyzer.NewRegistry()
	registry.Register(&ast.Analyzer{})
	registry.Register(regex.NewAnalyzer())

	sigSet, err := registry.AnalyzeFiles(files)
	if err != nil {
		t.Fatalf("analyze files: %v", err)
	}

	scorer := state.NewScorer()
	return scorer.InterpretSignalsWithIntent(sigSet, nil, nil, nil)
}

func mapSymbolsByName(symbols []state.Symbol) map[string]state.Symbol {
	byName := make(map[string]state.Symbol, len(symbols))
	for _, symbol := range symbols {
		byName[symbol.Name] = symbol
	}
	return byName
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
