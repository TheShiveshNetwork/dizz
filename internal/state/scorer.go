package state

import (
	"math"
	"sort"
	"time"

	"github.com/TheShiveshNetwork/dizz/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// Scorer interprets signals into state
type Scorer struct {
	// Mathematical instability scoring parameters
	tau                 float64 // Time decay constant (in days)
	percentileThreshold float64 // Percentile threshold for instability (0-100)
}

// NewScorer creates a new scorer with mathematical instability scoring
func NewScorer() *Scorer {
	return &Scorer{
		tau:                 30.0, // 30 days time decay constant
		percentileThreshold: 90.0, // 90th percentile for instability
	}
}

// calculateIntentWeight calculates intent weighting for enhanced scoring
func (s *Scorer) calculateIntentWeight(symbol *Symbol, intentState *IntentState) float64 {
	if intentState == nil {
		return 1.0
	}

	intentWeight := 1.0
	now := time.Now()

	for _, intent := range intentState.GetActiveIntents() {
		// Calculate proximity based on scope matching
		proximity := s.calculateProximity(symbol, intent)
		if proximity > 0 {
			// Apply severity and proximity weighting
			severityWeight := float64(intent.Severity) / 3.0                              // Normalize to 0-1
			recencyWeight := math.Exp(-now.Sub(intent.UpdatedAt).Hours() / (24.0 * 30.0)) // 30-day decay

			intentWeight += severityWeight * proximity * recencyWeight
		}
	}

	return intentWeight
}

// calculateProximity measures how close an intent is to a symbol's code scope
func (s *Scorer) calculateProximity(symbol *Symbol, intent Intent) float64 {
	// Direct file match
	if symbol.File == intent.Scope {
		return 1.0
	}

	// Parse scope format (file:line-line) and check if symbol is within range
	if len(intent.Scope) > 0 {
		// Simple file matching for now - could be enhanced with line range parsing
		if symbol.File == intent.Scope {
			return 1.0
		}
	}

	return 0.0
}

// Enhanced scoring: score_enhanced(f) = score(f) × intent_weight
func (s *Scorer) calculateEnhancedScore(symbol *Symbol, intentState *IntentState) float64 {
	baseScore := symbol.InstabilityScore
	intentWeight := s.calculateIntentWeight(symbol, intentState)
	return baseScore * intentWeight
}

// calculateInstabilityScore computes the mathematical instability score for a symbol
func (s *Scorer) calculateInstabilityScore(symbol *Symbol) float64 {
	// Get detailed commit history for the function
	commits, err := integrations.GetFunctionCommits(symbol.File, symbol.Line, symbol.EndLine, 50)
	if err != nil || len(commits) == 0 {
		return 0.0
	}

	now := time.Now()

	// 1. Group commits into activity windows (24h)
	// This prevents penalizing multiple small commits for a single logical change (burst activity)
	type activityEvent struct {
		time       time.Time
		changeSize int
	}

	events := make([]activityEvent, 0, len(commits))
	if len(commits) > 0 {
		// git log returns newest first, so commits[0] is the latest
		currentEvent := activityEvent{time: commits[0].Time, changeSize: commits[0].ChangeSize}
		for i := 1; i < len(commits); i++ {
			// If within 24 hours of the current event, group them
			if currentEvent.time.Sub(commits[i].Time).Hours() < 24.0 {
				currentEvent.changeSize += commits[i].ChangeSize
			} else {
				events = append(events, currentEvent)
				currentEvent = activityEvent{time: commits[i].Time, changeSize: commits[i].ChangeSize}
			}
		}
		events = append(events, currentEvent)
	}

	// 2. Calculate raw instability score with exponential decay
	var score float64
	for _, event := range events {
		// Calculate time difference in days
		deltaT := now.Sub(event.time).Hours() / 24.0

		// Apply exponential decay: change_size_i * exp(-Δt_i / τ)
		decay := math.Exp(-deltaT / s.tau)
		contribution := float64(event.changeSize) * decay
		score += contribution
	}

	// 3. Apply "Stability through Silence" (TSLC - Time Since Last Change)
	// If a function hasn't been touched in a long time, it's likely "settled" and correct.
	// We aggressively discount the score if it has been untouched for more than 1x tau.
	if len(commits) > 0 {
		lastChangeAge := now.Sub(commits[0].Time).Hours() / 24.0

		if lastChangeAge > s.tau {
			// Multiplier decreases exponentially after tau.
			// At 1x tau, multiplier is 1.0
			// At 2x tau, multiplier is ~0.36
			// At 3x tau, multiplier is ~0.13
			silenceMultiplier := math.Exp(-(lastChangeAge - s.tau) / s.tau)
			score *= silenceMultiplier
		}
	}

	return score
}

// percentile computes the given percentile (0-100) from a sorted slice.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	n := float64(len(sorted))
	rank := (p / 100.0) * (n - 1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))
	if lower == upper {
		return sorted[lower]
	}
	weight := rank - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// Score interprets a symbol's state from signals
