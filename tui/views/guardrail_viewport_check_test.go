package views

import (
	"strings"
	"testing"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	tea "github.com/charmbracelet/bubbletea"
)

func TestGuardrailSubRenderSmallViewport(t *testing.T) {
	m := NewConfigsModel()
	m.loading = false
	m.cfg = &client.ProjectConfig{
		Guardrails: []client.ConfigGuardrail{
			{ID: "gr-1", Paths: []string{"cmd/**", "internal/**"}, RequireAll: true, Action: "block", Reason: "must always run preflight"},
			{ID: "gr-2", Paths: []string{"**/*.go"}, RequireAll: false, Action: "warn", Reason: "lint before commit"},
		},
	}

	fieldOf := func(key string) int {
		for i, f := range configFields {
			if f.key == key {
				return i
			}
		}
		return -1
	}

	m.fieldIdx = fieldOf("guardrails")
	m.Update(specialKey(tea.KeyEnter))
	if !m.guardrailSub {
		t.Fatal("enter should expand guardrail sub-edit")
	}

	c := render.NewCanvas(100, 24)
	m.Render(c)
	rendered := c.String()

	subFields := []string{
		"id: gr-1",
		"paths: cmd/**, internal/**",
		"require_all: true",
		"action: block",
		"reason: must always run preflight",
	}
	for _, want := range subFields {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expanded guardrail sub-field %q not rendered in small viewport:\n%s", want, rendered)
		}
	}
	if !strings.Contains(rendered, "▸") {
		t.Fatalf("selected guardrail item should be marked with cursor:\n%s", rendered)
	}
	if strings.Contains(rendered, "↓ more") && strings.Index(rendered, "reason: must always run preflight") > strings.Index(rendered, "Guardrails:") {
		t.Logf("viewport scrolls remaining items below")
	}
}
