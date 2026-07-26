package app

import (
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/tui/client"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

func (m *Model) renderIntentsModal(c *render.Canvas) {
	if m.modalType == "add" {
		m.renderAddModal(c)
	} else if m.modalType == "resolve" {
		m.renderResolveModal(c)
	}
}

func (m *Model) renderResolveModal(c *render.Canvas) {
	msg := ""
	id := ""
	active := m.activeIntents()
	if m.resolveIdx < len(active) {
		in := active[m.resolveIdx]
		id = in.ID
		msg = in.Message
	}

	bodyW := 52
	if len(msg)+10 > bodyW {
		bodyW = len(msg) + 10
		if bodyW > 70 {
			bodyW = 70
		}
	}

	maxMsgW := bodyW - 4
	msgLines := wrapText(msg, maxMsgW)
	if len(msgLines) == 0 {
		msgLines = []string{""}
	}

	maxNoteW := bodyW - 14
	if maxNoteW < 10 {
		maxNoteW = 10
	}
	noteDisplayLines := noteWrapLines(m.addNote, maxNoteW)
	noteVisualH := len(noteDisplayLines)
	if noteVisualH < 1 {
		noteVisualH = 1
	}

	bodyH := 11 + len(msgLines) - 1 + noteVisualH - 1
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

	cursorLine, cursorCol := noteCursorPosition(m.addNote, m.noteCursor, maxNoteW)

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

	if cy < c.Height()-2 {
		btnW := 10
		spacing := 2
		totalBtnW := btnW*3 + spacing*2
		btnStartX := cx + bodyW - totalBtnW - 2
		ui.RenderButton(c, btnStartX, cy, btnW, "Cancel", m.focusIdx == 1, false, false)
		ui.RenderButton(c, btnStartX+btnW+spacing, cy, btnW, "Resolve", m.focusIdx == 2, false, true)
		ui.RenderButton(c, btnStartX+btnW*2+spacing*2, cy, btnW, "Close", m.focusIdx == 3, true, false)
	}
}

func (m *Model) renderAddModal(c *render.Canvas) {
	bodyW := 52
	bodyH := 9
	cx, cy := ui.RenderModalBox(c, "Add Intent", bodyW, bodyH)

	help := "Tab/↑↓=switch  ←→=change  Enter=submit  Esc=cancel"
	c.SetContent(cx, cy, ui.StyleHelp, help)
	cy += 2

	intentTypes := []string{"todo", "fixme", "refactor", "question", "hack", "temporary"}

	typ := intentTypes[m.addType]
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
	emptyMsg := m.addMsg == ""
	if msgFocused && emptyMsg {
		msgStyle = ui.StyleBtnFocus
	} else if emptyMsg {
		msgStyle = ui.StyleBtn
	}
	c.SetContent(cx, cy, render.StyleDefault, "Message:  ")
	display := m.addMsg
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
		cxPos := cx + 10 + m.addMsgCursor
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
	for i := 0; i <= 3; i++ {
		s := ui.StyleBtn
		if sevFocused && i == m.addSev {
			s = ui.StyleBtnFocus
		}
		mark := fmt.Sprintf(" %d ", i)
		c.SetContent(cx+10+i*6, cy, s, mark)
	}
	sevLabels := []string{"none", "low", "medium", "high"}
	sevColor := render.SeverityColor(m.addSev)
	c.SetContent(cx+10+24, cy, sevColor, sevLabels[m.addSev])
	c.SetContent(cx+10+24+len(sevLabels[m.addSev])+2, cy, ui.StyleHelp, "←→ to change")
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

func (m *Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	intentTypes := []string{"todo", "fixme", "refactor", "question", "hack", "temporary"}
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
		m.addNote = ""
		m.noteCursor = 0
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
				m.addType = (m.addType - 1 + len(intentTypes)) % len(intentTypes)
			case 1:
				if m.addMsgCursor > 0 {
					m.addMsgCursor--
				}
			case 2:
				if m.addSev > 0 {
					m.addSev--
				}
			}
		}
		return m, nil

	case "right":
		if m.modalType == "add" {
			switch m.focusIdx {
			case 0:
				m.addType = (m.addType + 1) % len(intentTypes)
			case 1:
				if m.addMsgCursor < len(m.addMsg) {
					m.addMsgCursor++
				}
			case 2:
				if m.addSev < 3 {
					m.addSev++
				}
			}
		}
		return m, nil

	case "backspace":
		if m.modalType == "add" && m.focusIdx == 1 && m.addMsgCursor > 0 {
			m.addMsg = m.addMsg[:m.addMsgCursor-1] + m.addMsg[m.addMsgCursor:]
			m.addMsgCursor--
		}
		return m, nil

	case "home":
		if m.modalType == "add" && m.focusIdx == 1 {
			m.addMsgCursor = 0
		}
		return m, nil

	case "end":
		if m.modalType == "add" && m.focusIdx == 1 {
			m.addMsgCursor = len(m.addMsg)
		}
		return m, nil

	default:
		if m.modalType == "add" && m.focusIdx == 1 && len(msg.String()) == 1 {
			ch := msg.String()[0]
			before := m.addMsg[:m.addMsgCursor]
			after := m.addMsg[m.addMsgCursor:]
			m.addMsg = before + string(ch) + after
			m.addMsgCursor++
		}
		return m, nil
	}
}

const maxNoteLines = 5

