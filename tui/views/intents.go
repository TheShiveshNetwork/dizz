package views

import (
	"fmt"
	"sort"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
)

var intentTypes = []string{"todo", "fixme", "refactor", "question", "hack", "temporary"}
var severityLabels = []string{"none", "low", "medium", "high"}

var typePriority = map[string]int{
	"fixme":     0,
	"todo":      1,
	"question":  2,
	"refactor":  3,
	"hack":      4,
	"temporary": 5,
}

type IntentsModel struct {
	intents     []dizzclient.Intent
	table       ui.TableModel
	filter      ui.Filter
	stateFilter *ui.StateFilter
	sevFilter   *ui.StateFilter
	typeFilter  *ui.StateFilter
	loading     bool
	err         string

	showModal     bool
	modalType     string
	focusIdx      int
	validationErr string

	intentType int
	inputMsg   string
	msgCursor  int
	inputSev   int

	resolveIdx int
	sortMode   int // 0=severity, 1=type priority
}

func NewIntentsModel() *IntentsModel {
	m := &IntentsModel{
		loading: true,
		table: ui.TableModel{
			Cols: []ui.Column{
				{Name: "Type", Width: 10, Sortable: false},
				{Name: "Severity", Width: 8, Sortable: true},
				{Name: "Message", Width: 80, Sortable: false},
			},
			ColGap: 2,
		},
		stateFilter: ui.NewStateFilter(map[string]string{
			"a": "active",
			"s": "resolved",
		}),
		sevFilter: ui.NewStateFilter(map[string]string{
			"0": "0",
			"1": "1",
			"2": "2",
			"3": "3",
		}),
		typeFilter: ui.NewStateFilter(map[string]string{
			"t": "todo",
			"f": "fixme",
			"r": "refactor",
			"q": "question",
			"h": "hack",
			"m": "temporary",
		}),
	}
	m.stateFilter.SetValue("active")
	m.table.SortCol = 1
	m.table.SortAsc = false
	return m
}

func (m *IntentsModel) Init() tea.Cmd {
	return m.refresh()
}

func (m *IntentsModel) refresh() tea.Cmd {
	return func() tea.Msg {
		intents, err := dizzclient.ListIntents()
		if err != nil {
			return intentsMsg{err: err.Error()}
		}
		return intentsMsg{intents: intents}
	}
}

type intentsMsg struct {
	intents []dizzclient.Intent
	err     string
}

func (m *IntentsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case intentsMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.intents = msg.intents
			m.err = ""
			m.buildTable()
		}
	}

	return m, nil
}

func (m *IntentsModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showModal {
		return m.handleModalKey(msg)
	}

	if consumed, changed := m.filter.HandleKey(msg.String()); consumed {
		if changed {
			m.buildTable()
		}
		return m, nil
	}

	if msg.String() == "x" {
		m.stateFilter.Reset()
		m.sevFilter.Reset()
		m.typeFilter.Reset()
		m.buildTable()
		return m, nil
	}

	if handled, changed := m.typeFilter.HandleKey(msg.String()); handled {
		if changed {
			m.buildTable()
		}
		return m, nil
	}

	if handled, changed := m.sevFilter.HandleKey(msg.String()); handled {
		if changed {
			m.buildTable()
		}
		return m, nil
	}

	if handled, changed := m.stateFilter.HandleKey(msg.String()); handled {
		if changed {
			m.buildTable()
		}
		return m, nil
	}

	switch msg.String() {
	case "up":
		m.table.PrevRow()
	case "down":
		m.table.NextRow()
	case "enter":
		if len(m.table.Rows) > 0 && m.table.Selected < len(m.table.Rows) {
			m.resolveIdx = m.table.Selected
			m.showModal = true
			m.modalType = "resolve"
			m.focusIdx = 0
		}
	case "i":
		m.showModal = true
		m.modalType = "add"
		m.focusIdx = 0
		m.intentType = 0
		m.inputMsg = ""
		m.msgCursor = 0
		m.inputSev = 1
		m.validationErr = ""
	case "g":
		m.sortMode = 0
		m.table.SortCol = 1
		m.table.SortAsc = false
		m.buildTable()
	case "p":
		m.sortMode = 1
		m.table.SortAsc = !m.table.SortAsc
		m.buildTable()
	}

	return m, nil
}

