package views

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type SnapshotsModel struct {
	snapshots []dizzclient.SnapshotInfo
	filter    ui.Filter
	selected  int
	loading   bool
	err       string
	lastMsg   string

	diffContent string
	diffOffset  int
	showDiff    bool

	checkoutInfo string
	showCheckout bool

	showPruneModal bool
	pruneInput     string
}

func NewSnapshotsModel() *SnapshotsModel {
	return &SnapshotsModel{loading: true}
}

func (m *SnapshotsModel) Init() tea.Cmd {
	return m.refresh()
}

func (m *SnapshotsModel) refresh() tea.Cmd {
	return func() tea.Msg {
		entries, err := dizzclient.SnapshotList()
		if err != nil {
			return snapshotsMsg{err: err.Error()}
		}
		return snapshotsMsg{entries: entries}
	}
}

type snapshotsMsg struct {
	entries []dizzclient.SnapshotInfo
	err     string
}

type diffMsg struct {
	content string
	err     string
}

type createMsg struct {
	result string
	err    string
}

type checkoutMsg struct {
	info string
	err  string
}

type pruneMsg struct {
	result string
	err    string
}

func (m *SnapshotsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case snapshotsMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.snapshots = msg.entries
			m.err = ""
			if m.selected >= len(m.snapshots) {
				m.selected = 0
			}
		}

	case diffMsg:
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.diffContent = msg.content
			m.diffOffset = 0
			m.showDiff = true
		}
		m.lastMsg = ""

	case createMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.lastMsg = msg.result
			return m, m.refresh()
		}

	case checkoutMsg:
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.checkoutInfo = msg.info
			m.showCheckout = true
		}
		m.lastMsg = ""

	case pruneMsg:
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.lastMsg = msg.result
			return m, m.refresh()
		}
	}

	return m, nil
}

func (m *SnapshotsModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showDiff {
		switch msg.String() {
		case "esc", "enter", "d", "q":
			m.showDiff = false
		case "up":
			if m.diffOffset > 0 {
				m.diffOffset--
			}
		case "down":
			lines := strings.Split(m.diffContent, "\n")
			if m.diffOffset < len(lines)-1 {
				m.diffOffset++
			}
		}
		return m, nil
	}

	if m.showCheckout {
		if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "q" {
			m.showCheckout = false
		}
		return m, nil
	}

	if m.showPruneModal {
		return m.handlePruneKey(msg)
	}

	if consumed, changed := m.filter.HandleKey(msg.String()); consumed {
		if changed {
			m.selected = 0
		}
		return m, nil
	}

	switch msg.String() {
	case "up":
		m.selected--
		if m.selected < 0 {
			m.selected = 0
		}
	case "down":
		m.selected++
		if m.selected >= len(m.snapshots) {
			m.selected = len(m.snapshots) - 1
		}
	case "c":
		m.loading = true
		m.lastMsg = ""
		return m, m.runCreate()
	case "d":
		return m, m.runDiff()
	case "enter":
		return m, m.runCheckout()
	case "p":
		m.showPruneModal = true
		m.pruneInput = "10"
		return m, nil
	case "r":
		m.lastMsg = ""
		return m, m.refresh()
	}

	return m, nil
}

func (m *SnapshotsModel) handlePruneKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showPruneModal = false
		return m, nil
	case "enter":
		keep := 10
		fmt.Sscanf(m.pruneInput, "%d", &keep)
		if keep < 1 {
			keep = 1
		}
		m.showPruneModal = false
		m.lastMsg = fmt.Sprintf("Pruning snapshots (keep=%d)...", keep)
		return m, m.runPrune(keep)
	case "backspace":
		if len(m.pruneInput) > 0 {
			m.pruneInput = m.pruneInput[:len(m.pruneInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			ch := rune(msg.String()[0])
			if unicode.IsDigit(ch) {
				m.pruneInput += string(ch)
			}
		}
	}
	return m, nil
}

func (m *SnapshotsModel) runCreate() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		result, err := dizzclient.SnapshotCreate()
		if err != nil {
			return createMsg{err: err.Error()}
		}
		return createMsg{result: result}
	}
}

func (m *SnapshotsModel) runDiff() tea.Cmd {
	return func() tea.Msg {
		content, err := dizzclient.SnapshotDiff()
		if err != nil {
			return diffMsg{err: err.Error()}
		}
		return diffMsg{content: content}
	}
}

func (m *SnapshotsModel) runCheckout() tea.Cmd {
	if len(m.snapshots) == 0 || m.selected >= len(m.snapshots) {
		return nil
	}
	hash := m.snapshots[m.selected].Hash
	return func() tea.Msg {
		info, err := dizzclient.SnapshotCheckout(hash)
		if err != nil {
			return checkoutMsg{err: err.Error()}
		}
		return checkoutMsg{info: info}
	}
}

func (m *SnapshotsModel) runPrune(keep int) tea.Cmd {
	return func() tea.Msg {
		result, err := dizzclient.SnapshotPrune(keep)
		if err != nil {
			return pruneMsg{err: err.Error()}
		}
		return pruneMsg{result: result}
	}
}

