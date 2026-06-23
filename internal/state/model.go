package state

import (
	"fmt"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/integrations"
)

// SymbolState represents the interpreted state of a code symbol
type SymbolState string

const (
	Active    SymbolState = "active"    // Used and stable
	Planned   SymbolState = "planned"   // Not fully implemented
	Unused    SymbolState = "unused"    // Declared but not called
	Unstable  SymbolState = "unstable"  // High churn
	Abandoned SymbolState = "abandoned" // Low churn, not used
)

// Symbol represents a tracked code symbol (function, class, etc.)
type Symbol struct {
	Name       string      `json:"name"`
	File       string      `json:"file"`
	Line       int         `json:"line"`
	Column     int         `json:"column"`
	EndLine    int         `json:"end_line,omitempty"` // end of symbol
	EndColumn  int         `json:"end_column,omitempty"`
	Type       string      `json:"type"` // "function", "class", "module"
	Language   string      `json:"language"`
	State      SymbolState `json:"state"`
	Confidence float64     `json:"confidence"`

	// Context
	IsDefined bool `json:"is_defined"`
	IsCalled  bool `json:"is_called"`
	HasTodo   bool `json:"has_todo"`

	// Time dimension
	LastTouched      *time.Time `json:"last_touched,omitempty"`
	ChurnCount       int        `json:"churn_count,omitempty"`
	InstabilityScore float64    `json:"instability_score,omitempty"`

	// Intent
	IntentMarker string `json:"intent_marker,omitempty"`

	// Analysis accuracy — one of "ast", "lexical", "regex".
	// Populated from signal source metadata; omitted for legacy state files.
	SignalSource string `json:"signal_source,omitempty"`
}

type Todo struct {
	File     string     `json:"file"`
	Line     int        `json:"line"`
	Text     string     `json:"text"`
	Type     string     `json:"type"`
	Language string     `json:"language"`
	AddedAt  *time.Time `json:"added_at,omitempty"`
	Resolved bool       `json:"resolved"`
}

// FileContext represents context about a file
type FileContext struct {
	Path         string    `json:"path"`
	Language     string    `json:"language"`
	LastModified time.Time `json:"last_modified"`
	ChurnCount   int       `json:"churn_count"`
	Symbols      []string  `json:"symbols"` // Symbol names in this file
}

// ProjectState represents the complete interpreted state of the project
type ProjectState struct {
	UpdatedAt time.Time              `json:"updated_at"`
	GitCommit *integrations.Commit   `json:"git_commit,omitempty"`
	Symbols   []Symbol               `json:"symbols,omitempty"`
	Todos     []Todo                 `json:"todos,omitempty"`
	Files     []FileContext          `json:"files,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// projectStateRaw is used for unmarshaling to handle backward compatibility
type projectStateRaw struct {
	UpdatedAt time.Time              `json:"updated_at"`
	GitCommit interface{}            `json:"git_commit,omitempty"`
	Symbols   []Symbol               `json:"symbols,omitempty"`
	Todos     []Todo                 `json:"todos,omitempty"`
	Files     []FileContext          `json:"files,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
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
	TotalSymbols int                 `json:"total_symbols"`
	ByState      map[SymbolState]int `json:"by_state"`
	ActiveTodos  int                 `json:"active_todos"`
	TotalFiles   int                 `json:"total_files"`
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

// IntentType represents the type of intent
type IntentType string

const (
	IntentTodo IntentType = "todo"
	Fixme      IntentType = "fixme"
	Refactor   IntentType = "refactor"
	Question   IntentType = "question"
	Hack       IntentType = "hack"
	Temporary  IntentType = "temporary"
)

// IntentStatus represents the status of an intent
type IntentStatus string

const (
	IntentActive   IntentStatus = "active"
	IntentResolved IntentStatus = "resolved"
	IntentDeferred IntentStatus = "deferred"
)

// Intent represents a human-authored intent
type Intent struct {
	ID             string       `json:"id"`
	Type           IntentType   `json:"type"`
	Message        string       `json:"message"`
	Scope          string       `json:"scope"` // file:line-line format
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	CreatedBy      string       `json:"created_by"`
	Severity       int          `json:"severity"`   // 0-3
	Confidence     float64      `json:"confidence"` // 0.0-1.0
	Status         IntentStatus `json:"status"`
	Tags           []string     `json:"tags,omitempty"`
	RelatedCommits []string     `json:"related_commits,omitempty"`
	BlockedBy      []string     `json:"blocked_by,omitempty"`
	Resolution     *Resolution  `json:"resolution,omitempty"`
}

// Resolution represents how an intent was resolved
type Resolution struct {
	Method      string    `json:"method"` // "fixed", "wontfix", "duplicate", etc.
	Description string    `json:"description"`
	ResolvedAt  time.Time `json:"resolved_at"`
	ResolvedBy  string    `json:"resolved_by"`
}

// Assumption represents a recorded assumption
type Assumption struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Scope       string    `json:"scope"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
	Confidence  float64   `json:"confidence"`
}

// KnownRisk represents a known risk
type KnownRisk struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Scope       string    `json:"scope"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
	Severity    int       `json:"severity"`
	Mitigation  string    `json:"mitigation,omitempty"`
}

