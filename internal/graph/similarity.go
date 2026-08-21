package graph

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/state"
)

// stopwords are high-frequency function words that carry no topical meaning.
// They are dropped during tokenization so they cannot dominate the score.
var stopwords = map[string]bool{
	"a": true, "an": true, "the": true, "to": true, "of": true, "in": true,
	"on": true, "at": true, "for": true, "and": true, "or": true, "with": true,
	"from": true, "by": true, "we": true, "our": true, "us": true, "it": true,
	"its": true, "this": true, "that": true, "is": true, "are": true, "be": true,
	"as": true, "do": true, "does": true, "if": true, "then": true, "so": true,
	"will": true, "can": true, "could": true, "should": true, "have": true,
	"has": true, "had": true, "not": true, "no": true, "but": true,
}

// tokenize lowercases s, splits it on non-alphanumeric runs, drops stopwords
// and single-character tokens, and stems every remaining token. The returned
// slice preserves the original order (including duplicates).
func tokenize(s string) []string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if stopwords[f] || len(f) == 1 {
			continue
		}
		out = append(out, stem(f))
	}
	return out
}

// stem is a light, deterministic suffix stripper. It canonicalizes the most
// common English plural and verb inflections so that related words share one
// stem: tokens -> token, caching -> cach == cache, refactoring -> refactor,
// resolved -> resolv == resolve. It is deliberately conservative: uncommon or
// ambiguous suffixes are left untouched rather than risking false matches.
func stem(tok string) string {
	if len(tok) <= 3 {
		return tok
	}
	t := tok
	switch {
	case strings.HasSuffix(t, "ies") && len(t) > 4:
		t = t[:len(t)-3] + "y" // abilities -> abilit
	case strings.HasSuffix(t, "sses"):
		t = t[:len(t)-2] // classes -> class
	case strings.HasSuffix(t, "ss"), strings.HasSuffix(t, "us"), strings.HasSuffix(t, "is"):
	case strings.HasSuffix(t, "es"):
		t = t[:len(t)-2] // caches -> cache
	case strings.HasSuffix(t, "s"):
		t = t[:len(t)-1] // tokens -> token
	}
	switch {
	case strings.HasSuffix(t, "ing") && len(t) > 5:
		t = t[:len(t)-3] // caching -> cach, refactoring -> refactor
	case strings.HasSuffix(t, "ed") && len(t) > 4:
		t = t[:len(t)-2] // resolved -> resolv
	case strings.HasSuffix(t, "ly") && len(t) > 4:
		t = t[:len(t)-2] // quickly -> quick
	}
	if strings.HasSuffix(t, "e") && len(t) > 3 && !strings.HasSuffix(t, "ee") && !strings.HasSuffix(t, "ie") {
		t = t[:len(t)-1] // cache -> cach, make -> mak
	}
	return t
}

