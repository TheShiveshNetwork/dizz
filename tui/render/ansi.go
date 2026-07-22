package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func styleANSI(s tcell.Style) string {
	var parts []string

	fg, bg, attrs := s.Decompose()

	if fg != tcell.ColorDefault {
		r, g, b := fg.RGB()
		parts = append(parts, fmt.Sprintf("38;2;%d;%d;%d", r, g, b))
	}

	if bg != tcell.ColorDefault {
		r, g, b := bg.RGB()
		parts = append(parts, fmt.Sprintf("48;2;%d;%d;%d", r, g, b))
	} else if len(parts) > 0 {
		parts = append(parts, "49")
	}

	if attrs&tcell.AttrBold != 0 {
		parts = append(parts, "1")
	}
	if attrs&tcell.AttrDim != 0 {
		parts = append(parts, "2")
	}
	if attrs&tcell.AttrBlink != 0 {
		parts = append(parts, "5")
	}

	if len(parts) == 0 {
		return "\033[0m"
	}
	return "\033[" + strings.Join(parts, ";") + "m"
}

func (c *Canvas) ANSIFrame() string {
	var b strings.Builder
	for y := 0; y < c.height; y++ {
		if y > 0 {
			b.WriteString("\r\n")
		}
		for x := 0; x < c.width; x++ {
			cell := c.cells[y][x]
			ansi := styleANSI(cell.Style)
			b.WriteString(ansi)
			b.WriteRune(cell.Char)
		}
	}
	b.WriteString("\033[0m")
	return b.String()
}

func (c *Canvas) ANSIDiff(prev *Canvas) string {
	if prev == nil || prev.width != c.width || prev.height != c.height {
		return c.ANSIFrame()
	}

	var b strings.Builder
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			if c.cells[y][x] != prev.cells[y][x] {
				cell := c.cells[y][x]
				ansi := styleANSI(cell.Style)
				b.WriteString(fmt.Sprintf("\033[%d;%dH%s%c", y+1, x+1, ansi, cell.Char))
			}
		}
	}
	if b.Len() > 0 {
		b.WriteString("\033[0m")
	}
	return b.String()
}
