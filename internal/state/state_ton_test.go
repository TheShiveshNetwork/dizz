package state

import (
	"strings"
	"testing"
)

func TestMarshalStateTON_Basic(t *testing.T) {
	ps := NewProjectState()
	ps.AddSymbol(Symbol{Name: "Foo", File: "main.go", Line: 10, State: Active, Confidence: 0.9})
	ps.AddSymbol(Symbol{Name: "Bar", File: "util.go", Line: 20, State: Unused, Confidence: 0.5})

	is := NewIntentState()
	is.AddIntent(Intent{
		ID: "int_001", Type: Fixme, Severity: 3, Status: IntentActive, Message: "Fix bug",
	})

	data, hash := MarshalStateTON(ps, is, []string{"abc123"})
	out := string(data)

	if !strings.Contains(out, "# intents") {
		t.Fatalf("missing intents section:\n%s", out)
	}
	if !strings.Contains(out, "# symbols:unused") {
		t.Fatalf("missing unused section:\n%s", out)
	}
	if !strings.Contains(out, "abc123") {
		t.Fatalf("missing snapshot hash:\n%s", out)
	}
	if !strings.Contains(out, "hash|") {
		t.Fatalf("missing content hash:\n%s", out)
	}
	if hash == "" {
		t.Fatal("expected non-empty content hash")
	}
}

func TestMarshalStateTON_Empty(t *testing.T) {
	ps := NewProjectState()
	is := NewIntentState()

	data, _ := MarshalStateTON(ps, is, nil)
	out := string(data)

	if !strings.Contains(out, "project|") {
		t.Fatalf("expected project header, got:\n%s", out)
	}
	if strings.Contains(out, "# intents") {
		t.Fatalf("expected no intents section for empty state:\n%s", out)
	}
	if strings.Contains(out, "# symbols:") {
		t.Fatalf("expected no symbol sections for empty state:\n%s", out)
	}
}

func TestVerifyStateTONHash_Valid(t *testing.T) {
	ps := NewProjectState()
	ps.AddSymbol(Symbol{Name: "Foo", File: "f.go", Line: 1, State: Active, Confidence: 1.0})

	data, _ := MarshalStateTON(ps, nil, nil)

	if !VerifyStateTONHash(data) {
		t.Fatal("expected valid hash verification")
	}
}

func TestVerifyStateTONHash_Invalid(t *testing.T) {
	ps := NewProjectState()
	data, _ := MarshalStateTON(ps, nil, nil)

	tampered := strings.Replace(string(data), "hash|", "hash|x", 1)

	if VerifyStateTONHash([]byte(tampered)) {
		t.Fatal("expected hash verification to fail on tampered data")
	}
}

func TestMarshalStateTON_SymbolSections(t *testing.T) {
	ps := NewProjectState()
	ps.AddSymbol(Symbol{Name: "Active", File: "a.go", Line: 1, State: Active, Confidence: 0.9})
	ps.AddSymbol(Symbol{Name: "Planned", File: "p.go", Line: 2, State: Planned, Confidence: 0.7})
	ps.AddSymbol(Symbol{Name: "Unstable", File: "u.go", Line: 3, State: Unstable, Confidence: 0.6})
	ps.AddSymbol(Symbol{Name: "Unused", File: "x.go", Line: 4, State: Unused, Confidence: 0.5})
	ps.AddSymbol(Symbol{Name: "Abandoned", File: "z.go", Line: 5, State: Abandoned, Confidence: 0.4})

	data, _ := MarshalStateTON(ps, nil, nil)
	out := string(data)

	sections := []string{
		"# symbols:unstable",
		"# symbols:unused",
		"# symbols:abandoned",
		"# symbols:planned",
		"# symbols:active",
	}
	for _, s := range sections {
		if !strings.Contains(out, s) {
			t.Fatalf("missing section %q:\n%s", s, out)
		}
	}
}

func TestMarshalStateTON_Todos(t *testing.T) {
	ps := NewProjectState()
	ps.AddTodo(Todo{File: "main.go", Line: 42, Text: "TODO: implement", Type: "TODO"})
	ps.AddTodo(Todo{File: "lib.go", Line: 10, Text: "FIXME: fix bug", Type: "FIXME"})

	data, _ := MarshalStateTON(ps, nil, nil)
	out := string(data)

	if !strings.Contains(out, "# todos") {
		t.Fatalf("missing todos section:\n%s", out)
	}
	if !strings.Contains(out, "implement") {
		t.Fatalf("missing todo text:\n%s", out)
	}
	if !strings.Contains(out, "fix bug") {
		t.Fatalf("missing fixme text:\n%s", out)
	}
}

func TestVerifyStateTONHash_NoHashLine(t *testing.T) {
	if VerifyStateTONHash([]byte("project|test|git:no git\n")) {
		t.Fatal("expected false for data without hash line")
	}
}

func TestVerifyStateTONHash_Empty(t *testing.T) {
	if VerifyStateTONHash([]byte{}) {
		t.Fatal("expected false for empty data")
	}
}

func TestMarshalStateTON_NoResolvedTodos(t *testing.T) {
	ps := NewProjectState()
	ps.AddTodo(Todo{File: "a.go", Line: 1, Text: "TODO: active", Type: "TODO"})
	ps.AddTodo(Todo{File: "b.go", Line: 2, Text: "TODO: resolved", Type: "TODO", Resolved: true})

	data, _ := MarshalStateTON(ps, nil, nil)
	out := string(data)

	if !strings.Contains(out, "active") {
		t.Fatalf("expected active todo in output:\n%s", out)
	}
	if strings.Contains(out, "resolved") {
		t.Fatalf("resolved todos should not appear:\n%s", out)
	}
}
