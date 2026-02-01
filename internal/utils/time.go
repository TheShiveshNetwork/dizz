package utils

import (
	"time"
	"fmt"

	"github.com/TheShiveshNetwork/dizz/internal/ui"
)

type TimeDisplay struct {
	Text	string
	Color	string
}

func FormatTime(timeSince time.Duration) TimeDisplay {
	switch {
	case timeSince < time.Minute:
		return TimeDisplay{
			Text:  "just now",
			Color: ui.BrightGreen,
		}

	case timeSince < time.Hour:
		mins := int(timeSince.Minutes())
		return TimeDisplay{
			Text:  fmt.Sprintf("%d minutes ago", mins),
			Color: ui.BrightGreen,
		}

	case timeSince < 24*time.Hour:
		hrs := int(timeSince.Hours())
		return TimeDisplay{
			Text:  fmt.Sprintf("%d hours ago", hrs),
			Color: ui.BrightYellow,
		}

	case timeSince < 7*24*time.Hour:
		days := int(timeSince.Hours() / 24)
		return TimeDisplay{
			Text:  fmt.Sprintf("%d days ago", days),
			Color: ui.BrightYellow,
		}

	case timeSince < 30*24*time.Hour:
		weeks := int(timeSince.Hours() / 24 / 7)
		return TimeDisplay{
			Text:  fmt.Sprintf("%d weeks ago", weeks),
			Color: ui.BrightRed,
		}

	default:
		months := int(timeSince.Hours() / 24 / 30)
		return TimeDisplay{
			Text:  fmt.Sprintf("%d months ago", months),
			Color: ui.BrightRed,
		}
	}
}

