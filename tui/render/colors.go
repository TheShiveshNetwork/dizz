package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	ColorGreen  = tcell.NewRGBColor(int32(80), int32(200), int32(120))
	ColorYellow = tcell.NewRGBColor(int32(220), int32(200), int32(60))
	ColorRed    = tcell.NewRGBColor(int32(220), int32(80), int32(80))
	ColorCyan   = tcell.NewRGBColor(int32(80), int32(180), int32(220))
	ColorDim    = tcell.NewRGBColor(int32(100), int32(100), int32(100))
	ColorBright = tcell.NewRGBColor(int32(255), int32(255), int32(255))

	StyleDefault   = tcell.StyleDefault
	StyleBold      = tcell.StyleDefault.Bold(true)
	StyleDim       = tcell.StyleDefault.Dim(true)
	StyleHighlight = tcell.StyleDefault.Bold(true).Foreground(ColorBright)
	StyleMuted     = tcell.StyleDefault.Foreground(ColorDim)
	StyleSuccess   = tcell.StyleDefault.Foreground(ColorGreen)
	StyleWarning   = tcell.StyleDefault.Foreground(ColorYellow)
	StyleError     = tcell.StyleDefault.Foreground(ColorRed)
	StyleInfo      = tcell.StyleDefault.Foreground(ColorCyan)
)

func StateColor(state string) tcell.Style {
	switch state {
	case "active":
		return StyleSuccess
	case "planned":
		return StyleWarning
	case "unstable":
		return StyleError
	case "unused":
		return StyleInfo
	case "abandoned":
		return StyleMuted
	default:
		return StyleDefault
	}
}

func SeverityColor(sev int) tcell.Style {
	switch {
	case sev >= 3:
		return StyleError.Bold(true)
	case sev >= 2:
		return StyleWarning
	case sev >= 1:
		return StyleInfo
	default:
		return StyleMuted
	}
}

func TodoTypeStyle(typ string) tcell.Style {
	switch strings.ToUpper(typ) {
	case "TODO":
		fg := tcell.NewRGBColor(255, 255, 255)
		bg := tcell.NewRGBColor(40, 100, 180)
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	case "FIXME":
		fg := tcell.NewRGBColor(255, 255, 255)
		bg := tcell.NewRGBColor(180, 50, 50)
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	default:
		fg := tcell.ColorWhite
		bg := tcell.ColorGray
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}
}

func IntentTypeStyle(typ string) tcell.Style {
	var fg, bg tcell.Color
	switch typ {
	case "todo":
		fg = tcell.NewRGBColor(255, 255, 255)
		bg = tcell.NewRGBColor(40, 100, 180)
	case "fixme":
		fg = tcell.NewRGBColor(255, 255, 255)
		bg = tcell.NewRGBColor(180, 50, 50)
	case "refactor":
		fg = tcell.NewRGBColor(255, 255, 255)
		bg = tcell.NewRGBColor(50, 140, 80)
	case "question":
		fg = tcell.NewRGBColor(30, 30, 30)
		bg = tcell.NewRGBColor(200, 180, 60)
	case "hack":
		fg = tcell.NewRGBColor(255, 255, 255)
		bg = tcell.NewRGBColor(140, 60, 160)
	case "temporary":
		fg = tcell.NewRGBColor(200, 200, 200)
		bg = tcell.NewRGBColor(80, 80, 80)
	default:
		fg = tcell.ColorWhite
		bg = tcell.ColorGray
	}
	return tcell.StyleDefault.Foreground(fg).Background(bg)
}
