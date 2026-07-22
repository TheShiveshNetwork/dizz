package views

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
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

var dizzArtBase = []string{
	`                  ⣀⣿⣷⣤⣀                              ⣀⣤⣷⣿⣀`,
	`                  ⣤█⣿⣷⣤⣷⣷⣷⣷⣀                    ⣤⣷⣿⣷⣷⣤⣿⣿▓⣀`,
	`                  ⣤⣿⣀░▓⣿⣀ ⣀⣤⣿⣷⣀                ⣤⣿⣿⣷⣀ ⣀⣿█░⣀⣿⣀`,
	`                  ⣤⣿  ⣤▓█░⣤  ⣀⣷⣿⣷▒▓░⣤░▓▒⣤░▓▒⣷⣿⣷⣀  ⣤▒█▓⣤  ⣿⣤`,
	`                  ⣤⣿   ⣀▒██▒⣤    ▒█░ ▒█▒ ⣿█▓    ⣤▒██▒⣀    ⣿⣤`,
	`                  ⣀░    ⣤▓⣿⣀     ▒█░ ▒█▒ ⣿█▒     ⣀░█⣤    ⣀░⣀`,
	`                   ⣤░  ⣤⣤⣀        ▒█░ ▒█▒ ⣿█▒        ⣀⣤⣤  ⣿⣤`,
	`                    ⣿⣷⣤⣀          ⣤░⣤ ░█▒ ⣤░⣤         ⣀⣤⣤⣿`,
	`                    ⣀░                ⣀⣤⣀                ⣀▒⣀`,
	`                    ⣿⣤                                   ⣷⣷`,
	`                   ⣀░                                     ░⣀`,
	`                   ⣷░⣤⣤⣤⣤⣤⣀  ⣤░░⣿⣤        ⣀⣤░░░⣤⣀ ⣀⣤⣤⣤⣤⣤░⣀`,
	`                ⣀⣷⣷⣿█████░⣀⣀▒⣿⣀⣀⣤░⣿        ▒⣿⣤⣀⣤⣿░ ⣀░█████⣷⣷⣷⣀`,
	`                ⣀⣷⣷▓▒░░⣤            ⣷⣷⣷⣿⣤            ⣤▒▒▓█⣷⣷⣀`,
	`                ⣀⣀ ⣷█▓▒⣤            ⣤⣿⣿⣿⣀            ⣤⣿⣷░⣿ ⣀⣀`,
	`                  ⣀⣷⣷▒▒░⣤      ⣀⣷⣀   ⣤▒⣀   ⣀⣷        ⣀░▓░⣷⣷⣀`,
	`                  ⣀⣀ ⣀⣷⣿⣀       ⣤⣿⣷⣷⣷⣷⣀⣷⣷⣷⣷⣷⣀       ⣀⣿⣷  ⣀⣀`,
	`                       ⣀⣷⣿⣤                        ⣤⣿⣷⣀`,
	`                       ⣤▒██▓░⣷⣤⣀                ⣀⣤⣿░▓██▒⣷`,
	`                     ⣀⣿██▓▒▒▓██▒⣷⣷⣿⣷⣷⣷⣷⣷⣷⣿⣷⣷░███▒░▒██░⣀`,
	`                     ⣿⣿⣿⣿⣀   ⣀⣤                ⣷⣀   ⣀⣿⣿⣿░`,
	`                  ⣀⣷⣿⣷⣷⣷⣿⣷⣀                    ⣀⣷⣿⣿⣷⣷⣿⣿⣤`,
	`                ⣀⣿⣷⣀     ⣀▒⣤                    ⣤░⣀     ⣀⣷⣿⣀`,
	`               ⣀░⣀        ⣀⣀⣷░⣀                  ⣀⣿⣷⣤⣀        ⣀⣿⣤`,
	`          ⣀░░░░▒⣀        ⣀▒⣤░⣤                  ⣀░⣤▒⣀        ⣀░░░░░⣤`,
	`        ⣀▒█▓▒█░ ░ ░  ⣀⣿ ⣀███▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓███⣤ ⣷⣤░  ▒ ⣿██▒██⣤`,
	`        ⣤▓█▓░▒██▒▓⣷⣤▓⣤⣷▒⣤░███⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿▓██▒⣤░⣿⣤▓⣷⣤█▒██▓░░██⣷`,
	`       ⣤██░░█░▒███████████▒▓█░░██░░█▒░▓█░░█▓░▒█░░████████████▒⣿░░░░██⣿`,
	`      ⣷██▒⣿░▓▒░░█▓░▒█▒▒██▒▒▒█▒░▓█░░▓▓░▒█▒░▓█░░█▓░▒█████▒▓█░░█▓░░▓▒⣿░▓█▒⣀`,
	`     ░██░⣿▒▒⣿░██▒░█▓░▒█▓⣿▒█▓░▒█▒⣿▒█░░▓█░░▓█░⣿▓█▒░▓█░⣿▓█▒░▒█▒░▓█▒⣿░▒▒⣿▓█▓⣀`,
	`  ⣀▒█▓▒▒▓▒▒▒▓▒▒▓▓▒░▓▒▒▒▓▓▒▒▓▓▓░▓▒▒░░▒█▒░▒█▓░▒█▓░░██░░▓█░░▓█▒▒▓█░░██▒░▒█▓⣤`,
	` ⣀▒█░░▒█▒⣿▓█▒░▓█▒░▓█░⣿▓█▒⣿▒▒⣿⣿⣿░⣿⣿⣿███▓⣿░█▓⣿⣿██░⣿▒█▓⣿░█▓⣿░█████▓⣿░██▒░▒█▓⣤`,
	`⣀▓████████████████████████████████████████████████████████████████████████⣷`,
	`⣿█████████████████████████████████████████████████████████████████████████▒`,
	`⣿█████████████████████████████████████████████████████████████████████████▒`,
}