func (s *Scorer) Score(symbol *Symbol) {
	// Priority rules (in order):
	// 1. Explicit intent markers override everything
	// 2. Todos indicate planned work
	// 3. Calculate instability score (mathematical scoring determines final state)

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

	// Rule 3: Calculate and store instability score for mathematical analysis
	instabilityScore := s.calculateInstabilityScore(symbol)
	symbol.InstabilityScore = instabilityScore

	// Don't set state here - mathematical scoring will determine final state
	// This ensures instability detection applies to ALL symbols (including called ones)
}

// applyMathematicalScoring applies mathematical scoring for instability, abandonment, and unused states
func (s *Scorer) applyMathematicalScoring(symbols []*Symbol) {
	if len(symbols) == 0 {
		return
	}

	// 1. INSTABILITY: Collect instability scores for all symbols
	instabilityScores := make([]float64, 0, len(symbols))
	for _, symbol := range symbols {
		if symbol.InstabilityScore > 0 {
			instabilityScores = append(instabilityScores, symbol.InstabilityScore)
		}
	}

	// 2. ABANDONMENT: Calculate time-since-last-modified for unused symbols
	ages := make([]float64, 0, len(symbols))
	for _, symbol := range symbols {
		if !symbol.IsCalled && symbol.IntentMarker == "" && !symbol.HasTodo {
			if symbol.LastTouched != nil {
				age := time.Since(*symbol.LastTouched).Hours() / 24.0
				ages = append(ages, age)
			}
		}
	}

	// Compute only the percentiles we actually use (75th and 90th)
	instabilityThreshold := 0.0
	ageThreshold := 0.0

	if len(instabilityScores) > 0 {
		sortableInstability := make([]float64, len(instabilityScores))
		copy(sortableInstability, instabilityScores)
		sort.Float64s(sortableInstability)
		instabilityThreshold = percentile(sortableInstability, 90.0)
	}

	if len(ages) > 0 {
		sortableAges := make([]float64, len(ages))
		copy(sortableAges, ages)
		sort.Float64s(sortableAges)
		ageThreshold = percentile(sortableAges, 75.0)
	}

	// Apply mathematical scoring to ALL symbols
	for _, symbol := range symbols {
		// Skip if already classified by explicit intent markers or todos
		if symbol.State != "" && (symbol.IntentMarker != "" || symbol.HasTodo) {
			continue
		}

		// HasTodo symbols skipped by carry-forward need Planned set here.
		if symbol.HasTodo {
			symbol.State = Planned
			symbol.Confidence = 0.7
			continue
		}

		// Same carry-forward issue for intent markers.
		if symbol.IntentMarker != "" {
			switch symbol.IntentMarker {
			case "planned":
				symbol.State = Planned
				symbol.Confidence = 1.0
			case "active":
				symbol.State = Active
				symbol.Confidence = 1.0
			case "deprecated":
				symbol.State = Abandoned
				symbol.Confidence = 1.0
			}
			continue
		}

		// UNSTABLE: High instability score (90th percentile) - APPLIES TO ALL SYMBOLS
		if instabilityThreshold > 0 && symbol.InstabilityScore >= instabilityThreshold {
			symbol.State = Unstable
			symbol.Confidence = 0.8 // High confidence for mathematical instability
			continue
		}

		// ABANDONED: Old and never called (only for uncalled symbols)
		if !symbol.IsCalled && symbol.LastTouched != nil && ageThreshold > 0 {
			age := time.Since(*symbol.LastTouched).Hours() / 24.0
			if age >= ageThreshold {
				symbol.State = Abandoned
				symbol.Confidence = s.abandonedConfidence(symbol)
				continue
			}
		}

		// UNUSED: Not called but not old enough to be abandoned.
		// For symbols extracted with low-accuracy regex (TierC), call detection
		// is imperfect.  We still report the symbol as unused but reduce the
		// confidence so downstream consumers and users understand the limitation.
		if !symbol.IsCalled {
			symbol.State = Unused
			symbol.Confidence = s.unusedConfidence(symbol)
			continue
		}

		// ACTIVE: Called and not unstable (default for remaining symbols)
		if symbol.IsCalled {
			symbol.State = Active
			symbol.Confidence = 0.9
		}
	}
}

