package app

import (
	"fmt"

	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func (m *Model) renderIntentsMini(c *render.Canvas, x, y, w, h int, focused bool) {
	active := m.activeIntents()

	header := fmt.Sprintf("Intents [%d]", len(active))
	headerStyle := render.StyleHighlight.Bold(true)
	if focused {
		headerStyle = render.StyleHighlight.Bold(true).Foreground(tcell.NewRGBColor(100, 200, 255))
	}
	c.SetContent(x+1, y, headerStyle, header)

	if h < 2 {
		return
	}
	listY := y + 1
	visible := h - 2
	if visible < 1 {
		visible = 1
	}

	for i := 0; i < visible && m.intentsOff+i < len(active); i++ {
		in := active[m.intentsOff+i]
		rowY := listY + i
		if rowY > y+h-1 {
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
			for runewidth.StringWidth(msg)+1 > availW {
				msg = msg[:len(msg)-1]
			}
			msg += "..."
		}
		msgStyle := render.StyleDefault
		if focused && m.intentsOff+i == m.intentsSel {
			msgStyle = render.StyleHighlight
		}
		c.SetContent(msgX, rowY, msgStyle, msg)
	}

	if m.intentsOff > 0 {
		c.SetCell(x, listY, render.StyleDim, '\u25B2')
	}
	if m.intentsOff+visible < len(active) {
		indY := listY + visible - 1
		if indY >= listY && indY < y+h {
			c.SetCell(x, indY, render.StyleDim, '\u25BC')
		}
	}

	hintY := y + h - 1
	if focused && h >= 6 {
		hint := "\u2191\u2193=nav enter=resolve i=add"
		c.SetContent(x+1, hintY, render.StyleDim, hint)
	}
}
