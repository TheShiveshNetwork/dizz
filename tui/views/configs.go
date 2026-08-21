package views

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
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
	{label: "Version", key: "version", kind: "version"},
	{label: "Project Name", key: "project_name", kind: "text"},
	{label: "Description", key: "description", kind: "text"},
	{label: "Commands", key: "commands", kind: "map"},
	{label: "Instructions", key: "instructions", kind: "instructions"},
	{label: "Guardrails", key: "guardrails", kind: "guardrails"},
	{label: "Severity Scale", key: "severity_scale", kind: "map"},
	{label: "Agent Defaults", key: "agent_defaults", kind: "agent_defaults"},
	{label: "Links", key: "links", kind: "map"},
	{label: "Include", key: "include", kind: "list"},
	{label: "Exclude", key: "exclude", kind: "list"},
}

var guardrailSubLabels = []string{"id", "paths", "require_all", "action", "reason"}

var instrScopeRe = regexp.MustCompile(`^(.*?)\s*\[([^\]]+)\]\s*$`)

type ConfigsModel struct {
	cfg          *client.ProjectConfig
	loading      bool
	err          string
	fieldIdx     int
	editing      bool
	editIdx      int
	editSub      int
	cursor       int
	inputBuf     string
	listIdx      int
	guardrailSub bool
	subIdx       int
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
			m.dirty = false
			m.editing = false
			m.guardrailSub = false
			m.inputBuf = ""
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

func (m *ConfigsModel) isItemList(f configField) bool {
	switch f.kind {
	case "list", "map", "instructions", "guardrails", "agent_defaults":
		return true
	}
	return false
}

func (m *ConfigsModel) itemCount(key string) int {
	switch key {
	case "include", "exclude":
		return len(m.listFor(key))
	case "instructions":
		return len(m.cfg.ParseInstructions())
	case "guardrails":
		return len(m.cfg.Guardrails)
	case "agent_defaults":
		return 2
	case "commands", "severity_scale", "links":
		ref := m.mapRef(key)
		if ref == nil || *ref == nil {
			return 0
		}
		return len(*ref)
	}
	return 0
}

func (m *ConfigsModel) fieldRows(f configField, avail int) int {
	var rows int
	switch f.kind {
	case "version", "text":
		rows = 2
	default:
		rows = 2 + m.itemCount(f.key)
		if f.kind == "guardrails" && m.guardrailSub && configFields[m.fieldIdx].key == f.key {
			rows += 5
		}
	}
	if rows > avail {
		rows = avail
	}
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m *ConfigsModel) clampScroll(contentH int) {
	if contentH < 1 {
		contentH = 1
	}
	if m.fieldIdx < m.scrollOffset {
		m.scrollOffset = m.fieldIdx
		return
	}
	rows := 0
	lastVisible := m.scrollOffset
	for i := m.scrollOffset; i < len(configFields); i++ {
		fr := m.fieldRows(configFields[i], contentH-rows)
		if rows+fr > contentH {
			break
		}
		rows += fr
		lastVisible = i
	}
	if m.fieldIdx > lastVisible {
		m.scrollOffset = m.fieldIdx
	}
}

func (m *ConfigsModel) hasMoreBelow(contentH int) bool {
	if contentH < 1 {
		contentH = 1
	}
	rows := 0
	for i := m.scrollOffset; i < len(configFields); i++ {
		fr := m.fieldRows(configFields[i], contentH-rows)
		if rows+fr > contentH {
			return true
		}
		rows += fr
	}
	return false
}

func (m *ConfigsModel) moveField(down bool) {
	if down {
		m.fieldIdx++
		if m.fieldIdx >= len(configFields) {
			m.fieldIdx = 0
		}
	} else {
		m.fieldIdx--
		if m.fieldIdx < 0 {
			m.fieldIdx = len(configFields) - 1
		}
	}
	m.listIdx = 0
	m.guardrailSub = false
	m.subIdx = 0
}

func (m *ConfigsModel) moveCursor(down bool) {
	f := configFields[m.fieldIdx]
	if !m.isItemList(f) {
		m.moveField(down)
		return
	}
	n := m.itemCount(f.key)
	if down {
		if m.listIdx < n-1 {
			m.listIdx++
			return
		}
		m.moveField(true)
		return
	}
	if m.listIdx > 0 {
		m.listIdx--
		return
	}
	m.moveField(false)
	pf := configFields[m.fieldIdx]
	if m.isItemList(pf) {
		m.listIdx = m.itemCount(pf.key) - 1
		if m.listIdx < 0 {
			m.listIdx = 0
		}
	}
}

func (m *ConfigsModel) moveList(back bool) {
	n := m.itemCount(configFields[m.fieldIdx].key)
	if n == 0 {
		return
	}
	if back {
		m.listIdx--
		if m.listIdx < 0 {
			m.listIdx = n - 1
		}
	} else {
		m.listIdx++
		if m.listIdx >= n {
			m.listIdx = 0
		}
	}
}

func (m *ConfigsModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.cfg == nil || m.saving {
		return m, nil
	}
	if m.editing {
		return m.handleEditKey(msg)
	}

	f := configFields[m.fieldIdx]

	if f.kind == "guardrails" && m.guardrailSub {
		switch msg.String() {
		case "up":
			m.subIdx--
			if m.subIdx < 0 {
				m.subIdx = len(guardrailSubLabels) - 1
			}
		case "down":
			m.subIdx++
			if m.subIdx >= len(guardrailSubLabels) {
				m.subIdx = 0
			}
		case "esc":
			m.guardrailSub = false
			m.subIdx = 0
		case "enter":
			if m.subIdx == 2 {
				m.toggleGuardrailRequireAll()
			} else {
				m.editing = true
				m.editIdx = m.listIdx
				m.editSub = m.subIdx
				m.inputBuf = m.guardrailSubValue(m.subIdx)
				m.cursor = len(m.inputBuf)
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

	switch msg.String() {
	case "up", "down":
		m.moveCursor(msg.String() == "down")

	case "enter":
		switch f.kind {
		case "version":
		case "text":
			m.editing = true
			m.editIdx = -1
			m.editSub = -1
			m.inputBuf = m.getFieldValue(f.key)
			m.cursor = len(m.inputBuf)
		case "list", "map", "instructions", "agent_defaults":
			m.editing = true
			m.editSub = -1
			if m.itemCount(f.key) == 0 {
				m.editIdx = -1
				m.inputBuf = ""
			} else {
				m.editIdx = m.listIdx
				m.inputBuf = m.itemEditBuf(f.key, m.listIdx)
			}
			m.cursor = len(m.inputBuf)
		case "guardrails":
			if len(m.cfg.Guardrails) > 0 {
				m.guardrailSub = true
				m.subIdx = 0
			}
		}

	case "a":
		switch f.kind {
		case "list", "map", "instructions":
			m.editing = true
			m.editIdx = -1
			m.editSub = -1
			m.inputBuf = ""
			m.cursor = 0
		case "guardrails":
			m.cfg.Guardrails = append(m.cfg.Guardrails, client.ConfigGuardrail{Action: "warn"})
			m.listIdx = len(m.cfg.Guardrails) - 1
			m.guardrailSub = true
			m.subIdx = 0
			m.dirty = true
		}

	case "d":
		m.deleteItem(f)

	case "tab", "shift+tab":
		if m.isItemList(f) {
			m.moveList(msg.String() == "shift+tab")
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
	switch msg.String() {
	case "esc":
		m.editing = false
		m.inputBuf = ""
		m.editIdx = -1
		m.editSub = -1

	case "enter", "ctrl+s":
		if m.commitEdit() && msg.String() == "ctrl+s" && m.dirty {
			m.saving = true
			m.saveMsg = ""
			return m, m.saveConfig()
		}

	case "up", "down":
		if m.commitEdit() {
			m.moveCursor(msg.String() == "down")
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

	case "delete":
		if m.cursor < len(m.inputBuf) {
			m.inputBuf = m.inputBuf[:m.cursor] + m.inputBuf[m.cursor+1:]
		}

	default:
		for _, r := range msg.Runes {
			m.inputBuf = m.inputBuf[:m.cursor] + string(r) + m.inputBuf[m.cursor:]
			m.cursor++
		}
	}

	return m, nil
}

func (m *ConfigsModel) commitEdit() bool {
	f := configFields[m.fieldIdx]
	val := m.inputBuf
	var ok bool
	switch f.kind {
	case "text":
		m.setFieldValue(f.key, val)
		m.dirty = true
		ok = true
	case "list":
		ok = m.commitListEdit(val)
	case "map":
		ok = m.commitMapEdit(val)
	case "instructions":
		ok = m.commitInstructionEdit(val)
	case "agent_defaults":
		ok = m.commitAgentDefaultsEdit(val)
	case "guardrails":
		ok = m.commitGuardrailSubEdit(val)
	}
	if ok {
		m.editing = false
		m.inputBuf = ""
		m.editIdx = -1
		m.editSub = -1
	}
	return ok
}

func (m *ConfigsModel) commitListEdit(val string) bool {
	key := configFields[m.fieldIdx].key
	if m.editIdx < 0 {
		if val == "" {
			return false
		}
		m.appendListItem(key, val)
		m.listIdx = m.itemCount(key) - 1
	} else if m.editIdx < m.itemCount(key) {
		m.setListItem(key, m.editIdx, val)
		m.listIdx = m.editIdx
	} else {
		return false
	}
	m.dirty = true
	return true
}

func (m *ConfigsModel) commitMapEdit(val string) bool {
	key := configFields[m.fieldIdx].key
	ref := m.mapRef(key)
	if ref == nil {
		return false
	}
	if *ref == nil {
		*ref = map[string]string{}
	}
	parts := strings.SplitN(val, "=", 2)
	k := strings.TrimSpace(parts[0])
	if k == "" {
		return false
	}
	v := ""
	if len(parts) > 1 {
		v = strings.TrimSpace(parts[1])
	}
	if m.editIdx >= 0 {
		keys := sortedMapKeys(*ref)
		if m.editIdx >= len(keys) {
			return false
		}
		delete(*ref, keys[m.editIdx])
	}
	(*ref)[k] = v
	if idx := mapKeyIndex(*ref, k); idx >= 0 {
		m.listIdx = idx
	}
	m.dirty = true
	return true
}

func (m *ConfigsModel) commitInstructionEdit(val string) bool {
	existing := m.cfg.ParseInstructions()
	rule, scope := parseInstructionText(val)
	if rule == "" && m.editIdx < 0 {
		return false
	}
	if m.editIdx < 0 {
		existing = append(existing, client.ConfigInstruction{Rule: rule, Scope: scope})
	} else if m.editIdx < len(existing) {
		existing[m.editIdx] = client.ConfigInstruction{Rule: rule, Scope: scope}
	} else {
		return false
	}
	data, err := marshalInstructions(existing)
	if err != nil {
		return false
	}
	m.cfg.Instructions = data
	m.dirty = true
	return true
}

func (m *ConfigsModel) commitAgentDefaultsEdit(val string) bool {
	ad := m.cfg.AgentDefaults
	if ad == nil {
		ad = &client.ConfigAgentDefaults{}
		m.cfg.AgentDefaults = ad
	}
	switch m.editIdx {
	case 0:
		ad.DefaultLens = strings.TrimSpace(val)
	case 1:
		n, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return false
		}
		ad.MinSeverity = n
	default:
		return false
	}
	m.dirty = true
	return true
}

func (m *ConfigsModel) commitGuardrailSubEdit(val string) bool {
	if m.listIdx >= len(m.cfg.Guardrails) || m.editSub < 0 {
		return false
	}
	g := &m.cfg.Guardrails[m.listIdx]
	switch m.editSub {
	case 0:
		g.ID = strings.TrimSpace(val)
	case 1:
		var paths []string
		for _, p := range strings.Split(val, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				paths = append(paths, p)
			}
		}
		g.Paths = paths
	case 3:
		g.Action = strings.TrimSpace(val)
	case 4:
		g.Reason = val
	default:
		return false
	}
	m.dirty = true
	return true
}

func (m *ConfigsModel) toggleGuardrailRequireAll() {
	if m.listIdx >= len(m.cfg.Guardrails) {
		return
	}
	m.cfg.Guardrails[m.listIdx].RequireAll = !m.cfg.Guardrails[m.listIdx].RequireAll
	m.dirty = true
}

func (m *ConfigsModel) deleteItem(f configField) {
	key := f.key
	switch f.kind {
	case "list":
		items := m.listFor(key)
		if m.listIdx >= len(items) {
			return
		}
		m.setListItems(key, append(items[:m.listIdx], items[m.listIdx+1:]...))
	case "map":
		keys := m.mapKeys(key)
		if m.listIdx >= len(keys) {
			return
		}
		ref := m.mapRef(key)
		delete(*ref, keys[m.listIdx])
	case "instructions":
		items := m.cfg.ParseInstructions()
		if m.listIdx >= len(items) {
			return
		}
		remaining := append(items[:m.listIdx:m.listIdx], items[m.listIdx+1:]...)
		data, err := marshalInstructions(remaining)
		if err != nil {
			return
		}
		m.cfg.Instructions = data
	case "guardrails":
		if m.listIdx >= len(m.cfg.Guardrails) {
			return
		}
		m.cfg.Guardrails = append(m.cfg.Guardrails[:m.listIdx], m.cfg.Guardrails[m.listIdx+1:]...)
		m.guardrailSub = false
		m.subIdx = 0
	default:
		return
	}
	m.dirty = true
	m.clampListIdx(key)
}

func (m *ConfigsModel) clampListIdx(key string) {
	n := m.itemCount(key)
	if n == 0 {
		m.listIdx = 0
		return
	}
	if m.listIdx >= n {
		m.listIdx = n - 1
	}
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

func (m *ConfigsModel) listFor(key string) []string {
	switch key {
	case "include":
		return m.cfg.Include
	case "exclude":
		return m.cfg.Exclude
	}
	return nil
}

func (m *ConfigsModel) setListItems(key string, items []string) {
	switch key {
	case "include":
		m.cfg.Include = items
	case "exclude":
		m.cfg.Exclude = items
	}
}

func (m *ConfigsModel) setListItem(key string, idx int, val string) {
	switch key {
	case "include":
		if idx < len(m.cfg.Include) {
			m.cfg.Include[idx] = val
		}
	case "exclude":
		if idx < len(m.cfg.Exclude) {
			m.cfg.Exclude[idx] = val
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

func (m *ConfigsModel) mapRef(key string) *map[string]string {
	switch key {
	case "commands":
		return &m.cfg.Commands
	case "severity_scale":
		return &m.cfg.SeverityScale
	case "links":
		return &m.cfg.Links
	}
	return nil
}

func (m *ConfigsModel) mapKeys(key string) []string {
	ref := m.mapRef(key)
	if ref == nil || *ref == nil {
		return nil
	}
	return sortedMapKeys(*ref)
}

func (m *ConfigsModel) itemEditBuf(key string, i int) string {
	switch key {
	case "include", "exclude":
		items := m.listFor(key)
		if i >= 0 && i < len(items) {
			return items[i]
		}
	case "instructions":
		items := m.cfg.ParseInstructions()
		if i >= 0 && i < len(items) {
			if items[i].Scope != "" {
				return items[i].Rule + " [" + items[i].Scope + "]"
			}
			return items[i].Rule
		}
	case "commands", "severity_scale", "links":
		keys := m.mapKeys(key)
		ref := m.mapRef(key)
		if i >= 0 && i < len(keys) {
			return keys[i] + "=" + (*ref)[keys[i]]
		}
	case "agent_defaults":
		ad := m.cfg.AgentDefaults
		if ad == nil {
			ad = &client.ConfigAgentDefaults{}
		}
		if i == 0 {
			return ad.DefaultLens
		}
		if i == 1 {
			return strconv.Itoa(ad.MinSeverity)
		}
	}
	return ""
}

func (m *ConfigsModel) itemDisplay(key string, i int) string {
	if key == "agent_defaults" {
		ad := m.cfg.AgentDefaults
		if ad == nil {
			ad = &client.ConfigAgentDefaults{}
		}
		if i == 0 {
			return "default_lens: " + ad.DefaultLens
		}
		if i == 1 {
			return "min_severity: " + strconv.Itoa(ad.MinSeverity)
		}
		return ""
	}
	if key == "guardrails" {
		if i < 0 || i >= len(m.cfg.Guardrails) {
			return ""
		}
		g := m.cfg.Guardrails[i]
		s := fmt.Sprintf("[%s] %s", g.Action, g.Reason)
		if g.ID != "" {
			s = fmt.Sprintf("[%s] %s (%s)", g.Action, g.Reason, g.ID)
		}
		return s
	}
	return m.itemEditBuf(key, i)
}

func (m *ConfigsModel) guardrailSubValue(idx int) string {
	if m.listIdx >= len(m.cfg.Guardrails) {
		return ""
	}
	g := m.cfg.Guardrails[m.listIdx]
	switch idx {
	case 0:
		return g.ID
	case 1:
		return strings.Join(g.Paths, ", ")
	case 2:
		return strconv.FormatBool(g.RequireAll)
	case 3:
		return g.Action
	case 4:
		return g.Reason
	}
	return ""
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

func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mapKeyIndex(m map[string]string, target string) int {
	for i, k := range sortedMapKeys(m) {
		if k == target {
			return i
		}
	}
	return -1
}

func parseInstructionText(s string) (rule, scope string) {
	trimmed := strings.TrimSpace(s)
	if m := instrScopeRe.FindStringSubmatch(trimmed); m != nil {
		return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	}
	return trimmed, ""
}

func marshalInstructions(insts []client.ConfigInstruction) ([]byte, error) {
	arr := make([]interface{}, 0, len(insts))
	for _, inst := range insts {
		if inst.Scope == "" {
			arr = append(arr, inst.Rule)
		} else {
			arr = append(arr, inst)
		}
	}
	return json.Marshal(arr)
}

func (m *ConfigsModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	if m.loading {
		c.SetContent((w-len("Loading config..."))/2, h/2, render.StyleMuted, "Loading config...")
		return
	}

	if m.cfg == nil {
		if m.err != "" {
			c.SetContent(2, 2, render.StyleError, m.err)
		} else {
			c.SetContent(2, 2, render.StyleMuted, "No config found")
		}
		return
	}

	header := "Project Configuration"
	c.SetContent(2, 0, render.StyleHighlight.Bold(true), header)

	if m.saveMsg != "" {
		c.SetContent(w-2-len(m.saveMsg), 0, render.StyleSuccess.Bold(true), m.saveMsg)
	}

	hintY := 1
	switch {
	case m.err != "":
		c.SetContent(2, hintY, render.StyleError, runewidth.Truncate(m.err, w-4, "..."))
	case m.dirty && !m.editing:
		c.SetContent(2, hintY, render.StyleWarning, "Unsaved changes")
		btnText := " [ Ctrl+S ] Save "
		btnStyle := ui.StyleBtnFocusSuccess
		btnX := w - 2 - runewidth.StringWidth(btnText)
		if btnX > 2+len("Unsaved changes")+2 {
			c.SetContent(btnX, hintY, btnStyle, btnText)
		}
	default:
		f := configFields[m.fieldIdx]
		hint := "up/down=nav enter=edit s=save r=refresh"
		if f.kind == "version" {
			hint = "up/down=nav (read-only) r=refresh"
		} else if m.isItemList(f) {
			hint = "up/down=item enter=edit a=add d=delete s=save r=refresh"
		}
		if f.kind == "guardrails" && m.guardrailSub {
			hint = "up/down=field enter=edit a=add d=delete esc=back s=save"
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
		rows := m.renderField(c, configFields[fi], fi == m.fieldIdx, y)
		y += rows
	}

	if m.scrollOffset > 0 {
		scrollHint := fmt.Sprintf("↑ %d more", m.scrollOffset)
		c.SetContent(w-2-runewidth.StringWidth(scrollHint), contentTop, render.StyleDim, scrollHint)
	}
	if m.hasMoreBelow(contentH) {
		scrollHint := "↓ more"
		c.SetContent(w-2-runewidth.StringWidth(scrollHint), h-2, render.StyleDim, scrollHint)
	}
}

func (m *ConfigsModel) renderField(c *render.Canvas, f configField, selected bool, y int) int {
	lines := 0
	labelStyle := render.StyleDefault
	if selected {
		labelStyle = render.StyleHighlight.Bold(true)
	}
	c.SetContent(2, y, labelStyle, f.label+":")
	y++
	lines++

	switch f.kind {
	case "version":
		valStyle := render.StyleMuted
		if selected {
			valStyle = render.StyleInfo
		}
		c.SetContent(4, y, valStyle, runewidth.Truncate(m.cfg.Version, c.Width()-6, "..."))
		return lines + 1

	case "text":
		val := m.getFieldValue(f.key)
		if val == "" {
			val = "(empty)"
		}
		valStyle := render.StyleDefault
		if selected && !m.editing {
			valStyle = render.StyleInfo
		}
		maxValW := c.Width() - 6
		if m.editing && selected {
			display := runewidth.Truncate(m.inputBuf, maxValW, "")
			c.SetContent(4, y, ui.StyleBtnFocus, display)
			cursorX := 4 + runewidth.StringWidth(display)
			if cursorX >= 4+maxValW {
				cursorX = 4 + maxValW - 1
			}
			if cursorX < c.Width()-1 {
				c.SetCell(cursorX, y, render.StyleHighlight, '_')
			}
		} else {
			c.SetContent(4, y, valStyle, runewidth.Truncate(val, maxValW, "..."))
		}
		return lines + 1

	case "list", "map", "instructions", "agent_defaults":
		n := m.itemCount(f.key)
		avail := c.Height() - y - 1
		if avail < 1 {
			avail = 1
		}
		itemStart := 0
		if selected && n > avail {
			itemStart = m.listIdx - avail + 1
			if itemStart < 0 {
				itemStart = 0
			}
		}
		if n == 0 {
			c.SetContent(4, y, render.StyleMuted, "(none)")
			y++
			lines++
		} else {
			for i := itemStart; i < n && i-itemStart < avail; i++ {
				if y >= c.Height()-1 {
					break
				}
				itemStyle := render.StyleDefault
				prefix := "  "
				if selected && i == m.listIdx {
					itemStyle = render.StyleHighlight
					prefix = "▸ "
				}
				display := runewidth.Truncate(prefix+m.itemDisplay(f.key, i), c.Width()-6, "...")
				c.SetContent(4, y, itemStyle, display)
				y++
				lines++
			}
		}
		if m.editing && selected && y < c.Height()-1 {
			prompt := "  "
			if m.editIdx < 0 {
				prompt = "  + "
			}
			m.renderInputLine(c, 4, y, c.Width()-6, prompt)
			y++
			lines++
		}
		return lines + 1

	case "guardrails":
		n := len(m.cfg.Guardrails)
		avail := c.Height() - y - 1
		if avail < 1 {
			avail = 1
		}
		itemStart := 0
		if selected && n > avail {
			if m.guardrailSub {
				itemStart = m.listIdx
			} else {
				itemStart = m.listIdx - avail + 1
				if itemStart < 0 {
					itemStart = 0
				}
			}
		}
		if n == 0 {
			c.SetContent(4, y, render.StyleMuted, "(none)")
			y++
			lines++
		} else {
			for i := itemStart; i < n && i-itemStart < avail; i++ {
				if y >= c.Height()-1 {
					break
				}
				expanded := selected && m.guardrailSub && i == m.listIdx
				itemStyle := render.StyleDefault
				prefix := "  "
				if selected && i == m.listIdx {
					itemStyle = render.StyleHighlight
					prefix = "▸ "
				}
				c.SetContent(4, y, itemStyle, runewidth.Truncate(prefix+m.itemDisplay("guardrails", i), c.Width()-6, "..."))
				y++
				lines++
				if expanded {
					sub := []string{
						"id: " + m.cfg.Guardrails[i].ID,
						"paths: " + strings.Join(m.cfg.Guardrails[i].Paths, ", "),
						"require_all: " + strconv.FormatBool(m.cfg.Guardrails[i].RequireAll),
						"action: " + m.cfg.Guardrails[i].Action,
						"reason: " + m.cfg.Guardrails[i].Reason,
					}
					for si, sv := range sub {
						if y >= c.Height()-1 {
							break
						}
						if m.editing && m.editSub == si {
							trunc := runewidth.Truncate(m.inputBuf, c.Width()-6, "")
							c.SetContent(4, y, ui.StyleBtnFocus, trunc)
							cursorX := 4 + runewidth.StringWidth(trunc)
							if cursorX >= c.Width()-1 {
								cursorX = c.Width() - 2
							}
							if cursorX < c.Width()-1 {
								c.SetCell(cursorX, y, render.StyleHighlight, '_')
							}
						} else {
							subStyle := render.StyleDim
							subPrefix := "   "
							if m.subIdx == si {
								subStyle = render.StyleInfo
								subPrefix = "   ▸ "
							}
							c.SetContent(4, y, subStyle, runewidth.Truncate(subPrefix+sv, c.Width()-6, "..."))
						}
						y++
						lines++
					}
				}
			}
		}
		return lines + 1
	}

	return lines
}

func (m *ConfigsModel) renderInputLine(c *render.Canvas, x, y, maxW int, prompt string) {
	display := runewidth.Truncate(m.inputBuf, maxW-len(prompt), "")
	c.SetContent(x, y, ui.StyleBtnFocus, prompt+display)
	cursorX := x + runewidth.StringWidth(prompt) + runewidth.StringWidth(display)
	if cursorX >= x+maxW {
		cursorX = x + maxW - 1
	}
	if cursorX < c.Width()-1 {
		c.SetCell(cursorX, y, render.StyleHighlight, '_')
	}
}

// @dizz-ignore-abandoned
func (m *ConfigsModel) View() string { return "" }

func (m *ConfigsModel) InputMode() bool { return m.editing }

func (m *ConfigsModel) WantsTab() bool {
	return !m.editing && !m.saving && m.cfg != nil && m.isItemList(configFields[m.fieldIdx])
}

func (m *ConfigsModel) IsModalActive() bool { return false }