func (m *SnapshotsModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	if m.loading && len(m.snapshots) == 0 {
		c.SetContent((w-len("Loading snapshots..."))/2, h/2, render.StyleMuted, "Loading snapshots...")
		return
	}

	if m.err != "" {
		c.SetContent(2, 2, render.StyleError, m.err)
	}

	header := "Snapshots"
	c.SetContent(2, 0, render.StyleHighlight.Bold(true), header)
	countStr := fmt.Sprintf("%d snapshots/deltas", len(m.snapshots))
	c.SetContent(w-2-len(countStr), 0, render.StyleMuted, countStr)

	filterHint := "c=create d=delta enter=checkout p=prune r=refresh  /=search"
	c.SetContent(2, 1, render.StyleDim, filterHint)

	if m.lastMsg != "" {
		c.SetContent(2, 2, render.StyleInfo, m.lastMsg)
	}

	y := 3
	overflow := false
	for i, snap := range m.snapshots {
		if y >= h-2 {
			overflow = true
			break
		}

		if m.filter.Active() && m.filter.Query != "" {
			if !strings.Contains(strings.ToLower(snap.Hash), strings.ToLower(m.filter.Query)) &&
				!strings.Contains(strings.ToLower(snap.Kind), strings.ToLower(m.filter.Query)) {
				continue
			}
		}

		style := render.StyleMuted
		prefix := "  "
		if i == m.selected {
			style = render.StyleHighlight
			prefix = "▸ "
		}

		line := fmt.Sprintf("%s%s  %s  %-8s %s", prefix, snap.Timestamp, snap.Hash, snap.Kind, snap.Size)
		c.SetContent(2, y, style, line)

		y++
	}

	if len(m.snapshots) == 0 {
		c.SetContent((w-26)/2, h/2, render.StyleMuted, "No snapshots yet. Press 'c' to create one.")
	} else if overflow {
		c.SetCell(2, h-2, render.StyleDim, '▼')
		c.SetContent(3, h-2, render.StyleDim, "more items...")
	}

	if m.showDiff {
		m.renderDiff(c)
	}

	if m.showCheckout {
		m.renderCheckout(c)
	}

	if m.showPruneModal {
		m.renderPruneModal(c)
	}

	m.filter.Render(c, 2, h-1)
}

func (m *SnapshotsModel) renderDiff(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	bodyW := w - 8
	if bodyW < 20 {
		bodyW = 20
	}
	bodyH := h - 6
	if bodyH < 5 {
		bodyH = 5
	}
	cx, cy := ui.RenderModalBox(c, "Snapshot Delta", bodyW, bodyH)

	maxLines := bodyH - 2
	lines := strings.Split(m.diffContent, "\n")
	y := 0
	for i := m.diffOffset; i < len(lines) && y < maxLines; i++ {
		line := lines[i]
		style := render.StyleDefault
		if strings.HasPrefix(line, "+") {
			style = render.StyleSuccess
		} else if strings.HasPrefix(line, "-") {
			style = render.StyleError
		} else if strings.HasPrefix(line, "@@") {
			style = render.StyleInfo
		}
		if len(line) > bodyW {
			line = line[:bodyW]
		}
		c.SetContent(cx, cy+y, style, line)
		y++
	}

	scrollHint := fmt.Sprintf("esc=close  ↑↓=scroll  %d/%d", m.diffOffset+1, len(lines))
	if len(scrollHint) > bodyW {
		scrollHint = fmt.Sprintf("↑↓ %d/%d", m.diffOffset+1, len(lines))
	}
	c.SetContent(cx, cy+maxLines, ui.StyleHelp, scrollHint)
}

func (m *SnapshotsModel) renderCheckout(c *render.Canvas) {
	lines := strings.Split(m.checkoutInfo, "\n")
	contentLines := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			contentLines++
		}
	}
	if contentLines < 3 {
		contentLines = 3
	}

	bodyW := 48
	bodyH := contentLines + 3
	cx, cy := ui.RenderModalBox(c, "Snapshot Checkout", bodyW, bodyH)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) > bodyW {
			line = line[:bodyW]
		}
		c.SetContent(cx, cy+i, render.StyleDefault, line)
	}

	c.SetContent(cx, cy+contentLines, ui.StyleHelp, "esc=close")
}

func (m *SnapshotsModel) renderPruneModal(c *render.Canvas) {
	bodyW := 44
	bodyH := 7
	cx, cy := ui.RenderModalBox(c, "Prune Snapshots", bodyW, bodyH)

	help := "Enter=confirm  Esc=cancel"
	c.SetContent(cx, cy, ui.StyleHelp, help)
	cy += 2

	c.SetContent(cx, cy, render.StyleDefault, "Keep last N checkpoints: ")
	input := m.pruneInput
	if input == "" {
		input = "10"
	}
	c.SetContent(cx+25, cy, render.StyleHighlight, input+"_")
	cy += 2

	c.SetContent(cx, cy, render.StyleDim, fmt.Sprintf("Currently %d snapshots, will prune older ones", len(m.snapshots)))
}

// @dizz-ignore-unused
func (m *SnapshotsModel) View() string { return "" }

func (m *SnapshotsModel) InputMode() bool { return m.filter.Active() || m.showPruneModal }
