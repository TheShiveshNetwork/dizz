package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
)

type SymbolsModel struct {
	symbols     []client.Symbol
	table       ui.TableModel
	filter      ui.Filter
	stateFilter *ui.StateFilter
	showDetail  bool
	loading     bool
	err         string
}

func NewSymbolsModel() *SymbolsModel {
	return &SymbolsModel{
		loading: true,
		table: ui.TableModel{
			Cols: []ui.Column{
				{Name: "Name", Width: 28, Sortable: true},
				{Name: "State", Width: 10, Sortable: true},
				{Name: "File", Width: 42, Sortable: false},
			},
		},
		stateFilter: ui.NewStateFilter(map[string]string{
			"a": "active",
			"p": "planned",
			"u": "unstable",
			"s": "unused",
			"d": "abandoned",
		}),
	}
}

func (m *SymbolsModel) Init() tea.Cmd {
	return m.refresh()
}

func (m *SymbolsModel) refresh() tea.Cmd {
	return func() tea.Msg {
		symbols, err := client.LogDump()
		if err != nil {
			return symbolsMsg{err: err.Error()}
		}
		return symbolsMsg{symbols: symbols}
	}
}

type symbolsMsg struct {
	symbols []client.Symbol
	err     string
}

func (m *SymbolsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case symbolsMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.symbols = msg.symbols
			m.err = ""
			m.buildTable()
		}

	case ui.RefreshTick:
		return m, m.refresh()
	}

	return m, nil
}