const (
	handRowStart       = 21
	handRowEnd         = 25
	handFrameCount     = 10
	handAnimIntervalMs = 130
	handShiftAmplitude = 1.0
)

// Specific horizontal index ranges for left and right paws
// so that shifting paws doesn't affect adjacent elements or the keyboard
type pawBounds struct {
	leftStart, leftEnd   int
	rightStart, rightEnd int
}

var pawBoundsByRow = map[int]pawBounds{
	21: {leftStart: 18, leftEnd: 35, rightStart: 55, rightEnd: 72},
	22: {leftStart: 16, leftEnd: 31, rightStart: 51, rightEnd: 66},
	23: {leftStart: 15, leftEnd: 30, rightStart: 50, rightEnd: 65},
	24: {leftStart: 10, leftEnd: 28, rightStart: 48, rightEnd: 68},
	25: {leftStart: 8, leftEnd: 25, rightStart: 45, rightEnd: 65},
}

var dizzArtFrames = generateTypingFrames(dizzArtBase)

func padRuneRow(row string, width int) string {
	rs := []rune(row)
	if len(rs) >= width {
		return string(rs[:width])
	}
	return row + strings.Repeat(" ", width-len(rs))
}

func normalizeArt(lines []string) ([]string, int) {
	width := 0
	for _, l := range lines {
		if w := len([]rune(l)); w > width {
			width = w
		}
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = padRuneRow(l, width)
	}
	return out, width
}

func shiftSegment(rowRunes []rune, start, end, offset int) {
	if start < 0 || end > len(rowRunes) || start >= end {
		return
	}
	length := end - start
	segment := make([]rune, length)
	copy(segment, rowRunes[start:end])

	shifted := make([]rune, length)
	for i := range shifted {
		shifted[i] = ' '
	}
	for i, r := range segment {
		dst := i + offset
		if dst >= 0 && dst < length {
			shifted[dst] = r
		}
	}
	copy(rowRunes[start:end], shifted)
}

func generateTypingFrames(base []string) [][]string {
	normalized, _ := normalizeArt(base)

	frames := make([][]string, handFrameCount)
	for f := 0; f < handFrameCount; f++ {
		phase := 2 * math.Pi * float64(f) / float64(handFrameCount)
		offset := int(math.Round(handShiftAmplitude * math.Sin(phase)))

		frame := make([]string, len(normalized))
		copy(frame, normalized)

		for r := handRowStart; r <= handRowEnd; r++ {
			bounds, ok := pawBoundsByRow[r]
			if !ok {
				continue
			}
			rowRunes := []rune(normalized[r])

			// Apply horizontal shift strictly within paw boundaries
			shiftSegment(rowRunes, bounds.leftStart, bounds.leftEnd, offset)
			shiftSegment(rowRunes, bounds.rightStart, bounds.rightEnd, -offset)

			frame[r] = string(rowRunes)
		}
		frames[f] = frame
	}
	return frames
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

	frame := m.frameTick % int64(len(dizzArtFrames))
	artLines := dizzArtFrames[frame]
	artH := len(artLines)
	maxW := 0
	for _, line := range artLines {
		lw := runewidth.StringWidth(line)
		if lw > maxW {
			maxW = lw
		}
	}
	startX := (w - maxW) / 2
	if startX < 2 {
		startX = 2
	}
	startY := (h - artH) / 2
	if startY < 4 {
		startY = 4
	}
	for i, line := range artLines {
		destY := startY + i
		if destY >= h-4 {
			break
		}
		c.SetContent(startX, destY, render.StyleDefault, line)
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

func (m *DashboardModel) View() string { return "" }
