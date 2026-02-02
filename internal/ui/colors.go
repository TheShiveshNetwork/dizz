package ui

// ANSI color codes
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"

	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

// Background colors
const (
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgGray    = "\033[100m"
)

func Colorize(text, color string) string {
	return color + text + Reset
}

func Success(text string) string {
	return Colorize(text, BrightGreen)
}

func Warning(text string) string {
	return Colorize(text, BrightYellow)
}

func Error(text string) string {
	return Colorize(text, BrightRed)
}

func Info(text string) string {
	return Colorize(text, BrightCyan)
}

func Muted(text string) string {
	return Colorize(text, Gray)
}

func Header(text string) string {
	return Bold + BrightWhite + text + Reset
}

func Highlight(text string) string {
	return Bold + BrightBlue + text + Reset
}
