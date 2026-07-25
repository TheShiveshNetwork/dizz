package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
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

	inputNote  string
	noteCursor int

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
			"o": "open",
			"s": "resolved",
			"c": "closed",
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

type intentActionDoneMsg struct {
	err error
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

	case intentActionDoneMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
		}
		return m, m.refresh()

	case ui.RefreshTick:
		return m, m.refresh()
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
			m.inputNote = ""
			m.noteCursor = 0
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
		totalFields = 4
	}

	if m.modalType == "resolve" && m.focusIdx == 0 {
		return m.handleNoteKey(msg)
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

const maxNoteLines = 5

func (m *IntentsModel) handleNoteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxNoteW := 38
	if maxNoteW < 10 {
		maxNoteW = 10
	}

	switch msg.String() {
	case "esc":
		m.showModal = false
		m.validationErr = ""
		m.inputNote = ""
		m.noteCursor = 0
		return m, nil

	case "tab":
		m.focusIdx = 1
		return m, nil

	case "shift+tab":
		m.focusIdx = 3
		return m, nil

	case "enter":
		noteLines := strings.Split(m.inputNote, "\n")
		totalVisual := 0
		for _, l := range noteLines {
			totalVisual += visualLines(l, maxNoteW)
		}
		if totalVisual < maxNoteLines {
			before := m.inputNote[:m.noteCursor]
			after := m.inputNote[m.noteCursor:]
			m.inputNote = before + "\n" + after
			m.noteCursor++
		}
		return m, nil

	case "left":
		if m.noteCursor > 0 {
			m.noteCursor--
		}
		return m, nil

	case "right":
		if m.noteCursor < len(m.inputNote) {
			m.noteCursor++
		}
		return m, nil

	case "up":
		lines := strings.Split(m.inputNote[:m.noteCursor], "\n")
		if len(lines) <= 1 {
			m.noteCursor = 0
		} else {
			curLine := lines[len(lines)-1]
			prevLine := lines[len(lines)-2]
			col := runewidth.StringWidth(curLine)
			newPos := len(m.inputNote[:m.noteCursor]) - len(curLine) - 1
			if col > runewidth.StringWidth(prevLine) {
				newPos -= runewidth.StringWidth(prevLine)
				newPos += runewidth.StringWidth(prevLine)
			} else {
				newPos -= col
				newPos += col
			}
			offset := 0
			for i := 0; i < len(lines)-2; i++ {
				offset += len(lines[i]) + 1
			}
			targetW := runewidth.StringWidth(curLine)
			pos := offset
			w := 0
			for pos < offset+len(prevLine) && w < targetW {
				w += runewidth.RuneWidth(rune(m.inputNote[pos]))
				if w <= targetW {
					pos++
				}
			}
			m.noteCursor = pos
		}
		return m, nil

	case "down":
		lines := strings.Split(m.inputNote, "\n")
		curLineStart := 0
		for i := 0; i < len(lines)-1; i++ {
			curLineStart += len(lines[i]) + 1
			if curLineStart > m.noteCursor {
				curLineStart -= len(lines[i]) + 1
				break
			}
		}
		curLineEnd := m.noteCursor
		if curLineEnd > len(m.inputNote) {
			curLineEnd = len(m.inputNote)
		}
		curLineText := m.inputNote[curLineStart:curLineEnd]
		curColW := runewidth.StringWidth(curLineText)

		nextLineStart := curLineEnd
		if nextLineStart < len(m.inputNote) && m.inputNote[nextLineStart] == '\n' {
			nextLineStart++
		}
		if nextLineStart >= len(m.inputNote) {
			m.noteCursor = len(m.inputNote)
			return m, nil
		}
		nextLineEnd := nextLineStart
		for nextLineEnd < len(m.inputNote) && m.inputNote[nextLineEnd] != '\n' {
			nextLineEnd++
		}
		pos := nextLineStart
		w := 0
		for pos < nextLineEnd && w < curColW {
			w += runewidth.RuneWidth(rune(m.inputNote[pos]))
			if w <= curColW {
				pos++
			}
		}
		m.noteCursor = pos
		return m, nil

	case "home":
		lines := strings.Split(m.inputNote[:m.noteCursor], "\n")
		offset := 0
		for i := 0; i < len(lines)-1; i++ {
			offset += len(lines[i]) + 1
		}
		m.noteCursor = offset
		return m, nil

	case "end":
		lines := strings.Split(m.inputNote, "\n")
		idx := len(lines) - 1
		for i := 0; i <= idx; i++ {
			if i == idx {
				offset := 0
				for j := 0; j < i; j++ {
					offset += len(lines[j]) + 1
				}
				m.noteCursor = offset + len(lines[i])
				break
			}
		}
		return m, nil

	case "backspace":
		if m.noteCursor > 0 {
			m.inputNote = m.inputNote[:m.noteCursor-1] + m.inputNote[m.noteCursor:]
			m.noteCursor--
		}
		return m, nil

	default:
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			before := m.inputNote[:m.noteCursor]
			after := m.inputNote[m.noteCursor:]
			m.inputNote = before + string(ch) + after
			m.noteCursor++
		}
		return m, nil
	}
}

