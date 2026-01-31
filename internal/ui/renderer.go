package ui

import (
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/state"
)

type RenderArgs struct {
	Title    string
	Subtitle string

	Symbols []state.Symbol
	ShowAll bool

	MaxPerFile int
	ShowChurn  bool
}

func RenderSymbolGroup(args RenderArgs) {
	if len(args.Symbols) == 0 {
		return
	}

	groupState := args.Symbols[0].State

	// ---- Header ----
	title := args.Title
	fmt.Println(
		ColorByState(title, groupState),
		Muted(fmt.Sprintf("(%s)", args.Subtitle)),
	)

	groups := GroupSymbolsByFile(args.Symbols)

	for _, group := range groups {
		// File line
		fmt.Printf(
			"  %s %s\n",
			ColorByState(group.File, groupState),
			Muted(fmt.Sprintf("(%d items)", group.Count)),
		)

		limit := args.MaxPerFile
		if args.ShowAll {
			limit = len(group.Symbols)
		}

		for i, sym := range group.Symbols {
			if i >= limit {
				fmt.Printf(
					Muted("     ... and %d more\n"),
					len(group.Symbols)-limit,
				)
				break
			}
			fmt.Printf(
				"     • %s %d:%d",
				ColorByState(sym.Name, sym.State),
				sym.Line,
				sym.Column,
			)

			if args.ShowChurn {
				fmt.Printf(
					" %s",
					Error(fmt.Sprintf("(churn: %d)", sym.ChurnCount)),
				)
			}

			fmt.Println()
		}

		fmt.Println()
	}
}

func RenderTodos(todos []state.Todo) {
	if len(todos) == 0 {
		return
	}

	fmt.Println(Highlight("━━ 📝 TODOS"))

	limit := 3
	if len(todos) < limit {
		limit = len(todos)
	}

	for i := 0; i < limit; i++ {
		todo := todos[i]

		// Location (always muted)
		fmt.Printf(
			"  %s\n",
			Muted(fmt.Sprintf("%s:%d", todo.File, todo.Line)),
		)

		renderTodoBlock(todo.Text)
	}

	if len(todos) > limit {
		fmt.Printf(
			Muted("     ... and %d more\n"),
			len(todos)-limit,
		)
	}

	fmt.Println()
}

func ColorByState(text string, s state.SymbolState) string {
	switch s {
	case state.Planned:
		return Warning(text)
	case state.Unstable:
		return Error(text)
	case state.Unused:
		return Info(text)
	case state.Active:
		return Success(text)
	case state.Abandoned:
		return Muted(text)
	default:
		return text
	}
}

func colorizeTodoText(text string) string {
	upper := strings.ToUpper(text)

	switch {
	case strings.Contains(upper, "FIX"),
		strings.Contains(upper, "BUG"),
		strings.Contains(upper, "HACK"):
		return Error(text)

	case strings.Contains(upper, "TODO"),
		strings.Contains(upper, "TBD"):
		return Warning(text)

	case strings.Contains(upper, "NOTE"),
		strings.Contains(upper, "INFO"):
		return Info(text)

	default:
		return Highlight(text)
	}
}

// FIXME modularize these functions
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
		header = Error(" FIX ")
		bg = BgRed
		fg = White

	case strings.Contains(upper, "TODO"),
		strings.Contains(upper, "TBD"):
		header = Warning(" TODO ")
		bg = BgYellow
		fg = Black

	case strings.Contains(upper, "NOTE"),
		strings.Contains(upper, "INFO"):
		header = Info(" NOTE ")
		bg = BgCyan
		fg = Black

	default:
		header = Highlight(" TODO ")
		bg = BgGray
		fg = White
	}

	content := " " + text + " "

	// Render block
	println("     " + header)
	println("     " + fg + bg + content + Reset)
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
