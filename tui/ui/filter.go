package ui

import (
	"strings"
	"unicode"

	"github.com/TheShiveshNetwork/dizz/tui/render"
)

type Filter struct {
	Mode  bool
	Query string
}

func (f *Filter) Active() bool {
	return f.Mode
}

func (f *Filter) HandleKey(key string) (consumed bool, changed bool) {
	if !f.Mode {
		if key == "/" {
			f.Mode = true
			f.Query = ""
			return true, true
		}
		return false, false
	}

	switch key {
	case "esc", "enter":
		f.Mode = false
		return true, true
	case "backspace":
		if len(f.Query) > 0 {
			f.Query = f.Query[:len(f.Query)-1]
			return true, true
		}
		return true, false
	case "space":
		f.Query += " "
		return true, true
	default:
		if len(key) == 1 {
			ch := rune(key[0])
			if unicode.IsPrint(ch) {
				f.Query += string(ch)
				return true, true
			}
		}
		return true, false
	}
}

func (f *Filter) MatchesAny(texts ...string) bool {
	if !f.Mode || f.Query == "" {
		return true
	}
	q := strings.ToLower(f.Query)
	for _, text := range texts {
		if strings.Contains(strings.ToLower(text), q) {
			return true
		}
	}
	return false
}

func (f *Filter) Render(c *render.Canvas, x, y int) {
	if !f.Mode {
		return
	}
	prompt := "/" + f.Query + "_"
	c.SetContent(x, y, render.StyleSuccess, prompt)
}
