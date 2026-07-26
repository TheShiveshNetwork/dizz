package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type configField struct {
	label string
	key   string
	kind  string
}

var configFields = []configField{
	{label: "Project Name", key: "project_name", kind: "text"},
	{label: "Description", key: "description", kind: "text"},
	{label: "Instructions", key: "instructions", kind: "instructions"},
	{label: "Guardrails", key: "guardrails", kind: "guardrails"},
	{label: "Include", key: "include", kind: "list"},
	{label: "Exclude", key: "exclude", kind: "list"},
}

type ConfigsModel struct {
	cfg          *client.ProjectConfig
	loading      bool
	err          string
	fieldIdx     int
	editing      bool
	cursor       int
	inputBuf     string
	listIdx      int
	dirty        bool
	saving       bool
	saveMsg      string
	scrollOffset int
}

func NewConfigsModel() *ConfigsModel {
	return &ConfigsModel{loading: true}
}

func (m *ConfigsModel) Init() tea.Cmd {
	return m.refresh()
}

func (m *ConfigsModel) refresh() tea.Cmd {
	return func() tea.Msg {
		cfg, err := client.LoadProjectConfig()
		if err != nil {
			return configLoadedMsg{err: err.Error()}
		}
		return configLoadedMsg{cfg: cfg}
	}
}

type configLoadedMsg struct {
	cfg *client.ProjectConfig
	err string
}

type configSavedMsg struct {
	err string
}

func (m *ConfigsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case configLoadedMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.cfg = msg.cfg
			m.err = ""
		}

	case configSavedMsg:
		m.saving = false
		if msg.err != "" {
			m.err = msg.err
			m.saveMsg = ""
		} else {
			m.saveMsg = "Saved!"
			m.dirty = false
		}

	case ui.RefreshTick:
		return m, m.refresh()
	}

	return m, nil
}

func (m *ConfigsModel) fieldRows(f configField) int {
	fieldRows := 1
	if f.kind != "text" {
		items := m.getListItems(f.key)
		fieldRows += len(items)
		if fieldRows < 1 {
			fieldRows = 1
		}
	}
	return fieldRows
}

func (m *ConfigsModel) clampScroll(contentH int) {
	if m.fieldIdx < m.scrollOffset {
		m.scrollOffset = m.fieldIdx
	}
	for {
		avail := contentH
		if avail < 1 {
			avail = 1
		}
		rows := 0
		for i := m.scrollOffset; i < len(configFields); i++ {
			fr := m.fieldRows(configFields[i])
			if rows+fr > avail {
				break
			}
			rows += fr
		}
		if m.fieldIdx >= m.scrollOffset+rows && rows > 0 {
			m.scrollOffset++
		} else {
			break
		}
	}
}

func (m *ConfigsModel) hasMoreBelow(contentH int) bool {
	avail := contentH
	if avail < 1 {
		avail = 1
	}
	rows := 0
	for i := m.scrollOffset; i < len(configFields); i++ {
		fr := m.fieldRows(configFields[i])
		if rows+fr > avail {
			return true
		}
		rows += fr
	}
	return false
}

