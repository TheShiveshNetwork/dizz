package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIntentMarshalTON_Basic(t *testing.T) {
	is := NewIntentState()
	is.AddIntent(Intent{
		ID:        "int_123",
		Type:      Fixme,
		Severity:  3,
		Status:    IntentActive,
		Message:   "Fix the bug",
		Scope:     "project",
		CreatedBy: "user",
		Tags:      []string{"performance", "security"},
	})

	data, err := is.MarshalTON()
	if err != nil {
		t.Fatal(err)
	}

	out := string(data)
	if !strings.Contains(out, "int_123") {
		t.Fatalf("missing intent id: %s", out)
	}
	if !strings.Contains(out, "fixme") {
		t.Fatalf("missing intent type: %s", out)
	}
	if !strings.Contains(out, "Fix the bug") {
		t.Fatalf("missing intent message: %s", out)
	}
	if !strings.Contains(out, "performance,security") {
		t.Fatalf("missing tags: %s", out)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header + 1 record, got %d lines", len(lines))
	}
}

func TestIntentMarshalTON_Resolved(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	is := NewIntentState()
	is.AddIntent(Intent{
		ID:        "int_456",
		Type:      IntentTodo,
		Severity:  2,
		Status:    IntentResolved,
		Message:   "Add tests",
		Scope:     "project",
		CreatedBy: "agent",
		Resolution: &Resolution{
			Method:      "fixed",
			Description: "Added tests",
			ResolvedAt:  now,
			ResolvedBy:  "agent",
		},
	})

	data, err := is.MarshalTON()
	if err != nil {
		t.Fatal(err)
	}

	out := string(data)
	if !strings.Contains(out, "resolved") {
		t.Fatalf("missing resolved status: %s", out)
	}
	if !strings.Contains(out, "fixed") {
		t.Fatalf("missing resolution method: %s", out)
	}
}

func TestIntentMarshalTON_EmptyState(t *testing.T) {
	is := NewIntentState()
	data, err := is.MarshalTON()
	if err != nil {
		t.Fatal(err)
	}

	out := string(data)
	if !strings.Contains(out, "id|type") {
		t.Fatalf("expected header even with no intents: %s", out)
	}
}

func TestIntentRoundTripTON(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	is := NewIntentState()

	is.AddIntent(Intent{
		ID:        "int_001",
		Type:      Fixme,
		Severity:  3,
		Status:    IntentActive,
		Message:   "Critical bug",
		Scope:     "module",
		CreatedBy: "user",
		Tags:      []string{"urgent"},
	})

	is.AddIntent(Intent{
		ID:        "int_002",
		Type:      Refactor,
		Severity:  1,
		Status:    IntentResolved,
		Message:   "Clean up",
		Scope:     "project",
		CreatedBy: "agent",
		Resolution: &Resolution{
			Method:      "wontfix",
			Description: "Not needed",
			ResolvedAt:  now,
			ResolvedBy:  "agent",
		},
	})

	data, err := is.MarshalTON()
	if err != nil {
		t.Fatal(err)
	}

	is2, err := UnmarshalIntentStateTON(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v\ninput: %s", err, string(data))
	}

	if is2.Version != 1.0 {
		t.Fatalf("expected version 1.0, got %f", is2.Version)
	}
	if len(is2.Intents) != 2 {
		t.Fatalf("expected 2 intents, got %d", len(is2.Intents))
	}

	if is2.Intents[0].ID != "int_001" || is2.Intents[0].Type != Fixme {
		t.Fatalf("unexpected first intent: %+v", is2.Intents[0])
	}
	if is2.Intents[0].Severity != 3 {
		t.Fatalf("expected severity 3, got %d", is2.Intents[0].Severity)
	}
	if is2.Intents[0].Status != IntentActive {
		t.Fatalf("expected active, got %s", is2.Intents[0].Status)
	}
	if len(is2.Intents[0].Tags) != 1 || is2.Intents[0].Tags[0] != "urgent" {
		t.Fatalf("unexpected tags: %v", is2.Intents[0].Tags)
	}

	if is2.Intents[1].ID != "int_002" || is2.Intents[1].Type != Refactor {
		t.Fatalf("unexpected second intent: %+v", is2.Intents[1])
	}
	if is2.Intents[1].Status != IntentResolved {
		t.Fatalf("expected resolved, got %s", is2.Intents[1].Status)
	}
	if is2.Intents[1].Resolution == nil {
		t.Fatal("expected resolution")
	}
	if is2.Intents[1].Resolution.Method != "wontfix" {
		t.Fatalf("expected wontfix, got %s", is2.Intents[1].Resolution.Method)
	}
}

