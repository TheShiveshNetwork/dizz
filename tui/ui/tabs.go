package ui

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
)

type Tab struct {
	Name string
	Key  string
}

var Tabs = []Tab{
	{Name: "Dashboard", Key: "1"},
	{Name: "Symbols", Key: "2"},
	{Name: "Intents", Key: "3"},
	{Name: "TODOs", Key: "4"},
	{Name: "Snapshots", Key: "5"},
	{Name: "Configs", Key: "6"},
}

func RenderSidebar(c *render.Canvas, active int, focused bool, x, y, width int) {
	for i, tab := range Tabs {
		style := render.StyleDefault
		mark := "  "
		if i == active {
			style = render.StyleHighlight
			if focused {
				mark = "▸ "
			} else {
				mark = " >"
			}
		}
		label := mark + " " + tab.Name
		c.SetContent(x, y+i, style, label)
	}
}
