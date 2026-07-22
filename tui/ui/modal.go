package ui

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/gdamore/tcell/v2"
)

var (
	StyleOverlay   = tcell.StyleDefault.Background(tcell.NewRGBColor(20, 20, 20)).Dim(true)
	StyleModalBg   = tcell.StyleDefault.Background(tcell.NewRGBColor(35, 35, 35))
	StyleModalBdr  = tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 120, 180))
	StyleBtn       = tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(50, 50, 50))
	StyleBtnFocus  = tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 0, 0)).Background(tcell.NewRGBColor(100, 200, 255)).Bold(true)
	StyleBtnDanger = tcell.StyleDefault.Foreground(tcell.NewRGBColor(255, 200, 200)).Background(tcell.NewRGBColor(80, 40, 40))
	StyleHelp      = tcell.StyleDefault.Foreground(tcell.NewRGBColor(140, 140, 160))
)

func RenderModalBox(c *render.Canvas, title string, bodyW, bodyH int) (cx, cy int) {
	w, h := c.Width(), c.Height()

	boxW := bodyW + 4
	boxH := bodyH + 4
	if boxW > w {
		boxW = w
	}
	if boxH > h {
		boxH = h
	}

	boxX := (w - boxW) / 2
	boxY := (h - boxH) / 2

	for y := 2; y < h-2; y++ {
		for x := 0; x < w; x++ {
			c.SetCell(x, y, StyleOverlay, ' ')
		}
	}

	c.Rect(boxX, boxY, boxW, boxH, StyleModalBg, ' ')

	border := StyleModalBdr
	for x := boxX + 1; x < boxX+boxW-1; x++ {
		c.SetCell(x, boxY, border, '─')
		c.SetCell(x, boxY+boxH-1, border, '─')
	}
	for y := boxY + 1; y < boxY+boxH-1; y++ {
		c.SetCell(boxX, y, border, '│')
		c.SetCell(boxX+boxW-1, y, border, '│')
	}
	c.SetCell(boxX, boxY, border, '┌')
	c.SetCell(boxX+boxW-1, boxY, border, '┐')
	c.SetCell(boxX, boxY+boxH-1, border, '└')
	c.SetCell(boxX+boxW-1, boxY+boxH-1, border, '┘')

	if title != "" {
		titleStyle := render.StyleHighlight.Bold(true)
		c.SetContent(boxX+2, boxY, titleStyle, " "+title+" ")
	}

	return boxX + 2, boxY + 2
}

func RenderButton(c *render.Canvas, x, y, w int, label string, focused bool, danger bool) {
	style := StyleBtn
	if focused {
		style = StyleBtnFocus
	} else if danger {
		style = StyleBtnDanger
	}

	bg := tcell.NewRGBColor(35, 35, 35)
	if focused {
		bg = tcell.NewRGBColor(100, 200, 255)
	}

	padding := w - len(label) - 2
	if padding < 1 {
		padding = 1
	}
	leftPad := padding / 2

	for i := 0; i < w; i++ {
		c.SetCell(x+i, y, style.Background(bg), ' ')
	}
	c.SetCell(x, y, style.Background(bg), '[')
	c.SetCell(x+w-1, y, style.Background(bg), ']')
	c.SetContent(x+1+leftPad, y, style.Background(bg).Bold(focused), label)
}
