package state

import (
	"testing"
	"time"
)

func TestDiffSnapshots_NoChanges(t *testing.T) {
	now := time.Now()
	prev := NewProjectState()
	prev.UpdatedAt = now
	prev.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10, State: Active, Confidence: 0.9})

	next := NewProjectState()
	next.UpdatedAt = now
	next.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10, State: Active, Confidence: 0.9})

	d := DiffSnapshots(prev, next)
	if !d.IsEmpty() {
		t.Fatalf("expected empty delta, got added=%d removed=%d changed=%d",
			len(d.SymbolsAdded), len(d.SymbolsRemoved), len(d.SymbolsChanged))
	}
}

func TestDiffSnapshots_AddedSymbol(t *testing.T) {
	prev := NewProjectState()
	prev.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10})

	next := NewProjectState()
	next.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10})
	next.AddSymbol(Symbol{Name: "Bar", File: "b.go", Line: 20})

	d := DiffSnapshots(prev, next)
	if len(d.SymbolsAdded) != 1 || d.SymbolsAdded[0].Name != "Bar" {
		t.Fatalf("expected Bar added, got %+v", d.SymbolsAdded)
	}
	if d.IsEmpty() {
		t.Fatalf("expected IsEmpty to be false when changes exist")
	}
}

func TestDiffSnapshots_RemovedSymbol(t *testing.T) {
	prev := NewProjectState()
	prev.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10})
	prev.AddSymbol(Symbol{Name: "Bar", File: "b.go", Line: 20})

	next := NewProjectState()
	next.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10})

	d := DiffSnapshots(prev, next)
	if len(d.SymbolsRemoved) != 1 || d.SymbolsRemoved[0].Name != "Bar" {
		t.Fatalf("expected Bar removed, got %+v", d.SymbolsRemoved)
	}
}

func TestDiffSnapshots_ChangedSymbol(t *testing.T) {
	prev := NewProjectState()
	prev.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10, State: Active, Confidence: 0.9})

	next := NewProjectState()
	next.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10, State: Unstable, Confidence: 0.7})

	d := DiffSnapshots(prev, next)
	if len(d.SymbolsChanged) != 1 || d.SymbolsChanged[0].Name != "Foo" {
		t.Fatalf("expected Foo changed, got %+v", d.SymbolsChanged)
	}
}

func TestDiffSnapshots_Todos(t *testing.T) {
	prev := NewProjectState()
	prev.AddTodo(Todo{File: "a.go", Line: 10, Text: "TODO: old", Type: "TODO"})

	next := NewProjectState()
	next.AddTodo(Todo{File: "a.go", Line: 10, Text: "TODO: old", Type: "TODO"})
	next.AddTodo(Todo{File: "b.go", Line: 20, Text: "TODO: new", Type: "TODO"})

	d := DiffSnapshots(prev, next)
	if len(d.TodosAdded) != 1 {
		t.Fatalf("expected 1 todo added, got %d", len(d.TodosAdded))
	}
}

func TestApplyDelta_NoChanges(t *testing.T) {
	base := NewProjectState()
	base.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 10, State: Active})
	base.AddTodo(Todo{File: "a.go", Line: 10, Text: "TODO", Type: "TODO"})

	d := &SnapshotDelta{CreatedAt: time.Now()}
	result := ApplyDelta(base, d)

	if len(result.Symbols) != 1 {
		t.Fatalf("expected 1 symbol preserved, got %d", len(result.Symbols))
	}
	if len(result.Todos) != 1 {
		t.Fatalf("expected 1 todo preserved, got %d", len(result.Todos))
	}
}

func TestApplyDelta_AddRemoveChange(t *testing.T) {
	base := NewProjectState()
	base.AddSymbol(Symbol{Name: "Keep", File: "a.go", Line: 1, State: Active})
	base.AddSymbol(Symbol{Name: "Remove", File: "b.go", Line: 2, State: Unused})
	base.AddSymbol(Symbol{Name: "Change", File: "c.go", Line: 3, State: Active, Confidence: 0.9})

	d := &SnapshotDelta{
		CreatedAt: time.Now(),
		SymbolsAdded: []Symbol{
			{Name: "New", File: "d.go", Line: 4, State: Active},
		},
		SymbolsRemoved: []Symbol{
			{Name: "Remove", File: "b.go", Line: 2},
		},
		SymbolsChanged: []Symbol{
			{Name: "Change", File: "c.go", Line: 3, State: Unstable, Confidence: 0.7},
		},
		TodosAdded: []Todo{
			{File: "e.go", Line: 5, Text: "new todo", Type: "TODO"},
		},
	}

	result := ApplyDelta(base, d)

	if len(result.Symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(result.Symbols))
	}

	found := make(map[string]Symbol)
	for _, s := range result.Symbols {
		found[s.Name] = s
	}

	if _, ok := found["Keep"]; !ok {
		t.Fatal("expected Keep to survive")
	}
	if _, ok := found["Remove"]; ok {
		t.Fatal("expected Remove to be deleted")
	}
	if s, ok := found["Change"]; !ok {
		t.Fatal("expected Change to exist")
	} else if s.State != Unstable {
		t.Fatalf("expected Change state=Unstable, got %s", s.State)
	}
	if _, ok := found["New"]; !ok {
		t.Fatal("expected New to be added")
	}

	if len(result.Todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(result.Todos))
	}
}

func TestApplyDelta_RemoveAdded(t *testing.T) {
	base := NewProjectState()
	base.AddSymbol(Symbol{Name: "Foo", File: "a.go", Line: 1, State: Active})

	d := &SnapshotDelta{
		CreatedAt:      time.Now(),
		SymbolsAdded:   []Symbol{{Name: "Bar", File: "b.go", Line: 2, State: Active}},
		SymbolsRemoved: []Symbol{{Name: "Foo", File: "a.go", Line: 1}},
	}

	result := ApplyDelta(base, d)
	if len(result.Symbols) != 1 {
		t.Fatalf("expected 1 symbol (only Bar), got %d", len(result.Symbols))
	}
	if result.Symbols[0].Name != "Bar" {
		t.Fatalf("expected Bar, got %s", result.Symbols[0].Name)
	}
}

func TestDiffApplyRoundTrip(t *testing.T) {
	prev := NewProjectState()
	prev.AddSymbol(Symbol{Name: "A", File: "a.go", Line: 1, State: Active, Confidence: 0.9})
	prev.AddSymbol(Symbol{Name: "B", File: "b.go", Line: 2, State: Planned, Confidence: 0.7})
	prev.AddTodo(Todo{File: "a.go", Line: 10, Text: "TODO: do A", Type: "TODO"})

	next := NewProjectState()
	next.AddSymbol(Symbol{Name: "A", File: "a.go", Line: 1, State: Active, Confidence: 0.95})
	next.AddSymbol(Symbol{Name: "C", File: "c.go", Line: 3, State: Unused, Confidence: 0.5})
	next.AddTodo(Todo{File: "a.go", Line: 10, Text: "TODO: do A", Type: "TODO"})
	next.AddTodo(Todo{File: "c.go", Line: 20, Text: "TODO: handle C", Type: "TODO"})

	d := DiffSnapshots(prev, next)

	result := ApplyDelta(prev, d)

	if len(result.Symbols) != 2 {
		t.Fatalf("expected 2 symbols (A changed, C added), got %d", len(result.Symbols))
	}
	if len(result.Todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(result.Todos))
	}
}

func TestCheckpointInterval(t *testing.T) {
	if CheckpointInterval != 10 {
		t.Fatalf("expected CheckpointInterval=10, got %d", CheckpointInterval)
	}
}
