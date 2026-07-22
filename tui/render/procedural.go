package render

import (
	"fmt"
	"math/rand"
	"time"
)

type Star struct {
	X, Y  int
	Phase int64
	Seed  int64
}

func GenerateStarField(w, h int, density float64, seed int64) []Star {
	rng := rand.New(rand.NewSource(seed))
	total := int(float64(w*h) * density)
	stars := make([]Star, total)
	for i := range stars {
		stars[i] = Star{
			X:     rng.Intn(w),
			Y:     rng.Intn(h),
			Phase: rng.Int63n(4000),
			Seed:  rng.Int63n(1000),
		}
	}
	return stars
}

func GetStarChar(s Star, now int64) (rune, bool) {
	t := (now + s.Phase + s.Seed) % 4000
	var ch rune
	var visible bool
	switch {
	case t < 1000:
		ch, visible = '★', true
	case t < 2000:
		ch, visible = '·', true
	case t < 2500:
		ch, visible = ' ', false
	case t < 3500:
		ch, visible = '·', true
	default:
		ch, visible = '★', true
	}
	return ch, visible
}

func GetMoonPhase(iteration int) string {
	phases := []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	return phases[iteration%len(phases)]
}

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

func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		d := int(d.Hours() / 24)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	}
}

type Meteor struct {
	X, Y      int
	Length    int
	Speed     int
	StartTime int64
}

type TrailPoint struct {
	X, Y  int
	Char  rune
	State string
}

func GenerateMeteorShower(w, h, count int, seed int64) []Meteor {
	rng := rand.New(rand.NewSource(seed))
	meteors := make([]Meteor, count)
	for i := range meteors {
		side := rng.Intn(2)
		var x, y int
		if side == 0 {
			x = rng.Intn(w / 3)
			y = rng.Intn(h)
		} else {
			x = w - rng.Intn(w/3) - 1
			y = rng.Intn(h)
		}
		meteors[i] = Meteor{
			X:         x,
			Y:         y,
			Length:    rng.Intn(6) + 3,
			Speed:     rng.Intn(3) + 1,
			StartTime: rng.Int63n(10000),
		}
	}
	return meteors
}

func GetMeteorTrail(m Meteor, now int64, w, h int) []TrailPoint {
	elapsed := int((now - m.StartTime) / 50)
	if elapsed < 0 {
		return nil
	}
	dist := elapsed * m.Speed
	dx := int(float64(dist) * 0.7)
	dy := dist

	headX := m.X - dx
	headY := m.Y + dy

	if headY < 0 || headY >= h {
		return nil
	}

	trail := make([]TrailPoint, 0, m.Length)
	for i := 0; i < m.Length; i++ {
		tx := headX + int(float64(i)*0.3)
		ty := headY - i
		if ty < 0 || ty >= h || tx < 0 || tx >= w {
			continue
		}
		ch := rune('░')
		st := "dim"
		if i == 0 {
			ch = rune('█')
			st = "bright"
		} else if i < 3 {
			ch = rune('▒')
			st = "dim"
		}
		trail = append(trail, TrailPoint{X: tx, Y: ty, Char: ch, State: st})
	}
	return trail
}
