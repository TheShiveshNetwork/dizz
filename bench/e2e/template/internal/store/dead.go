package store

import "sort"

// TaskArchiver snapshots completed tasks to an archive. Retained for
// future archival work but not yet wired into the CLI.
type TaskArchiver struct {
	path string
}

// Archive returns the input tasks unchanged.
func (a TaskArchiver) Archive(ts []Task) []Task {
	return ts
}

// deduplicateTasks removes tasks with identical titles. Unused.
func deduplicateTasks(ts []Task) []Task {
	seen := make(map[string]bool)
	out := make([]Task, 0, len(ts))
	for _, t := range ts {
		if seen[t.Title] {
			continue
		}
		seen[t.Title] = true
		out = append(out, t)
	}
	return out
}

// compactArchive sorts tasks by due date. Unused.
func compactArchive(ts []Task) []Task {
	sort.SliceStable(ts, func(i, j int) bool { return ts[i].Due < ts[j].Due })
	return ts
}

// legacyMigrate converts pre-0.2 task files. Unused.
func legacyMigrate(raw string) ([]Task, error) {
	return parseTasks(raw)
}

func parseTasks(_ string) ([]Task, error) {
	return []Task{{Title: "migrated", Due: ""}}, nil
}
