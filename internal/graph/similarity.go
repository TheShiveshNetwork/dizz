package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/state"
)

// Similarity returns the overlap coefficient between two texts: the size of
// the shared lowercased alphanumeric token set divided by the size of the
// smaller token set. It returns 0 when either text has no tokens. A score of
// 1 means one text's tokens are a subset of the other's.
func Similarity(a, b string) float64 {
	sa := tokenSet(a)
	sb := tokenSet(b)
	if len(sa) == 0 || len(sb) == 0 {
		return 0
	}
	inter := 0
	for t := range sa {
		if sb[t] {
			inter++
		}
	}
	minSize := len(sa)
	if len(sb) < minSize {
		minSize = len(sb)
	}
	return float64(inter) / float64(minSize)
}

func tokenSet(s string) map[string]bool {
	out := make(map[string]bool)
	for _, f := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}) {
		out[f] = true
	}
	return out
}

// addIntentSimilarityEdges links every intent (active or resolved) to the most
// similar other intents and unresolved todos via RELATED_TO edges whose weight
// is the text similarity score. Intents are project-level entities, so they
// are only matched against other intents and todos, never symbols or files.
func addIntentSimilarityEdges(g *Graph, is *state.IntentState, todos []state.Todo, threshold float64, topK int) {
	if is == nil || threshold <= 0 || topK <= 0 {
		return
	}
	type candidate struct {
		id   string
		text string
	}
	cands := make([]candidate, 0, len(is.Intents)+len(todos))
	for _, intent := range is.Intents {
		cands = append(cands, candidate{id: IntentID(intent.ID), text: intent.Message})
	}
	for _, todo := range todos {
		if todo.Resolved {
			continue
		}
		rel := relPathOf(g.ProjectRoot, todo.File)
		cands = append(cands, candidate{id: TodoID(rel, todo.Line), text: todo.Text})
	}
	if len(cands) < 2 {
		return
	}
	type scored struct {
		idx int
		sim float64
	}
	for i := range cands {
		matches := make([]scored, 0, len(cands)-1)
		for j := range cands {
			if i == j {
				continue
			}
			sim := Similarity(cands[i].text, cands[j].text)
			if sim >= threshold {
				matches = append(matches, scored{idx: j, sim: sim})
			}
		}
		sort.Slice(matches, func(a, b int) bool {
			if matches[a].sim != matches[b].sim {
				return matches[a].sim > matches[b].sim
			}
			return cands[matches[a].idx].id < cands[matches[b].idx].id
		})
		if len(matches) > topK {
			matches = matches[:topK]
		}
		for _, m := range matches {
			sim := m.sim
			g.AddEdge(&Edge{
				Type:   EdgeRelatedTo,
				From:   cands[i].id,
				To:     cands[m.idx].id,
				Weight: sim,
				Attrs: map[string]string{
					"similarity": fmt.Sprintf("%.3f", sim),
				},
				Rationale: Rationale{
					Confidence: sim,
					Evidence:   fmt.Sprintf("text similarity %.3f", sim),
					SourceType: "similarity",
				},
			})
		}
	}
}
