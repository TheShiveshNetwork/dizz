package store

import (
	"encoding/json"
	"os"
	"time"
)

// Task is a single work item tracked by taskforge.
type Task struct {
	Title    string `json:"title"`
	Due      string `json:"due,omitempty"`
	Priority string `json:"priority,omitempty"`
	Done     bool   `json:"done,omitempty"`
	Created  string `json:"created"`
}

// DefaultTasks returns the seeded task list used when no taskforge.json exists.
func DefaultTasks() []Task {
	return []Task{
		{Title: "write release notes", Due: "2026-08-03", Priority: "high", Created: time.Now().Format(time.RFC3339)},
		{Title: "refactor store package", Due: "", Priority: "medium", Created: time.Now().Format(time.RFC3339)},
		{Title: "ship notifications", Due: "2026-08-01", Priority: "high", Created: time.Now().Format(time.RFC3339)},
		{Title: "update landing page", Due: "2026-08-10", Priority: "low", Created: time.Now().Format(time.RFC3339)},
	}
}

func dataFile() string {
	return "taskforge.json"
}

// Load returns tasks from the working directory file, or the defaults.
func Load() ([]Task, error) {
	raw, err := os.ReadFile(dataFile())
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultTasks(), nil
		}
		return nil, err
	}
	var ts []Task
	if err := json.Unmarshal(raw, &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// Add appends a task and persists it to the working directory.
func Add(title, due string) error {
	ts, err := Load()
	if err != nil {
		return err
	}
	ts = append(ts, Task{Title: title, Due: due, Priority: "medium", Created: time.Now().Format(time.RFC3339)})
	raw, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile(), raw, 0644)
}
