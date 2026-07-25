package views

import (
	"fmt"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
)

type DashboardModel struct {
	summary   *dizzclient.Summary
	frameTick int64
	loading   bool
	err       string
}

func NewDashboardModel() *DashboardModel {
	return &DashboardModel{
		loading: true,
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return m.refresh()
}

func (m *DashboardModel) refresh() tea.Cmd {
	return func() tea.Msg {
		s, err := dizzclient.Status()
		if err != nil {
			return dashboardMsg{err: err.Error()}
		}
		return dashboardMsg{summary: s}
	}
}

type dashboardMsg struct {
	summary *dizzclient.Summary
	err     string
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			return m, m.refresh()
		}

	case dashboardMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.summary = msg.summary
			m.err = ""
		}

	case time.Time:
		m.frameTick++

	case ui.RefreshTick:
		return m, m.refresh()
	}

	return m, nil
}

func (m *DashboardModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	if m.loading {
		msg := "Loading..."
		c.SetContent((w-len(msg))/2, h/2, render.StyleMuted, msg)
		return
	}

	if m.err != "" {
		c.SetContent(2, 2, render.StyleError, m.err)
		return
	}

	title := "Hey, Dizzie"
	c.SetContent((w-len(title))/2, 0, render.StyleHighlight.Bold(true), title)

	s := m.summary
	if s == nil {
		return
	}

	codeScore := 0
	if s.TotalSymbols > 0 {
		codeScore = (s.Active * 100) / s.TotalSymbols
	}
	scoreStr := fmt.Sprintf("Code Score: ● %d%%", codeScore)
	c.SetContent(w-len(scoreStr)-4, 1, render.StyleHighlight, scoreStr)

	barW := 20
	if barW > w-30 {
		barW = w - 30
	}
	if barW < 5 {
		barW = 5
	}

	type barItem struct {
		label string
		count int
		style tcell.Style
	}
	items := []barItem{
		{"Active", s.Active, render.StyleSuccess},
		{"Planned", s.Planned, render.StyleWarning},
		{"Unstable", s.Unstable, render.StyleError},
		{"Unused", s.Unused, render.StyleInfo},
		{"Abandoned", s.Abandoned, render.StyleMuted},
	}

	maxCount := s.Active
	for _, it := range items {
		if it.count > maxCount {
			maxCount = it.count
		}
	}
	if maxCount < 1 {
		maxCount = 1
	}

	shortCodes := map[string]string{
		"Active": "ACT", "Planned": "PLN", "Unstable": "UNS",
		"Unused": "UNU", "Abandoned": "ABD",
	}
	barX := w - barW - 4
	barY := 3
	for i, it := range items {
		rowY := barY + i
		if rowY >= h-6 {
			break
		}
		bar := render.RenderBar(it.count, maxCount, barW)
		c.SetContent(barX, rowY, it.style, bar)
		if code, ok := shortCodes[it.label]; ok {
			c.SetContent(barX+barW+1, rowY, render.StyleDim, code)
		}
	}

	_, artH := render.DizzArtDimensions()
	startX := w / 2
	startY := (h - artH) / 2
	if startY < 4 {
		startY = 4
	}
	render.RenderDizzie(c, m.frameTick, startX, startY)

	if s.ActiveTodos > 0 {
		todoMsg := fmt.Sprintf("TODOs: %d", s.ActiveTodos)
		c.SetContent(4, h-3, render.StyleWarning, todoMsg)
	}

	intentMsg := fmt.Sprintf("Intents: %d", s.Intents)
	c.SetContent(4, h-2, render.StyleInfo, intentMsg)

	help := "Press ? for help  r=refresh  q=quit"
	if len(help) > w-2 {
		help = help[:w-2]
	}
	c.SetContent((w-len(help))/2, h-1, render.StyleMuted, help)
}

// @dizz-ignore-unused
func (m *DashboardModel) View() string { return "" }