func (m *IntentsModel) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalFields := 5
	if m.modalType == "resolve" {
		totalFields = 2
	}

	switch msg.String() {
	case "esc":
		m.showModal = false
		m.validationErr = ""
		return m, nil

	case "tab", "down":
		m.focusIdx = (m.focusIdx + 1) % totalFields
		return m, nil

	case "shift+tab", "up":
		m.focusIdx = (m.focusIdx - 1 + totalFields) % totalFields
		return m, nil

	case "enter":
		return m.handleModalEnter()

	case "left":
		if m.modalType == "add" {
			switch m.focusIdx {
			case 0:
				m.intentType = (m.intentType - 1 + len(intentTypes)) % len(intentTypes)
			case 1:
				if m.msgCursor > 0 {
					m.msgCursor--
				}
			case 2:
				if m.inputSev > 0 {
					m.inputSev--
				}
			}
		}
		return m, nil

	case "right":
		if m.modalType == "add" {
			switch m.focusIdx {
			case 0:
				m.intentType = (m.intentType + 1) % len(intentTypes)
			case 1:
				if m.msgCursor < len(m.inputMsg) {
					m.msgCursor++
				}
			case 2:
				if m.inputSev < 3 {
					m.inputSev++
				}
			}
		}
		return m, nil

	case "backspace":
		if m.modalType == "add" && m.focusIdx == 1 && m.msgCursor > 0 {
			m.inputMsg = m.inputMsg[:m.msgCursor-1] + m.inputMsg[m.msgCursor:]
			m.msgCursor--
		}
		return m, nil

	case "home":
		if m.modalType == "add" && m.focusIdx == 1 {
			m.msgCursor = 0
		}
		return m, nil

	case "end":
		if m.modalType == "add" && m.focusIdx == 1 {
			m.msgCursor = len(m.inputMsg)
		}
		return m, nil

	default:
		if m.modalType == "add" && m.focusIdx == 1 && len(msg.String()) == 1 {
			ch := msg.String()[0]
			before := m.inputMsg[:m.msgCursor]
			after := m.inputMsg[m.msgCursor:]
			m.inputMsg = before + string(ch) + after
			m.msgCursor++
		}
		return m, nil
	}
}

func (m *IntentsModel) handleModalEnter() (tea.Model, tea.Cmd) {
	if m.modalType == "resolve" {
		if m.focusIdx == 1 {
			idx := m.resolveIdx
			if idx < len(m.intents) {
				id := m.intents[idx].ID
				go dizzclient.IntentResolve(id)
			}
			m.showModal = false
			m.validationErr = ""
			return m, m.refresh()
		}
		m.showModal = false
		m.validationErr = ""
		return m, nil
	}

	if m.modalType == "add" {
		if m.focusIdx == 4 {
			if m.inputMsg == "" {
				m.validationErr = "Message is required"
				return m, nil
			}
			go dizzclient.IntentAdd(m.inputMsg, intentTypes[m.intentType], m.inputSev, nil)
			m.showModal = false
			m.validationErr = ""
			return m, m.refresh()
		}
		if m.focusIdx == 3 {
			m.showModal = false
			m.validationErr = ""
		}
		return m, nil
	}

	return m, nil
}

func (m *IntentsModel) buildHeader() string {
	parts := "Intents"
	sf := m.stateFilter.Value()
	if sf == "active" {
		parts += " [Active]"
	} else if sf == "resolved" {
		parts += " [Resolved]"
	} else if sf == "" && m.sevFilter.Value() == "" && m.typeFilter.Value() == "" {
		parts += " [All]"
	}
	if sv := m.sevFilter.Value(); sv != "" {
		idx := int(sv[0] - '0')
		if idx >= 0 && idx <= 3 {
			parts += " [sev:" + severityLabels[idx] + "]"
		}
	}
	if tv := m.typeFilter.Value(); tv != "" {
		parts += " [" + tv + "]"
	}
	if m.sortMode == 1 {
		if m.table.SortAsc {
			parts += " [type▲]"
		} else {
			parts += " [type▼]"
		}
	}
	return parts
}

func (m *IntentsModel) sortRows() {
	sort.SliceStable(m.table.Rows, func(i, j int) bool {
		inI, okI := m.table.Rows[i].Data.(dizzclient.Intent)
		inJ, okJ := m.table.Rows[j].Data.(dizzclient.Intent)
		if !okI || !okJ {
			return false
		}
		if m.sortMode == 1 {
			pI := typePriority[string(inI.Type)]
			pJ := typePriority[string(inJ.Type)]
			if m.table.SortAsc {
				return pI < pJ
			}
			return pI > pJ
		}
		if m.table.SortAsc {
			return inI.Severity < inJ.Severity
		}
		return inI.Severity > inJ.Severity
	})
}

