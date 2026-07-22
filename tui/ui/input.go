package ui

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
)

type TextInput struct {
	Text        string
	Cursor      int
	Placeholder string
	Active      bool
	Label       string
}

func (ti *TextInput) InsertRune(ch rune) {
	before := ti.Text[:ti.Cursor]
	after := ti.Text[ti.Cursor:]
	ti.Text = before + string(ch) + after
	ti.Cursor++
}

func (ti *TextInput) DeleteRune() {
	if ti.Cursor > 0 {
		ti.Text = ti.Text[:ti.Cursor-1] + ti.Text[ti.Cursor:]
		ti.Cursor--
	}
}

func (ti *TextInput) CursorLeft() {
	if ti.Cursor > 0 {
		ti.Cursor--
	}
}

func (ti *TextInput) CursorRight() {
	if ti.Cursor < len(ti.Text) {
		ti.Cursor++
	}
}

func (ti *TextInput) MoveHome() { ti.Cursor = 0 }
func (ti *TextInput) MoveEnd()  { ti.Cursor = len(ti.Text) }

func (ti *TextInput) Render(c *render.Canvas, x, y, width int) {
	style := render.StyleDim
	labelStyle := render.StyleHighlight
	if ti.Active {
		style = render.StyleDefault
		labelStyle = render.StyleSuccess
	}
	if ti.Label != "" {
		c.SetContent(x, y, labelStyle, ti.Label+": ")
		x += len(ti.Label) + 2
	}
	display := ti.Text
	if display == "" && !ti.Active {
		display = ti.Placeholder
	}
	if len(display) > width-2 {
		display = display[:width-2]
	}
	c.SetContent(x, y, style, display)
	if ti.Active {
		cx := x + ti.Cursor
		if cx < x+width-2 {
			c.SetCell(cx, y, render.StyleHighlight, '_')
		}
	}
}
