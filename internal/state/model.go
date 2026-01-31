package state

import "time"

// SymbolState represents the interpreted state of a code symbol
type SymbolState string

const (
	Active      SymbolState = "active"       // Used and stable
	Planned     SymbolState = "planned"      // Has TODO, not fully implemented
	Unused      SymbolState = "unused"       // Declared but not called
	Unstable    SymbolState = "unstable"     // High churn
	Abandoned   SymbolState = "abandoned"    // Low churn, not used
)

// Symbol represents a tracked code symbol (function, class, etc.)
type Symbol struct {
	Name       string      `json:"name"`
	File       string      `json:"file"`
	Type       string      `json:"type"`        // "function", "class", "module"
	Language   string      `json:"language"`
	State      SymbolState `json:"state"`
	Confidence float64     `json:"confidence"`
	
	// Context
	IsDefined  bool        `json:"is_defined"`
	IsCalled   bool        `json:"is_called"`
	HasTodo    bool        `json:"has_todo"`
	
	// Time dimension
	LastTouched  *time.Time `json:"last_touched,omitempty"`
	ChurnCount   int        `json:"churn_count,omitempty"`
	
	// Intent
	IntentMarker string     `json:"intent_marker,omitempty"`
}

// Todo represents a tracked TODO item
type Todo struct {
	File      string     `json:"file"`
	Line      int        `json:"line"`
	Text      string     `json:"text"`
	Type      string     `json:"type"`      // "TODO", "FIXME", etc.
	Language  string     `json:"language"`
	AddedAt   *time.Time `json:"added_at,omitempty"`
	Resolved  bool       `json:"resolved"`
}

// FileContext represents context about a file
type FileContext struct {
	Path         string     `json:"path"`
	Language     string     `json:"language"`
	LastModified time.Time  `json:"last_modified"`
	ChurnCount   int        `json:"churn_count"`
	Symbols      []string   `json:"symbols"`      // Symbol names in this file
}

// ProjectState represents the complete interpreted state of the project
type ProjectState struct {
	UpdatedAt    time.Time              `json:"updated_at"`
	GitCommit    string                 `json:"git_commit,omitempty"`
	Symbols      []Symbol               `json:"symbols"`
	Todos        []Todo                 `json:"todos"`
	Files        []FileContext          `json:"files"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NewProjectState creates a new project state
func NewProjectState() *ProjectState {
	return &ProjectState{
		UpdatedAt: time.Now(),
		Symbols:   make([]Symbol, 0),
		Todos:     make([]Todo, 0),
		Files:     make([]FileContext, 0),
		Metadata:  make(map[string]interface{}),
	}
}

// AddSymbol adds a symbol to the state
func (ps *ProjectState) AddSymbol(symbol Symbol) {
	ps.Symbols = append(ps.Symbols, symbol)
}

// AddTodo adds a todo to the state
func (ps *ProjectState) AddTodo(todo Todo) {
	ps.Todos = append(ps.Todos, todo)
}

// AddFile adds file context
func (ps *ProjectState) AddFile(file FileContext) {
	ps.Files = append(ps.Files, file)
}

// GetSymbolsByState returns all symbols in a given state
func (ps *ProjectState) GetSymbolsByState(state SymbolState) []Symbol {
	var result []Symbol
	for _, sym := range ps.Symbols {
		if sym.State == state {
			result = append(result, sym)
		}
	}
	return result
}

// GetActiveTodos returns unresolved todos
func (ps *ProjectState) GetActiveTodos() []Todo {
	var result []Todo
	for _, todo := range ps.Todos {
		if !todo.Resolved {
			result = append(result, todo)
		}
	}
	return result
}

// Summary returns a quick summary of the state
type Summary struct {
	TotalSymbols    int            `json:"total_symbols"`
	ByState         map[SymbolState]int `json:"by_state"`
	ActiveTodos     int            `json:"active_todos"`
	TotalFiles      int            `json:"total_files"`
}

// GetSummary returns a summary of the project state
func (ps *ProjectState) GetSummary() Summary {
	summary := Summary{
		TotalSymbols: len(ps.Symbols),
		ByState:      make(map[SymbolState]int),
		ActiveTodos:  len(ps.GetActiveTodos()),
		TotalFiles:   len(ps.Files),
	}
	
	for _, sym := range ps.Symbols {
		summary.ByState[sym.State]++
	}
	
	return summary
}
