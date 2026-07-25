package app

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	"github.com/TheShiveshNetwork/dizz/tui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type AnimationTick time.Time

type ModalState interface {
	IsModalActive() bool
	RenderModal(*render.Canvas)
}

type Model struct {
	currentTab int
	views      []tea.Model
	focusZone  string // "sidebar", "intents", "main"

	now         time.Time
	width       int
	height      int
	showHelp    bool
	statusData  ui.StatusData
	err         string
	initialized bool

	intents    []dizzclient.Intent
	intentsSel int
	intentsOff int

	showModal     bool
	modalType     string
	resolveIdx    int
	focusIdx      int
	addMsg        string
	addMsgCursor  int
	addType       int
	addSev        int
	validationErr string
	addNote       string
	noteCursor    int
}

func NewModel() *Model {
	m := &Model{
		currentTab:  0,
		focusZone:   "sidebar",
		width:       80,
		height:      24,
		showHelp:    false,
		now:         time.Now(),
		initialized: dizzclient.IsDizzInitialized(),
	}

	m.views = []tea.Model{
		views.NewDashboardModel(),
		views.NewSymbolsModel(),
		views.NewIntentsModel(),
		views.NewTodosModel(),
		views.NewSnapshotsModel(),
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	if !m.initialized {
		return nil
	}
	cmds := []tea.Cmd{m.tick(), m.refreshStatus(), m.refreshIntents()}
	for _, v := range m.views {
		cmds = append(cmds, v.Init())
	}
	return tea.Batch(cmds...)
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return AnimationTick(t)
	})
}

func (m *Model) refreshStatus() tea.Cmd {
	return func() tea.Msg {
		s, err := dizzclient.Status()
		if err != nil {
			return statusMsg{err: err.Error()}
		}
		if s.ProjectName == "" {
			if root, err := dizzclient.FindDizzRoot(); err == nil {
				s.ProjectName = filepath.Base(root)
			}
		}
		return statusMsg{
			summary: s,
			version: "0.1.0",
		}
	}
}

func (m *Model) refreshIntents() tea.Cmd {
	return func() tea.Msg {
		intents, err := dizzclient.ListIntents()
		if err != nil {
			return intentsMsg{err: err.Error()}
		}
		return intentsMsg{intents: intents}
	}
}

type statusMsg struct {
	summary *dizzclient.Summary
	err     string
	version string
}