func (m *SymbolsModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	consumed, changed := m.filter.HandleKey(msg.String())
	if consumed {
		if changed {
			m.buildTable()
		}
		return m, nil
	}

	switch msg.String() {
	case "up":
		m.table.PrevRow()
		m.showDetail = false
	case "down":
		m.table.NextRow()
		m.showDetail = false
	case "r":
		m.loading = true
		return m, m.refresh()
	case "1":
		m.table.SortBy(0)
		m.buildTable()
	case "2":
		m.table.SortBy(1)
		m.buildTable()
	case "enter":
		m.toggleDetail()
	default:
		if handled, changed := m.stateFilter.HandleKey(msg.String()); handled {
			if changed {
				m.buildTable()
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *SymbolsModel) toggleDetail() {
	if m.showDetail {
		m.showDetail = false
	} else if len(m.table.Rows) > 0 && m.table.Selected < len(m.table.Rows) {
		m.showDetail = true
	}
}

func (m *SymbolsModel) buildTable() {
	filtered := m.symbols

	if m.stateFilter.Active() {
		var f []client.Symbol
		for _, s := range filtered {
			if s.State == m.stateFilter.Value() {
				f = append(f, s)
			}
		}
		filtered = f
	} else {
		var f []client.Symbol
		for _, s := range filtered {
			if s.State != "active" {
				f = append(f, s)
			}
		}
		filtered = f
	}

	if m.filter.Active() && m.filter.Query != "" {
		var f []client.Symbol
		for _, s := range filtered {
			if m.filter.MatchesAny(s.Name, client.RelPath(s.File), s.State) {
				f = append(f, s)
			}
		}
		filtered = f
	}

	m.table.Rows = make([]ui.Row, len(filtered))
	for i, s := range filtered {
		relFile := client.RelPath(s.File)
		fileLine := fmt.Sprintf("%s:%d", relFile, s.Line)
		if len(fileLine) > 42 {
			fileLine = "..." + fileLine[len(fileLine)-39:]
		}

		stateStyle := render.StateColor(s.State)

		m.table.Rows[i] = ui.Row{
			Cells: []string{s.Name, s.State, fileLine},
			Style: stateStyle,
			Data:  s,
		}
	}

	m.sortRows()

	if m.table.Selected >= len(m.table.Rows) {
		m.table.Selected = len(m.table.Rows) - 1
		if m.table.Selected < 0 {
			m.table.Selected = 0
		}
	}
}

func (m *SymbolsModel) sortRows() {
	col := m.table.SortCol
	if col < 0 || col >= len(m.table.Cols) || !m.table.Cols[col].Sortable {
		return
	}
	sort.SliceStable(m.table.Rows, func(i, j int) bool {
		a, b := "", ""
		if col < len(m.table.Rows[i].Cells) {
			a = m.table.Rows[i].Cells[col]
		}
		if col < len(m.table.Rows[j].Cells) {
			b = m.table.Rows[j].Cells[col]
		}
		if m.table.SortAsc {
			return a < b
		}
		return a > b
	})
}

func (m *SymbolsModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	if m.loading {
		msg := "Loading symbols..."
		c.SetContent((w-len(msg))/2, h/2, render.StyleMuted, msg)
		return
	}

	if m.err != "" {
		c.SetContent(2, 2, render.StyleError, m.err)
		return
	}

	header := m.buildHeader()
	c.SetContent(2, 0, render.StyleHighlight.Bold(true), header)

	countStr := fmt.Sprintf("%d symbols", len(m.table.Rows))
	c.SetContent(w-2-len(countStr), 0, render.StyleMuted, countStr)

	filterHint := "1=sort:name 2=sort:state  a/p/u/s/d=filter x=clear  /=search  enter=detail"
	c.SetContent(2, 1, render.StyleDim, filterHint)

	tableH := h
	if m.showDetail && len(m.table.Rows) > 0 && m.table.Selected < len(m.table.Rows) {
		if sym, ok := m.table.Rows[m.table.Selected].Data.(client.Symbol); ok {
			tableH = h - detailHeight(sym) - 2
		}
	}
	if tableH < 3 {
		tableH = 3
	}

	nameW := 28
	stateW := 10
	fileW := w - 4 - nameW - stateW
	if fileW < 20 {
		nameW = w - 4 - stateW - 20
		if nameW < 10 {
			nameW = 10
		}
		fileW = w - 4 - nameW - stateW
	}
	m.table.Cols[0].Width = nameW
	m.table.Cols[1].Width = stateW
	m.table.Cols[2].Width = fileW

	ui.RenderTable(c, &m.table, 2, 2, w-4, tableH)

	if m.showDetail && len(m.table.Rows) > 0 && m.table.Selected < len(m.table.Rows) {
		m.renderDetail(c, tableH+1)
	}

	m.filter.Render(c, 2, h-1)
}

func (m *SymbolsModel) buildHeader() string {
	parts := []string{"Symbols"}
	if m.stateFilter.Active() {
		parts = append(parts, " ["+m.stateFilter.Value()+"]")
	} else {
		parts = append(parts, " [non-active]")
	}
	if m.filter.Query != "" {
		parts = append(parts, " /"+m.filter.Query)
	}
	return strings.Join(parts, "")
}

func detailHeight(sym client.Symbol) int {
	n := 4
	if sym.Type != "" {
		n++
	}
	if sym.Language != "" {
		n++
	}
	if sym.ChurnCount > 0 {
		n++
	}
	if sym.Confidence > 0 {
		n++
	}
	if sym.HasTodo {
		n++
	}
	return n
}

func (m *SymbolsModel) renderDetail(c *render.Canvas, y int) {
	w := c.Width()
	row := m.table.Rows[m.table.Selected]
	sym, ok := row.Data.(client.Symbol)
	if !ok {
		return
	}

	dh := detailHeight(sym)
	border := render.StyleMuted

	maxY := y + dh
	if maxY >= c.Height() {
		maxY = c.Height() - 1
	}

	c.SetCell(2, y, border, '┌')
	title := fmt.Sprintf(" %s ", sym.Name)
	c.SetContent(3, y, render.StyleHighlight.Bold(true), title)
	for x := 3 + len(title); x < w-3; x++ {
		c.SetCell(x, y, border, '─')
	}
	c.SetCell(w-3, y, border, '┐')
	y++

	detail := []struct {
		label    string
		value    string
		valStyle tcell.Style
	}{
		{"File", fmt.Sprintf("%s:%d", client.RelPath(sym.File), sym.Line), render.StyleInfo},
		{"State", sym.State, render.StateColor(sym.State)},
	}

	if sym.Type != "" {
		detail = append(detail, struct {
			label    string
			value    string
			valStyle tcell.Style
		}{"Type", sym.Type, render.StyleInfo})
	}
	if sym.Language != "" {
		detail = append(detail, struct {
			label    string
			value    string
			valStyle tcell.Style
		}{"Lang", sym.Language, render.StyleInfo})
	}
	if sym.ChurnCount > 0 {
		detail = append(detail, struct {
			label    string
			value    string
			valStyle tcell.Style
		}{"Churn", fmt.Sprintf("%d changes", sym.ChurnCount), render.StyleWarning})
	}
	if sym.Confidence > 0 {
		detail = append(detail, struct {
			label    string
			value    string
			valStyle tcell.Style
		}{"Conf", fmt.Sprintf("%.0f%%", sym.Confidence*100), render.StyleSuccess})
	}
	if sym.HasTodo {
		detail = append(detail, struct {
			label    string
			value    string
			valStyle tcell.Style
		}{"Marker", "TODO/FIXME present", render.StyleError})
	}

	for _, d := range detail {
		if y >= c.Height()-1 || y >= maxY {
			break
		}
		c.SetCell(2, y, border, '│')
		labelW := 7
		c.SetContent(3, y, tcell.StyleDefault.Bold(true), fmt.Sprintf("%-*s", labelW, d.label))
		c.SetContent(3+labelW, y, d.valStyle, d.value)
		c.SetCell(w-3, y, border, '│')
		y++
	}

	if y < c.Height() && y <= maxY {
		c.SetCell(2, y, border, '└')
		for x := 3; x < w-3; x++ {
			c.SetCell(x, y, border, '─')
		}
		c.SetCell(w-3, y, border, '┘')
	}
}

// @dizz-ignore-unused
func (m *SymbolsModel) View() string { return "" }

func (m *SymbolsModel) InputMode() bool { return m.filter.Active() }
