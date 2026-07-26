package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type DashboardModel struct {
	summary   *client.Summary
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
		s, err := client.Status()
		if err != nil {
			return dashboardMsg{err: err.Error()}
		}
		return dashboardMsg{summary: s}
	}
}

type dashboardMsg struct {
	summary *client.Summary
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
	scoreStr := fmt.Sprintf("Code Score: %d%%", codeScore)
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
	barX := w - barW - 8
	barY := 3
	for i, it := range items {
		rowY := barY + i
		if rowY >= h-6 {
			break
		}
		xOff := 0
		if i == 0 {
			xOff = -1
		}
		bar := render.RenderBar(it.count, maxCount, barW)
		c.SetContent(barX+xOff, rowY, it.style, bar)
		if code, ok := shortCodes[it.label]; ok {
			c.SetContent(barX+xOff+barW+1, rowY, render.StyleDim, code)
		}
	}

	_, artH := render.DizzArtDimensions()
	startX := w / 2
	startY := (h - artH) / 2
	if startY < 4 {
		startY = 4
	}
	render.RenderDizzie(c, m.frameTick, startX, startY)

	nextItems := m.buildNextActions()
	if len(nextItems) > 0 {
		bulbStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(255, 200, 50)).Bold(true)
		itemStyle := render.StyleDefault
		nextW := barX - 4
		if nextW < 20 {
			nextW = 20
		}
		if nextW > 45 {
			nextW = 45
		}

		headerText := "What to do next"
		c.SetContent(2, barY, bulbStyle, "\u26A1 "+headerText)

		for i, item := range nextItems {
			rowY := barY + 1 + i
			if rowY >= h-5 {
				break
			}
			wrapped := wrapText(item, nextW-3)
			for j, line := range wrapped {
				if rowY+j >= h-5 {
					break
				}
				prefix := "  "
				if j == 0 {
					prefix = "\u2022 "
				}
				display := prefix + line
				if runewidth.StringWidth(display) > nextW {
					display = display[:nextW-3] + "..."
				}
				c.SetContent(2, rowY+j, itemStyle, display)
			}
			if len(wrapped) > 1 {
				rowY += len(wrapped) - 1
			}
		}
	}

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

func (m *DashboardModel) buildNextActions() []string {
	s := m.summary
	if s == nil {
		return nil
	}

	var items []string

	if s.Planned > 0 {
		items = append(items, fmt.Sprintf("Implement %d planned symbol%s", s.Planned, plural(s.Planned)))
	}
	if s.Unstable > 0 {
		items = append(items, fmt.Sprintf("Stabilize %d high-churn symbol%s", s.Unstable, plural(s.Unstable)))
	}
	if s.Unused > 0 {
		items = append(items, fmt.Sprintf("Connect or remove %d unused symbol%s", s.Unused, plural(s.Unused)))
	}
	if s.Abandoned > 0 {
		items = append(items, fmt.Sprintf("Review %d abandoned symbol%s for cleanup", s.Abandoned, plural(s.Abandoned)))
	}
	if s.ActiveTodos > 0 {
		items = append(items, fmt.Sprintf("Address %d TODO/FIXME comment%s", s.ActiveTodos, plural(s.ActiveTodos)))
	}
	if s.Intents > 0 {
		items = append(items, fmt.Sprintf("Resolve %d open intent%s", s.Intents, plural(s.Intents)))
	}
	if len(items) == 0 {
		items = append(items, "All clean! Consider adding tests or docs")
	}
	return items
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func wrapText(text string, maxWidth int) []string {
	if maxWidth < 10 {
		return []string{text}
	}
	if runewidth.StringWidth(text) <= maxWidth {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var current strings.Builder
	currentW := 0

	for _, word := range words {
		wordW := runewidth.StringWidth(word)
		if current.Len() == 0 {
			if wordW > maxWidth {
				lines = append(lines, word[:maxWidth])
				continue
			}
			current.WriteString(word)
			currentW = wordW
		} else if currentW+1+wordW <= maxWidth {
			current.WriteString(" ")
			current.WriteString(word)
			currentW += 1 + wordW
		} else {
			lines = append(lines, current.String())
			current.Reset()
			currentW = 0
			if wordW > maxWidth {
				lines = append(lines, word[:maxWidth])
				continue
			}
			current.WriteString(word)
			currentW = wordW
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

// @dizz-ignore-unused
func (m *DashboardModel) View() string { return "" }
