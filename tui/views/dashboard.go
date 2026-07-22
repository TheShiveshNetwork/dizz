package views

import (
	"fmt"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
)

type DashboardModel struct {
	summary *dizzclient.Summary
	stars   []render.Star
	meteors []render.Meteor
	now     time.Time
	seed    int64
	loading bool
	err     string
}

func NewDashboardModel() *DashboardModel {
	m := &DashboardModel{
		seed:    time.Now().UnixNano(),
		loading: true,
		now:     time.Now(),
	}
	m.stars = render.GenerateStarField(80, 24, 0.035, m.seed)
	m.meteors = render.GenerateMeteorShower(80, 24, 5, m.seed)
	return m
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

var catArt = []string{
	`         /\_/\   `,
	`        ( o.o )  `,
	`         > ^ <   `,
	`        +-----+  `,
	`        |  @  |  `,
	`        +-----+  `,
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			return m, m.refresh()
		}

	case tea.WindowSizeMsg:
		m.meteors = render.GenerateMeteorShower(msg.Width, msg.Height, 5, m.seed)
		m.stars = render.GenerateStarField(msg.Width, msg.Height, 0.035, m.seed)

	case dashboardMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.summary = msg.summary
			m.err = ""
		}

	case time.Time:
		m.now = time.Time(msg)
	}

	return m, nil
}

func (m *DashboardModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()
	now := m.now.UnixMilli()

	if m.loading {
		msg := "Loading..."
		c.SetContent((w-len(msg))/2, h/2, render.StyleMuted, msg)
		return
	}

	for _, star := range m.stars {
		if star.X < w && star.Y < h {
			ch, visible := render.GetStarChar(star, now)
			if visible {
				style := render.StyleDim
				if ch == '★' {
					style = render.StyleBold
				}
				c.SetCell(star.X, star.Y, style, ch)
			}
		}
	}

	for _, meteor := range m.meteors {
		trail := render.GetMeteorTrail(meteor, now, w, h)
		for _, tp := range trail {
			style := render.StyleDim
			if tp.State == "bright" {
				style = render.StyleBold
			}
			if tp.X >= 0 && tp.X < w && tp.Y >= 0 && tp.Y < h {
				c.SetCell(tp.X, tp.Y, style, tp.Char)
			}
		}
	}

	if m.err != "" {
		c.SetContent(2, 2, render.StyleError, m.err)
		return
	}

	title := "dizz"
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

	barX := w - barW - 4
	barY := 3
	for i, it := range items {
		rowY := barY + i
		if rowY >= h-6 {
			break
		}
		label := fmt.Sprintf("  %s %d", it.label, it.count)
		c.SetContent(4, rowY, it.style, label)
		bar := render.RenderBar(it.count, maxCount, barW)
		c.SetContent(barX, rowY, it.style, bar)
	}

	catY := barY
	catX := 4
	if catX+14 < w {
		for i, line := range catArt {
			if catY+i >= h-4 {
				break
			}
			for x, ch := range line {
				c.SetCell(catX+x, catY+i, render.StyleHighlight, ch)
			}
		}
	}

	if s.ActiveTodos > 0 {
		todoMsg := fmt.Sprintf("TODOs: %d", s.ActiveTodos)
		c.SetContent(4, h-3, render.StyleWarning, todoMsg)
	}

	intentMsg := fmt.Sprintf("Intents: %d", s.Planned)
	c.SetContent(4, h-2, render.StyleInfo, intentMsg)

	help := "Press ? for help  r=refresh  q=quit"
	if len(help) > w-2 {
		help = help[:w-2]
	}
	c.SetContent((w-len(help))/2, h-1, render.StyleMuted, help)
}

func (m *DashboardModel) View() string { return "" }
