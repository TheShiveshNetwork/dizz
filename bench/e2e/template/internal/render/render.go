package render

import (
	"fmt"

	"taskforge/internal/store"
)

// List prints tasks with their state and priority.
func List(ts []store.Task, all, verbose bool) {
	shown := 0
	for _, t := range ts {
		if t.Done && !all {
			continue
		}
		state := " "
		if t.Done {
			state = "x"
		}
		fmt.Printf("[%s] %s (priority=%s)\n", state, t.Title, t.Priority)
		shown++
	}
	if verbose {
		fmt.Printf("listed %d tasks\n", shown)
	}
}

// Plan prints tasks ordered as supplied.
func Plan(ts []store.Task) {
	for _, t := range ts {
		due := t.Due
		if due == "" {
			due = "no-due"
		}
		fmt.Printf("%s | %s | %s\n", due, t.Priority, t.Title)
	}
}
