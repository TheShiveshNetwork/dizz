package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type Cell struct {
	Char  rune
	Style tcell.Style
}

type Canvas struct {
	cells  [][]Cell
	width  int
	height int
}

func NewCanvas(w, h int) *Canvas {
	cells := make([][]Cell, h)
	for y := range cells {
		cells[y] = make([]Cell, w)
		for x := range cells[y] {
			cells[y][x] = Cell{Char: ' ', Style: tcell.StyleDefault}
		}
	}
	return &Canvas{cells: cells, width: w, height: h}
}

func (c *Canvas) Width() int  { return c.width }
func (c *Canvas) Height() int { return c.height }

func (c *Canvas) InBounds(x, y int) bool {
	return x >= 0 && x < c.width && y >= 0 && y < c.height
}

func (c *Canvas) SetCell(x, y int, style tcell.Style, ch rune) {
	if !c.InBounds(x, y) {
		return
	}
	c.cells[y][x] = Cell{Char: ch, Style: style}
}

func (c *Canvas) SetContent(x, y int, style tcell.Style, text string) {
	cx := x
	for _, ch := range text {
		w := runewidth.RuneWidth(ch)
		if w < 1 {
			w = 1
		}
		c.SetCell(cx, y, style, ch)
		cx += w
	}
}

func (c *Canvas) SetContentBounded(x, y, max int, style tcell.Style, text string) {
	cx := x
	for _, ch := range text {
		w := runewidth.RuneWidth(ch)
		if w < 1 {
			w = 1
		}
		if cx+w > x+max {
			break
		}
		c.SetCell(cx, y, style, ch)
		cx += w
	}
}

func (c *Canvas) Fill(style tcell.Style, ch rune) {
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			c.cells[y][x] = Cell{Char: ch, Style: style}
		}
	}
}

func (c *Canvas) Rect(x, y, w, h int, style tcell.Style, fill rune) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.SetCell(x+dx, y+dy, style, fill)
		}
	}
}

func (c *Canvas) HLine(x0, x1, y int, style tcell.Style, ch rune) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
		c.SetCell(x, y, style, ch)
	}
}

func (c *Canvas) Sub(x, y, w, h int) *Canvas {
	sub := NewCanvas(w, h)
	for dy := 0; dy < h && y+dy < c.height; dy++ {
		for dx := 0; dx < w && x+dx < c.width; dx++ {
			sub.cells[dy][dx] = c.cells[y+dy][x+dx]
		}
	}
	return sub
}

func (c *Canvas) Blit(src *Canvas, x, y int) {
	for sy := 0; sy < src.height; sy++ {
		for sx := 0; sx < src.width; sx++ {
			c.SetCell(x+sx, y+sy, src.cells[sy][sx].Style, src.cells[sy][sx].Char)
		}
	}
}

func (c *Canvas) String() string {
	var sb strings.Builder
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			ch := c.cells[y][x].Char
			if ch == 0 {
				ch = ' '
			}
			sb.WriteRune(ch)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
