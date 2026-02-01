package render

import (
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

func RenderTodos(todos []state.Todo) {
	if len(todos) == 0 {
		return
	}

	fmt.Println(ui.Highlight("━━ 📝 TODOS"))

	limit := 3
	if len(todos) < limit {
		limit = len(todos)
	}

	for i := 0; i < limit; i++ {
		todo := todos[i]

		// Location (always muted)
		fmt.Printf(
			"  %s\n",
			ui.Muted(fmt.Sprintf("%s:%d", todo.File, todo.Line)),
		)

		renderTodoBlock(todo.Text)
	}

	if len(todos) > limit {
		fmt.Printf(
			ui.Muted("     ... and %d more\n"),
			len(todos)-limit,
		)
	}

	fmt.Println()
}

func renderTodoBlock(raw string) {
	text := normalizeTodoText(raw)
	upper := strings.ToUpper(text)

	var header string
	var fg string
	var bg string

	switch {
	case strings.Contains(upper, "FIX"),
		strings.Contains(upper, "BUG"),
		strings.Contains(upper, "HACK"):
		header = ui.Error(" FIX ")
		bg = ui.BgRed
		fg = ui.White

	case strings.Contains(upper, "TODO"),
		strings.Contains(upper, "TBD"):
		header = ui.Warning(" TODO ")
		bg = ui.BgYellow
		fg = ui.Black

	case strings.Contains(upper, "NOTE"),
		strings.Contains(upper, "INFO"):
		header = ui.Info(" NOTE ")
		bg = ui.BgCyan
		fg = ui.Black

	default:
		header = ui.Highlight(" TODO ")
		bg = ui.BgGray
		fg = ui.White
	}

	content := " " + text + " "

	// Render block
	println("     " + header)
	println("     " + fg + bg + content + ui.Reset)
}

func normalizeTodoText(text string) string {
	t := strings.TrimSpace(text)

	// Strip common comment prefixes
	prefixes := []string{
		"//", "#", "/*", "*/", "*",
	}

	for _, p := range prefixes {
		if strings.HasPrefix(t, p) {
			t = strings.TrimSpace(strings.TrimPrefix(t, p))
		}
	}

	return t
}

