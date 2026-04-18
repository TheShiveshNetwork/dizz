package render

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
)

func RenderTodos(todos []state.Todo) {
	if len(todos) == 0 {
		return
	}

	fmt.Println(ui.Highlight("━━ INTENTS & TODOS"))

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

		renderTodoBlock(todo.Type, todo.Text)
	}

	if len(todos) > limit {
		fmt.Printf(
			ui.Muted("     ... and %d more\n"),
			len(todos)-limit,
		)
	}

	fmt.Println()
}

// RenderTodoList renders all todos in a list format similar to intent list
func RenderTodoList(todos []state.Todo) {
	if len(todos) == 0 {
		fmt.Println("No active todos found.")
		return
	}

	fmt.Println()
	fmt.Println(ui.Highlight("━━ CODE TODOS & FIXMES"))
	fmt.Println()

	for _, todo := range todos {
		// Location (always muted)
		fmt.Printf(
			"  %s\n",
			ui.Muted(fmt.Sprintf("%s:%d", todo.File, todo.Line)),
		)

		renderTodoBlock(todo.Type, todo.Text)
	}

	fmt.Printf(ui.Muted("  Total: %d todos found in code\n"), len(todos))
	fmt.Println()
}

func renderIntentBlock(intent state.Intent) {
	var header string
	var bg string
	var fg string

	switch intent.Type {
	case state.Fixme, state.Hack:
		header = " " + strings.ToUpper(string(intent.Type)) + " "
		bg = ui.BgRed
		fg = ui.White

	case state.IntentTodo:
		header = " " + strings.ToUpper(string(intent.Type)) + " "
		bg = ui.BgYellow
		fg = ui.Black

	case state.Question:
		header = " " + strings.ToUpper(string(intent.Type)) + " "
		bg = ui.BgCyan
		fg = ui.Black

	case state.Refactor:
		header = " " + strings.ToUpper(string(intent.Type)) + " "
		bg = ui.BgMagenta
		fg = ui.Black

	case state.Temporary:
		header = " " + strings.ToUpper(string(intent.Type)) + " "
		bg = ui.BgGray
		fg = ui.White

	default:
		header = " " + strings.ToUpper(string(intent.Type)) + " "
		bg = ui.BgGray
		fg = ui.White
	}

	var severityColor func(string) string
	switch {
	case intent.Severity >= 3:
		severityColor = ui.Error
	case intent.Severity >= 2:
		severityColor = ui.Warning
	case intent.Severity >= 1:
		severityColor = ui.Info
	default:
		severityColor = ui.Muted
	}

	fmt.Printf("     %s %s %s\n",
		ui.Colorize(header, fg+bg),
		severityColor(intent.Message),
		ui.Muted("("+intent.ID+")"),
	)

	// Render metadata (all muted)
	var metadata []string
	metadata = append(metadata, fmt.Sprintf("Severity: %d", intent.Severity))

	// Format time using utils.FormatTime
	timeSince := time.Since(intent.CreatedAt)
	timeDisplay := utils.FormatTime(timeSince)
	metadata = append(metadata, fmt.Sprintf("Created: %s", timeDisplay.Text))

	if len(intent.Tags) > 0 {
		metadata = append(metadata, fmt.Sprintf("Tags: %s", strings.Join(intent.Tags, ", ")))
	}

	fmt.Printf("     %s\n", ui.Muted(strings.Join(metadata, ", ")))
	fmt.Println()
}

// RenderIntents renders human-authored intents with enhanced UI
func RenderIntents(intents []state.Intent) {
	if len(intents) == 0 {
		return
	}

	// Sort intents by severity in decreasing order
	sort.Slice(intents, func(i, j int) bool {
		return intents[i].Severity > intents[j].Severity
	})

	// limit := 5
	// if len(intents) < limit {
	// 	limit = len(intents)
	// }

	limit := len(intents)
	for i := 0; i < limit; i++ {
		intent := intents[i]
		renderIntentBlock(intent)
	}

	if len(intents) > limit {
		fmt.Printf(
			ui.Muted("     ... and %d more\n"),
			len(intents)-limit,
		)
	}

	fmt.Println()
}

// RenderTodosAndIntents renders both todos and intents in a unified view
func RenderTodosAndIntents(todos []state.Todo, intents []state.Intent) {
	hasTodos := len(todos) > 0
	hasIntents := len(intents) > 0

	if !hasTodos && !hasIntents {
		return
	}

	// Render todos first (existing behavior)
	if hasTodos {
		RenderTodos(todos)
	}

	// Render intents with enhanced UI
	if hasIntents {
		RenderIntents(intents)
	}
}

func renderTodoBlock(todoType string, raw string) {
	text := normalizeTodoText(raw)
	header, fg, bg := classifyTodoStyle(todoType, text)

	content := " " + text + " "

	// Render block
	println("     " + header)
	println("     " + fg + bg + content + ui.Reset)
}

func classifyTodoStyle(todoType string, text string) (header string, fg string, bg string) {
	normalizedType := strings.ToUpper(strings.TrimSpace(todoType))
	switch normalizedType {
	case "FIXME", "FIX", "BUG", "HACK":
		return ui.Error(" FIX "), ui.White, ui.BgRed
	case "TODO", "TBD":
		return ui.Warning(" TODO "), ui.Black, ui.BgYellow
	case "NOTE", "INFO":
		return ui.Info(" NOTE "), ui.Black, ui.BgCyan
	}

	upper := strings.ToUpper(text)
	switch {
	case strings.Contains(upper, "FIX"),
		strings.Contains(upper, "BUG"),
		strings.Contains(upper, "HACK"):
		return ui.Error(" FIX "), ui.White, ui.BgRed
	case strings.Contains(upper, "TODO"),
		strings.Contains(upper, "TBD"):
		return ui.Warning(" TODO "), ui.Black, ui.BgYellow
	case strings.Contains(upper, "NOTE"),
		strings.Contains(upper, "INFO"):
		return ui.Info(" NOTE "), ui.Black, ui.BgCyan
	default:
		return ui.Highlight(" TODO "), ui.White, ui.BgGray
	}
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