func visualLines(text string, maxW int) int {
	if maxW < 1 {
		maxW = 1
	}
	if text == "" {
		return 1
	}
	w := runewidth.StringWidth(text)
	lines := w / maxW
	if w%maxW != 0 {
		lines++
	}
	if lines < 1 {
		lines = 1
	}
	return lines
}

func (m *IntentsModel) handleModalEnter() (tea.Model, tea.Cmd) {
	if m.modalType == "resolve" {
		if m.focusIdx == 2 {
			idx := m.resolveIdx
			note := m.inputNote
			m.showModal = false
			m.validationErr = ""
			m.inputNote = ""
			m.noteCursor = 0
			if idx < len(m.intents) {
				id := m.intents[idx].ID
				return m, func() tea.Msg {
					return intentActionDoneMsg{err: dizzclient.IntentResolve(id, note)}
				}
			}
			return m, m.refresh()
		}
		if m.focusIdx == 3 {
			idx := m.resolveIdx
			note := m.inputNote
			m.showModal = false
			m.validationErr = ""
			m.inputNote = ""
			m.noteCursor = 0
			if idx < len(m.intents) {
				id := m.intents[idx].ID
				return m, func() tea.Msg {
					return intentActionDoneMsg{err: dizzclient.IntentClose(id, note)}
				}
			}
			return m, m.refresh()
		}
		m.showModal = false
		m.validationErr = ""
		m.inputNote = ""
		m.noteCursor = 0
		return m, nil
	}

	if m.modalType == "add" {
		if m.focusIdx == 4 {
			if m.inputMsg == "" {
				m.validationErr = "Message is required"
				return m, nil
			}
			msg := m.inputMsg
			typ := intentTypes[m.intentType]
			sev := m.inputSev
			m.showModal = false
			m.validationErr = ""
			return m, func() tea.Msg {
				return intentActionDoneMsg{err: dizzclient.IntentAdd(msg, typ, sev, nil)}
			}
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
	if sf == "open" {
		parts += " [Open]"
	} else if sf == "resolved" {
		parts += " [Resolved]"
	} else if sf == "closed" {
		parts += " [Closed]"
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
	if sf == "open" {
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
	} else if sf == "closed" {
		var f []dizzclient.Intent
		for _, in := range filtered {
			if in.Status == "closed" {
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
		maxMsg := m.table.Cols[2].Width - 2
		if maxMsg < 10 {
			maxMsg = 10
		}
		if len(msg) > maxMsg {
			if maxMsg > 3 {
				msg = msg[:maxMsg-3] + "..."
			} else {
				msg = msg[:maxMsg]
			}
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
		filterHint := "g=sev p=type  o=open s=res c=clo  0-3=sev  t/f/r/q/h/m=type  x=clear  /=search  i=add"
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
	maxMsgW := bodyW - 20
	if maxMsgW < 10 {
		maxMsgW = 10
	}
	if len(display) > maxMsgW {
		display = display[:maxMsgW]
	}
	c.SetContentBounded(cx+10, cy, maxMsgW, msgStyle, display)
	if msgFocused {
		cxPos := cx + 10 + m.msgCursor
		if cxPos >= cx+10+maxMsgW {
			cxPos = cx + 10 + maxMsgW - 1
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

	ui.RenderButton(c, btnStartX, btnY, btnW, "Cancel", m.focusIdx == 3, false, false)
	ui.RenderButton(c, btnStartX+btnW+spacing, btnY, btnW, "Add", m.focusIdx == 4, false, false)
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

	maxMsgW := bodyW - 4
	msgLines := wordWrap(msg, maxMsgW)
	if len(msgLines) == 0 {
		msgLines = []string{""}
	}

	maxNoteW := bodyW - 14
	if maxNoteW < 10 {
		maxNoteW = 10
	}
	noteDisplayLines := noteWrapLines(m.inputNote, maxNoteW)
	noteVisualH := len(noteDisplayLines)
	if noteVisualH < 1 {
		noteVisualH = 1
	}

	bodyH := 10 + len(msgLines) - 1 + noteVisualH - 1
	cx, cy := ui.RenderModalBox(c, "Resolve Intent", bodyW, bodyH)

	help := "Tab=switch  Enter=confirm  Esc=cancel"
	c.SetContent(cx, cy, ui.StyleHelp, help)
	cy += 2

	prompt := fmt.Sprintf("Resolve %s?", id)
	c.SetContent(cx+(bodyW-len(prompt))/2, cy, render.StyleWarning, prompt)
	cy++

	for _, line := range msgLines {
		c.SetContent(cx+(bodyW-runewidth.StringWidth(line))/2, cy, render.StyleInfo, line)
		cy++
	}
	cy++

	noteFocused := m.focusIdx == 0
	noteStyle := ui.StyleBtn
	if noteFocused {
		noteStyle = ui.StyleBtnFocus
	}
	c.SetContent(cx, cy, render.StyleDefault, "Note:     ")

	cursorLine, cursorCol := noteCursorPosition(m.inputNote, m.noteCursor, maxNoteW)

	for i, line := range noteDisplayLines {
		lineY := cy + i
		c.SetContentBounded(cx+10, lineY, maxNoteW, noteStyle, line)
		if noteFocused && i == cursorLine {
			cxPos := cx + 10 + cursorCol
			if cxPos >= cx+10+maxNoteW {
				cxPos = cx + 10 + maxNoteW - 1
			}
			if cxPos < c.Width()-1 {
				c.SetCell(cxPos, lineY, render.StyleHighlight, '_')
			}
		}
	}
	cy += noteVisualH + 1

	btnW := 10
	spacing := 2
	totalBtnW := btnW*3 + spacing*2
	btnStartX := cx + bodyW - totalBtnW - 2

	ui.RenderButton(c, btnStartX, cy, btnW, "Cancel", m.focusIdx == 1, false, false)
	ui.RenderButton(c, btnStartX+btnW+spacing, cy, btnW, "Resolve", m.focusIdx == 2, false, true)
	ui.RenderButton(c, btnStartX+btnW*2+spacing*2, cy, btnW, "Close", m.focusIdx == 3, true, false)
}

// @dizz-ignore-unused
func (m *IntentsModel) View() string { return "" }

func (m *IntentsModel) InputMode() bool { return m.showModal || m.filter.Active() }

func (m *IntentsModel) IsModalActive() bool { return m.showModal }

func (m *IntentsModel) RenderModal(c *render.Canvas) {
	m.renderModal(c)
}

func wordWrap(text string, maxW int) []string {
	if maxW < 1 {
		maxW = 1
	}
	if text == "" {
		return []string{""}
	}
	var lines []string
	remaining := text
	for len(remaining) > 0 {
		if runewidth.StringWidth(remaining) <= maxW {
			lines = append(lines, remaining)
			break
		}
		idx := strings.LastIndex(remaining[:maxW], " ")
		if idx <= 0 {
			idx = maxW
			for idx < len(remaining) && runewidth.StringWidth(remaining[:idx+1]) <= maxW {
				idx++
			}
		}
		lines = append(lines, remaining[:idx])
		remaining = strings.TrimPrefix(remaining[idx:], " ")
	}
	return lines
}

func noteWrapLines(text string, maxW int) []string {
	if maxW < 1 {
		maxW = 1
	}
	if text == "" {
		return []string{""}
	}
	var allLines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			allLines = append(allLines, "")
			continue
		}
		allLines = append(allLines, wordWrap(paragraph, maxW)...)
	}
	return allLines
}

func noteCursorPosition(text string, cursor int, maxW int) (line int, col int) {
	if cursor > len(text) {
		cursor = len(text)
	}
	_before := text[:cursor]
	lines := strings.Split(_before, "\n")
	line = len(lines) - 1
	lastLine := lines[line]
	col = runewidth.StringWidth(lastLine)
	if col >= maxW {
		col = maxW - 1
	}
	return line, col
}