func TestUnmarshalTON_Empty(t *testing.T) {
	_, err := UnmarshalIntentStateTON([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestUnmarshalTON_OnlyHeader(t *testing.T) {
	data := []byte("# intents\nid|type|sev|status|msg\n")
	is, err := UnmarshalIntentStateTON(data)
	if err != nil {
		t.Fatal(err)
	}
	if is.Version != 1.0 {
		t.Fatalf("expected version 1.0, got %f", is.Version)
	}
	if len(is.Intents) != 0 {
		t.Fatalf("expected 0 intents, got %d", len(is.Intents))
	}
}

func TestUnmarshalTON_ResolvedIntent(t *testing.T) {
	data := []byte("# intents\nid|type|sev|status|msg|resolution\nint_99|todo|1|resolved|done task|fixed:Completed:1781776800:user\n")
	is, err := UnmarshalIntentStateTON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(is.Intents) != 1 {
		t.Fatalf("expected 1 intent, got %d", len(is.Intents))
	}
	if is.Intents[0].Status != IntentResolved {
		t.Fatalf("expected resolved status")
	}
	if is.Intents[0].Resolution.Method != "fixed" {
		t.Fatalf("expected method fixed, got %s", is.Intents[0].Resolution.Method)
	}
}

func TestUnmarshalTON_Tags(t *testing.T) {
	data := []byte("# intents\nid|type|sev|status|msg|tags\nint_1|refactor|2|active|refactor|a,b,c\n")
	is, err := UnmarshalIntentStateTON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(is.Intents[0].Tags) != 3 {
		t.Fatalf("expected 3 tags, got %v", is.Intents[0].Tags)
	}
	if is.Intents[0].Tags[0] != "a" || is.Intents[0].Tags[1] != "b" || is.Intents[0].Tags[2] != "c" {
		t.Fatalf("unexpected tags: %v", is.Intents[0].Tags)
	}
}

func TestUnmarshalTON_MessageWithPipes(t *testing.T) {
	data := []byte("# intents\nid|type|sev|status|msg\nint_1|fixme|3|active|hello \\| world\n")
	is, err := UnmarshalIntentStateTON(data)
	if err != nil {
		t.Fatal(err)
	}
	if is.Intents[0].Message != "hello | world" {
		t.Fatalf("expected 'hello | world', got %q", is.Intents[0].Message)
	}
}

func TestUnmarshalTON_ExtraWhitespace(t *testing.T) {
	data := []byte("  \n  # intents  \n  id|type|sev|status|msg  \n  int_1|fixme|3|active|test  \n")
	is, err := UnmarshalIntentStateTON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(is.Intents) != 1 {
		t.Fatalf("expected 1 intent, got %d", len(is.Intents))
	}
}

func TestIntentRoundTripTON_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "intent.ton")

	is := NewIntentState()
	is.AddIntent(Intent{
		ID:       "int_001",
		Type:     Fixme,
		Severity: 3,
		Status:   IntentActive,
		Message:  "Test with pipes: a | b | c",
		Scope:    "project",
		Tags:     []string{"x", "y"},
	})

	data, err := is.MarshalTON()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	readback, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	is2, err := UnmarshalIntentStateTON(readback)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if is2.Intents[0].Message != "Test with pipes: a | b | c" {
		t.Fatalf("message mismatch: %q", is2.Intents[0].Message)
	}
	if is2.Intents[0].Scope != "project" {
		t.Fatalf("scope mismatch: %q", is2.Intents[0].Scope)
	}
	if len(is2.Intents[0].Tags) != 2 || is2.Intents[0].Tags[1] != "y" {
		t.Fatalf("tags mismatch: %v", is2.Intents[0].Tags)
	}
}

func TestIntentRoundTripTON_AllFields(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	is := NewIntentState()
	is.AddIntent(Intent{
		ID:         "int_all",
		Type:       Hack,
		Severity:   0,
		Status:     IntentDeferred,
		Message:    "Temp workaround until lib upgrade",
		Scope:      "src/worker.go:42-45",
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
		CreatedBy:  "agent",
		Confidence: 0.5,
		Tags:       nil,
	})

	data, err := is.MarshalTON()
	if err != nil {
		t.Fatal(err)
	}

	is2, err := UnmarshalIntentStateTON(data)
	if err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, string(data))
	}

	intent := is2.Intents[0]
	if intent.ID != "int_all" {
		t.Fatalf("id: got %q", intent.ID)
	}
	if intent.Type != Hack {
		t.Fatalf("type: got %s", intent.Type)
	}
	if intent.Severity != 0 {
		t.Fatalf("severity: got %d", intent.Severity)
	}
	if intent.Status != IntentDeferred {
		t.Fatalf("status: got %s", intent.Status)
	}
	if intent.Message != "Temp workaround until lib upgrade" {
		t.Fatalf("message: got %q", intent.Message)
	}
	if intent.Scope != "src/worker.go:42-45" {
		t.Fatalf("scope: got %q", intent.Scope)
	}
	if intent.CreatedBy != "agent" {
		t.Fatalf("created_by: got %q", intent.CreatedBy)
	}
}

func TestIntentUnmarshalTONRealFile(t *testing.T) {
	content := `# intents
id|type|sev|status|msg|scope|tags|created_by|resolution
int_1770020361|refactor|2|active|Refactor scoring system to use weights|project|performance,architecture|user|
int_1770047724|fixme|3|active|improve the speed of the cli|project||user|
	int_1770020251|fixme|3|resolved|Fix cache invalidation race condition|project||user|fixed:Resolved via CLI:1768789081:user
`
	is, err := UnmarshalIntentStateTON([]byte(content))
	if err != nil {
		t.Fatal(err)
	}

	if len(is.Intents) != 3 {
		t.Fatalf("expected 3 intents, got %d", len(is.Intents))
	}

	if is.Intents[0].ID != "int_1770020361" {
		t.Fatalf("unexpected id: %s", is.Intents[0].ID)
	}
	if is.Intents[0].Tags[0] != "performance" {
		t.Fatalf("unexpected tag: %s", is.Intents[0].Tags[0])
	}

	if is.Intents[1].ID != "int_1770047724" {
		t.Fatalf("unexpected id: %s", is.Intents[1].ID)
	}
	if len(is.Intents[1].Tags) != 0 {
		t.Fatalf("expected empty tags, got %v", is.Intents[1].Tags)
	}

	if is.Intents[2].ID != "int_1770020251" {
		t.Fatalf("unexpected id: %s", is.Intents[2].ID)
	}
	if is.Intents[2].Status != IntentResolved {
		t.Fatalf("expected resolved")
	}
	if is.Intents[2].Resolution.Method != "fixed" {
		t.Fatalf("unexpected resolution method: %s", is.Intents[2].Resolution.Method)
	}
}
