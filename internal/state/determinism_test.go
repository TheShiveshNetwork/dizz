package state

import (
	"testing"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// TestInstabilityScoreSerializedLosslessly guards against the regression where
// InstabilityScore was truncated to 4 decimals on disk. Near-equal git-derived
// scores (e.g. 2.99982 / 2.99985 / 2.99988) collapsed to a single value, which
// flipped the 90th-percentile Unstable classification between runs. The score
// must survive a TON marshal/unmarshal round-trip with full precision.
func TestInstabilityScoreSerializedLosslessly(t *testing.T) {
	ps := NewProjectState()
	now := time.Now()
	values := []float64{2.99982, 2.99985, 2.99988}
	for i, v := range values {
		ps.AddSymbol(Symbol{
			Name:             "fn" + string(rune('A'+i)),
			File:             "/proj/main.go",
			Line:             10 + i*10,
			EndLine:          12 + i*10,
			Type:             "function",
			Language:         "go",
			State:            Unstable,
			InstabilityScore: v,
			LastTouched:      &now,
		})
	}

	data, err := ps.MarshalTON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	round, err := UnmarshalProjectStateTON(data)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(round.Symbols) != len(values) {
		t.Fatalf("symbol count mismatch: got %d want %d", len(round.Symbols), len(values))
	}
	for i, want := range values {
		got := round.Symbols[i].InstabilityScore
		if got != want {
			t.Fatalf("InstabilityScore precision lost: symbol %d got %v want %v", i, got, want)
		}
	}
}

// buildThreeSymbolSignals returns a signal set with a called helper, an uncalled
// helper, and an uncalled entry point (main) — the shape from issue #30.
func buildThreeSymbolSignals() *signals.SignalSet {
	ss := &signals.SignalSet{}
	add := func(name string, line, end int, called bool) {
		def := signals.NewSignal(signals.FunctionDefined, "/proj/main.go").
			WithName(name).WithRange(line, 1, end, 2).WithLanguage("go")
		ss.Add(*def)
		if called {
			call := signals.NewSignal(signals.FunctionCalled, "/proj/main.go").
				WithName(name)
			ss.Add(*call)
		}
	}
	add("main", 5, 7, false)
	add("usedHelper", 9, 11, true)
	add("unusedHelper", 13, 15, false)
	return ss
}

// TestClassificationDeterministicAcrossRuns guards against the run-order
// dependency: LastTouched is now supplied via gitMeta before classification
// (not carried from a previous run's post-score save), so an uncalled, old
// symbol is classified identically whether or not a previous state exists.
func TestClassificationDeterministicAcrossRuns(t *testing.T) {
	ss := buildThreeSymbolSignals()

	old := time.Now().Add(-200 * 24 * time.Hour)
	gitMeta := map[string]*GitMeta{
		"/proj/main.go::main":         {LastTouched: &old, ChurnCount: 1},
		"/proj/main.go::unusedHelper": {LastTouched: &old, ChurnCount: 1},
		"/proj/main.go::usedHelper":   {LastTouched: &old, ChurnCount: 5},
	}

	scorer := NewScorer()
	first := scorer.InterpretSignalsWithIntent(ss, nil, nil, gitMeta)

	byName := map[string]Symbol{}
	for _, s := range first.Symbols {
		byName[s.Name] = s
	}
	if byName["unusedHelper"].State != Abandoned {
		t.Fatalf("first run: want unusedHelper=abandoned, got %q", byName["unusedHelper"].State)
	}
	if byName["usedHelper"].State != Active {
		t.Fatalf("first run: want usedHelper=active, got %q", byName["usedHelper"].State)
	}

	// Second run with the previous state present (simulating a later command).
	second := scorer.InterpretSignalsWithIntent(ss, nil, first, gitMeta)

	for _, name := range []string{"main", "usedHelper", "unusedHelper"} {
		if byName[name].State != secondSymbolState(second, name) {
			t.Fatalf("classification differs between runs for %s: first=%q second=%q",
				name, byName[name].State, secondSymbolState(second, name))
		}
	}
}

func secondSymbolState(ps *ProjectState, name string) SymbolState {
	for _, s := range ps.Symbols {
		if s.Name == name {
			return s.State
		}
	}
	return ""
}
