package state

// ScoreFunction determines the state and confidence of a function based on signals
// This is intentionally simple to start - can be enhanced later
func ScoreFunction(isCalled bool, hasTodo bool) (FunctionState, float64) {
	// Priority: TODO > Called > Declared
	
	if hasTodo {
		// Function has a TODO - it's planned work
		return Planned, 0.3
	}
	
	if isCalled {
		// Function is called somewhere - it's in use
		return Used, 0.9
	}
	
	// Function exists but isn't called and has no TODO
	// Could be dead code or just not connected yet
	return Unused, 0.6
}

// SuggestNextAction recommends what to work on based on state
func SuggestNextAction(functions []Function) string {
	// Count states
	plannedCount := 0
	unusedCount := 0
	
	var firstPlanned, firstUnused *Function
	
	for i := range functions {
		fn := &functions[i]
		switch fn.State {
		case Planned:
			plannedCount++
			if firstPlanned == nil {
				firstPlanned = fn
			}
		case Unused:
			unusedCount++
			if firstUnused == nil {
				firstUnused = fn
			}
		}
	}
	
	// Priority: Implement planned work first
	if plannedCount > 0 && firstPlanned != nil {
		return "Implement " + firstPlanned.Name
	}
	
	// Then connect unused functions
	if unusedCount > 0 && firstUnused != nil {
		return "Connect or remove " + firstUnused.Name
	}
	
	return "All functions are in use"
}
