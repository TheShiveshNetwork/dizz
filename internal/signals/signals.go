package signals

// SignalType represents the type of fact extracted from code
type SignalType string

const (
	// Structure signals
	FunctionDefined SignalType = "function_defined"
	FunctionCalled  SignalType = "function_called"
	ImportFound     SignalType = "import_found"

	// Intent signals
	TodoFound    SignalType = "todo_found"
	TodoRemoved  SignalType = "todo_removed"
	IntentMarker SignalType = "intent_marker"

	// Time signals
	FileTouched  SignalType = "file_touched"
	FileModified SignalType = "file_modified"
)

// Signal represents a single fact extracted from the codebase
type Signal struct {
	Type       SignalType             `json:"type"`
	Name       string                 `json:"name,omitempty"`
	File       string                 `json:"file"`
	Line       int                    `json:"line,omitempty"`
	Column     int                    `json:"column,omitempty"`
	EndLine    int                    `json:"end_line,omitempty"`
	EndColumn  int                    `json:"end_column,omitempty"`
	Language   string                 `json:"language,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
}

// NewSignal creates a new signal with default confidence
func NewSignal(sigType SignalType, file string) *Signal {
	return &Signal{
		Type:       sigType,
		File:       file,
		Confidence: 1.0,
		Metadata:   make(map[string]interface{}),
	}
}

// WithName sets the name of the signal
func (s *Signal) WithName(name string) *Signal {
	s.Name = name
	return s
}

func (s *Signal) WithRange(
	line, col, endLine, endCol int,
) *Signal {
	return s.
			WithLine(line).
			WithColumn(col).
			WithEndLine(endLine).
			WithEndColumn(endCol)
}

// WithLine sets the line number
func (s *Signal) WithLine(line int) *Signal {
	s.Line = line
	return s
}

// WithColumn sets the column number
func (s *Signal) WithColumn(col int) *Signal {
	s.Column = col
	return s
}

// WithEndLine sets the ending line number
func (s *Signal) WithEndLine(line int) *Signal {
	s.EndLine = line
	return s
}

// WithEndColumn sets the ending column number
func (s *Signal) WithEndColumn(col int) *Signal {
	s.EndColumn = col
	return s
}

// WithLanguage sets the language
func (s *Signal) WithLanguage(lang string) *Signal {
	s.Language = lang
	return s
}

// WithMeta adds metadata
func (s *Signal) WithMeta(key string, value interface{}) *Signal {
	s.Metadata[key] = value
	return s
}

// WithConfidence sets confidence level
func (s *Signal) WithConfidence(conf float64) *Signal {
	s.Confidence = conf
	return s
}

// SignalSet is a collection of signals from analysis
type SignalSet struct {
	Signals []Signal `json:"signals"`
}

// Add appends a signal to the set
func (ss *SignalSet) Add(signal Signal) {
	ss.Signals = append(ss.Signals, signal)
}

// Filter returns signals matching a predicate
func (ss *SignalSet) Filter(predicate func(Signal) bool) []Signal {
	var result []Signal
	for _, sig := range ss.Signals {
		if predicate(sig) {
			result = append(result, sig)
		}
	}
	return result
}

// ByType returns all signals of a specific type
func (ss *SignalSet) ByType(sigType SignalType) []Signal {
	return ss.Filter(func(s Signal) bool {
		return s.Type == sigType
	})
}

// ByFile returns all signals from a specific file
func (ss *SignalSet) ByFile(file string) []Signal {
	return ss.Filter(func(s Signal) bool {
		return s.File == file
	})
}
