package state

type FunctionState string

const (
	Used    FunctionState = "used"
	Unused  FunctionState = "unused"
	Planned FunctionState = "planned"
)

type Function struct {
	Name       string        `json:"name"`
	File       string        `json:"file"`
	State      FunctionState `json:"state"`
	Confidence float64       `json:"confidence"`
}

type ProjectState struct {
	UpdatedAt string     `json:"updated_at"`
	Functions []Function `json:"functions"`
	Todos     []string   `json:"todos"`
}
