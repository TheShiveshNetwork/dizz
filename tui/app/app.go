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
	currentTab   int
	views        []tea.Model
	focusSidebar bool

	now         time.Time
	width       int
	height      int
	showHelp    bool
	statusData  ui.StatusData
	err         string
	initialized bool
}

func NewModel() *Model {
	m := &Model{
		currentTab:   0,
		focusSidebar: true,
		width:        80,
		height:       24,
		showHelp:     false,
		now:          time.Now(),
		initialized:  dizzclient.IsDizzInitialized(),
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
	cmds := []tea.Cmd{m.tick(), m.refreshStatus()}
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

type statusMsg struct {
	summary *dizzclient.Summary
	err     string
	version string
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
		m.focusSidebar = !m.focusSidebar
		m.showHelp = false
		return m, nil

	case "shift+tab":
		m.focusSidebar = !m.focusSidebar
		m.showHelp = false
		return m, nil

	case "up":
		if m.focusSidebar {
			m.currentTab = (m.currentTab - 1 + len(m.views)) % len(m.views)
			m.showHelp = false
			return m, nil
		}

	case "down":
		if m.focusSidebar {
			m.currentTab = (m.currentTab + 1) % len(m.views)
			m.showHelp = false
			return m, nil
		}

	case "enter":
		if m.focusSidebar {
			m.focusSidebar = false
			m.showHelp = false
			return m, nil
		}

	case "1", "2", "3", "4", "5":
		if m.focusSidebar {
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

	if m.initialized && m.currentTab < len(m.views) {
		var cmd tea.Cmd
		m.views[m.currentTab], cmd = m.views[m.currentTab].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) isInInputMode() bool {
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

	sidebarW := 16
	mainTop := 2
	mainBottom := m.height - 3
	contentW := m.width - sidebarW - 1
	contentH := mainBottom - mainTop + 1
	if contentH < 1 {
		contentH = 1
	}

	ui.RenderSidebar(c, m.currentTab, m.focusSidebar, 0, mainTop, sidebarW)
	sepStyle := render.StyleMuted
	if m.focusSidebar {
		sepStyle = render.StyleHighlight
	}
	if mainBottom >= mainTop {
		for y := mainTop; y <= mainBottom; y++ {
			c.SetCell(sidebarW, y, sepStyle, '│')
		}
	}

	contentCanvas := render.NewCanvas(contentW, contentH)
	view := m.views[m.currentTab]
	if viewWithRender, ok := view.(interface{ Render(*render.Canvas) }); ok {
		viewWithRender.Render(contentCanvas)
	}
	c.Blit(contentCanvas, sidebarW+1, mainTop)

	if ms, ok := view.(ModalState); ok && ms.IsModalActive() {
		ms.RenderModal(c)
	}

	statusY := m.height - 2
	inputMode := false
	if m.currentTab < len(m.views) {
		if im, ok := m.views[m.currentTab].(interface{ InputMode() bool }); ok {
			inputMode = im.InputMode()
		}
	}
	ui.RenderStatusBar(c, m.statusData, m.focusSidebar, statusY, m.width, inputMode)

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

func (m *Model) renderInitScreen(c *render.Canvas) {
	title := "dizz Terminal UI"
	subtitle := "Run 'dizz init' in a project to get started"
	help := "Press q to quit"

	c.SetContent((m.width-len(title))/2, m.height/2-2, render.StyleHighlight, title)
	c.SetContent((m.width-len(subtitle))/2, m.height/2, render.StyleMuted, subtitle)
	c.SetContent((m.width-len(help))/2, m.height/2+2, render.StyleDim, help)
}
