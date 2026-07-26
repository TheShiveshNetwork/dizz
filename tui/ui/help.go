package ui

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/gdamore/tcell/v2"
)

type KeyBinding struct {
	Key  string
	Desc string
}

var GlobalBindings = []KeyBinding{
	{Key: "Tab", Desc: "Toggle focus sidebar / main"},
	{Key: "↑↓", Desc: "Sidebar: switch tab | Main: navigate"},
	{Key: "1-5", Desc: "Jump to tab"},
	{Key: "Enter", Desc: "Select / detail"},
	{Key: "Esc", Desc: "Back / cancel"},
	{Key: "R", Desc: "Refresh"},
	{Key: "?", Desc: "Toggle help"},
	{Key: "q", Desc: "Quit"},
}

var ViewBindings = map[string][]KeyBinding{
	"symbols": {
		{Key: "/", Desc: "Search symbols"},
		{Key: "Enter", Desc: "Toggle detail panel"},
		{Key: "1/2", Desc: "Sort by name/state"},
		{Key: "a/p/u/s/d", Desc: "Filter by state"},
		{Key: "x", Desc: "Clear filter"},
	},
	"intents": {
		{Key: "/", Desc: "Search intents"},
		{Key: "i", Desc: "Add intent modal"},
		{Key: "r", Desc: "Resolve confirm modal"},
		{Key: "a/s", Desc: "Toggle filter"},
		{Key: "x", Desc: "Clear filter"},
	},
	"snapshots": {
		{Key: "/", Desc: "Search snapshots"},
		{Key: "c", Desc: "Create snapshot"},
		{Key: "d", Desc: "View delta"},
		{Key: "enter", Desc: "Checkout snapshot"},
		{Key: "p", Desc: "Prune old snapshots"},
		{Key: "r", Desc: "Refresh list"},
	},
	"todos": {
		{Key: "/", Desc: "Search todos"},
		{Key: "Enter", Desc: "Toggle collapse file"},
		{Key: "a", Desc: "Toggle all"},
		{Key: "t f", Desc: "Filter TODO / FIXME"},
		{Key: "x", Desc: "Clear filter"},
	},
}

func RenderHelp(c *render.Canvas, activeTab string, width, height int) {
	c.Fill(tcell.StyleDefault.Foreground(tcell.NewRGBColor(30, 30, 30)).Background(tcell.NewRGBColor(200, 200, 200)), ' ')

	style := tcell.StyleDefault.Background(tcell.NewRGBColor(200, 200, 200)).Foreground(tcell.NewRGBColor(20, 20, 20))
	keyStyle := style.Bold(true)
	titleStyle := style.Bold(true)

	c.SetContent(2, 1, titleStyle, "Keyboard Shortcuts")
	c.HLine(2, width-3, 3, style, '─')

	y := 5
	c.SetContent(2, y, titleStyle, "Global")
	y += 2
	for _, kb := range GlobalBindings {
		c.SetContent(4, y, keyStyle, kb.Key)
		c.SetContent(4+len(kb.Key)+2, y, style, kb.Desc)
		y++
	}

	if bindings, ok := ViewBindings[activeTab]; ok {
		y++
		c.SetContent(2, y, titleStyle, activeTab)
		y += 2
		for _, kb := range bindings {
			c.SetContent(4, y, keyStyle, kb.Key)
			c.SetContent(4+len(kb.Key)+2, y, style, kb.Desc)
			y++
		}
	}

	closeStyle := style.Bold(true)
	closeText := "Press ? or Esc to close"
	c.SetContent((width-len(closeText))/2, height-2, closeStyle, closeText)
}

func RenderHelpOverlay(c *render.Canvas, activeTab string) {
	w, h := c.Width(), c.Height()
	overlay := render.NewCanvas(w, h)
	RenderHelp(overlay, activeTab, w, h)
	c.Blit(overlay, 0, 0)
}
