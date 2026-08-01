package store

import "testing"

func TestDefaultTasksNonEmpty(t *testing.T) {
	ts := DefaultTasks()
	if len(ts) == 0 {
		t.Fatal("DefaultTasks returned nothing")
	}
}

func TestDefaultTasksHavePriorities(t *testing.T) {
	for _, task := range DefaultTasks() {
		if task.Title == "" {
			t.Fatal("task with empty title")
		}
		if task.Priority == "" {
			t.Fatalf("task %q has no priority", task.Title)
		}
	}
}
