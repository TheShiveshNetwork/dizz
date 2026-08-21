package plan

import (
	"errors"
	"sort"
	"strings"
	"time"

	"taskforge/internal/store"
)

// Sort orders tasks by the given key ("priority" or "due").
func Sort(ts []store.Task, by string) ([]store.Task, error) {
	switch by {
	case "priority":
		return sortByPriority(ts), nil
	case "due":
		return sortByDue(ts, time.Now()), nil
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
		return 2
	case "medium":
		return 1
	default:
		return 0
	}
}

// Remind returns tasks due on or before now.
func Remind(ts []store.Task, now time.Time) ([]store.Task, error) {
	var due []store.Task
	for _, t := range ts {
		d, err := time.Parse("2006-01-02", t.Due)
		if err != nil {
			continue
		}
		if !d.After(now) {
			due = append(due, t)
		}
	}
	return sortByDue(due, now), nil
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
