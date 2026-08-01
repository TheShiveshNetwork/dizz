package plan

import (
	"errors"
	"sort"
	"strings"
	"time"

	"taskforge/internal/store"
)

// TODO: due-date sorting is not wired into Sort yet.
// TODO: implement Remind, which returns tasks due on or before now.
// FIXME: priority sorting ranks "high" tasks the same as "medium".

// Sort orders tasks by the given key ("priority" or "due").
func Sort(ts []store.Task, by string) ([]store.Task, error) {
	switch by {
	case "priority":
		return sortByPriority(ts), nil
	case "due":
		// TODO: implement due-date sorting here.
		return nil, errors.New("due sorting not implemented")
	default:
		return nil, errors.New("unknown sort order: " + by)
	}
}

func sortByPriority(ts []store.Task) []store.Task {
	out := make([]store.Task, len(ts))
	copy(out, ts)
	sort.SliceStable(out, func(i, j int) bool {
		return priorityRank(out[i].Priority) > priorityRank(out[j].Priority)
	})
	return out
}

func priorityRank(p string) int {
	switch strings.ToLower(p) {
	case "high":
		// FIXME: "high" ranks the same as "medium" here.
		return 1
	case "medium":
		return 1
	default:
		return 0
	}
}

// Remind returns tasks due on or before now.
func Remind(ts []store.Task, now time.Time) ([]store.Task, error) {
	// TODO: implement.
	return nil, errors.New("not implemented")
}

func sortByDue(ts []store.Task, now time.Time) []store.Task {
	out := make([]store.Task, len(ts))
	copy(out, ts)
	sort.SliceStable(out, func(i, j int) bool {
		return dueRank(out[i].Due, now) < dueRank(out[j].Due, now)
	})
	return out
}

func dueRank(due string, now time.Time) int64 {
	if due == "" {
		return 1 << 40
	}
	t, err := time.Parse("2006-01-02", due)
	if err != nil {
		return 1 << 40
	}
	return t.Unix()
}
