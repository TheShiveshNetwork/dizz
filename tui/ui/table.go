package ui

import (
	"fmt"

	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/gdamore/tcell/v2"
)

type Column struct {
	Name     string
	Width    int
	Sortable bool
}

type Row struct {
	Cells      []string
	Style      tcell.Style
	CellStyles []tcell.Style
	Data       interface{}
}

type TableModel struct {
	Cols     []Column
	Rows     []Row
	Selected int
	Offset   int
	SortCol  int
	SortAsc  bool
	ColGap   int
}

func (t *TableModel) SortBy(col int) {
	if col < 0 || col >= len(t.Cols) || !t.Cols[col].Sortable {
		return
	}
	if t.SortCol == col {
		t.SortAsc = !t.SortAsc
	} else {
		t.SortCol = col
		t.SortAsc = true
	}
}

func (t *TableModel) PrevRow() {
	t.Selected--
	if t.Selected < 0 {
		t.Selected = 0
	}
	t.ensureVisible()
}

func (t *TableModel) NextRow() {
	t.Selected++
	if t.Selected >= len(t.Rows) {
		t.Selected = len(t.Rows) - 1
	}
	t.ensureVisible()
}

func (t *TableModel) ensureVisible() {
	if t.Selected < t.Offset {
		t.Offset = t.Selected
	}
	if t.Selected >= t.Offset+10 {
		t.Offset = t.Selected - 9
	}
}

func RenderTable(c *render.Canvas, table *TableModel, x, y, w, h int) {
	if len(table.Cols) == 0 {
		return
	}

	cx := x
	for ci, col := range table.Cols {
		name := col.Name
		if col.Sortable {
			arrow := ' '
			if table.SortCol == ci {
				if table.SortAsc {
					arrow = '▲'
				} else {
					arrow = '▼'
				}
			}
			name = name + " " + string(arrow)
		}
		c.SetContentBounded(cx, y, col.Width, render.StyleHighlight.Bold(true), name)
		cx += col.Width + table.ColGap
	}

	sepStyle := render.StyleDim
	for sx := x; sx < x+w; sx++ {
		c.SetCell(sx, y+1, sepStyle, '─')
	}

	visibleRows := h - 3
	if visibleRows < 0 {
		visibleRows = 0
	}

	lastRendered := 0
	for i := 0; i < visibleRows && table.Offset+i < len(table.Rows); i++ {
		row := table.Rows[table.Offset+i]
		ry := y + 2 + i
		lastRendered = table.Offset + i
		isSelected := table.Offset+i == table.Selected

		rowStyle := row.Style
		if isSelected {
			rowStyle = rowStyle.Bold(true)
		}

		cx := x
		for ci, col := range table.Cols {
			cellText := ""
			if ci < len(row.Cells) {
				cellText = row.Cells[ci]
			}
			cellStyle := rowStyle
			if ci < len(row.CellStyles) {
				cellStyle = row.CellStyles[ci]
				if isSelected {
					cellStyle = cellStyle.Bold(true)
				}
			}
			c.SetContentBounded(cx, ry, col.Width, cellStyle, cellText)
			cx += col.Width + table.ColGap
		}
	}

	if table.Offset > 0 {
		c.SetCell(x, y+2, render.StyleDim, '▲')
	}

	if lastRendered < len(table.Rows)-1 {
		indicatorY := y + 2 + visibleRows - 1
		if indicatorY > y+2 {
			indicatorY--
		}
		c.SetCell(x, indicatorY, render.StyleDim, '▼')
		more := fmt.Sprintf("%d more", len(table.Rows)-lastRendered-1)
		c.SetContent(x+1, indicatorY, render.StyleDim, more)
	}
}
