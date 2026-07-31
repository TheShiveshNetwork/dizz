package views

import (
	"testing"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMultiRuneInput(t *testing.T) {
	m := NewConfigsModel()
	m.loading = false
	m.cfg = &client.ProjectConfig{Include: []string{}}

	m.fieldIdx = 0
	for i, f := range configFields {
		if f.key == "include" {
			m.fieldIdx = i
			break
		}
	}
	m.Update(key("a"))
	if !m.editing {
		t.Fatal("expected add mode")
	}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("docs/**")}
	m.Update(msg)
	if m.inputBuf != "docs/**" {
		t.Fatalf("multi-rune input: got %q", m.inputBuf)
	}
	m.Update(specialKey(tea.KeyEnter))
	if len(m.cfg.Include) != 1 || m.cfg.Include[0] != "docs/**" {
		t.Fatalf("include: %v", m.cfg.Include)
	}
}
