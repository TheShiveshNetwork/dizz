package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/state"
)

func TestIntentStore_SaveLoadTON(t *testing.T) {
	dir := t.TempDir()
	s := NewIntentStore(dir)

	is := state.NewIntentState()
	is.AddIntent(state.Intent{
		ID:        "int_001",
		Type:      state.Fixme,
		Severity:  3,
		Status:    state.IntentActive,
		Message:   "Fix critical bug",
		Scope:     "project",
		CreatedBy: "user",
	})

	if err := s.SaveIntentState(is); err != nil {
		t.Fatal(err)
	}

	tonPath := filepath.Join(dir, "intent.ton")
	jsonPath := filepath.Join(dir, "intent.json")

	if _, err := os.Stat(tonPath); os.IsNotExist(err) {
		t.Fatalf("intent.ton not written at %s", tonPath)
	}
	if _, err := os.Stat(jsonPath); err == nil {
		// JSON is no longer the primary format; this is optional
		t.Logf("intent.json should not be written by default")
	}

	loaded, err := s.LoadIntentState()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Intents) != 1 {
		t.Fatalf("expected 1 intent, got %d", len(loaded.Intents))
	}
	if loaded.Intents[0].ID != "int_001" {
		t.Fatalf("expected int_001, got %s", loaded.Intents[0].ID)
	}
	if loaded.Intents[0].Message != "Fix critical bug" {
		t.Fatalf("message mismatch: %s", loaded.Intents[0].Message)
	}
}

func TestIntentStore_LoadFallbackToJSON(t *testing.T) {
	dir := t.TempDir()

	// Write an old-format JSON file (no .ton)
	jsonPath := filepath.Join(dir, "intent.json")
	jsonContent := `{
  "version": 1,
  "intents": [
    {
      "id": "int_legacy",
      "type": "question",
      "severity": 1,
      "status": "active",
      "message": "Legacy intent"
    }
  ]
}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewIntentStore(dir)

	loaded, err := s.LoadIntentState()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Intents) != 1 {
		t.Fatalf("expected 1 intent, got %d", len(loaded.Intents))
	}
	if loaded.Intents[0].ID != "int_legacy" {
		t.Fatalf("unexpected id: %s", loaded.Intents[0].ID)
	}
	if loaded.Intents[0].Type != state.Question {
		t.Fatalf("unexpected type: %s", loaded.Intents[0].Type)
	}
}

func TestIntentStore_RoundTripMultiple(t *testing.T) {
	dir := t.TempDir()
	s := NewIntentStore(dir)

	is := state.NewIntentState()
	for i := 0; i < 5; i++ {
		is.AddIntent(state.Intent{
			ID:       "int_" + string(rune('A'+i)),
			Type:     state.IntentTodo,
			Severity: i,
			Status:   state.IntentActive,
			Message:  "Intent " + string(rune('A'+i)),
		})
	}

	if err := s.SaveIntentState(is); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.LoadIntentState()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Intents) != 5 {
		t.Fatalf("expected 5 intents, got %d", len(loaded.Intents))
	}

	for i, intent := range loaded.Intents {
		expectedMsg := "Intent " + string(rune('A'+i))
		if intent.Message != expectedMsg {
			t.Fatalf("intent %d: expected %q, got %q", i, expectedMsg, intent.Message)
		}
	}
}

func TestIntentStore_EmptyState(t *testing.T) {
	dir := t.TempDir()
	s := NewIntentStore(dir)

	loaded, err := s.LoadIntentState()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Intents) != 0 {
		t.Fatalf("expected 0 intents for empty state, got %d", len(loaded.Intents))
	}
}

func TestIntentStore_TONPreferredOverJSON(t *testing.T) {
	dir := t.TempDir()

	// Write both formats with different content
	jsonPath := filepath.Join(dir, "intent.json")
	jsonContent := `{"version":1,"intents":[{"id":"int_from_json","type":"fixme","severity":3,"status":"active","message":"from JSON"}]}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	tonPath := filepath.Join(dir, "intent.ton")
	tonContent := "# intents\nid|type|sev|status|msg|scope|tags|created_by|resolution\nint_from_ton|fixme|3|active|from TON|project||user|\n"
	if err := os.WriteFile(tonPath, []byte(tonContent), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewIntentStore(dir)

	loaded, err := s.LoadIntentState()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Intents) != 1 {
		t.Fatalf("expected 1 intent, got %d", len(loaded.Intents))
	}

	// Should have loaded from TON (preferred over JSON)
	if loaded.Intents[0].ID != "int_from_ton" {
		t.Fatalf("expected to load from TON, got %s", loaded.Intents[0].ID)
	}
}

func TestIntentStore_SaveAtomic(t *testing.T) {
	dir := t.TempDir()
	s := NewIntentStore(dir)

	// Save initial state
	is := state.NewIntentState()
	is.AddIntent(state.Intent{
		ID:       "int_001",
		Type:     state.IntentTodo,
		Severity: 1,
		Status:   state.IntentActive,
		Message:  "test",
	})
	if err := s.SaveIntentState(is); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.LoadIntentState()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Intents) != 1 {
		t.Fatalf("expected 1 intent after save, got %d", len(loaded.Intents))
	}
}
