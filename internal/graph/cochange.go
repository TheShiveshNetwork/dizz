package graph

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// CommitInfo is a single commit with the files it touched.
type CommitInfo struct {
	Hash  string
	Files []string // absolute paths, resolved against the git root
}

// GitLogWithFiles runs `git log --name-only` and returns one CommitInfo per
// commit, newest first. Only files that exist in the working tree relative to
// the given projectRoot are resolved to absolute paths; everything else is
// dropped. This is the only git work the graph does, and it is opt-in (co-
// change queries only), so normal graph loads never pay for it.
func GitLogWithFiles(projectRoot string, maxCommits int) ([]CommitInfo, error) {
	args := []string{"log", "--name-only", "--format=%H", "--max-count=" + itoa(maxCommits)}
	cmd := exec.Command("git", args...)
	cmd.Dir = projectRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	gitRoot := projectRoot
	if rb, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		gitRoot = strings.TrimSpace(string(rb))
	}

	var commits []CommitInfo
	var current *CommitInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if current != nil && isFullHash(line) {
			commits = append(commits, *current)
			current = &CommitInfo{Hash: line}
			continue
		}
		if isFullHash(line) {
			current = &CommitInfo{Hash: line}
			continue
		}
		if current == nil {
			continue
		}
		// File name relative to the git root.
		f := filepath.Clean(filepath.Join(gitRoot, line))
		if strings.HasPrefix(f, projectRoot) {
			current.Files = append(current.Files, f)
		}
	}
	if current != nil {
		commits = append(commits, *current)
	}
	return commits, nil
}

func isFullHash(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

// CoChange is a hidden coupling between two files: they were modified in the
// same commits more often than chance would suggest, with no import relation.
type CoChange struct {
	FileA         string
	FileB         string
	CoOccurrences int
	CommitsA      int
	CommitsB      int
	Jaccard       float64
}

// ComputeCoChanges finds file pairs that co-change across the given commits.
// Only files present in relevantFiles are considered (files already in the
// graph), and only files touched by at least minCommits commits are treated as
// candidates to filter noise. Pairs must reach the given minJaccard threshold.
func ComputeCoChanges(commits []CommitInfo, relevantFiles map[string]bool, minJaccard float64, minCommits int) []CoChange {
	fileCommits := make(map[string]map[string]bool)
	for _, c := range commits {
		for _, f := range c.Files {
			if !relevantFiles[f] {
				continue
			}
			set := fileCommits[f]
			if set == nil {
				set = make(map[string]bool)
				fileCommits[f] = set
			}
			set[c.Hash] = true
		}
	}

	// Candidate files: touched in at least minCommits distinct commits.
	candidates := make([]string, 0, len(fileCommits))
	for f, set := range fileCommits {
		if len(set) >= minCommits {
			candidates = append(candidates, f)
		}
	}
	sort.Strings(candidates)

	var out []CoChange
	for i := 0; i < len(candidates); i++ {
		a := candidates[i]
		setA := fileCommits[a]
		for j := i + 1; j < len(candidates); j++ {
			b := candidates[j]
			setB := fileCommits[b]
			inter := 0
			for h := range setA {
				if setB[h] {
					inter++
				}
			}
			if inter == 0 {
				continue
			}
			union := len(setA) + len(setB) - inter
			jaccard := float64(inter) / float64(union)
			if jaccard >= minJaccard {
				out = append(out, CoChange{
					FileA:         a,
					FileB:         b,
					CoOccurrences: inter,
					CommitsA:      len(setA),
					CommitsB:      len(setB),
					Jaccard:       jaccard,
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Jaccard != out[j].Jaccard {
			return out[i].Jaccard > out[j].Jaccard
		}
		return out[i].FileA+out[i].FileB < out[j].FileA+out[j].FileB
	})
	return out
}
