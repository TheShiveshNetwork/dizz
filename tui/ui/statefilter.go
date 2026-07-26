package ui

type StateFilter struct {
	current string
	keys    map[string]string
}

func NewStateFilter(mapping map[string]string) *StateFilter {
	keys := make(map[string]string, len(mapping))
	for k, v := range mapping {
		keys[k] = v
	}
	return &StateFilter{keys: keys}
}

func (f *StateFilter) HandleKey(key string) (handled bool, changed bool) {
	if state, ok := f.keys[key]; ok {
		if f.current == state {
			f.current = ""
		} else {
			f.current = state
		}
		return true, true
	}
	if key == "x" {
		if f.current != "" {
			f.current = ""
			return true, true
		}
		return true, false
	}
	return false, false
}

func (f *StateFilter) Active() bool {
	return f.current != ""
}

func (f *StateFilter) Value() string {
	return f.current
}

func (f *StateFilter) Reset() {
	f.current = ""
}

func (f *StateFilter) SetValue(v string) {
	f.current = v
}
