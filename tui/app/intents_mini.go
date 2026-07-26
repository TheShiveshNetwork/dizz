package app

import (
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func (m *Model) renderIntentsMini(c *render.Canvas, x, y, w, h int, intentsFocused, todosFocused bool) {
	active := m.activeIntents()

	intentsH := m.intentsSpace(len(active), h)
	todosH := h - intentsH

	renderIntents := func(topY, avail int) {
		header := fmt.Sprintf("Active Intents [%d]", len(active))
		headerStyle := render.StyleHighlight.Bold(true)
		if intentsFocused {
			headerStyle = render.StyleHighlight.Bold(true).Foreground(tcell.NewRGBColor(100, 200, 255))
		}
		c.SetContent(x+1, topY, headerStyle, header)

		if avail < 2 {
			return
		}
		listY := topY + 1
		visible := avail - 2
		if visible < 1 {
			visible = 1
		}

		for i := 0; i < visible && m.intentsOff+i < len(active); i++ {
			in := active[m.intentsOff+i]
			rowY := listY + i
			if rowY > topY+avail-1 {
				break
			}

			typStyle := render.IntentTypeStyle(string(in.Type))
			label := " " + string(in.Type) + " "
			c.SetContent(x+1, rowY, typStyle, label)

			msgX := x + 1 + len(label)
			availW := x + w - msgX - 1
			if availW < 3 {
				availW = 3
			}
			msg := in.Message
			if runewidth.StringWidth(msg) > availW {
				for runewidth.StringWidth(msg)+3 > availW && len(msg) > 0 {
					msg = msg[:len(msg)-1]
				}
				msg += "..."
			}
			msgStyle := render.StyleDefault
			if intentsFocused && m.intentsOff+i == m.intentsSel {
				msgStyle = render.StyleHighlight
			}
			c.SetContent(msgX, rowY, msgStyle, msg)
		}

		if m.intentsOff > 0 {
			c.SetCell(x, listY, render.StyleDim, '\u25B2')
		}
		if m.intentsOff+visible < len(active) {
			indY := listY + visible - 1
			if indY >= listY && indY < topY+avail {
				c.SetCell(x, indY, render.StyleDim, '\u25BC')
			}
		}
	}

	renderTodos := func(topY, avail int) {
		if avail < 1 {
			return
		}

		unresolved := m.unresolvedTodos()
		header := fmt.Sprintf("TODOs [%d]", len(unresolved))
		headerStyle := render.StyleHighlight.Bold(true)
		if todosFocused {
			headerStyle = render.StyleHighlight.Bold(true).Foreground(tcell.NewRGBColor(100, 200, 255))
		}
		c.SetContent(x+1, topY, headerStyle, header)

		if avail < 2 || len(unresolved) == 0 {
			return
		}
		listY := topY + 1
		visible := avail - 2
		if visible < 1 {
			visible = 1
		}

		for i := 0; i < visible && m.todosOff+i < len(unresolved); i++ {
			td := unresolved[m.todosOff+i]
			rowY := listY + i
			if rowY > topY+avail-1 {
				break
			}

			typStyle := render.TodoTypeStyle(td.Type)
			label := " " + strings.ToUpper(td.Type) + " "
			c.SetContent(x+1, rowY, typStyle, label)

			msgX := x + 1 + len(label)
			availW := x + w - msgX - 1
			if availW < 3 {
				availW = 3
			}
			msg := td.Text
			if runewidth.StringWidth(msg) > availW {
				for runewidth.StringWidth(msg)+3 > availW && len(msg) > 0 {
					msg = msg[:len(msg)-1]
				}
				msg += "..."
			}
			msgStyle := render.StyleDefault
			if todosFocused && m.todosOff+i == m.todosSel {
				msgStyle = render.StyleHighlight
			}
			c.SetContent(msgX, rowY, msgStyle, msg)
		}

		if m.todosOff > 0 {
			c.SetCell(x, listY, render.StyleDim, '\u25B2')
		}
		if m.todosOff+visible < len(unresolved) {
			indY := listY + visible - 1
			if indY >= listY && indY < topY+avail {
				c.SetCell(x, indY, render.StyleDim, '\u25BC')
			}
		}
	}

	if intentsH > 0 {
		renderIntents(y, intentsH)
	}

	if todosH > 0 && intentsH > 0 {
		sepY := y + intentsH
		if sepY < y+h {
			for sx := x; sx < x+w; sx++ {
				c.SetCell(sx, sepY, render.StyleMuted, '\u2500')
			}
		}
	}

	if todosH > 0 {
		todosTop := y + intentsH
		if intentsH > 0 {
			todosTop++
			todosH--
		}
		if todosH > 0 {
			renderTodos(todosTop, todosH)
		}
	}

	hintY := y + h - 1
	if (intentsFocused || todosFocused) && h >= 6 {
		hint := "\u2191\u2193=nav enter=resolve i=add"
		c.SetContent(x+1, hintY, render.StyleDim, hint)
	}
}

func (m *Model) intentsSpace(activeCount, totalH int) int {
	if activeCount == 0 && len(m.unresolvedTodos()) > 0 {
		return 0
	}
	half := totalH / 2
	if half < 3 {
		half = 3
	}
	if half > totalH-2 {
		half = totalH - 2
	}
	return half
}
