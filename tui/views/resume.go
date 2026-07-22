package views

import (
	"fmt"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	tea "github.com/charmbracelet/bubbletea"
)

type ResumeModel struct {
	summary  *dizzclient.Summary
	planned  []dizzclient.Symbol
	symbols  []dizzclient.Symbol
	loadedAt time.Time
	loading  bool
	err      string
}

func NewResumeModel() *ResumeModel {
	return &ResumeModel{loading: true}
}

func (m *ResumeModel) Init() tea.Cmd {
	return m.refresh()
}

func (m *ResumeModel) refresh() tea.Cmd {
	return func() tea.Msg {
		symbols, err := dizzclient.LogDump()
		if err != nil {
			return resumeMsg{err: err.Error()}
		}
		summary, err := dizzclient.Status()
		if err != nil {
			return resumeMsg{err: err.Error()}
		}
		return resumeMsg{symbols: symbols, summary: summary}
	}
}

type resumeMsg struct {
	symbols []dizzclient.Symbol
	summary *dizzclient.Summary
	err     string
}

func (m *ResumeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "r" {
			m.loading = true
			return m, m.refresh()
		}

	case resumeMsg:
		m.loading = false
		m.loadedAt = time.Now()
		if msg.err != "" {
			m.err = msg.err
		} else if msg.symbols != nil {
			m.symbols = msg.symbols
			m.summary = msg.summary
			m.err = ""

			var planned []dizzclient.Symbol
			for _, s := range msg.symbols {
				if s.State == "planned" {
					planned = append(planned, s)
				}
			}
			m.planned = planned
		}
	}

	return m, nil
}

func (m *ResumeModel) Render(c *render.Canvas) {
	w := c.Width()

	if m.loading {
		c.SetContent((w-20)/2, c.Height()/2, render.StyleMuted, "Loading project state...")
		return
	}

	if m.err != "" {
		c.SetContent((w-len(m.err)-4)/2, c.Height()/2, render.StyleError, m.err)
		return
	}

	s := m.summary
	if s == nil {
		return
	}

	y := 2

	timeSince := "Just now"
	if !m.loadedAt.IsZero() {
		ago := time.Since(m.loadedAt)
		if ago < time.Minute {
			timeSince = "Less than a minute ago"
		} else if ago < time.Hour {
			timeSince = fmt.Sprintf("%d minutes ago", int(ago.Minutes()))
		} else if ago < 24*time.Hour {
			timeSince = fmt.Sprintf("%d hours ago", int(ago.Hours()))
		} else {
			timeSince = fmt.Sprintf("%d days ago", int(ago.Hours()/24))
		}
	}

	title := fmt.Sprintf("Resume — %s", timeSince)
	c.SetContent((w-len(title))/2, y, render.StyleHighlight, title)
	y += 2

	issues := s.Planned + s.Unstable + s.Unused + s.Abandoned

	summaryStr := fmt.Sprintf("✓ %d symbols working well", s.Active)
	c.SetContent(2, y, render.StyleSuccess, summaryStr)
	y++

	if issues > 0 {
		issueStr := fmt.Sprintf("⚠ %d items need attention", issues)
		c.SetContent(2, y, render.StyleWarning, issueStr)
		y++
	}

	y++

	if len(m.planned) > 0 {
		c.SetContent(2, y, render.StyleHighlight, "Planned work items:")
		y++
		limit := 5
		if len(m.planned) < limit {
			limit = len(m.planned)
		}
		for i := 0; i < limit; i++ {
			if y >= c.Height()-2 {
				break
			}
			sym := m.planned[i]
			fileLine := fmt.Sprintf("%s:%d", dizzclient.RelPath(sym.File), sym.Line)
			if len(fileLine) > w-20 {
				fileLine = "..." + fileLine[len(fileLine)-w+23:]
			}
			line := fmt.Sprintf("  • %s  %s", sym.Name, fileLine)
			c.SetContent(2, y, render.StyleInfo, line)
			y++
		}
		if len(m.planned) > limit {
			more := fmt.Sprintf("  ... and %d more", len(m.planned)-limit)
			c.SetContent(2, y, render.StyleMuted, more)
			y++
		}
		y++
	}

	y++
	c.SetContent(2, y, render.StyleHighlight, "Quick Summary")
	y++
	c.SetContent(4, y, render.StyleMuted, fmt.Sprintf("Total symbols: %d", s.TotalSymbols))
	y++
	c.SetContent(4, y, render.StyleMuted, fmt.Sprintf("Files tracked: %d", s.TotalFiles))
	y++
	c.SetContent(4, y, render.StyleMuted, fmt.Sprintf("Active TODOs: %d", s.ActiveTodos))
}

func (m *ResumeModel) View() string { return "" }