// unusedConfidence returns an appropriate confidence for the Unused state based
// on how reliably call-sites were detected for this symbol's language.
func (s *Scorer) unusedConfidence(sym *Symbol) float64 {
	switch sym.SignalSource {
	case "ast":
		return 0.9 // AST-backed call detection — very reliable
	case "lexical":
		return 0.65 // Structural regex — reasonably reliable
	case "regex":
		return 0.4 // Low-accuracy fallback — treat with scepticism
	default:
		return 0.6 // Legacy / unknown source
	}
}

// abandonedConfidence follows the same tier-based logic as unusedConfidence.
func (s *Scorer) abandonedConfidence(sym *Symbol) float64 {
	switch sym.SignalSource {
	case "ast":
		return 0.85
	case "lexical":
		return 0.6
	case "regex":
		return 0.35
	default:
		return 0.7
	}
}

// InterpretSignalsWithIntent converts signals with intent enhancement.
// When prevState is non-nil, InstabilityScore is carried forward for symbols
// whose name and location are unchanged, avoiding redundant git calls.
func (s *Scorer) InterpretSignalsWithIntent(sigSet *signals.SignalSet, intentState *IntentState, prevState *ProjectState) *ProjectState {
	ps := NewProjectState()

	// Build previous symbol index for git-data carry-forward
	prevIndex := make(map[string]*Symbol)
	if prevState != nil {
		for i := range prevState.Symbols {
			sym := &prevState.Symbols[i]
			prevIndex[sym.File+"::"+sym.Name] = sym
		}
	}

	// Build symbol index
	symbolIndex := make(map[string]*Symbol)

	// Process function definitions
	for _, sig := range sigSet.ByType(signals.FunctionDefined) {
		key := sig.File + "::" + sig.Name

		if _, exists := symbolIndex[key]; !exists {
			var source string
			if src, ok := sig.Metadata["source_tier"].(string); ok && src != "" {
				source = src
			} else {
				// Derive default from language: Go has AST analyzer, others use regex fallback
				if sig.Language == "go" {
					source = "ast"
				} else {
					source = "regex"
				}
			}
			symbolIndex[key] = &Symbol{
				Name:         sig.Name,
				File:         sig.File,
				Line:         sig.Line,
				Column:       sig.Column,
				EndLine:      sig.EndLine,
				EndColumn:    sig.EndColumn,
				Type:         "function",
				Language:     sig.Language,
				IsDefined:    true,
				IsCalled:     false,
				HasTodo:      false,
				ChurnCount:   0,
				SignalSource: source,
			}
		}
	}

	// Build indexes for O(1) lookups
	nameIndex := make(map[string][]*Symbol)
	fileIndex := make(map[string][]*Symbol)
	for _, symbol := range symbolIndex {
		nameIndex[symbol.Name] = append(nameIndex[symbol.Name], symbol)
		fileIndex[symbol.File] = append(fileIndex[symbol.File], symbol)
	}

	// Process function calls
	for _, sig := range sigSet.ByType(signals.FunctionCalled) {
		if symbols, ok := nameIndex[sig.Name]; ok {
			for _, symbol := range symbols {
				symbol.IsCalled = true
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
	for file, todos := range todosByFile {
		if symbols, ok := fileIndex[file]; ok {
			for _, symbol := range symbols {
				for _, todo := range todos {
					if todo.Line >= symbol.Line && todo.Line <= symbol.EndLine {
						symbol.HasTodo = true
						break
					}
				}
			}
		}
	}

	// Match intent markers to symbols in the same file and line range
	for _, sig := range sigSet.ByType(signals.IntentMarker) {
		if symbols, ok := fileIndex[sig.File]; ok {
			for _, symbol := range symbols {
				if sig.Line >= symbol.Line && sig.Line <= symbol.EndLine {
					if markerType, ok := sig.Metadata["marker_type"].(string); ok && markerType == "state" {
						if value, ok := sig.Metadata["value"].(string); ok {
							symbol.IntentMarker = value
						}
					}
				}
			}
		}
	}

	// Process intent ignore markers
	ignoreSignals := sigSet.ByType(signals.IgnoreFlag)
	for _, sig := range ignoreSignals {
		if symbols, ok := fileIndex[sig.File]; ok {
			for _, symbol := range symbols {
				matched := false
				if symbolName, ok := sig.Metadata["symbol_name"].(string); ok && symbolName == symbol.Name {
					matched = true
				} else if sig.Line >= symbol.Line && sig.Line <= symbol.EndLine {
					matched = true
				}
				if matched {
					if ignoreType, ok := sig.Metadata["ignore_type"].(string); ok {
						switch ignoreType {
						case "unstable", "unused", "abandoned":
							symbol.State = Active
							symbol.Confidence = 1.0
							symbol.IntentMarker = "ignored"
						}
					}
					break
				}
			}
		}
	}

	// Score all symbols (basic scoring), carrying forward previous InstabilityScore
	// for unchanged symbols to avoid redundant git calls.
	for _, symbol := range symbolIndex {
		if prev, ok := prevIndex[symbol.File+"::"+symbol.Name]; ok &&
			prev.Line == symbol.Line && prev.EndLine == symbol.EndLine {
			symbol.InstabilityScore = prev.InstabilityScore
			symbol.ChurnCount = prev.ChurnCount
			symbol.LastTouched = prev.LastTouched
		} else {
			s.Score(symbol)
		}
	}

	// Apply mathematical scoring (percentile-based)
	symbolSlice := make([]*Symbol, 0, len(symbolIndex))
	for _, symbol := range symbolIndex {
		symbolSlice = append(symbolSlice, symbol)
	}

	// Apply enhanced scoring with intent weighting
	for _, symbol := range symbolSlice {
		enhancedScore := s.calculateEnhancedScore(symbol, intentState)
		// Note: We could add a new field but will reuse InstabilityScore for now
		if intentState != nil {
			symbol.InstabilityScore = enhancedScore
		}
	}

	s.applyMathematicalScoring(symbolSlice)

	// Add all symbols to project state
	for _, symbol := range symbolIndex {
		ps.AddSymbol(*symbol)
	}

	return ps
}

// SuggestNextAction recommends what to work on based on Priority: Planned > Unstable > Unused
func SuggestNextAction(ps *ProjectState) string {
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

	return "Nothing to do yet. All symbols are active and stable"
}