func (m *ConfigsModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.cfg == nil {
		return m, nil
	}

	if m.saving {
		return m, nil
	}

	if m.editing {
		return m.handleEditKey(msg)
	}

	switch msg.String() {
	case "up":
		m.fieldIdx--
		if m.fieldIdx < 0 {
			m.fieldIdx = len(configFields) - 1
		}
		m.listIdx = 0
	case "down":
		m.fieldIdx++
		if m.fieldIdx >= len(configFields) {
			m.fieldIdx = 0
		}
		m.listIdx = 0
	case "enter":
		f := configFields[m.fieldIdx]
		if f.kind == "text" {
			m.editing = true
			m.inputBuf = m.getFieldValue(f.key)
			m.cursor = len(m.inputBuf)
		} else if f.kind == "list" || f.kind == "instructions" || f.kind == "guardrails" {
			m.editing = true
			m.inputBuf = ""
			m.cursor = 0
		}
	case "e":
		f := configFields[m.fieldIdx]
		if f.kind == "list" {
			m.editing = true
			m.inputBuf = ""
			m.cursor = 0
		}
	case "d":
		f := configFields[m.fieldIdx]
		if f.kind == "list" {
			items := m.getListItems(f.key)
			if m.listIdx < len(items) {
				m.removeListItem(f.key, m.listIdx)
				if m.listIdx >= len(m.getListItems(f.key)) {
					m.listIdx = len(m.getListItems(f.key)) - 1
				}
				if m.listIdx < 0 {
					m.listIdx = 0
				}
				m.dirty = true
			}
		} else if f.kind == "instructions" {
			items := m.cfg.ParseInstructions()
			if m.listIdx < len(items) {
				remaining := append(items[:m.listIdx:m.listIdx], items[m.listIdx+1:]...)
				var newInstructions []byte
				if len(remaining) > 0 {
					newInstructions, _ = jsonMarshal(remaining)
				}
				m.cfg.Instructions = newInstructions
				if m.listIdx >= len(remaining) {
					m.listIdx = len(remaining) - 1
				}
				if m.listIdx < 0 {
					m.listIdx = 0
				}
				m.dirty = true
			}
		} else if f.kind == "guardrails" {
			if m.listIdx < len(m.cfg.Guardrails) {
				m.cfg.Guardrails = append(m.cfg.Guardrails[:m.listIdx], m.cfg.Guardrails[m.listIdx+1:]...)
				if m.listIdx >= len(m.cfg.Guardrails) {
					m.listIdx = len(m.cfg.Guardrails) - 1
				}
				if m.listIdx < 0 {
					m.listIdx = 0
				}
				m.dirty = true
			}
		}
	case "tab":
		f := configFields[m.fieldIdx]
		if f.kind == "list" || f.kind == "instructions" || f.kind == "guardrails" {
			items := m.getListItems(f.key)
			if len(items) > 0 {
				m.listIdx++
				if m.listIdx >= len(items) {
					m.listIdx = 0
				}
			}
		}
	case "shift+tab":
		f := configFields[m.fieldIdx]
		if f.kind == "list" || f.kind == "instructions" || f.kind == "guardrails" {
			items := m.getListItems(f.key)
			if len(items) > 0 {
				m.listIdx--
				if m.listIdx < 0 {
					m.listIdx = len(items) - 1
				}
			}
		}
	case "s", "ctrl+s":
		if m.dirty {
			m.saving = true
			m.saveMsg = ""
			return m, m.saveConfig()
		}
	case "r":
		return m, m.refresh()
	}

	return m, nil
}

func (m *ConfigsModel) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := configFields[m.fieldIdx]

	switch msg.String() {
	case "esc":
		m.editing = false
		m.inputBuf = ""
		return m, nil

	case "ctrl+s":
		if f.kind == "text" {
			m.setFieldValue(f.key, m.inputBuf)
		} else if f.kind == "list" && m.inputBuf != "" {
			m.appendListItem(f.key, m.inputBuf)
		} else if f.kind == "instructions" && m.inputBuf != "" {
			existing := m.cfg.ParseInstructions()
			existing = append(existing, client.ConfigInstruction{Rule: m.inputBuf})
			newData, _ := jsonMarshal(existing)
			m.cfg.Instructions = newData
		} else if f.kind == "guardrails" && m.inputBuf != "" {
			parts := strings.SplitN(m.inputBuf, "|", 2)
			action := strings.TrimSpace(parts[0])
			reason := ""
			if len(parts) > 1 {
				reason = strings.TrimSpace(parts[1])
			}
			if action == "" {
				action = "warn"
			}
			m.cfg.Guardrails = append(m.cfg.Guardrails, client.ConfigGuardrail{
				Action: action,
				Reason: reason,
			})
		}
		m.dirty = true
		m.editing = false
		m.inputBuf = ""
		m.saving = true
		m.saveMsg = ""
		return m, m.saveConfig()

	case "enter":
		if f.kind == "text" {
			m.setFieldValue(f.key, m.inputBuf)
			m.dirty = true
			m.editing = false
			m.inputBuf = ""
		} else if f.kind == "list" {
			if m.inputBuf != "" {
				m.appendListItem(f.key, m.inputBuf)
				m.dirty = true
			}
			m.editing = false
			m.inputBuf = ""
		} else if f.kind == "instructions" {
			if m.inputBuf != "" {
				existing := m.cfg.ParseInstructions()
				existing = append(existing, client.ConfigInstruction{Rule: m.inputBuf})
				newData, _ := jsonMarshal(existing)
				m.cfg.Instructions = newData
				m.dirty = true
			}
			m.editing = false
			m.inputBuf = ""
		} else if f.kind == "guardrails" {
			if m.inputBuf != "" {
				parts := strings.SplitN(m.inputBuf, "|", 2)
				action := strings.TrimSpace(parts[0])
				reason := ""
				if len(parts) > 1 {
					reason = strings.TrimSpace(parts[1])
				}
				if action == "" {
					action = "warn"
				}
				m.cfg.Guardrails = append(m.cfg.Guardrails, client.ConfigGuardrail{
					Action: action,
					Reason: reason,
				})
				m.dirty = true
			}
			m.editing = false
			m.inputBuf = ""
		}

	case "up":
		if f.kind == "text" {
			m.setFieldValue(f.key, m.inputBuf)
			m.dirty = true
			m.editing = false
			m.inputBuf = ""
			m.fieldIdx--
			if m.fieldIdx < 0 {
				m.fieldIdx = len(configFields) - 1
			}
			m.listIdx = 0
		}

	case "down":
		if f.kind == "text" {
			m.setFieldValue(f.key, m.inputBuf)
			m.dirty = true
			m.editing = false
			m.inputBuf = ""
			m.fieldIdx++
			if m.fieldIdx >= len(configFields) {
				m.fieldIdx = 0
			}
			m.listIdx = 0
		}

	case "left":
		if m.cursor > 0 {
			m.cursor--
		}

	case "right":
		if m.cursor < len(m.inputBuf) {
			m.cursor++
		}

	case "home":
		m.cursor = 0

	case "end":
		m.cursor = len(m.inputBuf)

	case "backspace":
		if m.cursor > 0 {
			m.inputBuf = m.inputBuf[:m.cursor-1] + m.inputBuf[m.cursor:]
			m.cursor--
		}

	default:
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			before := m.inputBuf[:m.cursor]
			after := m.inputBuf[m.cursor:]
			m.inputBuf = before + string(ch) + after
			m.cursor++
		}
	}

	return m, nil
}

