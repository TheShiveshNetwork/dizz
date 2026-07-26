package app

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	"github.com/TheShiveshNetwork/dizz/tui/views"
	tea "github.com/charmbracelet/bubbletea"
)

type AnimationTick time.Time

const refreshInterval = 30

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

	intents    []client.Intent
	intentsSel int
	intentsOff int

	todos    []client.Todo
	todosSel int
	todosOff int

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
	tickCount     int

	lastStatusHash  string
	lastIntentsHash string
	lastTodosHash   string

	viewLoaded []bool
}

type batchRefreshMsg struct {
	summary     *client.Summary
	intents     []client.Intent
	todos       []client.Todo
	err         string
	statusHash  string
	intentsHash string
	todosHash   string
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
		initialized: client.IsDizzInitialized(),
		version:     version,
	}

	m.views = []tea.Model{
		views.NewDashboardModel(),
		views.NewSymbolsModel(),
		views.NewIntentsModel(),
		views.NewTodosModel(),
		views.NewSnapshotsModel(),
		views.NewConfigsModel(),
	}
	m.viewLoaded = make([]bool, len(m.views))

	return m
}

func (m *Model) Init() tea.Cmd {
	if !m.initialized {
		return m.tick()
	}
	m.viewLoaded[0] = true
	return tea.Batch(m.tick(), m.refreshBatch(), m.views[0].Init())
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
		if m.initialized {
			m.tickCount++
			if m.tickCount%refreshInterval == 0 {
				return m, tea.Batch(m.tick(), m.refreshBatch())
			}
			var cmd tea.Cmd
			m.views[m.currentTab], cmd = m.views[m.currentTab].Update(time.Time(msg))
			return m, tea.Batch(m.tick(), cmd)
		}
		m.initFrameTick++
		return m, m.tick()

	case batchRefreshMsg:
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.err = ""
			changed := false
			if msg.statusHash != m.lastStatusHash && msg.summary != nil {
				m.statusData = ui.StatusData{
					ProjectName: msg.summary.ProjectName,
					Branch:      msg.summary.Branch,
					Commit:      msg.summary.Commit,
					Version:     m.version,
				}
				m.lastStatusHash = msg.statusHash
				changed = true
			}
			if msg.intentsHash != m.lastIntentsHash {
				m.intents = msg.intents
				m.lastIntentsHash = msg.intentsHash
				changed = true
			}
			if msg.todosHash != m.lastTodosHash {
				m.todos = msg.todos
				m.lastTodosHash = msg.todosHash
				changed = true
			}
			if changed {
				var cmds []tea.Cmd
				for i, v := range m.views {
					if !m.viewLoaded[i] {
						continue
					}
					newV, cmd := v.Update(ui.RefreshTick{})
					m.views[i] = newV
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
			}
		}

	case intentActionDoneMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
		}
		return m, m.refreshBatch()

	case initDoneMsg:
		m.initBusy = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.initialized = client.IsDizzInitialized()
		if m.initialized {
			m.viewLoaded[0] = true
			return m, tea.Batch(m.refreshBatch(), m.views[0].Init())
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
	isMainFocused := m.focusZone == "main"

	contentLeftX := 0

	ui.RenderSidebar(c, m.currentTab, isSidebarFocused, contentLeftX, mainTop, leftPanelW-contentLeftX)

	hideIntentsPanel := m.currentTab == 2 || m.currentTab == 5
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
		isIntentsFocused := m.focusZone == "intents"
		isTodosFocused := m.focusZone == "todos"
		m.renderIntentsMini(c, contentLeftX, intentsTop, leftPanelW-contentLeftX, intentsH, isIntentsFocused, isTodosFocused)
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
		case 5:
			tabName = "configs"
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
					return initDoneMsg{err: client.Initialize()}
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
		if m.intentsPanelHidden() {
			switch m.focusZone {
			case "sidebar", "intents", "todos":
				m.focusZone = "main"
			case "main":
				m.focusZone = "sidebar"
			}
		} else {
			switch m.focusZone {
			case "sidebar":
				m.focusZone = "intents"
			case "intents":
				m.focusZone = "todos"
			case "todos":
				m.focusZone = "main"
			case "main":
				m.focusZone = "sidebar"
			}
		}
		m.showHelp = false
		return m, nil

	case "shift+tab":
		if m.intentsPanelHidden() {
			switch m.focusZone {
			case "sidebar", "intents", "todos":
				m.focusZone = "main"
			case "main":
				m.focusZone = "sidebar"
			}
		} else {
			switch m.focusZone {
			case "sidebar":
				m.focusZone = "main"
			case "intents":
				m.focusZone = "sidebar"
			case "todos":
				m.focusZone = "intents"
			case "main":
				m.focusZone = "todos"
			}
		}
		m.showHelp = false
		return m, nil

	case "up":
		switch m.focusZone {
		case "sidebar":
			nextTab := (m.currentTab - 1 + len(m.views)) % len(m.views)
			return m, m.switchTab(nextTab)
		case "intents":
			m.intentsSel--
			if m.intentsSel < 0 {
				m.intentsSel = 0
			}
			m.ensureIntentsVisible()
			return m, nil
		case "todos":
			m.todosSel--
			if m.todosSel < 0 {
				m.todosSel = 0
			}
			m.ensureTodosVisible()
			return m, nil
		}

	case "down":
		switch m.focusZone {
		case "sidebar":
			nextTab := (m.currentTab + 1) % len(m.views)
			return m, m.switchTab(nextTab)
		case "intents":
			m.intentsSel++
			active := m.activeIntents()
			if m.intentsSel >= len(active) {
				m.intentsSel = len(active) - 1
			}
			m.ensureIntentsVisible()
			return m, nil
		case "todos":
			m.todosSel++
			unresolved := m.unresolvedTodos()
			if m.todosSel >= len(unresolved) {
				m.todosSel = len(unresolved) - 1
			}
			m.ensureTodosVisible()
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
		if m.focusZone == "intents" || m.focusZone == "todos" {
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

	case "1", "2", "3", "4", "5", "6":
		if m.focusZone == "sidebar" {
			tab := int(msg.String()[0] - '1')
			if tab >= 0 && tab < len(m.views) {
				return m, m.switchTab(tab)
			}
		}
		return m, nil
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

func (m *Model) switchTab(tab int) tea.Cmd {
	if tab == m.currentTab {
		return nil
	}
	m.currentTab = tab
	m.showHelp = false
	m.fixFocusZone()
	if !m.viewLoaded[tab] {
		m.viewLoaded[tab] = true
		return m.views[tab].Init()
	}
	return nil
}

func (m *Model) refreshBatch() tea.Cmd {
	return func() tea.Msg {
		type statusResult struct {
			summary *client.Summary
			err     error
		}
		type intentsResult struct {
			intents []client.Intent
			err     error
		}
		type todosResult struct {
			todos []client.Todo
			err   error
		}

		statusCh := make(chan statusResult, 1)
		intentsCh := make(chan intentsResult, 1)
		todosCh := make(chan todosResult, 1)

		go func() {
			s, err := client.Status()
			if err == nil && s.ProjectName == "" {
				if root, rerr := client.FindDizzRoot(); rerr == nil {
					s.ProjectName = filepath.Base(root)
				}
			}
			statusCh <- statusResult{summary: s, err: err}
		}()
		go func() {
			intents, err := client.ListIntents()
			intentsCh <- intentsResult{intents: intents, err: err}
		}()
		go func() {
			todos, err := client.ListTodos()
			todosCh <- todosResult{todos: todos, err: err}
		}()

		sr := <-statusCh
		ir := <-intentsCh
		tr := <-todosCh

		msg := batchRefreshMsg{}
		if sr.err != nil {
			msg.err = sr.err.Error()
			return msg
		}
		msg.summary = sr.summary
		msg.statusHash = hashSummary(sr.summary)
		if ir.err == nil {
			msg.intents = ir.intents
			msg.intentsHash = hashIntents(ir.intents)
		}
		if tr.err == nil {
			msg.todos = tr.todos
			msg.todosHash = hashTodos(tr.todos)
		}
		return msg
	}
}

func hashSummary(s *client.Summary) string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("%d-%d-%d-%d-%d-%d-%d-%d",
		s.TotalSymbols, s.Active, s.Planned, s.Unstable,
		s.Unused, s.Abandoned, s.ActiveTodos, s.Intents)
}

func hashIntents(intents []client.Intent) string {
	if len(intents) == 0 {
		return "0"
	}
	return fmt.Sprintf("%d-%s", len(intents), intents[len(intents)-1].ID)
}

func hashTodos(todos []client.Todo) string {
	if len(todos) == 0 {
		return "0"
	}
	return fmt.Sprintf("%d-%s", len(todos), todos[len(todos)-1].File)
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

func (m *Model) ensureTodosVisible() {
	listH := 8
	if m.todosSel < m.todosOff {
		m.todosOff = m.todosSel
	}
	if m.todosSel >= m.todosOff+listH {
		m.todosOff = m.todosSel - listH + 1
	}
}

func (m *Model) intentsPanelHidden() bool {
	return m.currentTab == 2 || m.currentTab == 5
}

func (m *Model) fixFocusZone() {
	if m.intentsPanelHidden() && (m.focusZone == "intents" || m.focusZone == "todos") {
		m.focusZone = "main"
	}
}

func (m *Model) activeIntents() []client.Intent {
	var active []client.Intent
	for _, in := range m.intents {
		if in.Status == "" || in.Status == "active" {
			active = append(active, in)
		}
	}
	return active
}

func (m *Model) unresolvedTodos() []client.Todo {
	var unresolved []client.Todo
	for _, td := range m.todos {
		if !td.Resolved {
			unresolved = append(unresolved, td)
		}
	}
	return unresolved
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
