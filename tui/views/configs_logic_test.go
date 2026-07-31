package views

import (
	"strconv"
	"testing"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	runes := []rune(s)
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
}

func specialKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func typeText(m *ConfigsModel, s string) {
	for _, r := range s {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestConfigsEditLogic(t *testing.T) {
	m := NewConfigsModel()
	m.loading = false
	m.cfg = &client.ProjectConfig{
		Version:       "1.0.0",
		ProjectName:   "dizz",
		Description:   "desc",
		Include:       []string{"**/*"},
		Exclude:       []string{"vendor/**", "node_modules/**"},
		Instructions:  nil,
		Guardrails:    []client.ConfigGuardrail{{ID: "gr-1", Paths: []string{"a"}, RequireAll: false, Action: "warn", Reason: "reason"}},
		SeverityScale: map[string]string{"1": "low"},
		Commands:      map[string]string{"build": "go build ./..."},
		AgentDefaults: &client.ConfigAgentDefaults{DefaultLens: "priority", MinSeverity: 1},
		Links:         map[string]string{"docs": "https://x"},
	}

	fieldOf := func(key string) int {
		for i, f := range configFields {
			if f.key == key {
				return i
			}
		}
		return -1
	}

	m.fieldIdx = fieldOf("exclude")
	m.Update(specialKey(tea.KeyTab))
	if m.listIdx != 1 {
		t.Fatalf("tab: want listIdx 1 got %d", m.listIdx)
	}

	m.Update(specialKey(tea.KeyEnter))
	if !m.editing || m.editIdx != 1 {
		t.Fatalf("enter on item: want editing editIdx=1, got editing=%v editIdx=%d", m.editing, m.editIdx)
	}
	if m.inputBuf != "node_modules/**" {
		t.Fatalf("edit buffer should prefill item text, got %q", m.inputBuf)
	}
	m.inputBuf = "**/out/**"
	m.Update(specialKey(tea.KeyEnter))
	if m.editing {
		t.Fatal("commit should exit editing")
	}
	if m.cfg.Exclude[1] != "**/out/**" {
		t.Fatalf("exclude[1] not updated: %q", m.cfg.Exclude[1])
	}

	m.Update(key("a"))
	if !m.editing || m.editIdx != -1 {
		t.Fatalf("a: want add mode, got editing=%v editIdx=%d", m.editing, m.editIdx)
	}
	typeText(m, "dist/**")
	m.Update(specialKey(tea.KeyEnter))
	if m.cfg.Exclude[len(m.cfg.Exclude)-1] != "dist/**" {
		t.Fatalf("append failed: %v", m.cfg.Exclude)
	}

	m.fieldIdx = fieldOf("commands")
	m.listIdx = 0
	m.Update(specialKey(tea.KeyEnter))
	if !m.editing || m.editIdx != 0 {
		t.Fatalf("map enter: editing=%v editIdx=%d", m.editing, m.editIdx)
	}
	if m.inputBuf != "build=go build ./..." {
		t.Fatalf("map edit buf: %q", m.inputBuf)
	}
	m.inputBuf = "build=go vet ./..."
	m.Update(specialKey(tea.KeyEnter))
	if m.cfg.Commands["build"] != "go vet ./..." {
		t.Fatalf("map value not updated: %v", m.cfg.Commands)
	}

	m.Update(key("a"))
	typeText(m, "lint=golangci-lint run")
	m.Update(specialKey(tea.KeyEnter))
	if m.cfg.Commands["lint"] != "golangci-lint run" {
		t.Fatalf("map add failed: %v", m.cfg.Commands)
	}

	m.fieldIdx = fieldOf("severity_scale")
	m.listIdx = 0
	m.Update(key("d"))
	if _, ok := m.cfg.SeverityScale["1"]; ok {
		t.Fatal("map delete failed")
	}

	m.fieldIdx = fieldOf("guardrails")
	m.Update(specialKey(tea.KeyEnter))
	if !m.guardrailSub {
		t.Fatal("guardrails enter should enter sub-edit")
	}
	for m.subIdx != 4 {
		m.Update(specialKey(tea.KeyDown))
	}
	m.Update(specialKey(tea.KeyEnter))
	if !m.editing || m.editSub != 4 {
		t.Fatalf("guardrail sub enter: editing=%v editSub=%d", m.editing, m.editSub)
	}
	m.inputBuf = "new reason"
	m.Update(specialKey(tea.KeyEnter))
	if m.cfg.Guardrails[0].Reason != "new reason" {
		t.Fatalf("guardrail reason not updated: %q", m.cfg.Guardrails[0].Reason)
	}
	m.Update(specialKey(tea.KeyEsc))
	if m.guardrailSub {
		t.Fatal("esc should exit guardrail sub-edit")
	}

	m.fieldIdx = fieldOf("agent_defaults")
	m.listIdx = 1
	m.Update(specialKey(tea.KeyEnter))
	m.inputBuf = "2"
	m.Update(specialKey(tea.KeyEnter))
	if m.cfg.AgentDefaults.MinSeverity != 2 {
		t.Fatalf("min_severity not updated: %d", m.cfg.AgentDefaults.MinSeverity)
	}

	m.fieldIdx = fieldOf("version")
	before := m.fieldIdx
	m.Update(specialKey(tea.KeyEnter))
	if m.editing {
		t.Fatal("version must not be editable")
	}
	if m.fieldIdx != before {
		t.Fatal("version enter should be a no-op")
	}

	m.fieldIdx = fieldOf("description")
	m.Update(specialKey(tea.KeyEnter))
	m.inputBuf = "new description"
	m.Update(specialKey(tea.KeyEnter))
	if m.cfg.Description != "new description" {
		t.Fatalf("description not updated: %q", m.cfg.Description)
	}

	c := render.NewCanvas(80, 24)
	m.Render(c)
}

func TestConfigsScrollingNavigation(t *testing.T) {
	m := NewConfigsModel()
	m.loading = false
	m.cfg = &client.ProjectConfig{
		Include: []string{"**/*"},
		Exclude: make([]string, 28),
	}
	for i := range m.cfg.Exclude {
		m.cfg.Exclude[i] = "pattern-" + strconv.Itoa(i)
	}

	fieldOf := func(key string) int {
		for i, f := range configFields {
			if f.key == key {
				return i
			}
		}
		return -1
	}

	m.fieldIdx = fieldOf("exclude")
	if !m.isItemList(configFields[m.fieldIdx]) {
		t.Fatal("exclude should be an item list")
	}

	// Moving down through all items reaches the last one
	for i := 0; i < 27; i++ {
		m.moveCursor(true)
	}
	if m.listIdx != 27 {
		t.Fatalf("down navigation: want listIdx 27 got %d", m.listIdx)
	}

	// Moving down past the last item exits to the next field (wraps to top)
	m.moveCursor(true)
	if m.fieldIdx != 0 || m.listIdx != 0 {
		t.Fatalf("down at end: want wrap to field 0, got field %d listIdx %d", m.fieldIdx, m.listIdx)
	}

	// Moving up from a text field wraps to the last field
	m.moveCursor(false)
	if m.fieldIdx != fieldOf("exclude") {
		t.Fatalf("up wrap: want exclude field, got %d", m.fieldIdx)
	}

	// Moving up at the top of the list exits to the previous field (include)
	m.moveCursor(false)
	if m.fieldIdx != fieldOf("include") {
		t.Fatalf("up at first item: want include field, got %d", m.fieldIdx)
	}

	// Within-field scrolling keeps the selected item visible in a short viewport
	m.fieldIdx = fieldOf("exclude")
	m.listIdx = 0
	for m.listIdx < 20 {
		m.moveCursor(true)
	}
	c := render.NewCanvas(80, 16)
	m.Render(c)
}
