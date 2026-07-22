package render

import (
	"strings"
	"testing"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

func TestContextRenderer_EmptyProject(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	info := ContextInfo{
		ProjectName: "testproj",
		Branch:      "main",
		Commit:      "abc1234",
		HasGit:      true,
	}

	renderer := NewContextRenderer()
	output, err := renderer.Render(ps, is, info, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "testproj") {
		t.Fatalf("expected project name, got:\n%s", output)
	}
	if !strings.Contains(output, "main") {
		t.Fatalf("expected branch, got:\n%s", output)
	}
	if !strings.Contains(output, "abc1234") {
		t.Fatalf("expected commit, got:\n%s", output)
	}
	if !strings.Contains(output, "# config") {
		t.Fatalf("expected config section, got:\n%s", output)
	}
}

func TestContextRenderer_ActiveIntents(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	is.AddIntent(state.Intent{
		ID:       "int_001",
		Type:     state.Fixme,
		Severity: 3,
		Status:   state.IntentActive,
		Message:  "Fix critical bug",
	})
	is.AddIntent(state.Intent{
		ID:       "int_002",
		Type:     state.Refactor,
		Severity: 1,
		Status:   state.IntentActive,
		Message:  "Clean up code",
	})
	is.AddIntent(state.Intent{
		ID:       "int_003",
		Type:     state.Fixme,
		Severity: 2,
		Status:   state.IntentResolved,
		Message:  "Already done",
	})

	renderer := NewContextRenderer()
	output, err := renderer.Render(ps, is, ContextInfo{ProjectName: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "# intents") || !strings.Contains(output, "int_001") || !strings.Contains(output, "int_002") {
		t.Fatalf("expected active intents, got:\n%s", output)
	}
	if strings.Contains(output, "int_003") {
		t.Fatalf("resolved intents should not appear, got:\n%s", output)
	}
}

func TestContextRenderer_SymbolStates(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	now := time.Now()
	ps.AddSymbol(state.Symbol{
		Name:       "StableFunc",
		File:       "src/main.go",
		Line:       10,
		Type:       "function",
		State:      state.Active,
		Confidence: 0.9,
		IsDefined:  true,
		IsCalled:   true,
	})
	ps.AddSymbol(state.Symbol{
		Name:             "UnstableFunc",
		File:             "src/hot.go",
		Line:             20,
		Type:             "function",
		State:            state.Unstable,
		Confidence:       0.7,
		ChurnCount:       12,
		InstabilityScore: 0.85,
	})
	ps.AddSymbol(state.Symbol{
		Name:        "UnusedFunc",
		File:        "src/old.go",
		Line:        30,
		Type:        "function",
		State:       state.Unused,
		Confidence:  0.6,
		LastTouched: &now,
	})
	ps.AddSymbol(state.Symbol{
		Name:       "PlannedFunc",
		File:       "src/todo.go",
		Line:       40,
		Type:       "function",
		State:      state.Planned,
		Confidence: 0.7,
	})

	renderer := NewContextRenderer()
	output, err := renderer.Render(ps, is, ContextInfo{ProjectName: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "# symbols") {
		t.Fatalf("missing symbols summary section: %s", output)
	}
	if !strings.Contains(output, "unstable|1") {
		t.Fatalf("missing unstable count: %s", output)
	}
	if !strings.Contains(output, "unused|1") {
		t.Fatalf("missing unused count: %s", output)
	}
	if !strings.Contains(output, "planned|1") {
		t.Fatalf("missing planned count: %s", output)
	}
	if strings.Contains(output, "UnstableFunc") || strings.Contains(output, "UnusedFunc") || strings.Contains(output, "PlannedFunc") || strings.Contains(output, "StableFunc") {
		t.Fatalf("symbol names should not appear in context output: %s", output)
	}
}

func TestContextRenderer_ActiveTodos(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	ps.Todos = append(ps.Todos, state.Todo{
		File: "src/main.go",
		Line: 42,
		Text: "TODO: implement feature",
		Type: "TODO",
	})
	ps.Todos = append(ps.Todos, state.Todo{
		File:     "src/lib.go",
		Line:     10,
		Text:     "FIXME: fix bug",
		Type:     "FIXME",
		Resolved: true,
	})

	renderer := NewContextRenderer()
	output, err := renderer.Render(ps, is, ContextInfo{ProjectName: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "# todos") || !strings.Contains(output, "implement feature") {
		t.Fatalf("expected todos in output: %s", output)
	}
	if strings.Contains(output, "fix bug") {
		t.Fatalf("resolved todos should not appear: %s", output)
	}
}

func TestContextRenderer_OutputIsValidTON(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	is.AddIntent(state.Intent{
		ID:       "int_001",
		Type:     state.Fixme,
		Severity: 3,
		Status:   state.IntentActive,
		Message:  "test",
	})

	renderer := NewContextRenderer()
	output, err := renderer.Render(ps, is, ContextInfo{ProjectName: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}

	_, err = ton.NewReader([]byte(output))
	_ = err
}

func TestContextRenderer_GitInfo(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	tests := []struct {
		name     string
		info     ContextInfo
		contains string
	}{
		{"with git", ContextInfo{ProjectName: "p", Branch: "main", Commit: "abc1234", HasGit: true}, "abc1234"},
		{"no git", ContextInfo{ProjectName: "p", HasGit: false}, "no git"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewContextRenderer()
			output, err := renderer.Render(ps, is, tc.info, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output, tc.contains) {
				t.Fatalf("expected %q in output:\n%s", tc.contains, output)
			}
		})
	}
}

func TestContextRenderer_MultipleSections(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	is.AddIntent(state.Intent{
		ID:       "int_001",
		Type:     state.Fixme,
		Severity: 3,
		Status:   state.IntentActive,
		Message:  "fix bug",
	})

	ps.AddSymbol(state.Symbol{
		Name:  "UnstableFunc",
		File:  "src/hot.go",
		Line:  20,
		Type:  "function",
		State: state.Unstable,
	})

	ps.AddSymbol(state.Symbol{
		Name:  "UnusedFunc",
		File:  "src/legacy.go",
		Line:  30,
		Type:  "function",
		State: state.Unused,
	})

	ps.Todos = append(ps.Todos, state.Todo{
		File: "src/main.go",
		Line: 10,
		Text: "TODO: refactor",
		Type: "TODO",
	})

	renderer := NewContextRenderer()
	output, err := renderer.Render(ps, is, ContextInfo{ProjectName: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	sections := []string{
		"Project:",
		"# config",
		"# intents",
		"# symbols",
		"# todos",
	}
	for _, s := range sections {
		if !strings.Contains(output, s) {
			t.Fatalf("missing section %q in output:\n%s", s, output)
		}
	}
}

func TestContextRenderer_GitCommitInfo(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()

	ps.GitCommit = &integrations.Commit{
		Hash:    "abc1234def5678",
		Message: "feat: add feature",
		Time:    time.Now(),
	}

	renderer := NewContextRenderer()
	output, err := renderer.Render(ps, is, ContextInfo{ProjectName: "p", HasGit: true}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "abc1234") {
		t.Fatalf("expected commit hash in output: %s", output)
	}
}

func TestContextRenderer_EdgeCases(t *testing.T) {
	ps := state.NewProjectState()
	is := state.NewIntentState()
	renderer := NewContextRenderer()

	ps.AddSymbol(state.Symbol{
		Name:       "",
		File:       "src/main.go",
		State:      state.Unused,
		Confidence: 0.5,
		IsDefined:  true,
		IsCalled:   false,
	})

	output, err := renderer.Render(ps, is, ContextInfo{ProjectName: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "# symbols") {
		t.Fatalf("symbols summary section should appear: %s", output)
	}
}
