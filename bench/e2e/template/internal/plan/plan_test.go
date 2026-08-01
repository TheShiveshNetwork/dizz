package plan

import (
	"testing"

	"taskforge/internal/store"
)

func TestSortByPriority(t *testing.T) {
	ordered, err := Sort(store.DefaultTasks(), "priority")
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}
	if len(ordered) == 0 {
		t.Fatal("Sort returned nothing")
	}
	// Highest-ranked task must be first.
	if ordered[0].Priority != "high" {
		t.Fatalf("expected high priority first, got %q", ordered[0].Priority)
	}
}

func TestSortUnknownOrder(t *testing.T) {
	if _, err := Sort(store.DefaultTasks(), "bogus"); err == nil {
		t.Fatal("expected error for unknown sort order")
	}
}
