package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	tea "github.com/charmbracelet/bubbletea"
)

func trimmed(c *render.Canvas) string {
	var sb strings.Builder
	for _, line := range strings.Split(strings.TrimRight(c.String(), "\n"), "\n") {
		sb.WriteString(strings.TrimRight(line, " "))
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func TestRenderPreview(t *testing.T) {
	cfg, err := client.LoadProjectConfig()
	if err != nil {
		t.Skip("no config: " + err.Error())
	}
	m := NewConfigsModel()
	m.loading = false
	m.cfg = cfg

	c := render.NewCanvas(100, 34)
	m.Render(c)
	fmt.Println("=== LIST VIEW ===")
	fmt.Println(trimmed(c))

	for i, f := range configFields {
		if f.key == "guardrails" {
			m.fieldIdx = i
			break
		}
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c = render.NewCanvas(100, 34)
	m.Render(c)
	fmt.Println("=== GUARDRAIL SUB-EDIT ===")
	fmt.Println(trimmed(c))

	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	for i, f := range configFields {
		if f.key == "commands" {
			m.fieldIdx = i
			break
		}
	}
	c = render.NewCanvas(100, 34)
	m.Render(c)
	fmt.Println("=== MAP VIEW ===")
	fmt.Println(trimmed(c))
}