type intentsMsg struct {
	intents []dizzclient.Intent
	err     string
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case AnimationTick:
		m.now = time.Time(msg)
		if m.initialized && m.currentTab < len(m.views) {
			var cmd tea.Cmd
			m.views[m.currentTab], cmd = m.views[m.currentTab].Update(time.Time(msg))
			return m, tea.Batch(m.tick(), cmd)
		}
		return m, m.tick()

	case statusMsg:
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.err = ""
			m.statusData = ui.StatusData{
				ProjectName: msg.summary.ProjectName,
				Branch:      msg.summary.Branch,
				Commit:      msg.summary.Commit,
				Version:     msg.version,
			}
		}

	case intentsMsg:
		if msg.err == "" {
			m.intents = msg.intents
		}

	case error:
		m.err = msg.Error()
	}

	if m.initialized {
		var cmds []tea.Cmd
		for i, v := range m.views {
			newV, cmd := v.Update(msg)
			m.views[i] = newV
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if !m.isInInputMode() {
			return m, tea.Quit
		}
	}

	if m.showModal {
		return m.handleModalKey(msg)
	}

	if m.initialized && m.currentTab < len(m.views) {
		if ms, ok := m.views[m.currentTab].(ModalState); ok && ms.IsModalActive() {
			var cmd tea.Cmd
			m.views[m.currentTab], cmd = m.views[m.currentTab].Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "tab":
		switch m.focusZone {
		case "sidebar":
			m.focusZone = "intents"
		case "intents":
			m.focusZone = "main"
		case "main":
			m.focusZone = "sidebar"
		}
		m.showHelp = false
		return m, nil

	case "shift+tab":
		switch m.focusZone {
		case "sidebar":
			m.focusZone = "main"
		case "intents":
			m.focusZone = "sidebar"
		case "main":
			m.focusZone = "intents"
		}
		m.showHelp = false
		return m, nil

	case "up":
		switch m.focusZone {
		case "sidebar":
			m.currentTab = (m.currentTab - 1 + len(m.views)) % len(m.views)
			m.showHelp = false
			return m, nil
		case "intents":
			m.intentsSel--
			if m.intentsSel < 0 {
				m.intentsSel = 0
			}
			m.ensureIntentsVisible()
			return m, nil
		}

	case "down":
		switch m.focusZone {
		case "sidebar":
			m.currentTab = (m.currentTab + 1) % len(m.views)
			m.showHelp = false
			return m, nil
		case "intents":
			m.intentsSel++
			active := m.activeIntents()
			if m.intentsSel >= len(active) {
				m.intentsSel = len(active) - 1
			}
			m.ensureIntentsVisible()
			return m, nil
		}

	case "enter":
		switch m.focusZone {
		case "sidebar":
			m.focusZone = "main"
			m.showHelp = false
			return m, nil
		case "intents":
			active := m.activeIntents()
			if len(active) > 0 && m.intentsSel < len(active) {
				m.resolveIdx = m.intentsSel
				m.showModal = true
				m.modalType = "resolve"
				m.focusIdx = 0
				m.addNote = ""
				m.noteCursor = 0
			}
			return m, nil
		}

	case "i":
		if m.focusZone == "intents" {
			m.showModal = true
			m.modalType = "add"
			m.focusIdx = 0
			m.addType = 0
			m.addMsg = ""
			m.addMsgCursor = 0
			m.addSev = 1
			m.validationErr = ""
			return m, nil
		}

	case "1", "2", "3", "4", "5":
		if m.focusZone == "sidebar" {
			tab := int(msg.String()[0] - '1')
			if tab >= 0 && tab < len(m.views) {
				m.currentTab = tab
				m.showHelp = false
			}
			return m, nil
		}
	}

	if m.showHelp && msg.String() == "esc" {
		m.showHelp = false
		return m, nil
	}

	if m.initialized && m.currentTab < len(m.views) && m.focusZone == "main" {
		var cmd tea.Cmd
		m.views[m.currentTab], cmd = m.views[m.currentTab].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) ensureIntentsVisible() {
	listH := 8
	if m.intentsSel < m.intentsOff {
		m.intentsOff = m.intentsSel
	}
	if m.intentsSel >= m.intentsOff+listH {
		m.intentsOff = m.intentsSel - listH + 1
	}
}

func (m *Model) activeIntents() []dizzclient.Intent {
	var active []dizzclient.Intent
	for _, in := range m.intents {
		if in.Status == "" || in.Status == "active" {
			active = append(active, in)
		}
	}
	return active
}

func (m *Model) isInInputMode() bool {
	if m.showModal {
		return true
	}
	if !m.initialized || m.currentTab >= len(m.views) {
		return false
	}
	if ms, ok := m.views[m.currentTab].(ModalState); ok && ms.IsModalActive() {
		return true
	}
	if im, ok := m.views[m.currentTab].(interface{ InputMode() bool }); ok {
		return im.InputMode()
	}
	return false
}

func wrapText(text string, maxW int) []string {
	if maxW < 1 {
		maxW = 1
	}
	var lines []string
	for len(text) > 0 {
		if runewidth.StringWidth(text) <= maxW {
			lines = append(lines, text)
			break
		}
		idx := len(text)
		for runewidth.StringWidth(text[:idx]) > maxW {
			idx--
		}
		if idx == 0 {
			idx = 1
		}
		lines = append(lines, text[:idx])
		text = text[idx:]
	}
	return lines
}

func (m *Model) View() string {
	c := render.NewCanvas(m.width, m.height)

	if !m.initialized {
		m.renderInitScreen(c)
		return c.ANSIFrame()
	}

	ui.RenderTopBar(c, m.statusData, 0, m.width)
	for x := 0; x < m.width; x++ {
		c.SetCell(x, 1, render.StyleMuted, '─')
	}

	leftPanelW := 30
	mainTop := 2
	mainBottom := m.height - 3
	contentW := m.width - leftPanelW - 1
	contentH := mainBottom - mainTop + 1
	if contentH < 1 {
		contentH = 1
	}

	isSidebarFocused := m.focusZone == "sidebar"
	isIntentsFocused := m.focusZone == "intents"
	isMainFocused := m.focusZone == "main"

	contentLeftX := 0

	// ── Sidebar tabs (top of left panel) ──
	ui.RenderSidebar(c, m.currentTab, isSidebarFocused, contentLeftX, mainTop, leftPanelW-contentLeftX)

	// ── Separator between tabs and intents (hide when intents tab is active) ──
	hideIntentsPanel := m.currentTab == 2
	if !hideIntentsPanel {
		sepY := mainTop + len(ui.Tabs)
		if sepY < mainBottom {
			for sx := contentLeftX; sx < leftPanelW; sx++ {
				c.SetCell(sx, sepY, render.StyleMuted, '─')
			}
		}

		// ── Intents mini view (bottom of left panel) ──
		intentsTop := sepY + 1
		intentsH := mainBottom - intentsTop + 1
		if intentsH < 3 {
			intentsH = 3
		}
		m.renderIntentsMini(c, contentLeftX, intentsTop, leftPanelW-contentLeftX, intentsH, isIntentsFocused)
	}

	// ── Right separator (between left panel and content) ──
	rightSepStyle := render.StyleMuted
	if isMainFocused {
		rightSepStyle = render.StyleHighlight
	}
	if mainBottom >= mainTop {
		for y := mainTop; y <= mainBottom; y++ {
			c.SetCell(leftPanelW, y, rightSepStyle, '│')
		}
	}

	// ── Main content (right panel) ──
	contentCanvas := render.NewCanvas(contentW, contentH)
	view := m.views[m.currentTab]
	if viewWithRender, ok := view.(interface{ Render(*render.Canvas) }); ok {
		viewWithRender.Render(contentCanvas)
	}
	c.Blit(contentCanvas, leftPanelW+1, mainTop)

	if ms, ok := view.(ModalState); ok && ms.IsModalActive() {
		ms.RenderModal(c)
	}

	if m.showModal {
		m.renderIntentsModal(c)
	}

	// ── Status bar ──
	statusY := m.height - 2
	inputMode := m.isInInputMode()
	ui.RenderStatusBar(c, m.statusData, m.focusZone, statusY, m.width, inputMode)

	if m.err != "" {
		c.SetContent(0, m.height-3, render.StyleError, m.err)
	}

	if m.showHelp && m.currentTab < len(m.views) {
		tabName := ""
		switch m.currentTab {
		case 0:
			tabName = "dashboard"
		case 1:
			tabName = "symbols"
		case 2:
			tabName = "intents"
		case 3:
			tabName = "todos"
		case 4:
			tabName = "snapshots"
		case 5:
		}
		ui.RenderHelpOverlay(c, tabName)
	}

	return c.ANSIFrame()
}

func (m *Model) renderIntentsMini(c *render.Canvas, x, y, w, h int, focused bool) {
	active := m.activeIntents()

	header := fmt.Sprintf("Intents [%d]", len(active))
	headerStyle := render.StyleHighlight.Bold(true)
	if focused {
		headerStyle = render.StyleHighlight.Bold(true).Foreground(tcell.NewRGBColor(100, 200, 255))
	}
	c.SetContent(x+1, y, headerStyle, header)

	if h < 2 {
		return
	}
	listY := y + 1
	visible := h - 2
	if visible < 1 {
		visible = 1
	}

	for i := 0; i < visible && m.intentsOff+i < len(active); i++ {
		in := active[m.intentsOff+i]
		rowY := listY + i
		if rowY > y+h-1 {
			break
		}

		typStyle := render.IntentTypeStyle(string(in.Type))
		label := " " + string(in.Type) + " "
		c.SetContent(x+1, rowY, typStyle, label)

		msgX := x + 1 + len(label)
		availW := x + w - msgX - 1
		if availW < 3 {
			availW = 3
		}
		msg := in.Message
		if runewidth.StringWidth(msg) > availW {
			for runewidth.StringWidth(msg)+1 > availW {
				msg = msg[:len(msg)-1]
			}
			msg += "…"
		}
		msgStyle := render.StyleDefault
		if focused && m.intentsOff+i == m.intentsSel {
			msgStyle = render.StyleHighlight
		}
		c.SetContent(msgX, rowY, msgStyle, msg)
	}

	if m.intentsOff > 0 {
		c.SetCell(x, listY, render.StyleDim, '▲')
	}
	if m.intentsOff+visible < len(active) {
		indY := listY + visible - 1
		if indY >= listY && indY < y+h {
			c.SetCell(x, indY, render.StyleDim, '▼')
		}
	}

	hintY := y + h - 1
	if focused && h >= 6 {
		hint := "↑↓=nav enter=resolve i=add"
		c.SetContent(x+1, hintY, render.StyleDim, hint)
	}
}

func (m *Model) renderIntentsModal(c *render.Canvas) {
	if m.modalType == "add" {
		m.renderAddModal(c)
	} else if m.modalType == "resolve" {
		m.renderResolveModal(c)
	}
}

func (m *Model) renderResolveModal(c *render.Canvas) {
	msg := ""
	id := ""
	active := m.activeIntents()
	if m.resolveIdx < len(active) {
		in := active[m.resolveIdx]
		id = in.ID
		msg = in.Message
	}

	bodyW := 52
	if len(msg)+10 > bodyW {
		bodyW = len(msg) + 10
		if bodyW > 70 {
			bodyW = 70
		}
	}
	bodyH := 11
	cx, cy := ui.RenderModalBox(c, "Resolve Intent", bodyW, bodyH)

	help := "Tab/↑↓=switch  Enter=confirm  Esc=cancel"
	c.SetContent(cx, cy, ui.StyleHelp, help)
	cy += 2

	prompt := fmt.Sprintf("Resolve %s?", id)
	c.SetContent(cx+(bodyW-len(prompt))/2, cy, render.StyleWarning, prompt)
	cy++

	maxMsgW := bodyW - 4
	lines := wrapText(msg, maxMsgW)
	for _, line := range lines {
		if cy >= c.Height()-4 {
			break
		}
		c.SetContent(cx+(bodyW-runewidth.StringWidth(line))/2, cy, render.StyleInfo, line)
		cy++
	}
	cy++

	noteFocused := m.focusIdx == 0
	noteStyle := ui.StyleBtn
	if noteFocused {
		noteStyle = ui.StyleBtnFocus
	}
	c.SetContent(cx, cy, render.StyleDefault, "Note:     ")
	display := m.addNote
	if display == "" {
		display = "(optional)"
	}
	maxNoteW := bodyW - 14
	if maxNoteW < 10 {
		maxNoteW = 10
	}
	if len(display) > maxNoteW {
		display = display[:maxNoteW]
	}
	c.SetContent(cx+10, cy, noteStyle, display)
	if noteFocused {
		cxPos := cx + 10 + m.noteCursor
		if display == "(optional)" {
			noteStyle = render.StyleDim
		}
		if cxPos < c.Width()-1 {
			c.SetCell(cxPos, cy, render.StyleHighlight, '_')
		}
	}
	cy += 2

	if cy < c.Height()-2 {
		btnW := 10
		spacing := 2
		totalBtnW := btnW*3 + spacing*2
		btnStartX := cx + bodyW - totalBtnW - 2
		ui.RenderButton(c, btnStartX, cy, btnW, "Cancel", m.focusIdx == 1, false)
		ui.RenderButton(c, btnStartX+btnW+spacing, cy, btnW, "Resolve", m.focusIdx == 2, false)
		ui.RenderButton(c, btnStartX+btnW*2+spacing*2, cy, btnW, "Close", m.focusIdx == 3, true)
	}
}

func (m *Model) renderAddModal(c *render.Canvas) {
	bodyW := 52
	bodyH := 9
	cx, cy := ui.RenderModalBox(c, "Add Intent", bodyW, bodyH)

	help := "Tab/↑↓=switch  ←→=change  Enter=submit  Esc=cancel"
	c.SetContent(cx, cy, ui.StyleHelp, help)
	cy += 2

	intentTypes := []string{"todo", "fixme", "refactor", "question", "hack", "temporary"}

	typ := intentTypes[m.addType]
	typeFocused := m.focusIdx == 0
	typeStyle := ui.StyleBtn
	if typeFocused {
		typeStyle = ui.StyleBtnFocus
	}
	c.SetContent(cx, cy, render.StyleDefault, "Type:     ")
	for i := range typ {
		c.SetCell(cx+10+i, cy, typeStyle, rune(typ[i]))
	}
	c.SetContent(cx+11+len(typ), cy, ui.StyleHelp, " ←→ to cycle")
	cy++

	msgFocused := m.focusIdx == 1
	msgStyle := ui.StyleBtn
	emptyMsg := m.addMsg == ""
	if msgFocused && emptyMsg {
		msgStyle = ui.StyleBtnFocus
	} else if emptyMsg {
		msgStyle = ui.StyleBtn
	}
	c.SetContent(cx, cy, render.StyleDefault, "Message:  ")
	display := m.addMsg
	if display == "" {
		display = "required"
	}
	if len(display) > bodyW-20 {
		display = display[:bodyW-20]
	}
	c.SetContent(cx+10, cy, msgStyle, display)
	if msgFocused {
		cxPos := cx + 10 + m.addMsgCursor
		if display == "required" {
			msgStyle = render.StyleDim
		}
		if cxPos < c.Width()-1 {
			c.SetCell(cxPos, cy, render.StyleHighlight, '_')
		}
	}
	cy++

	sevFocused := m.focusIdx == 2
	c.SetContent(cx, cy, render.StyleDefault, "Severity: ")
	for i := 0; i <= 3; i++ {
		s := ui.StyleBtn
		if sevFocused && i == m.addSev {
			s = ui.StyleBtnFocus
		}
		mark := fmt.Sprintf(" %d ", i)
		c.SetContent(cx+10+i*6, cy, s, mark)
	}
	sevLabels := []string{"none", "low", "medium", "high"}
	sevColor := render.SeverityColor(m.addSev)
	c.SetContent(cx+10+24, cy, sevColor, sevLabels[m.addSev])
	c.SetContent(cx+10+24+len(sevLabels[m.addSev])+2, cy, ui.StyleHelp, "←→ to change")
	if m.validationErr != "" {
		c.SetContent(cx, cy+1, render.StyleError, "⚠ "+m.validationErr)
	}
	cy += 2

	btnY := cy
	btnW := 10
	spacing := 2
	totalBtnW := btnW*2 + spacing
	btnStartX := cx + bodyW - totalBtnW - 2

	ui.RenderButton(c, btnStartX, btnY, btnW, "Cancel", m.focusIdx == 3, false)
	ui.RenderButton(c, btnStartX+btnW+spacing, btnY, btnW, "Add", m.focusIdx == 4, false)
}

func (m *Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	intentTypes := []string{"todo", "fixme", "refactor", "question", "hack", "temporary"}
	totalFields := 5
	if m.modalType == "resolve" {
		totalFields = 4
	}

	switch msg.String() {
	case "esc":
		m.showModal = false
		m.validationErr = ""
		m.addNote = ""
		m.noteCursor = 0
		return m, nil

	case "tab", "down":
		m.focusIdx = (m.focusIdx + 1) % totalFields
		return m, nil

	case "shift+tab", "up":
		m.focusIdx = (m.focusIdx - 1 + totalFields) % totalFields
		return m, nil

	case "enter":
		return m.handleModalEnter()

	case "left":
		if m.modalType == "add" {
			switch m.focusIdx {
			case 0:
				m.addType = (m.addType - 1 + len(intentTypes)) % len(intentTypes)
			case 1:
				if m.addMsgCursor > 0 {
					m.addMsgCursor--
				}
			case 2:
				if m.addSev > 0 {
					m.addSev--
				}
			}
		}
		return m, nil

	case "right":
		if m.modalType == "add" {
			switch m.focusIdx {
			case 0:
				m.addType = (m.addType + 1) % len(intentTypes)
			case 1:
				if m.addMsgCursor < len(m.addMsg) {
					m.addMsgCursor++
				}
			case 2:
				if m.addSev < 3 {
					m.addSev++
				}
			}
		}
		return m, nil

	case "backspace":
		if m.modalType == "add" && m.focusIdx == 1 && m.addMsgCursor > 0 {
			m.addMsg = m.addMsg[:m.addMsgCursor-1] + m.addMsg[m.addMsgCursor:]
			m.addMsgCursor--
		}
		if m.modalType == "resolve" && m.focusIdx == 0 && m.noteCursor > 0 {
			m.addNote = m.addNote[:m.noteCursor-1] + m.addNote[m.noteCursor:]
			m.noteCursor--
		}
		return m, nil

	case "home":
		if m.modalType == "add" && m.focusIdx == 1 {
			m.addMsgCursor = 0
		}
		if m.modalType == "resolve" && m.focusIdx == 0 {
			m.noteCursor = 0
		}
		return m, nil

	case "end":
		if m.modalType == "add" && m.focusIdx == 1 {
			m.addMsgCursor = len(m.addMsg)
		}
		if m.modalType == "resolve" && m.focusIdx == 0 {
			m.noteCursor = len(m.addNote)
		}
		return m, nil

	default:
		if m.modalType == "add" && m.focusIdx == 1 && len(msg.String()) == 1 {
			ch := msg.String()[0]
			before := m.addMsg[:m.addMsgCursor]
			after := m.addMsg[m.addMsgCursor:]
			m.addMsg = before + string(ch) + after
			m.addMsgCursor++
		}
		if m.modalType == "resolve" && m.focusIdx == 0 && len(msg.String()) == 1 {
			ch := msg.String()[0]
			before := m.addNote[:m.noteCursor]
			after := m.addNote[m.noteCursor:]
			m.addNote = before + string(ch) + after
			m.noteCursor++
		}
		return m, nil
	}
}

func (m *Model) handleModalEnter() (tea.Model, tea.Cmd) {
	if m.modalType == "resolve" {
		if m.focusIdx == 2 {
			active := m.activeIntents()
			if m.resolveIdx < len(active) {
				id := active[m.resolveIdx].ID
				go dizzclient.IntentResolve(id, m.addNote)
			}
			m.showModal = false
			m.validationErr = ""
			m.addNote = ""
			m.noteCursor = 0
			return m, m.refreshIntents()
		}
		if m.focusIdx == 3 {
			active := m.activeIntents()
			if m.resolveIdx < len(active) {
				id := active[m.resolveIdx].ID
				go dizzclient.IntentClose(id, m.addNote)
			}
			m.showModal = false
			m.validationErr = ""
			m.addNote = ""
			m.noteCursor = 0
			return m, m.refreshIntents()
		}
		m.showModal = false
		m.validationErr = ""
		m.addNote = ""
		m.noteCursor = 0
		return m, nil
	}

	if m.modalType == "add" {
		if m.focusIdx == 4 {
			if m.addMsg == "" {
				m.validationErr = "Message is required"
				return m, nil
			}
			intentTypes := []string{"todo", "fixme", "refactor", "question", "hack", "temporary"}
			go dizzclient.IntentAdd(m.addMsg, intentTypes[m.addType], m.addSev, nil)
			m.showModal = false
			m.validationErr = ""
			return m, m.refreshIntents()
		}
		if m.focusIdx == 3 {
			m.showModal = false
			m.validationErr = ""
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) renderInitScreen(c *render.Canvas) {
	title := "dizz Terminal UI"
	subtitle := "Run 'dizz init' in a project to get started"
	help := "Press q to quit"

	c.SetContent((m.width-len(title))/2, m.height/2-2, render.StyleHighlight, title)
	c.SetContent((m.width-len(subtitle))/2, m.height/2, render.StyleMuted, subtitle)
	c.SetContent((m.width-len(help))/2, m.height/2+2, render.StyleDim, help)
}
