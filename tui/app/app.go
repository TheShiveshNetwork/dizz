package app

import (
	"path/filepath"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	"github.com/TheShiveshNetwork/dizz/tui/views"
	tea "github.com/charmbracelet/bubbletea"
)

type AnimationTick time.Time

type ModalState interface {
	IsModalActive() bool
	RenderModal(*render.Canvas)
}

type Model struct {
	currentTab int
	views      []tea.Model
	focusZone  string

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

	initFrameTick int64
	initBusy      bool
	version       string
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

type intentActionDoneMsg struct {
	err error
}

type initDoneMsg struct {
	err error
}

func NewModel(version string) *Model {
	m := &Model{
		currentTab:  0,
		focusZone:   "sidebar",
		width:       80,
		height:      24,
		showHelp:    false,
		now:         time.Now(),
		initialized: dizzclient.IsDizzInitialized(),
		version:     version,
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
		return m.tick()
	}
	cmds := []tea.Cmd{m.tick(), m.refreshStatus(), m.refreshIntents()}
	for _, v := range m.views {
		cmds = append(cmds, v.Init())
	}
	return tea.Batch(cmds...)
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
		if !m.initialized {
			m.initFrameTick++
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

	case intentActionDoneMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
		}
		return m, m.refreshIntents()

	case initDoneMsg:
		m.initBusy = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.initialized = dizzclient.IsDizzInitialized()
		if m.initialized {
			cmds := []tea.Cmd{m.refreshStatus(), m.refreshIntents()}
			for _, v := range m.views {
				cmds = append(cmds, v.Init())
			}
			return m, tea.Batch(cmds...)
		}
		return m, nil

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

func (m *Model) View() string {
	c := render.NewCanvas(m.width, m.height)

	if !m.initialized {
		m.renderInitScreen(c)
		return c.ANSIFrame()
	}

	ui.RenderTopBar(c, m.statusData, 0, m.width)
	for x := 0; x < m.width; x++ {
		c.SetCell(x, 1, render.StyleMuted, '\u2500')
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

	ui.RenderSidebar(c, m.currentTab, isSidebarFocused, contentLeftX, mainTop, leftPanelW-contentLeftX)

	hideIntentsPanel := m.currentTab == 2
	if !hideIntentsPanel {
		sepY := mainTop + len(ui.Tabs)
		if sepY < mainBottom {
			for sx := contentLeftX; sx < leftPanelW; sx++ {
				c.SetCell(sx, sepY, render.StyleMuted, '\u2500')
			}
		}

		intentsTop := sepY + 1
		intentsH := mainBottom - intentsTop + 1
		if intentsH < 3 {
			intentsH = 3
		}
		m.renderIntentsMini(c, contentLeftX, intentsTop, leftPanelW-contentLeftX, intentsH, isIntentsFocused)
	}

	rightSepStyle := render.StyleMuted
	if isMainFocused {
		rightSepStyle = render.StyleHighlight
	}
	if mainBottom >= mainTop {
		for y := mainTop; y <= mainBottom; y++ {
			c.SetCell(leftPanelW, y, rightSepStyle, '\u2502')
		}
	}

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
		}
		ui.RenderHelpOverlay(c, tabName)
	}

	return c.ANSIFrame()
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

	if !m.initialized {
		switch msg.String() {
		case "enter":
			if !m.initBusy {
				m.initBusy = true
				m.err = ""
				return m, func() tea.Msg {
					return initDoneMsg{err: dizzclient.Initialize()}
				}
			}
		}
		return m, nil
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
			version: m.version,
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