func (m *ConfigsModel) getFieldValue(key string) string {
	switch key {
	case "project_name":
		return m.cfg.ProjectName
	case "description":
		return m.cfg.Description
	}
	return ""
}

func (m *ConfigsModel) setFieldValue(key, val string) {
	switch key {
	case "project_name":
		m.cfg.ProjectName = val
	case "description":
		m.cfg.Description = val
	}
}

func (m *ConfigsModel) getListItems(key string) []string {
	switch key {
	case "include":
		return m.cfg.Include
	case "exclude":
		return m.cfg.Exclude
	case "instructions":
		var items []string
		for _, inst := range m.cfg.ParseInstructions() {
			if inst.Scope != "" {
				items = append(items, inst.Rule+" ["+inst.Scope+"]")
			} else {
				items = append(items, inst.Rule)
			}
		}
		return items
	case "guardrails":
		var items []string
		for _, g := range m.cfg.Guardrails {
			items = append(items, fmt.Sprintf("[%s] %s", g.Action, g.Reason))
		}
		return items
	}
	return nil
}

func (m *ConfigsModel) removeListItem(key string, idx int) {
	switch key {
	case "include":
		if idx < len(m.cfg.Include) {
			m.cfg.Include = append(m.cfg.Include[:idx], m.cfg.Include[idx+1:]...)
		}
	case "exclude":
		if idx < len(m.cfg.Exclude) {
			m.cfg.Exclude = append(m.cfg.Exclude[:idx], m.cfg.Exclude[idx+1:]...)
		}
	}
}

func (m *ConfigsModel) appendListItem(key, val string) {
	switch key {
	case "include":
		m.cfg.Include = append(m.cfg.Include, val)
	case "exclude":
		m.cfg.Exclude = append(m.cfg.Exclude, val)
	}
}

func (m *ConfigsModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		err := client.SaveProjectConfig(m.cfg)
		return configSavedMsg{err: errStr(err)}
	}
}