func (m *IntentsModel) buildTable() {
	filtered := m.intents

	if m.filter.Active() && m.filter.Query != "" {
		var f []dizzclient.Intent
		for _, in := range filtered {
			if m.filter.MatchesAny(in.ID, string(in.Type), in.Message, in.Scope) {
				f = append(f, in)
			}
		}
		filtered = f
	}

	sf := m.stateFilter.Value()
	if sf == "active" {
		var f []dizzclient.Intent
		for _, in := range filtered {
			if in.Status == "active" || in.Status == "" {
				f = append(f, in)
			}
		}
		filtered = f
	} else if sf == "resolved" {
		var f []dizzclient.Intent
		for _, in := range filtered {
			if in.Status == "resolved" {
				f = append(f, in)
			}
		}
		filtered = f
	}

	if sv := m.sevFilter.Value(); sv != "" {
		sevVal := int(sv[0] - '0')
		var f []dizzclient.Intent
		for _, in := range filtered {
			if in.Severity == sevVal {
				f = append(f, in)
			}
		}
		filtered = f
	}

	if tv := m.typeFilter.Value(); tv != "" {
		var f []dizzclient.Intent
		for _, in := range filtered {
			if string(in.Type) == tv {
				f = append(f, in)
			}
		}
		filtered = f
	}

	m.table.Rows = make([]ui.Row, len(filtered))
	for i, in := range filtered {
		typ := string(in.Type)
		typDisplay := " " + typ + " "
		if len(typDisplay) > 10 {
			typDisplay = typDisplay[:10]
		}

		sev := in.Severity
		if sev < 0 {
			sev = 0
		} else if sev > 3 {
			sev = 3
		}
		sevText := severityLabels[sev]
		sevStyle := render.SeverityColor(sev)

		msg := " " + in.Message
		if len(msg) > 78 {
			msg = msg[:78]
		}

		typeStyle := render.IntentTypeStyle(typ)

		m.table.Rows[i] = ui.Row{
			Cells:      []string{typDisplay, sevText, msg},
			CellStyles: []tcell.Style{typeStyle, sevStyle, tcell.StyleDefault},
			Data:       in,
		}
	}

	m.sortRows()
}

func (m *IntentsModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	if m.loading {
		c.SetContent((w-len("Loading intents..."))/2, h/2, render.StyleMuted, "Loading intents...")
		return
	}

	if m.err != "" {
		c.SetContent(2, 2, render.StyleError, m.err)
		return
	}

	header := m.buildHeader()
	c.SetContent(2, 0, render.StyleHighlight.Bold(true), header)

	countStr := fmt.Sprintf("%d intents", len(m.table.Rows))
	c.SetContent(w-2-len(countStr), 0, render.StyleMuted, countStr)

	if !m.showModal {
		filterHint := "g=sev p=type  a=act s=res  0-3=sev  t/f/r/q/h/m=type  x=clear  /=search  i=add"
		c.SetContent(2, 1, render.StyleDim, filterHint)
	}

	tableH := h
	if tableH < 2 {
		tableH = 2
	}

	m.table.Cols[0].Width = 10
	m.table.Cols[1].Width = 8
	gapTotal := m.table.ColGap * (len(m.table.Cols) - 1)
	m.table.Cols[2].Width = w - 4 - 10 - 8 - gapTotal
	if m.table.Cols[2].Width < 10 {
		m.table.Cols[2].Width = 10
	}

	ui.RenderTable(c, &m.table, 2, 2, w-4, tableH)

	if len(m.table.Rows) > 0 {
		relY := m.table.Selected - m.table.Offset
		if relY >= 0 {
			visibleCount := tableH - 3
			if visibleCount < 0 {
				visibleCount = 0
			}
			if relY < visibleCount {
				c.SetContent(1, 4+relY, render.StyleHighlight.Bold(true), "▸")
			}
		}
	}

	if m.showModal {
		m.renderModal(c)
	}

	if !m.showModal {
		m.filter.Render(c, 2, h-1)
	}
}

func (m *IntentsModel) renderModal(c *render.Canvas) {
	if m.modalType == "add" {
		m.renderAddModal(c)
	} else if m.modalType == "resolve" {
		m.renderResolveModal(c)
	}
}

