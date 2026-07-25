package views

import (
	"fmt"
	"sort"

	"github.com/TheShiveshNetwork/dizz/tui/dizzclient"
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/TheShiveshNetwork/dizz/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type TodosModel struct {
	todos      []dizzclient.Todo
	filter     ui.Filter
	typeFilter *ui.StateFilter
	grouped    map[string][]dizzclient.Todo
	files      []string
	collapsed  map[string]bool
	selected   int
	loading    bool
	err        string
}

func NewTodosModel() *TodosModel {
	return &TodosModel{
		loading: true,
		typeFilter: ui.NewStateFilter(map[string]string{
			"t": "TODO",
			"f": "FIXME",
		}),
	}
}

func (m *TodosModel) Init() tea.Cmd {
	return m.refresh()
}

func (m *TodosModel) refresh() tea.Cmd {
	return func() tea.Msg {
		todos, err := dizzclient.ListTodos()
		if err != nil {
			return todosMsg{err: err.Error()}
		}
		return todosMsg{todos: todos}
	}
}

type todosMsg struct {
	todos []dizzclient.Todo
	err   string
}

func (m *TodosModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case todosMsg:
		m.loading = false
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.todos = msg.todos
			m.err = ""
			m.buildGroups()
		}

	case ui.RefreshTick:
		return m, m.refresh()
	}

	return m, nil
}

func (m *TodosModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if consumed, changed := m.filter.HandleKey(msg.String()); consumed {
		if changed {
			m.buildGroups()
		}
		return m, nil
	}

	if handled, changed := m.typeFilter.HandleKey(msg.String()); handled {
		if changed {
			m.buildGroups()
		}
		return m, nil
	}

	switch msg.String() {
	case "up":
		m.selected--
		if m.selected < 0 {
			m.selected = 0
		}
	case "down":
		m.selected++
		if m.selected >= len(m.files) {
			m.selected = len(m.files) - 1
		}
	case "enter":
		if m.selected < len(m.files) {
			file := m.files[m.selected]
			m.collapsed[file] = !m.collapsed[file]
		}
	case "a":
		anyCollapsed := false
		for _, file := range m.files {
			if m.collapsed[file] {
				anyCollapsed = true
				break
			}
		}
		expand := anyCollapsed
		for _, file := range m.files {
			m.collapsed[file] = !expand
		}
	case "r":
		return m, m.refresh()
	}

	return m, nil
}

func (m *TodosModel) buildGroups() {
	filtered := m.todos

	if m.filter.Active() && m.filter.Query != "" {
		var f []dizzclient.Todo
		for _, t := range filtered {
			if m.filter.MatchesAny(dizzclient.RelPath(t.File), t.Text, t.Type) {
				f = append(f, t)
			}
		}
		filtered = f
	}

	if tv := m.typeFilter.Value(); tv != "" {
		var f []dizzclient.Todo
		for _, t := range filtered {
			if t.Type == tv {
				f = append(f, t)
			}
		}
		filtered = f
	}

	m.grouped = make(map[string][]dizzclient.Todo)
	for _, t := range filtered {
		m.grouped[t.File] = append(m.grouped[t.File], t)
	}

	m.files = make([]string, 0, len(m.grouped))
	for f := range m.grouped {
		m.files = append(m.files, f)
	}
	sort.Strings(m.files)

	if m.collapsed == nil {
		m.collapsed = make(map[string]bool)
	}

	if m.selected >= len(m.files) {
		m.selected = 0
	}
}

func (m *TodosModel) Render(c *render.Canvas) {
	w, h := c.Width(), c.Height()

	if m.loading {
		c.SetContent((w-len("Loading todos..."))/2, h/2, render.StyleMuted, "Loading todos...")
		return
	}

	if m.err != "" {
		c.SetContent(2, 2, render.StyleError, m.err)
		return
	}

	header := "TODOs & FIXMEs"
	if tv := m.typeFilter.Value(); tv != "" {
		header += " [" + tv + "]"
	}
	if m.filter.Query != "" {
		header += " /" + m.filter.Query
	}
	c.SetContent(2, 0, render.StyleHighlight.Bold(true), header)

	countStr := fmt.Sprintf("%d todos, %d files", len(m.todos), len(m.files))
	c.SetContent(w-2-len(countStr), 0, render.StyleMuted, countStr)

	filterHint := "enter=toggle  a=all  t=to-do f=fixme x=clear  r=refresh  /=search"
	c.SetContent(2, 1, render.StyleDim, filterHint)

	y := 3
	overflow := false
outer:
	for fi, file := range m.files {
		if y >= h-1 {
			overflow = true
			break
		}

		todos := m.grouped[file]
		fileDisplay := dizzclient.RelPath(file)
		if len(fileDisplay) > w-10 {
			fileDisplay = "..." + fileDisplay[len(fileDisplay)-w+13:]
		}

		style := render.StyleHighlight
		prefix := "▼ "
		if m.collapsed[file] {
			prefix = "▶ "
		}
		if fi == m.selected {
			style = render.StyleHighlight.Bold(true)
			c.SetContent(2, y, style, "▸")
		}
		c.SetContent(3, y, style, fmt.Sprintf("%s%s", prefix, fileDisplay))
		y++

		if m.collapsed[file] {
			continue
		}

		for _, todo := range todos {
			if y >= h-1 {
				overflow = true
				break outer
			}
			todoStyle := render.StyleWarning
			if todo.Type == "FIXME" {
				todoStyle = render.StyleError
			}

			line := fmt.Sprintf("    %4d  %s  %s", todo.Line, todo.Type, todo.Text)
			if len(line) > w-3 {
				line = line[:w-3]
			}
			c.SetContent(3, y, todoStyle, line)
			y++
		}
	}

	if len(m.todos) == 0 {
		c.SetContent((w-20)/2, h/2, render.StyleSuccess, "No TODOs found!")
	} else if overflow {
		c.SetCell(2, h-2, render.StyleDim, '▼')
		c.SetContent(3, h-2, render.StyleDim, "more items below...")
	}

	m.filter.Render(c, 2, h-1)
}

// @dizz-ignore-unused
func (m *TodosModel) View() string { return "" }

func (m *TodosModel) InputMode() bool { return m.filter.Active() }