func (m *Model) handleNoteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxNoteW := 38
	if maxNoteW < 10 {
		maxNoteW = 10
	}

	switch msg.String() {
	case "esc":
		m.showModal = false
		m.validationErr = ""
		m.addNote = ""
		m.noteCursor = 0
		return m, nil

	case "tab":
		m.focusIdx = 1
		return m, nil

	case "shift+tab":
		m.focusIdx = 3
		return m, nil

	case "enter":
		noteLines := strings.Split(m.addNote, "\n")
		totalVisual := 0
		for _, l := range noteLines {
			totalVisual += noteVisualLines(l, maxNoteW)
		}
		if totalVisual < maxNoteLines {
			before := m.addNote[:m.noteCursor]
			after := m.addNote[m.noteCursor:]
			m.addNote = before + "\n" + after
			m.noteCursor++
		}
		return m, nil

	case "left":
		if m.noteCursor > 0 {
			m.noteCursor--
		}
		return m, nil

	case "right":
		if m.noteCursor < len(m.addNote) {
			m.noteCursor++
		}
		return m, nil

	case "up":
		lines := strings.Split(m.addNote[:m.noteCursor], "\n")
		if len(lines) <= 1 {
			m.noteCursor = 0
		} else {
			curLine := lines[len(lines)-1]
			prevLine := lines[len(lines)-2]
			targetW := runewidth.StringWidth(curLine)
			offset := 0
			for i := 0; i < len(lines)-2; i++ {
				offset += len(lines[i]) + 1
			}
			pos := offset
			w := 0
			for pos < offset+len(prevLine) && w < targetW {
				w += runewidth.RuneWidth(rune(m.addNote[pos]))
				if w <= targetW {
					pos++
				}
			}
			m.noteCursor = pos
		}
		return m, nil

	case "down":
		lines := strings.Split(m.addNote, "\n")
		curLineStart := 0
		for i := 0; i < len(lines)-1; i++ {
			curLineStart += len(lines[i]) + 1
			if curLineStart > m.noteCursor {
				curLineStart -= len(lines[i]) + 1
				break
			}
		}
		curLineEnd := m.noteCursor
		if curLineEnd > len(m.addNote) {
			curLineEnd = len(m.addNote)
		}
		curLineText := m.addNote[curLineStart:curLineEnd]
		curColW := runewidth.StringWidth(curLineText)

		nextLineStart := curLineEnd
		if nextLineStart < len(m.addNote) && m.addNote[nextLineStart] == '\n' {
			nextLineStart++
		}
		if nextLineStart >= len(m.addNote) {
			m.noteCursor = len(m.addNote)
			return m, nil
		}
		nextLineEnd := nextLineStart
		for nextLineEnd < len(m.addNote) && m.addNote[nextLineEnd] != '\n' {
			nextLineEnd++
		}
		pos := nextLineStart
		w := 0
		for pos < nextLineEnd && w < curColW {
			w += runewidth.RuneWidth(rune(m.addNote[pos]))
			if w <= curColW {
				pos++
			}
		}
		m.noteCursor = pos
		return m, nil

	case "home":
		lines := strings.Split(m.addNote[:m.noteCursor], "\n")
		offset := 0
		for i := 0; i < len(lines)-1; i++ {
			offset += len(lines[i]) + 1
		}
		m.noteCursor = offset
		return m, nil

	case "end":
		lines := strings.Split(m.addNote, "\n")
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
			m.addNote = m.addNote[:m.noteCursor-1] + m.addNote[m.noteCursor:]
			m.noteCursor--
		}
		return m, nil

	default:
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			before := m.addNote[:m.noteCursor]
			after := m.addNote[m.noteCursor:]
			m.addNote = before + string(ch) + after
			m.noteCursor++
		}
		return m, nil
	}
}

func (m *Model) handleModalEnter() (tea.Model, tea.Cmd) {
	if m.modalType == "resolve" {
		if m.focusIdx == 2 {
			active := m.activeIntents()
			note := m.addNote
			m.showModal = false
			m.validationErr = ""
			m.addNote = ""
			m.noteCursor = 0
			if m.resolveIdx < len(active) {
				id := active[m.resolveIdx].ID
				return m, func() tea.Msg {
					return intentActionDoneMsg{err: client.IntentResolve(id, note)}
				}
			}
			return m, m.refreshBatch()
		}
		if m.focusIdx == 3 {
			active := m.activeIntents()
			note := m.addNote
			m.showModal = false
			m.validationErr = ""
			m.addNote = ""
			m.noteCursor = 0
			if m.resolveIdx < len(active) {
				id := active[m.resolveIdx].ID
				return m, func() tea.Msg {
					return intentActionDoneMsg{err: client.IntentClose(id, note)}
				}
			}
			return m, m.refreshBatch()
		}
		m.showModal = false
		m.validationErr = ""
		m.addNote = ""
		m.noteCursor = 0
		return m, nil
	}

	if m.modalType == "add" {
		if m.focusIdx == 4 {
			if m.addMsg == "" {
				m.validationErr = "Message is required"
				return m, nil
			}
			intentTypes := []string{"todo", "fixme", "refactor", "question", "hack", "temporary"}
			msg := m.addMsg
			typ := intentTypes[m.addType]
			sev := m.addSev
			m.showModal = false
			m.validationErr = ""
			return m, func() tea.Msg {
				return intentActionDoneMsg{err: client.IntentAdd(msg, typ, sev, nil)}
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

func noteVisualLines(text string, maxW int) int {
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
		allLines = append(allLines, wrapText(paragraph, maxW)...)
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

func wrapText(text string, maxW int) []string {
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