// IntentState represents persistent human intent tracking
type IntentState struct {
	Version     float64      `json:"version"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Intents     []Intent     `json:"intents"`
	Assumptions []Assumption `json:"assumptions,omitempty"`
	KnownRisks  []KnownRisk  `json:"known_risks,omitempty"`
}

// NewIntentState creates a new intent state
func NewIntentState() *IntentState {
	now := time.Now()
	return &IntentState{
		Version:     1.0,
		CreatedAt:   now,
		UpdatedAt:   now,
		Intents:     make([]Intent, 0),
		Assumptions: make([]Assumption, 0),
		KnownRisks:  make([]KnownRisk, 0),
	}
}

// AddIntent adds an intent to the state
func (is *IntentState) AddIntent(intent Intent) {
	is.UpdatedAt = time.Now()
	is.Intents = append(is.Intents, intent)
}

// GetActiveIntents returns all active intents
func (is *IntentState) GetActiveIntents() []Intent {
	var result []Intent
	for _, intent := range is.Intents {
		if intent.Status == IntentActive {
			result = append(result, intent)
		}
	}
	return result
}

// @ignore-unused
// TODO: add intent filters
func (is *IntentState) GetIntentsByType(intentType IntentType) []Intent {
	var result []Intent
	for _, intent := range is.Intents {
		if intent.Type == intentType {
			result = append(result, intent)
		}
	}
	return result
}

// @ignore-unused
// TODO: add intent filters
func (is *IntentState) GetIntentsBySeverity(minSeverity int) []Intent {
	var result []Intent
	for _, intent := range is.Intents {
		if intent.Severity >= minSeverity {
			result = append(result, intent)
		}
	}
	return result
}

// ResolveIntent marks an intent as resolved
func (is *IntentState) ResolveIntent(id string, resolution Resolution) error {
	for i, intent := range is.Intents {
		if intent.ID == id {
			is.Intents[i].Status = IntentResolved
			is.Intents[i].Resolution = &resolution
			is.Intents[i].UpdatedAt = time.Now()
			is.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("intent not found: %s", id)
}

// GetIntentSummary returns a summary of the intent state
type IntentSummary struct {
	TotalIntents  int                  `json:"total_intents"`
	ActiveIntents int                  `json:"active_intents"`
	ByType        map[IntentType]int   `json:"by_type"`
	ByStatus      map[IntentStatus]int `json:"by_status"`
	HighSeverity  int                  `json:"high_severity"` // severity >= 3
}

// GetIntentSummary returns a summary of intents
func (is *IntentState) GetIntentSummary() IntentSummary {
	summary := IntentSummary{
		TotalIntents: len(is.Intents),
		ByType:       make(map[IntentType]int),
		ByStatus:     make(map[IntentStatus]int),
		HighSeverity: 0,
	}

	for _, intent := range is.Intents {
		summary.ByType[intent.Type]++
		summary.ByStatus[intent.Status]++

		if intent.Status == IntentActive {
			summary.ActiveIntents++
		}

		if intent.Severity >= 3 {
			summary.HighSeverity++
		}
	}

	return summary
}
