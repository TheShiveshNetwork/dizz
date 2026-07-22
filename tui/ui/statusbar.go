package ui

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
)

type StatusData struct {
	ProjectName string
	Branch      string
	Commit      string
	Dirty       bool
	Version     string
}

func RenderTopBar(c *render.Canvas, data StatusData, y, width int) {
	title := "dizz - The brain for your codebase"

	right := ""
	if data.ProjectName != "" {
		right = data.ProjectName
	}
	if data.Branch != "" {
		if right != "" {
			right += "  "
		}
		right += "@ " + data.Branch
	}

	c.SetContent(0, y, render.StyleHighlight, title)

	if right != "" {
		if len(right)+1 > width {
			right = right[:width-4] + "..."
		}
		c.SetContent(width-len(right), y, render.StyleMuted, right)
	}
}

func RenderStatusBar(c *render.Canvas, data StatusData, focusSidebar bool, y, width int, inputMode bool) {
	help := "? help  q quit"
	focus := "main"
	if focusSidebar {
		focus = "sidebar"
	}
	focusIndicator := "[" + focus + "]"
	modeStr := "NORMAL"
	modeStyle := render.StyleMuted
	if inputMode {
		modeStr = "INSERT"
		modeStyle = render.StyleSuccess.Bold(true)
	}

	for x := 0; x < width; x++ {
		c.SetCell(x, y, render.StyleMuted, '─')
	}
	c.SetContent(0, y+1, render.StyleMuted, help)
	c.SetContent(width-len(modeStr), y+1, modeStyle, modeStr)
	c.SetContent(width-len(modeStr)-len(focusIndicator)-1, y+1, render.StyleMuted, focusIndicator)
}
