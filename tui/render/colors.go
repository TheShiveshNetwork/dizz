package render

import "github.com/gdamore/tcell/v2"

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