func errStr(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (m *ConfigsModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	if m.loading {
		c.SetContent((w-len("Loading config..."))/2, h/2, render.StyleMuted, "Loading config...")
		return
	}

	if m.err != "" {
		c.SetContent(2, 2, render.StyleError, m.err)
		return
	}

	if m.cfg == nil {
		c.SetContent(2, 2, render.StyleMuted, "No config found")
		return
	}

	header := "Project Configuration"
	c.SetContent(2, 0, render.StyleHighlight.Bold(true), header)

	if m.saveMsg != "" {
		c.SetContent(w-2-len(m.saveMsg), 0, render.StyleSuccess.Bold(true), m.saveMsg)
	}

	hintY := 1
	if m.dirty && !m.editing {
		c.SetContent(2, hintY, render.StyleWarning, "Unsaved changes")
		btnText := " [ Ctrl+S ] Save "
		btnStyle := ui.StyleBtnFocusSuccess
		btnX := w - 2 - runewidth.StringWidth(btnText)
		if btnX > 2+len("Unsaved changes")+2 {
			c.SetContent(btnX, hintY, btnStyle, btnText)
		}
	} else {
		hint := "up/down=nav enter=edit s=save r=refresh"
		f := configFields[m.fieldIdx]
		if f.kind == "list" || f.kind == "instructions" || f.kind == "guardrails" {
			hint = "up/down=nav tab=next-item e=add d=delete enter=edit s=save r=refresh"
		}
		c.SetContent(2, hintY, render.StyleDim, hint)
	}

	separatorY := 2
	if separatorY < h {
		for x := 0; x < w; x++ {
			c.SetCell(x, separatorY, render.StyleMuted, '\u2500')
		}
	}

	contentTop := 3
	contentH := h - contentTop - 1
	if contentH < 1 {
		contentH = 1
	}

	m.clampScroll(contentH)

	y := contentTop
	for fi := m.scrollOffset; fi < len(configFields); fi++ {
		if y >= h-1 {
			break
		}

		f := configFields[fi]
		selected := fi == m.fieldIdx
		labelStyle := render.StyleDefault
		if selected {
			labelStyle = render.StyleHighlight.Bold(true)
		}

		c.SetContent(2, y, labelStyle, f.label+":")
		y++

		switch f.kind {
		case "text":
			val := m.getFieldValue(f.key)
			if val == "" {
				val = "(empty)"
			}
			valStyle := render.StyleDefault
			if selected && !m.editing {
				valStyle = render.StyleInfo
			}
			maxValW := w - 6
			if runewidth.StringWidth(val) > maxValW {
				val = val[:maxValW-3] + "..."
			}
			if m.editing && selected {
				display := m.inputBuf
				if runewidth.StringWidth(display) > maxValW {
					display = display[:maxValW]
				}
				c.SetContent(4, y, ui.StyleBtnFocus, display)
				cursorX := 4 + m.cursor
				if cursorX >= 4+maxValW {
					cursorX = 4 + maxValW - 1
				}
				if cursorX < w-1 {
					c.SetCell(cursorX, y, render.StyleHighlight, '_')
				}
			} else {
				c.SetContent(4, y, valStyle, val)
			}
			y++

		case "list", "instructions", "guardrails":
			items := m.getListItems(f.key)
			if len(items) == 0 {
				c.SetContent(4, y, render.StyleMuted, "(none)")
				y++
			} else {
				for ii, item := range items {
					if y >= h-1 {
						break
					}
					itemStyle := render.StyleDefault
					prefix := "  "
					if selected && ii == m.listIdx {
						itemStyle = render.StyleHighlight
						prefix = "▸ "
					}
					display := prefix + item
					maxW := w - 6
					if runewidth.StringWidth(display) > maxW {
						display = display[:maxW-3] + "..."
					}
					c.SetContent(4, y, itemStyle, display)
					y++
				}
			}
			if m.editing && selected && y < h-1 {
				prompt := "  + "
				display := m.inputBuf
				maxW := w - 6
				c.SetContent(4, y, ui.StyleBtnFocus, prompt+display)
				cursorX := 4 + len(prompt) + m.cursor
				if cursorX >= 4+maxW {
					cursorX = 4 + maxW - 1
				}
				if cursorX < w-1 {
					c.SetCell(cursorX, y, render.StyleHighlight, '_')
				}
				y++
			}
			y++
		}
	}

	if m.scrollOffset > 0 {
		scrollHint := fmt.Sprintf("↑ %d more", m.scrollOffset)
		c.SetContent(w-2-len(scrollHint), contentTop, render.StyleDim, scrollHint)
	}
	if m.hasMoreBelow(contentH) {
		scrollHint := "↓ more"
		c.SetContent(w-2-len(scrollHint), h-2, render.StyleDim, scrollHint)
	}
}

func (m *ConfigsModel) View() string { return "" }

func (m *ConfigsModel) InputMode() bool { return m.editing }

func (m *ConfigsModel) IsModalActive() bool { return false }
