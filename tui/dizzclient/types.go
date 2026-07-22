package dizzclient

type Symbol struct {
	Name       string  `json:"name"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	EndLine    int     `json:"end_line,omitempty"`
	Type       string  `json:"type"`
	Language   string  `json:"language"`
	State      string  `json:"state"`
	Confidence float64 `json:"confidence"`
	ChurnCount int     `json:"churn_count"`
	HasTodo    bool    `json:"has_todo"`
}

type Todo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Text     string `json:"text"`
	Type     string `json:"type"`
	Language string `json:"language"`
	Resolved bool   `json:"resolved"`
}

type Intent struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Message  string   `json:"message"`
	Scope    string   `json:"scope"`
	Severity int      `json:"severity"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags,omitempty"`
}

type SnapshotInfo struct {
	Hash      string
	Timestamp string
	Kind      string
	Size      string
}

type Summary struct {
	TotalSymbols int
	Active       int
	Planned      int
	Unstable     int
	Unused       int
	Abandoned    int
	ActiveTodos  int
	Intents      int
	TotalFiles   int
	ProjectName  string
	Branch       string
	Commit       string
}
