package render

func RenderBar(count, total, width int) string {
	if total <= 0 {
		return ""
	}
	filled := (count * width) / total
	if filled > width {
		filled = width
	}
	var bar string
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
