package ui

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/gdamore/tcell/v2"
)

type FormField struct {
	Label     string
	Value     string
	Focused   bool
	Validator func(string) string
}

type FormModel struct {
	Fields   []FormField
	Selected int
	Title    string
	SubmitFn func([]string)
	CancelFn func()
}

func (f *FormModel) PrevField() {
	f.Fields[f.Selected].Focused = false
	f.Selected--
	if f.Selected < 0 {
		f.Selected = 0
	}
	f.Fields[f.Selected].Focused = true
}

func (f *FormModel) NextField() {
	f.Fields[f.Selected].Focused = false
	f.Selected++
	if f.Selected >= len(f.Fields) {
		f.Selected = len(f.Fields) - 1
	}
	f.Fields[f.Selected].Focused = true
}

func (f *FormModel) TypeRune(ch rune) {
	field := &f.Fields[f.Selected]
	field.Value += string(ch)
}

func (f *FormModel) DeleteRune() {
	field := &f.Fields[f.Selected]
	if len(field.Value) > 0 {
		field.Value = field.Value[:len(field.Value)-1]
	}
}

func (f *FormModel) Values() []string {
	vals := make([]string, len(f.Fields))
	for i, field := range f.Fields {
		vals[i] = field.Value
	}
	return vals
}

func RenderForm(c *render.Canvas, form *FormModel, x, y, w int) {
	sepStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(80, 80, 80))
	bgStyle := tcell.StyleDefault.Background(tcell.NewRGBColor(30, 30, 30))

	for dy := -1; dy <= len(form.Fields)*2+2; dy++ {
		for dx := -2; dx < w+2; dx++ {
			c.SetCell(x+dx, y+dy, bgStyle, ' ')
		}
	}

	c.SetContent(x, y-1, render.StyleHighlight.Bold(true), form.Title)
	c.HLine(x, x+w-1, y, sepStyle, '─')

	for i, field := range form.Fields {
		fy := y + 2 + i*2
		c.SetContent(x, fy, render.StyleDim, field.Label+":")

		valStyle := render.StyleDefault
		if field.Focused {
			valStyle = render.StyleHighlight
		}
		val := field.Value
		if val == "" {
			val = "(empty)"
			valStyle = render.StyleDim
		}
		if len(val) > w-len(field.Label)-3 {
			val = val[:w-len(field.Label)-3]
		}
		c.SetContent(x+len(field.Label)+2, fy, valStyle, val)
	}

	help := "Tab: next  Enter: submit  Esc: cancel"
	c.SetContent(x, y+len(form.Fields)*2+2, render.StyleMuted, help)
}
