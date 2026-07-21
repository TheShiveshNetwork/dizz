package state

import (
	"time"
)

// SnapshotDelta represents the difference between two project states.
// Used for --diff mode to store only changes instead of full snapshots.
type SnapshotDelta struct {
	PrevHash     string    `json:"prev_hash"`
	IsCheckpoint bool      `json:"is_checkpoint"`
	Sequence     int       `json:"sequence"`
	CreatedAt    time.Time `json:"created_at"`

	SymbolsAdded   []Symbol `json:"symbols_added,omitempty"`
	SymbolsRemoved []Symbol `json:"symbols_removed,omitempty"`
	SymbolsChanged []Symbol `json:"symbols_changed,omitempty"`

	TodosAdded   []Todo `json:"todos_added,omitempty"`
	TodosRemoved []Todo `json:"todos_removed,omitempty"`
}

func (d *SnapshotDelta) IsEmpty() bool {
	return len(d.SymbolsAdded) == 0 &&
		len(d.SymbolsRemoved) == 0 &&
		len(d.SymbolsChanged) == 0 &&
		len(d.TodosAdded) == 0 &&
		len(d.TodosRemoved) == 0
}

type symKey struct {
	name string
	file string
}

func symKeyFor(s Symbol) symKey {
	return symKey{name: s.Name, file: s.File}
}

func indexSymbols(syms []Symbol) map[symKey]Symbol {
	idx := make(map[symKey]Symbol, len(syms))
	for _, s := range syms {
		idx[symKeyFor(s)] = s
	}
	return idx
}

func indexTodos(todos []Todo) map[string]Todo {
	idx := make(map[string]Todo, len(todos))
	for _, t := range todos {
		pos := t.File + ":" + itoa(t.Line)
		idx[pos] = t
	}
	return idx
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func symbolFieldsEqual(a, b Symbol) bool {
	return a.State == b.State &&
		a.Confidence == b.Confidence &&
		a.ChurnCount == b.ChurnCount &&
		a.InstabilityScore == b.InstabilityScore
}

// DiffSnapshots computes the delta between prev and next ProjectState.
func DiffSnapshots(prev, next *ProjectState) *SnapshotDelta {
	d := &SnapshotDelta{
		CreatedAt: time.Now(),
	}

	prevSyms := indexSymbols(prev.Symbols)
	nextSyms := indexSymbols(next.Symbols)

	for k, ns := range nextSyms {
		ps, exists := prevSyms[k]
		if !exists {
			d.SymbolsAdded = append(d.SymbolsAdded, ns)
		} else if !symbolFieldsEqual(ps, ns) {
			d.SymbolsChanged = append(d.SymbolsChanged, ns)
		}
	}

	for k, ps := range prevSyms {
		if _, exists := nextSyms[k]; !exists {
			d.SymbolsRemoved = append(d.SymbolsRemoved, ps)
		}
	}

	prevTodos := indexTodos(prev.Todos)
	nextTodos := indexTodos(next.Todos)

	for k, nt := range nextTodos {
		if _, exists := prevTodos[k]; !exists {
			d.TodosAdded = append(d.TodosAdded, nt)
		}
	}
	for k, pt := range prevTodos {
		if _, exists := nextTodos[k]; !exists {
			d.TodosRemoved = append(d.TodosRemoved, pt)
		}
	}

	return d
}

// ApplyDelta applies a delta to a base ProjectState, producing a new state.
func ApplyDelta(base *ProjectState, d *SnapshotDelta) *ProjectState {
	state := NewProjectState()
	state.UpdatedAt = d.CreatedAt
	state.GitCommit = base.GitCommit
	state.Metadata = base.Metadata

	removed := make(map[symKey]bool)
	for _, s := range d.SymbolsRemoved {
		removed[symKeyFor(s)] = true
	}

	changed := make(map[symKey]Symbol)
	for _, s := range d.SymbolsChanged {
		changed[symKeyFor(s)] = s
	}

	added := make(map[symKey]bool)
	for _, s := range d.SymbolsAdded {
		added[symKeyFor(s)] = true
	}

	for _, s := range base.Symbols {
		k := symKeyFor(s)
		if removed[k] {
			continue
		}
		if ch, ok := changed[k]; ok {
			state.AddSymbol(ch)
		} else {
			state.AddSymbol(s)
		}
	}

	for _, s := range d.SymbolsAdded {
		k := symKeyFor(s)
		if removed[k] {
			continue
		}
		if !added[k] {
			continue
		}
		state.AddSymbol(s)
	}

	removedTodos := make(map[string]bool)
	for _, t := range d.TodosRemoved {
		removedTodos[t.File+":"+itoa(t.Line)] = true
	}

	addedTodos := make(map[string]bool)
	for _, t := range d.TodosAdded {
		addedTodos[t.File+":"+itoa(t.Line)] = true
	}

	for _, t := range base.Todos {
		pos := t.File + ":" + itoa(t.Line)
		if !removedTodos[pos] {
			state.AddTodo(t)
		}
	}
	for _, t := range d.TodosAdded {
		pos := t.File + ":" + itoa(t.Line)
		if addedTodos[pos] {
			state.AddTodo(t)
		}
	}

	return state
}

const CheckpointInterval = 10
