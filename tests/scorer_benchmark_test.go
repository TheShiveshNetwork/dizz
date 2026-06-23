package benchmarks

import (
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

// ──────────────────────────────────────────────────────────────────────────────
// Scorer benchmarks — measure interpretation performance
// ──────────────────────────────────────────────────────────────────────────────

func BenchmarkScorerInterpretation(b *testing.B) {
	sigSet := buildBenchmarkSignalSet(500, 2000)
	scorer := state.NewScorer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps := scorer.InterpretSignalsWithIntent(sigSet, nil)
		if len(ps.Symbols) == 0 {
			b.Fatal("no symbols interpreted")
		}
	}
}

func BenchmarkScorerWithIntentState(b *testing.B) {
	sigSet := buildBenchmarkSignalSet(500, 2000)
	scorer := state.NewScorer()
	intentState := buildBenchmarkIntentState(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps := scorer.InterpretSignalsWithIntent(sigSet, intentState)
		if len(ps.Symbols) == 0 {
			b.Fatal("no symbols interpreted")
		}
	}
}

func BenchmarkScorerSummary(b *testing.B) {
	sigSet := buildBenchmarkSignalSet(500, 2000)
	scorer := state.NewScorer()
	ps := scorer.InterpretSignalsWithIntent(sigSet, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ps.GetSummary()
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Symbol state benchmarks
// ──────────────────────────────────────────────────────────────────────────────

func BenchmarkGetSymbolsByState(b *testing.B) {
	sigSet := buildBenchmarkSignalSet(500, 2000)
	scorer := state.NewScorer()
	ps := scorer.InterpretSignalsWithIntent(sigSet, nil)

	sizes := []int{10, 100, 500}
	for _, n := range sizes {
		subset := make([]state.Symbol, n)
		for i := 0; i < n && i < len(ps.Symbols); i++ {
			subset[i] = ps.Symbols[i]
		}
		ps2 := state.NewProjectState()
		for _, s := range subset {
			ps2.AddSymbol(s)
		}

		b.Run("active_"+itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ps2.GetSymbolsByState(state.Active)
			}
		})
		b.Run("unused_"+itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ps2.GetSymbolsByState(state.Unused)
			}
		})
	}
}

func BenchmarkSuggestNextAction(b *testing.B) {
	sigSet := buildBenchmarkSignalSet(500, 2000)
	scorer := state.NewScorer()
	ps := scorer.InterpretSignalsWithIntent(sigSet, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.SuggestNextAction(ps)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Intent state benchmarks
// ──────────────────────────────────────────────────────────────────────────────

func BenchmarkIntentStateOperations(b *testing.B) {
	intentState := buildBenchmarkIntentState(100)

	b.Run("GetActiveIntents", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = intentState.GetActiveIntents()
		}
	})

	b.Run("GetIntentsByType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = intentState.GetIntentsByType(state.IntentTodo)
		}
	})

	b.Run("GetIntentSummary", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = intentState.GetIntentSummary()
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// itoa is a minimal int→string for benchmark labels.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// buildBenchmarkSignalSet creates a signal set with nDefs definitions and
// nCalls call signals for benchmarking.
func buildBenchmarkSignalSet(nDefs, nCalls int) *signals.SignalSet {
	ss := &signals.SignalSet{}
	for i := 0; i < nDefs; i++ {
		sig := signals.NewSignal(signals.FunctionDefined, "/bench/file.go").
			WithName("func_"+itoa(i)).
			WithLine(10+i*5).
			WithEndLine(15+i*5).
			WithLanguage("go").
			WithMeta("source_tier", "ast").
			WithConfidence(1.0)
		ss.Add(*sig)
	}
	for i := 0; i < nCalls && i < nDefs; i++ {
		sig := signals.NewSignal(signals.FunctionCalled, "/bench/file.go").
			WithName("func_" + itoa(i)).
			WithLine(100 + i).
			WithLanguage("go")
		ss.Add(*sig)
	}
	for i := 0; i < nDefs/10; i++ {
		sig := signals.NewSignal(signals.TodoFound, "/bench/file.go").
			WithLine(200+i).
			WithLanguage("go").
			WithMeta("type", "TODO").
			WithMeta("text", "benchmark todo "+itoa(i))
		ss.Add(*sig)
	}
	return ss
}

// buildBenchmarkIntentState creates an intent state with n intents.
func buildBenchmarkIntentState(n int) *state.IntentState {
	is := state.NewIntentState()
	types := []state.IntentType{state.IntentTodo, state.Refactor, state.Fixme, state.Question, state.Hack}
	for i := 0; i < n; i++ {
		intent := state.Intent{
			ID:         "bench_intent_" + itoa(i),
			Type:       types[i%len(types)],
			Message:    "benchmark intent " + itoa(i),
			Scope:      "/bench/file.go",
			Severity:   i % 4,
			Status:     state.IntentActive,
			Confidence: 0.8,
		}
		is.AddIntent(intent)
	}
	return is
}
