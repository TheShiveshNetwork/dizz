package state

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompactProjectStateJSON(t *testing.T) {
	ps := NewProjectState()
	ps.AddSymbol(Symbol{
		Name:       "Foo",
		File:       "main.go",
		Line:       10,
		Type:       "function",
		Language:   "go",
		State:      Active,
		Confidence: 0.9,
		IsDefined:  true,
		IsCalled:   true,
	})

	data, err := json.Marshal(ps)
	if err != nil {
		t.Fatal(err)
	}

	output := string(data)

	// Compact JSON should not have indentation (no \n before keys)
	if strings.Contains(output, "\n    ") || strings.Contains(output, "\n  ") {
		t.Fatalf("compact JSON should not have indentation, got:\n%s", output)
	}

	// Compact JSON should not have trailing whitespace
	if strings.Contains(output, " \n") {
		t.Fatalf("compact JSON should not have trailing whitespace, got:\n%s", output)
	}

	// Verify it's valid JSON
	var decoded ProjectState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, output)
	}
}

func TestCompactJSON_RoundTrip(t *testing.T) {
	ps := NewProjectState()
	ps.AddSymbol(Symbol{
		Name:       "Bar",
		File:       "lib.go",
		Line:       5,
		Type:       "function",
		Language:   "go",
		State:      Unstable,
		Confidence: 0.6,
		IsDefined:  true,
		IsCalled:   false,
	})

	compact, err := json.Marshal(ps)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ProjectState
	if err := json.Unmarshal(compact, &decoded); err != nil {
		t.Fatal(err)
	}

	if len(decoded.Symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(decoded.Symbols))
	}
	if decoded.Symbols[0].Name != "Bar" {
		t.Fatalf("unexpected name: %s", decoded.Symbols[0].Name)
	}
	if decoded.Symbols[0].State != Unstable {
		t.Fatalf("unexpected state: %s", decoded.Symbols[0].State)
	}
}
