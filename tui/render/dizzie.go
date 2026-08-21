package render

import (
	"math"
	"strings"

	"github.com/mattn/go-runewidth"
)

var dizzArtBase = []string{
	`                  ⣀⣿⣷⣤⣀                              ⣀⣤⣷⣿⣀`,
	`                  ⣤█⣿⣷⣤⣷⣷⣷⣷⣀                    ⣤⣷⣿⣷⣷⣤⣿⣿▓⣀`,
	`                  ⣤⣿⣀░▓⣿⣀ ⣀⣤⣿⣷⣀                ⣤⣿⣿⣷⣀ ⣀⣿█░⣀⣿⣀`,
	`                  ⣤⣿  ⣤▓█░⣤  ⣀⣷⣿⣷▒▓░⣤░▓▒⣤░▓▒⣷⣿⣷⣀  ⣤▒█▓⣤  ⣿⣤`,
	`                  ⣤⣿   ⣀▒██▒⣤    ▒█░ ▒█▒ ⣿█▓    ⣤▒██▒⣀    ⣿⣤`,
	`                  ⣀░    ⣤▓⣿⣀     ▒█░ ▒█▒ ⣿█▒     ⣀░█⣤    ⣀░⣀`,
	`                   ⣤░  ⣤⣤⣀        ▒█░ ▒█▒ ⣿█▒        ⣀⣤⣤  ⣿⣤`,
	`                    ⣿⣷⣤⣀          ⣤░⣤ ░█▒ ⣤░⣤         ⣀⣤⣤⣿`,
	`                    ⣀░                ⣀⣤⣀                ⣀▒⣀`,
	`                    ⣿⣤                                   ⣷⣷`,
	`                   ⣀░                                     ░⣀`,
	`                   ⣷░⣤⣤⣤⣤⣤⣀  ⣤░░⣿⣤        ⣀⣤░░░⣤⣀ ⣀⣤⣤⣤⣤⣤░⣀`,
	`                ⣀⣷⣷⣿█████░⣀⣀▒⣿⣀⣀⣤░⣿        ▒⣿⣤⣀⣤⣿░ ⣀░█████⣷⣷⣷⣀`,
	`                ⣀⣷⣷▓▒░░⣤            ⣷⣷⣷⣿⣤            ⣤▒▒▓█⣷⣷⣀`,
	`                ⣀⣀ ⣷█▓▒⣤            ⣤⣿⣿⣿⣀            ⣤⣿⣷░⣿ ⣀⣀`,
	`                  ⣀⣷⣷▒▒░⣤      ⣀⣷⣀   ⣤▒⣀   ⣀⣷        ⣀░▓░⣷⣷⣀`,
	`                  ⣀⣀ ⣀⣷⣿⣀       ⣤⣿⣷⣷⣷⣷⣀⣷⣷⣷⣷⣷⣀       ⣀⣿⣷  ⣀⣀`,
	`                       ⣀⣷⣿⣤                        ⣤⣿⣷⣀`,
	`                       ⣤▒██▓░⣷⣤⣀                ⣀⣤⣿░▓██▒⣷`,
	`                     ⣀⣿██▓▒▒▓██▒⣷⣷⣿⣷⣷⣷⣷⣷⣷⣿⣷⣷░███▒░▒██░⣀`,
	`                     ⣿⣿⣿⣿⣀   ⣀⣤                ⣷⣀   ⣀⣿⣿⣿░`,
	`                  ⣀⣷⣿⣷⣷⣷⣿⣷⣀                    ⣀⣷⣿⣿⣷⣷⣿⣿⣤`,
	`                ⣀⣿⣷⣀     ⣀▒⣤                    ⣤░⣀     ⣀⣷⣿⣀`,
	`               ⣀░⣀        ⣀⣀⣷░⣀                  ⣀⣿⣷⣤⣀        ⣀⣿⣤`,
	`          ⣀░░░░▒⣀        ⣀▒⣤░⣤                  ⣀░⣤▒⣀        ⣀░░░░░⣤`,
	`        ⣀▒█▓▒█░ ░ ░  ⣀⣿ ⣀███▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓███⣤ ⣷⣤░  ▒ ⣿██▒██⣤`,
	`        ⣤▓█▓░▒██▒▓⣷⣤▓⣤⣷▒⣤░███⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿▓██▒⣤░⣿⣤▓⣷⣤█▒██▓░░██⣷`,
	`       ⣤██░░█░▒███████████▒▓█░░██░░█▒░▓█░░█▓░▒█░░████████████▒⣿░░░░██⣿`,
	`      ⣷██▒⣿░▓▒░░█▓░▒█▒▒██▒▒▒█▒░▓█░░▓▓░▒█▒░▓█░░█▓░▒█████▒▓█░░█▓░░▓▒⣿░▓█▒⣀`,
	`     ░██⣿▒▒⣿░██▒░█▓░▒█▓⣿▒█▓░▒█▒⣿▒█░░▓█░░▓█⣿⣿▓█▒░▓█⣿⣿█▒░▒█▒░▓█▒⣿░▒▒⣿▓█▓⣀`,
	`  ⣀▒█▓▒▒▓▒▒▒▓▒▒▓▓▒░▓▒▒▒▓▓▒▒▓▓▓░▓▒▒░░▒█▒░▒█▓░▒█▓░░██░░▓█░░▓█▒▒▓█░░██▒░▒█▓⣤`,
	` ⣀▒█░░▒█▒⣿▓█▒░▓█▒░▓█⣿⣿▓█▒⣿▒▒⣿⣿⣿░⣿⣿⣿███▓⣿░█▓⣿⣿██⣿⣿▒█▓⣿░█▓⣿░█████▓⣿░██▒░▒█▓⣤`,
	`⣀▓████████████████████████████████████████████████████████████████████████⣷`,
	`⣿█████████████████████████████████████████████████████████████████████████▒`,
	`⣿█████████████████████████████████████████████████████████████████████████▒`,
}