// Similarity returns the cosine similarity between the token vectors of a and
// b. Tokens are lowercased, stopword-filtered, and stemmed; all tokens are
// weighted equally (no corpus statistics). A score of 1 means identical token
// sets; 0 means none are shared. It returns 0 when either text has no tokens.
// @dizz-ignore-abandoned
func Similarity(a, b string) float64 {
	ta := tokenize(a)
	tb := tokenize(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	va := termVec(ta)
	vb := termVec(tb)
	return cosine(va, vb, vecNorm(va), vecNorm(vb))
}

type simVec struct {
	id    string
	terms map[string]float64
	norm  float64
}

type simMatch struct {
	idx int
	sim float64
}

// candidate pairs a node id with the text used for similarity. The text is the
// intent message plus its tags (repeated to double their weight). Scope and
// file paths are excluded: they are dominated by constant values ("project")
// and generic path tokens ("pkg", "go") that only add noise.
type candidate struct {
	id   string
	text string
}

func candidateText(msg string, tags []string) string {
	var b strings.Builder
	b.WriteString(msg)
	for _, tag := range tags {
		b.WriteByte(' ')
		b.WriteString(tag)
		b.WriteByte(' ')
		b.WriteString(tag)
	}
	return b.String()
}

func buildCandidates(projectRoot string, is *state.IntentState, todos []state.Todo) []candidate {
	cands := make([]candidate, 0, len(is.Intents)+len(todos))
	for _, intent := range is.Intents {
		cands = append(cands, candidate{
			id:   IntentID(intent.ID),
			text: candidateText(intent.Message, intent.Tags),
		})
	}
	for _, todo := range todos {
		if todo.Resolved {
			continue
		}
		rel := relPathOf(projectRoot, todo.File)
		cands = append(cands, candidate{
			id:   TodoID(rel, todo.Line),
			text: candidateText(todo.Text, nil),
		})
	}
	return cands
}

func termVec(tokens []string) map[string]float64 {
	out := make(map[string]float64, len(tokens))
	for _, t := range tokens {
		out[t]++
	}
	return out
}

func vecNorm(v map[string]float64) float64 {
	s := 0.0
	for _, w := range v {
		s += w * w
	}
	return math.Sqrt(s)
}

func cosine(va, vb map[string]float64, normA, normB float64) float64 {
	if len(va) > len(vb) {
		va, vb = vb, va
	}
	dot := 0.0
	for t, wa := range va {
		if wb, ok := vb[t]; ok {
			dot += wa * wb
		}
	}
	if dot == 0 {
		return 0
	}
	denom := normA * normB
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// idfVectors builds one term-weight vector per candidate. Each term is weighted
// tf * idf where tf is the term count in the candidate and idf = ln(1 + N/df)
// with df the number of candidates containing the term. Norms are precomputed
// so pairwise cosine is a single division.
func idfVectors(cands []candidate) []simVec {
	n := len(cands)
	df := make(map[string]int)
	tokens := make([][]string, n)
	for i, c := range cands {
		toks := tokenize(c.text)
		tokens[i] = toks
		seen := make(map[string]bool)
		for _, t := range toks {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}
	vecs := make([]simVec, n)
	for i := range cands {
		freq := make(map[string]int)
		for _, t := range tokens[i] {
			freq[t]++
		}
		v := simVec{id: cands[i].id, terms: make(map[string]float64, len(freq))}
		for t, c := range freq {
			idf := math.Log(1 + float64(n)/float64(df[t]))
			w := float64(c) * idf
			v.terms[t] = w
			v.norm += w * w
		}
		v.norm = math.Sqrt(v.norm)
		vecs[i] = v
	}
	return vecs
}

// buildPostings builds an inverted index term -> ascending doc indices. Only
// candidates that share at least one term are ever scored, which keeps the
// practical cost near-linear for sparse vocabularies (see ADR-0001).
func buildPostings(vecs []simVec) map[string][]int {
	p := make(map[string][]int)
	for i := range vecs {
		for t := range vecs[i].terms {
			p[t] = append(p[t], i)
		}
	}
	return p
}

// topMatches returns the topK candidates most similar to vecs[i] with
// similarity >= threshold, in descending order (tie-broken by node id).
func topMatches(vecs []simVec, postings map[string][]int, i int, threshold float64, topK int) []simMatch {
	terms := make([]string, 0, len(vecs[i].terms))
	for t := range vecs[i].terms {
		terms = append(terms, t)
	}
	sort.Strings(terms)

	acc := make(map[int]float64)
	for _, t := range terms {
		wi := vecs[i].terms[t]
		for _, j := range postings[t] {
			if j == i {
				continue
			}
			acc[j] += wi * vecs[j].terms[t]
		}
	}

	matches := make([]simMatch, 0, len(acc))
	for j, dot := range acc {
		denom := vecs[i].norm * vecs[j].norm
		if denom == 0 {
			continue
		}
		sim := dot / denom
		if sim >= threshold {
			matches = append(matches, simMatch{idx: j, sim: sim})
		}
	}
	sort.Slice(matches, func(a, b int) bool {
		if matches[a].sim != matches[b].sim {
			return matches[a].sim > matches[b].sim
		}
		return vecs[matches[a].idx].id < vecs[matches[b].idx].id
	})
	if len(matches) > topK {
		matches = matches[:topK]
	}
	return matches
}

// addIntentSimilarityEdges links every intent (active or resolved) to the most
// similar other intents and unresolved todos via RELATED_TO edges whose weight
// is the cosine similarity of the IDF-weighted token vectors. Intents are
// project-level entities, so they are only matched against other intents and
// todos, never symbols or files.
func addIntentSimilarityEdges(g *Graph, is *state.IntentState, todos []state.Todo, threshold float64, topK int) {
	if is == nil || threshold <= 0 || topK <= 0 {
		return
	}
	cands := buildCandidates(g.ProjectRoot, is, todos)
	if len(cands) < 2 {
		return
	}
	vecs := idfVectors(cands)
	postings := buildPostings(vecs)
	for i := range vecs {
		for _, m := range topMatches(vecs, postings, i, threshold, topK) {
			sim := m.sim
			g.AddEdge(&Edge{
				Type:   EdgeRelatedTo,
				From:   vecs[i].id,
				To:     vecs[m.idx].id,
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
