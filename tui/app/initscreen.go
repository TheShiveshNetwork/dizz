package app

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
)

func (m *Model) renderInitScreen(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	_, artH := render.DizzArtDimensions()
	artStartY := 2
	if artH+6 < h {
		artStartY = (h-artH)/2 - 4
	}
	if artStartY < 1 {
		artStartY = 1
	}
	render.RenderDizzie(c, m.initFrameTick, w/2, artStartY)

	popupW := 48
	if popupW > w-4 {
		popupW = w - 4
	}
	popupH := 8
	popupY := artStartY + artH + 1
	if popupY+popupH > h-1 {
		popupY = h - popupH - 1
	}
	if popupY < 1 {
		popupY = 1
	}

	cx, cy := ui.RenderModalBox(c, "Welcome to dizz", popupW, popupH)

	help := "Tab=switch  Enter=confirm  q=quit"
	c.SetContent(cx, cy, ui.StyleHelp, help)
	cy += 2

	msg := "No .dizz directory found."
	c.SetContent(cx+(popupW-4-len(msg))/2, cy, render.StyleMuted, msg)
	cy++

	sub := "Initialize this project to get started."
	c.SetContent(cx+(popupW-4-len(sub))/2, cy, render.StyleMuted, sub)
	cy += 2

	btnW := 14
	btnX := cx + (popupW-4-btnW)/2
	if m.initBusy {
		c.SetContent(cx+(popupW-4-14)/2, cy, render.StyleMuted, "Initializing...")
	} else {
		ui.RenderButton(c, btnX, cy, btnW, "Initialize", true, false, true)
	}
}
