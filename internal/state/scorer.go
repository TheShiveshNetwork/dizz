package state

import (
	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// Scorer interprets signals into state
type Scorer struct {
	// Configuration for scoring rules
	highChurnThreshold int
	abandonedThreshold int
}

// NewScorer creates a new scorer with default thresholds
func NewScorer() *Scorer {
	return &Scorer{
		highChurnThreshold: 5,  // Files changed >5 times = unstable
		abandonedThreshold: 10, // Not touched in 10+ commits = abandoned
	}
}

// Score interprets a symbol's state from signals
func (s *Scorer) Score(symbol *Symbol) {
	// Priority rules (in order):
	// 1. Explicit intent markers override everything
	// 2. Todos indicate planned work
	// 3. High churn indicates instability
	// 4. Called + stable = active
	// 5. Not called + old = abandoned
	// 6. Not called + new = unused

	// Rule 1: Intent markers
	if symbol.IntentMarker != "" {
		switch symbol.IntentMarker {
		case "planned":
			symbol.State = Planned
			symbol.Confidence = 1.0 // Explicit intent = high confidence
			return
		case "active":
			symbol.State = Active
			symbol.Confidence = 1.0
			return
		case "deprecated":
			symbol.State = Abandoned
			symbol.Confidence = 1.0
			return
		}
	}

	// Rule 2: Todos
	if symbol.HasTodo {
		symbol.State = Planned
		symbol.Confidence = 0.7
		return
	}

	// Rule 3: High churn
	if symbol.ChurnCount > s.highChurnThreshold {
		symbol.State = Unstable
		symbol.Confidence = 0.6
		return
	}

	// Rule 4: Active use
	if symbol.IsCalled {
		symbol.State = Active
		symbol.Confidence = 0.9
		return
	}

	// Rule 5: Abandoned (not used, old code)
	if !symbol.IsCalled && symbol.ChurnCount > s.abandonedThreshold {
		symbol.State = Abandoned
		symbol.Confidence = 0.7
		return
	}

	// Rule 6: Default unused
	symbol.State = Unused
	symbol.Confidence = 0.6
}

// InterpretSignals converts a signal set into project state
func (s *Scorer) InterpretSignals(sigSet *signals.SignalSet) *ProjectState {
	ps := NewProjectState()

	// Build symbol index
	symbolIndex := make(map[string]*Symbol)

	// Process function definitions
	for _, sig := range sigSet.ByType(signals.FunctionDefined) {
		key := sig.File + "::" + sig.Name

		if _, exists := symbolIndex[key]; !exists {
			symbolIndex[key] = &Symbol{
				Name:       sig.Name,
				File:       sig.File,
				Line:       sig.Line,
				Column:     sig.Column,
				EndLine:    sig.EndLine,
				EndColumn:  sig.EndColumn,
				Type:       "function",
				Language:   sig.Language,
				IsDefined:  true,
				IsCalled:   false,
				HasTodo:    false,
				ChurnCount: 0,
			}
		}
	}

	// Process function calls
	for _, sig := range sigSet.ByType(signals.FunctionCalled) {
		// Match calls to definitions
		for key, symbol := range symbolIndex {
			if symbol.Name == sig.Name {
				symbolIndex[key].IsCalled = true
			}
		}
	}

	// Process Todos
	todosByFile := make(map[string][]signals.Signal)
	for _, sig := range sigSet.ByType(signals.TodoFound) {
		todosByFile[sig.File] = append(todosByFile[sig.File], sig)

		// Add to project state
		todo := Todo{
			File:     sig.File,
			Line:     sig.Line,
			Language: sig.Language,
			Resolved: false,
		}

		if text, ok := sig.Metadata["text"].(string); ok {
			todo.Text = text
		}
		if todoType, ok := sig.Metadata["type"].(string); ok {
			todo.Type = todoType
		} else {
			todo.Type = "TODO"
		}

		ps.AddTodo(todo)
	}

	// Mark symbols with Todos (function-level association)
	for key, symbol := range symbolIndex {
		if todos, exists := todosByFile[symbol.File]; exists {
			for _, todo := range todos {
				if todo.Line >= symbol.Line && todo.Line <= symbol.EndLine {
					symbolIndex[key].HasTodo = true
					break
				}
			}
		}
	}

	// Process intent markers (function-level association)
	for _, sig := range sigSet.ByType(signals.IntentMarker) {
		// Match to symbols in the same file and line range
		for key, symbol := range symbolIndex {
			if symbol.File == sig.File && sig.Line >= symbol.Line && sig.Line <= symbol.EndLine {
				if markerType, ok := sig.Metadata["marker_type"].(string); ok && markerType == "state" {
					if value, ok := sig.Metadata["value"].(string); ok {
						symbolIndex[key].IntentMarker = value
					}
				}
			}
		}
	}

	// Score all symbols
	for _, symbol := range symbolIndex {
		s.Score(symbol)
		ps.AddSymbol(*symbol)
	}

	return ps
}

// SuggestNextAction recommends what to work on
func SuggestNextAction(ps *ProjectState) string {
	// Priority: Planned > Unstable > Unused

	planned := ps.GetSymbolsByState(Planned)
	if len(planned) > 0 {
		return "Implement " + planned[0].Name + " (" + planned[0].File + ")"
	}

	unstable := ps.GetSymbolsByState(Unstable)
	if len(unstable) > 0 {
		return "Stabilize " + unstable[0].Name + " (" + unstable[0].File + ")"
	}

	unused := ps.GetSymbolsByState(Unused)
	if len(unused) > 0 {
		return "Connect or remove " + unused[0].Name + " (" + unused[0].File + ")"
	}

	abandoned := ps.GetSymbolsByState(Abandoned)
	if len(abandoned) > 0 {
		return "Review " + abandoned[0].Name + " for removal (" + abandoned[0].File + ")"
	}

	return "All symbols are active and stable"
}