const (
	handRowStart       = 21
	handRowEnd         = 25
	handFrameCount     = 10
	handShiftAmplitude = 1.0
)

type pawBounds struct {
	leftStart, leftEnd   int
	rightStart, rightEnd int
}

var pawBoundsByRow = map[int]pawBounds{
	21: {leftStart: 18, leftEnd: 35, rightStart: 55, rightEnd: 72},
	22: {leftStart: 16, leftEnd: 31, rightStart: 51, rightEnd: 66},
	23: {leftStart: 15, leftEnd: 30, rightStart: 50, rightEnd: 65},
	24: {leftStart: 10, leftEnd: 28, rightStart: 48, rightEnd: 68},
	25: {leftStart: 8, leftEnd: 25, rightStart: 45, rightEnd: 65},
}

var dizzArtFrames = generateTypingFrames(dizzArtBase)

func padRuneRow(row string, width int) string {
	rs := []rune(row)
	if len(rs) >= width {
		return string(rs[:width])
	}
	return row + strings.Repeat(" ", width-len(rs))
}

func normalizeArt(lines []string) ([]string, int) {
	width := 0
	for _, l := range lines {
		if w := len([]rune(l)); w > width {
			width = w
		}
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = padRuneRow(l, width)
	}
	return out, width
}

func shiftSegment(rowRunes []rune, start, end, offset int) {
	if start < 0 || end > len(rowRunes) || start >= end {
		return
	}
	length := end - start
	segment := make([]rune, length)
	copy(segment, rowRunes[start:end])

	shifted := make([]rune, length)
	for i := range shifted {
		shifted[i] = ' '
	}
	for i, r := range segment {
		dst := i + offset
		if dst >= 0 && dst < length {
			shifted[dst] = r
		}
	}
	copy(rowRunes[start:end], shifted)
}

func generateTypingFrames(base []string) [][]string {
	normalized, _ := normalizeArt(base)

	frames := make([][]string, handFrameCount)
	for f := 0; f < handFrameCount; f++ {
		phase := 2 * math.Pi * float64(f) / float64(handFrameCount)
		offset := int(math.Round(handShiftAmplitude * math.Sin(phase)))

		frame := make([]string, len(normalized))
		copy(frame, normalized)

		for r := handRowStart; r <= handRowEnd; r++ {
			bounds, ok := pawBoundsByRow[r]
			if !ok {
				continue
			}
			rowRunes := []rune(normalized[r])

			shiftSegment(rowRunes, bounds.leftStart, bounds.leftEnd, offset)
			shiftSegment(rowRunes, bounds.rightStart, bounds.rightEnd, -offset)

			frame[r] = string(rowRunes)
		}
		frames[f] = frame
	}
	return frames
}

func DizzArtDimensions() (width, height int) {
	art := dizzArtFrames[0]
	height = len(art)
	for _, line := range art {
		if w := runewidth.StringWidth(line); w > width {
			width = w
		}
	}
	return width, height
}

func RenderDizzie(c *Canvas, frameTick int64, centerX, topY int) {
	frame := frameTick % int64(len(dizzArtFrames))
	artLines := dizzArtFrames[frame]

	maxW := 0
	for _, line := range artLines {
		if w := runewidth.StringWidth(line); w > maxW {
			maxW = w
		}
	}

	startX := centerX - maxW/2
	for i, line := range artLines {
		c.SetContent(startX, topY+i, StyleDefault, line)
	}
}