func (m *IntentsModel) renderAddModal(c *render.Canvas) {
	bodyW := 52
	bodyH := 9
	cx, cy := ui.RenderModalBox(c, "Add Intent", bodyW, bodyH)

	help := "Tab/↑↓=switch  ←→=change  Enter=submit  Esc=cancel"
	c.SetContent(cx, cy, ui.StyleHelp, help)
	cy += 2

	typ := intentTypes[m.intentType]
	typeFocused := m.focusIdx == 0
	typeStyle := ui.StyleBtn
	if typeFocused {
		typeStyle = ui.StyleBtnFocus
	}
	c.SetContent(cx, cy, render.StyleDefault, "Type:     ")
	for i := range typ {
		c.SetCell(cx+10+i, cy, typeStyle, rune(typ[i]))
	}
	c.SetContent(cx+11+len(typ), cy, ui.StyleHelp, " ←→ to cycle")
	cy++

	msgFocused := m.focusIdx == 1
	msgStyle := ui.StyleBtn
	emptyMsg := m.inputMsg == ""
	if msgFocused && emptyMsg {
		msgStyle = ui.StyleBtnFocus
	} else if emptyMsg {
		msgStyle = ui.StyleBtn
	}
	c.SetContent(cx, cy, render.StyleDefault, "Message:  ")
	display := m.inputMsg
	if display == "" {
		display = "required"
	}
	if len(display) > bodyW-20 {
		display = display[:bodyW-20]
	}
	c.SetContent(cx+10, cy, msgStyle, display)
	if msgFocused {
		cxPos := cx + 10 + m.msgCursor
		if display == "required" {
			msgStyle = render.StyleDim
		}
		if cxPos < c.Width()-1 {
			c.SetCell(cxPos, cy, render.StyleHighlight, '_')
		}
	}
	cy++

	sevFocused := m.focusIdx == 2
	c.SetContent(cx, cy, render.StyleDefault, "Severity: ")
	labels := []string{"none", "low", "medium", "high"}
	sevColor := render.StyleMuted
	label := labels[m.inputSev]
	sevColor = render.SeverityColor(m.inputSev)
	for i := 0; i <= 3; i++ {
		s := ui.StyleBtn
		if sevFocused && i == m.inputSev {
			s = ui.StyleBtnFocus
		}
		mark := fmt.Sprintf(" %d ", i)
		c.SetContent(cx+10+i*6, cy, s, mark)
	}
	c.SetContent(cx+10+24, cy, sevColor, label)
	c.SetContent(cx+10+24+len(label)+2, cy, ui.StyleHelp, "←→ to change")
	if m.validationErr != "" {
		c.SetContent(cx, cy+1, render.StyleError, "⚠ "+m.validationErr)
	}
	cy += 2

	btnY := cy
	btnW := 10
	spacing := 2
	totalBtnW := btnW*2 + spacing
	btnStartX := cx + bodyW - totalBtnW - 2

	ui.RenderButton(c, btnStartX, btnY, btnW, "Cancel", m.focusIdx == 3, false)
	ui.RenderButton(c, btnStartX+btnW+spacing, btnY, btnW, "Add", m.focusIdx == 4, false)
}

func (m *IntentsModel) renderResolveModal(c *render.Canvas) {
	msg := ""
	id := ""
	if m.resolveIdx < len(m.intents) {
		in := m.intents[m.resolveIdx]
		id = in.ID
		msg = in.Message
	}

	bodyW := 52
	if len(msg)+10 > bodyW {
		bodyW = len(msg) + 10
		if bodyW > 60 {
			bodyW = 60
		}
	}
	bodyH := 7
	cx, cy := ui.RenderModalBox(c, "Resolve Intent", bodyW, bodyH)

	prompt := fmt.Sprintf("Resolve %s?", id)
	c.SetContent(cx+(bodyW-len(prompt))/2, cy, render.StyleWarning, prompt)
	cy++

	if len(msg) > bodyW-4 {
		msg = msg[:bodyW-4]
	}
	c.SetContent(cx+(bodyW-len(msg))/2, cy, render.StyleInfo, msg)
	cy += 2

	btnW := 10
	spacing := 2
	totalBtnW := btnW*2 + spacing
	btnStartX := cx + bodyW - totalBtnW - 2

	ui.RenderButton(c, btnStartX, cy, btnW, "Cancel", m.focusIdx == 0, false)
	ui.RenderButton(c, btnStartX+btnW+spacing, cy, btnW, "OK", m.focusIdx == 1, false)
}

// @dizz-ignore-unused
func (m *IntentsModel) View() string { return "" }

func (m *IntentsModel) InputMode() bool { return m.showModal || m.filter.Active() }

func (m *IntentsModel) IsModalActive() bool { return m.showModal }

func (m *IntentsModel) RenderModal(c *render.Canvas) {
	m.renderModal(c)
}
