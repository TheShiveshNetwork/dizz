# dizzie — Terminal UI for dizz

dizzie is an interactive TUI for [dizz](https://github.com/TheShiveshNetwork/dizz), the state-aware project assistant.

## Usage

```bash
dizz ui        # auto-installs dizzie if missing, then launches
dizzie         # or run directly if already installed
```

## Tabs

| Tab | Description |
|-----|-------------|
| Dashboard | Animated project health overview with star field, meteors, code score, and symbol bar chart |
| Symbols | Searchable, filterable table of all project symbols with detail panel |
| Intents | Manage intents with a 3-step add wizard (type, message, severity) |
| Snapshots | Create and view project state snapshots |
| TODOs | File-grouped TODO/FIXME viewer |

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Toggle focus sidebar / main |
| `↑`/`↓` | Sidebar: switch tab | Main: navigate content |
| `1`–`5` | Jump to tab |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

### Symbols tab

| Key | Action |
|-----|--------|
| `↑`/`↓` | Navigate rows |
| `Enter` | Toggle detail panel (shows full file path, language, churn, confidence, etc.) |
| `/` | Enter search mode — type to filter by name/file/state (case-insensitive) |
| `Esc` / `Enter` | Exit search mode |
| `Backspace` | Delete last search character |
| `a` / `p` / `u` / `s` / `d` | Filter by state: active, planned, unstable, unused, abandoned |
| `x` | Clear state filter |
| `r` | Refresh data |

### Intents tab

| Key | Action |
|-----|--------|
| `i` | Add intent (3-step wizard: type+message → severity → submit) |
| `r` | Resolve selected intent |
| `a` / `s` | Filter: active / resolved (toggle) |
| `x` | Clear filter |
| `/` | Search by ID, type, message |
| `Tab` | Switch focus in wizard (type ↔ message) |
| `Enter` | Next step / submit |

### Snapshots tab

| Key | Action |
|-----|--------|
| `s` | Create snapshot |
| `d` | Create delta |
| `/` | Search by hash |
| `r` | Refresh |

### TODOs tab

| Key | Action |
|-----|--------|
| `/` | Search by file, text, type |
| `r` | Refresh |

## Status Bar

The bottom bar shows:

- **NORMAL** (dim gray) — view is in navigation mode
- **INSERT** (green bold) — view is in text input mode (search or wizard)
- `[sidebar]` / `[main]` — current focus indicator

Any view with text input implements `InputMode() bool` to signal INSERT mode. The app detects this via a type assertion — adding a new input-capable view requires only implementing the method.

## Reusable Filter Components

### `ui.Filter` — free-text search

```go
filter := ui.Filter{}
filter.HandleKey("/")       // activate
filter.HandleKey("text")    // build query
filter.MatchesAny(name, file, state) // case-insensitive match
filter.Render(c, x, y)      // draw "/query_" prompt
filter.Active()             // for InputMode status bar
```

Embed in any view. Call `HandleKey()` in the key handler before view-specific keys. Use `Matches()` / `MatchesAny()` in the build/filter pipeline.

### `ui.StateFilter` — key-driven state toggle

```go
sf := ui.NewStateFilter(map[string]string{
    "a": "active",
    "p": "planned",
})
sf.HandleKey("a")  // toggles "active" on/off
sf.HandleKey("x")  // always clears
sf.Value()         // current active state or ""
sf.Active()
```

Define key→value mappings at construction. Same key press toggles the filter off. `x` always clears. Used by Symbols and Intents views.

### `InputMode() bool` — status bar integration

Any view that enters a text-input mode (search, form wizard) implements:
```go
func (m *ViewName) InputMode() bool { return m.filter.Active() }
```
The app detects this via type assertion — no import coupling needed.

## Architecture

```
dizz CLI  →  exec.Command  →  dizzie TUI
```

Communication is through shell exec only — no Go-level import coupling between dizz and dizzie. The TUI calls `dizz status`, `dizz log --all`, `dizz context --intents`, `dizz intent add`, `dizz intent resolve`, `dizz snapshot --auto`, etc. and parses the plain-text output.

- **Bubble Tea** — app shell, key dispatch, animation tick, lifecycle
- **Custom Canvas** — cell-based frame buffer with ANSI output (not lipgloss)
- **tcell/v2** — Style and Color type system only (no tcell.Screen)
