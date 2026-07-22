# dizz TUI — Plan

## Overview

The dizz TUI is a **separate Go binary** (`dizz-tui`) that provides an interactive terminal interface for every dizz feature. It lives in its own Go module at `tui/` inside the dizz repo and is **never compiled into the `dizz` binary itself**.

A new `dizz ui` command acts as the bridge — it checks if `dizz-tui` is installed and either installs it or launches it.

---

## Architecture — Hybrid Rendering

The TUI uses a **two-layer hybrid architecture**:

| Layer | Library | Responsibility |
|-------|---------|----------------|
| **App shell** | BubbleTea | Model/Update/View loop, keyboard input, tab routing, lifecycle, side effects, forms |
| **Frame renderer** | tcell (type system) + custom cell buffer | Per-view frame building, cell grid, styles, frame diffing, procedural animation (stars, moons, meteors) |

```
 BubbleTea (tea.Model)
  │
  │  Update(msg) → receives key events, tick timers, resize
  │  View()     → builds full frame as ANSI string
  │                   │
  │                   ▼
  │         render.Canvas (cell buffer)
  │         ├── Cell{Char rune, Style tcell.Style}
  │         ├── SetCell(x, y, style, char)
  │         ├── Diff(prev Canvas) → []CellDelta
  │         ├── Render() → ANSI string (with \x1b[H + diffs)
  │         └── FullFrame() → ANSI string (full paint)
  │                   │
  │                   ▼
  │         views/dashboard.go, symbols.go, etc.
  │         Each view builds a Canvas frame:
  │           ├── Dashboard: star field, moon phases, bar charts
  │           ├── Symbols:  table with sortable columns
  │           ├── Intents:  list + form overlay
  │           └── ...
  │
  └── BubbleTea returns the ANSI string from View()
       → BubbleTea's internal diff engine flushes only changed cells
```

### Why hybrid?

- **BubbleTea** gives us a battle-tested app structure: Elm-architecture models, clean key dispatch, resize handling, `tea.Tick` for animation, `tea.Exec` for shelling out, and Huh for forms
- **tcell** gives us precise cell-level control: every character has an explicit `(rune, Style{fg, bg, attrs})` pair, identical to gnhf's `Cell{char, style, width}` approach
- The two integrate seamlessly: each view's `View()` method builds a `render.Canvas` (cell grid), calls `canvas.Render()` to produce an ANSI string, and returns that string. BubbleTea handles the terminal-level diffing internally

### Animation (gnhf-style)

```
tea.Tick (200ms)
  │
  ├── Updates model.Now timestamp
  ├── View() rebuilds the frame:
  │     ├── Stars twinkle (cell char flips based on time seed)
  │     ├── Meteors animate across screen (position = f(time))
  │     ├── Moon phase cycles (emoji changes per tick)
  │     └── Stats refresh (elapsed, tokens, commits)
  └── BubbleTea diffs the frame string → only changed ANSI codes flushed
```

---

## Communication with dizz

```
┌─────────────────────────────────────────────────┐
│                  dizz (CLI)                      │
│  cmd/ui.go → "dizz ui"                          │
│    │                                             │
│    ├── not installed → install dizz-tui binary   │
│    └── installed     → exec dizz-tui             │
└──────────────┬──────────────────────────────────┘
               │ exec (replaces process)
               ▼
┌─────────────────────────────────────────────────────┐
│              dizz-tui (TUI)                          │
│                                                      │
│  ┌─────────────────┐    ┌─────────────────────────┐ │
│  │ dizz/client.go   │    │ dizz/state.go            │ │
│  │ (exec.Command)   │    │ (reads .dizz/ files)     │ │
│  │                  │    │                          │ │
│  │ intent add       │    │ state.json (gzip JSON)   │ │
│  │ intent resolve   │    │ intent.ton (pipe-delim)  │ │
│  │ snapshot create  │    │ context.ton (TON format) │ │
│  │ snapshot diff    │    │                          │ │
│  └────────┬─────────┘    └────────────┬─────────────┘ │
│           │                           │               │
│           └───────────┬───────────────┘               │
│                       ▼                               │
│              views/*.go (BubbleTea Models)            │
└─────────────────────────────────────────────────────┘
```

| Channel | Direction | Format | Used For |
|---------|-----------|--------|----------|
| Shell exec (`client.go`) | dizz → TUI output | Text / exit code | Mutations: `dizz intent add`, `dizz intent resolve`, `dizz snapshot create` |
| File read (`state.go`, `intents.go`) | `.dizz/` files → TUI | JSON / TON | Reads: `state.json.gz`, `intent.ton`, `context.ton` |

No Go-level import coupling. The TUI only needs the `dizz` binary on `$PATH` and a valid `.dizz/` directory.

---

## `dizz ui` Command

```go
// cmd/ui.go — registered in dizz's own CLI
var uiCmd = &cobra.Command{
    Use:   "ui",
    Short: "Launch or install the dizz Terminal UI",
    Long:  `Opens the dizz interactive terminal UI.
If dizz-tui is not installed, installs it first,
then launches it as a child process.`,
    Run: func(cmd *cobra.Command, args []string) {
        runUI()
    },
}

func init() { rootCmd.AddCommand(uiCmd) }
```

### Install flow

1. Detect OS/arch
2. Try these paths in order: `~/.dizz/bin/dizz-tui`, `~/.local/bin/dizz-tui`, any `$PATH` dir
3. If not found, install:
   - **Download**: Pre-built binary from GitHub release (`dizz-tui-{os}-{arch}`) — same pattern as `dizz upgrade`
   - **Go install**: Fallback: `go install github.com/TheShiveshNetwork/dizz/tui@latest`
   - **Source**: If inside repo, build from `tui/`
4. Place binary in `~/.dizz/bin/`, add to `$PATH` if possible
5. Exec the binary

### Launch flow

```go
func runUI() {
    path, err := exec.LookPath("dizz-tui")
    if err != nil {
        installTUI()
        path, _ = exec.LookPath("dizz-tui")
    }
    c := exec.Command(path)
    c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
    c.Run()
}
```

---

## Project Structure

```
tui/                                    # Separate Go module
├── go.mod
├── go.sum
├── main.go                             # Entry point: tea.NewProgram(app.NewModel())
│
├── app/
│   ├── app.go                          # Top-level BubbleTea Model
│   │   type Model struct {
│   │       currentTab int
│   │       views      []tea.Model       # each view is a tea.Model
│   │       renderer   *render.Engine    # shared render engine (tcell screen)
│   │       status     StatusData
│   │       help       help.Model
│   │       err        error
│   │   }
│   │   Init()    → tea.Cmd (init views, start ticker)
│   │   Update()  → dispatch keys, tab switch, resize
│   │   View()    → delegate to active view's View()
│   └── app_test.go
│
├── views/
│   ├── dashboard.go                     # Tab 1: Animated health overview
│   │   builds Canvas with:
│   │   ├── star field (twinkle at 200ms tick)
│   │   ├── meteor shower (configurable frequency)
│   │   ├── moon phase strip (per iteration state)
│   │   ├── health score + symbol bar chart
│   │   ├── intent / todo counts
│   │   └── git status line
│   │
│   ├── symbols.go                       # Tab 2: Symbol explorer
│   │   builds Canvas with:
│   │   ├── tab header row (state filters)
│   │   ├── scrollable table (Name, Type, State, File:Line)
│   │   ├── sort indicators on column headers
│   │   ├── color-coded rows by state
│   │   └── detail panel on Enter (churn, last touched, intent marker)
│   │
│   ├── intents.go                       # Tab 3: Intent management
│   │   builds Canvas with:
│   │   ├── filtered list (active/resolved/all)
│   │   ├── severity color bars
│   │   ├── Huh form overlay on 'i' (message, type, severity, tags)
│   │   └── confirm dialog on 'r'
│   │
│   ├── snapshots.go                     # Tab 4: Snapshot timeline
│   │   builds Canvas with:
│   │   ├── reverse-chronological timeline
│   │   ├── hash badge, timestamp, git ref
│   │   ├── delta arrows between entries
│   │   └── detail panel with symbol diff
│   │
│   ├── todos.go                         # Tab 5: TODO/FIXME viewer
│   │   builds Canvas with:
│   │   ├── collapsible file groups
│   │   ├── line-number gutter
│   │   ├── TODO (yellow) / FIXME (red) coloring
│   │   └── Enter → $EDITOR file:line
│   │
│   └── resume.go                        # Tab 6: Resume view
│       builds Canvas with:
│       ├── time-away banner
│       ├── planned work items
│       ├── action suggestion
│       └── quick summary grid
│
├── render/                              # Cell-based render engine (tcell-inspired)
│   ├── canvas.go                        # Canvas — the core cell buffer
│   │   type Canvas struct {
│   │       cells  [][]Cell
│   │       width, height int
│   │   }
│   │   type Cell struct {
│   │       Char rune
│   │       Style tcell.Style
│   │   }
│   │   type Style = tcell.Style         # reuse tcell's Style type
│   │
│   │   New(w, h) → *Canvas
│   │   SetCell(x, y, s Style, r rune)
│   │   SetContent(x, y, s Style, text string)
│   │   Fill(s Style, r rune)
│   │   Line(x1, y1, x2, y2, s Style, r rune)
│   │   Rect(x, y, w, h, s Style, fill rune)
│   │   Sub(x, y, w, h) → *Canvas       # create sub-canvas for compositing
│   │
│   ├── diff.go                          # Frame differ
│   │   type CellDelta struct {
│   │       X, Y int
│   │       Cell Cell
│   │   }
│   │   func Diff(prev, next Canvas) []CellDelta
│   │   └── only returns cells that changed
│   │
│   ├── ansi.go                          # ANSI output builder
│   │   func (c Canvas) FullFrame() string
│   │   └── wraps with \x1b[H + ansi style codes per cell
│   │
│   ├── procedural.go                    # Procedural visual generators
│   │   type Star struct { X, Y int; Char rune; Phase int }
│   │   func GenerateStarField(w, h, density float64, seed int64) []Star
│   │   func GetStarState(s Star, now int64) string
│   │   │   └── returns "bright", "dim", or "hidden" based on time phase
│   │   │
│   │   type Meteor struct { X, Y, Length, Speed int; StartTime int64 }
│   │   func GenerateMeteorShower(w, h, count int, seed int64) []Meteor
│   │   func GetMeteorTrail(m Meteor, now int64) []TrailPoint
│   │   │   └── returns animated positions: head (bright) + fading tail (dim)
│   │   │
│   │   func GetMoonPhase(kind string) string
│   │   │   └── returns emoji: 🌑🌒🌓🌔🌕🌖🌗🌘 based on iteration state
│   │   │
│   │   func RenderBar(count, total, width int, fillStyle, emptyStyle Style) string
│   │   └── renders ████████░░░░ bar
│   │
│   └── colors.go                        # Color palette
│       ├── StateColor(state string) Style  # green=active, yellow=planned, etc.
│       ├── SeverityColor(sev int) Style    # red=3, yellow=2, green=1, dim=0
│       └── (maps to tcell.ColorXYZ values)
│
├── ui/                                  # TUI components (built on render.Canvas)
│   ├── help.go                          # Help overlay (Canvas-based)
│   │   ├── KeyBinding{Key, Desc}
│   │   ├── RenderHelp(canvas, bindings, page)
│   │   └── Toggle(model) → shows/hides overlay
│   │
│   ├── tabs.go                          # Tab bar (Canvas-based)
│   │   ├── Tab{Name, Key, StateIcon}
│   │   ├── RenderTabBar(canvas, active int, tabs []Tab, x, y, width int)
│   │   └── active tab highlighted, inactive dimmed
│   │
│   ├── statusbar.go                     # Status bar (Canvas-based)
│   │   ├── StatusData{Branch, Commit, Dirty, Project, Version}
│   │   ├── RenderStatusBar(canvas, data StatusData, x, y, width int)
│   │   └── git:branch | commit hash | ● dirty indicator
│   │
│   ├── table.go                         # Scrollable table (Canvas-based)
│   │   ├── Table{Cols, Rows, SortCol, SortDir, Selected}
│   │   ├── RenderTable(canvas, table, x, y, w, h)
│   │   ├── header row with sort arrows (▲/▼)
│   │   ├── color-coded rows
│   │   └── selection highlight + scroll
│   │
│   └── form.go                          # Inline form overlay
│       ├── FormField{Label, Value, Validator}
│       ├── RenderForm(canvas, fields, x, y, w)
│       └── handles focus, validation, submit
│
├── dizz/                                # dizz communication layer
│   ├── client.go                        # Shell exec wrappers
│   │   Status() intentAdd() intentResolve() snapshotCreate() Version()
│   ├── parser.go                        # Parse CLI output
│   │   ParseStatusOutput() ParseLogOutput() ParseSnapshotOutput()
│   ├── state.go                         # Read .dizz/state.json.gz → DTOs
│   │   ReadProjectState() ReadSummary() ReadTodos()
│   └── intents.go                       # Read .dizz/intent.ton → DTOs
│       ReadIntentState() ParseIntentsTON()
│
├── Makefile
│   build install run test lint clean
│
└── README.md
```

---

## Module Dependencies

```go
// tui/go.mod
module github.com/TheShiveshNetwork/dizz/tui

go 1.25

require (
    github.com/charmbracelet/bubbletea  v1.3.4   // app architecture, input, lifecycle
    github.com/charmbracelet/huh        v0.6.0   // forms (intent add)
    github.com/gdamore/tcell/v2         v2.8.0   // Style type, Color values, cell model
)
```

| Library | Role | Used for |
|---------|------|----------|
| `bubbletea` | App framework | Model/Update/View, key dispatch, `tea.Tick`, `tea.Exec`, resize events, lifecycle |
| `huh` | Forms | Intent add form (message, type, severity, tags) |
| `tcell/v2` | Cell/Style types only | `tcell.Style`, `tcell.Color` — used as the type system for `render.Cell.Style` |
| `stdlib` | Terminal rendering | Direct ANSI escape codes via `fmt.Fprintf` within the render engine |

**No `tcell.Screen`** is used — BubbleTea owns the terminal. tcell is imported solely for its `Style` and `Color` types, giving us a well-designed cell style system without reinventing it.

---

## Render Engine — How It Works

### Cell model (borrowed from tcell)

```go
// render/canvas.go
type Style = tcell.Style  // fg color + bg color + attributes (bold, dim, blink)

type Cell struct {
    Char rune
    Style Style
}

type Canvas struct {
    cells  [][]Cell
    width  int
    height int
}
```

### Frame pipeline

Each view's `View()` method follows this exact pipeline:

```go
func (m *DashboardModel) View() string {
    // 1. Create a fresh canvas at terminal size
    c := render.NewCanvas(m.width, m.height)

    // 2. Build the visual frame using Canvas primitives
    renderTitle(&c, m.projectName)
    renderStars(&c, m.stars, time.Now())
    renderMeteors(&c, m.meteors, time.Now())
    renderHealthScore(&c, m.healthScore)
    renderBarChart(&c, m.symbolCounts)
    renderStatusBar(&c, m.statusData)

    // 3. Diff against previous frame and return ANSI string
    return c.RenderDiff(m.prevCanvas)
}
```

### Diff + output

```go
func (c *Canvas) RenderDiff(prev *Canvas) string {
    var b strings.Builder
    for y := 0; y < c.height; y++ {
        for x := 0; x < c.width; x++ {
            if prev == nil || c.cells[y][x] != prev.cells[y][x] {
                // ANSI: move cursor to (y, x), write styled char
                fmt.Fprintf(&b, "\x1b[%d;%dH%s%c",
                    y+1, x+1,
                    ansiEscape(c.cells[y][x].Style),
                    c.cells[y][x].Char,
                )
            }
        }
    }
    return b.String()
}
```

BubbleTea will then flush this string to the terminal, only outputting the changed cells.

### Animation loop

```go
// app.go — animation tick
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.initViews(),
        tickEvery(200 * time.Millisecond),  // gnhf uses 200ms
    )
}

func tickEvery(d time.Duration) tea.Cmd {
    return tea.Tick(d, func(t time.Time) tea.Msg {
        return AnimationTick(t)
    })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case AnimationTick:
        m.now = time.Now()
        // Stars twinkle, meteors advance, moon phases cycle
        return m, tickEvery(200 * time.Millisecond)  // re-arm
    case tea.KeyMsg:
        // handle keys
    }
}
```

---

## Views — Detailed Specification

### Dashboard (Tab 1) — Animated Project Health

```
 ┌──────────────────────────────────────────────────────────────┐
 │   d i z z                                                    │
 │   ┌──┐┌──┐┌──┐┌──┐┌──┐                                      │
 │   │  ││  ││  ││  ││  │  Project: myproject                   │
 │   └──┘└──┘└──┘└──┘└──┘                                      │
 │                                                              │
 │             ╭───╮    ★        ★                              │
 │    ★    ◯   │ ○ │         ★      *        Code Score: ● 82% │
 │        *    ╰───╮    ★                              ████░░  │
 │    ★      *     │       *     ★                     ████░░  │
 │               ╰─╯    ★                              ██░░░░  │
 │    *     ★                   ★      ★               █░░░░░  │
 │                                                     ░░░░░░  │
 │   🌑🌒🌓🌔🌕🌖🌗🌘🌑🌒🌓🌔🌕🌖🌗🌘  TODOs: 8              │
 │                                                              │
 │   ✓ Active  42  ████████████████░░░░  Intents: 3 (1 high)   │
 │     Planned  5  ██░░░░░░░░░░░░░░░░░░                        │
 │     Unused   3  █░░░░░░░░░░░░░░░░░░░  11 items need         │
 │     Unstable 2  █░░░░░░░░░░░░░░░░░░░  attention             │
 │     Abandon  1  ░░░░░░░░░░░░░░░░░░░░                        │
 │   ────────────────────────                                   │
 │     Total: 53                                                │
 │                                                              │
 │   main  a1b2c3d  ●  Last commit: add auth middleware        │
 │                                                              │
 │         ★        ★            ★         ★                   │
 │            ★          ★               ★       ★             │
 │   ★              ★        ★    ★               ★            │
 │   [ctrl+c to stop, dizz-tui again to resume]                │
 └──────────────────────────────────────────────────────────────┘
```

- Star field twinkles at 200ms intervals (density: 3.5% of cells)
- Meteors streak across left/right edges (frequency configurable 0-5)
- Moon phase strip shows last 30 iteration states (🌑=fail, 🌕=success, 🌖=running)
- Content centered within a 63-char wide viewport
- Terminal title updates live: `dizz · 1.2k in · 4.5k out · 3 commits`

### Symbols (Tab 2) — Explorer

```
 ┌──────────────────────────────────────────────────────────────┐
 │  Symbols  [planned] [unstable] [unused] [abandoned] [active] │
 │ ──────────────────────────────────────────────────────────── │
 │  Name                  Type   State      File                 │
 │  ────────────────────────────────────────────                 │
 │  ▸ authenticateUser   fn     planned  ░ auth.go:42           │
 │  ▸ connectToServer    fn     unstable ░ process.go:712       │
 │  ▸ legacyParser       fn     abandoned░ parser.go:203         │
 │  ▸ unusedHelper       fn     unused   ░ utils.go:15          │
 │  ▸ validateInput      fn     active   ░ validate.go:88       │
 │  ────────────────────────────────────────────                 │
 │  [selected item detail]                                       │
 │  ────────────────────────────────────────────                 │
 │  authenticateUser                                             │
 │  File:    internal/auth/auth.go:42-78                         │
 │  Type:    function                                            │
 │  State:   planned                                             │
 │  Churn:   3 changes in 20 commits                             │
 │  Marker:  TODO: implement OAuth2 flow                         │
 │  Lang:    Go                                                  │
 └──────────────────────────────────────────────────────────────┘
```

- Built with `render.Canvas` + `ui.Table` component
- Arrow keys to scroll, `/` to filter, Enter for detail panel
- Color-coded state column: green, yellow, red, cyan, gray
- Sort by any column (clickable header)

### Intents (Tab 3) — Management

```
 ┌──────────────────────────────────────────────────────────────┐
 │  Intents  [active●] [resolved] [all]                         │
 │ ──────────────────────────────────────────────────────────── │
 │  ID           Type    Sev  Message                            │
 │  ────────────────────────────────────────────                 │
 │  int_001      fixme   ███ Fix critical auth bug               │
 │  int_002      todo    ██░ Refactor middleware                 │
 │  int_003      quest   █░░ Why is parser.so slow?              │
 │  int_004      hack    ░░░ Temp workaround for #142            │
 │                                                              │
 │  [i] Add intent    [r] Resolve selected                      │
 └──────────────────────────────────────────────────────────────┘
```

- Severity shown as 3-cell bar (███ = high, ░░░ = low)
- `i` opens a Huh form overlay: message, type (dropdown), severity (slider), tags
- `r` shows confirm dialog, calls `dizz intent resolve <id>`

### Snapshots (Tab 4) — Timeline

```
 ┌──────────────────────────────────────────────────────────────┐
 │  Snapshots          [s] Create  [d] Create delta             │
 │ ──────────────────────────────────────────────────────────── │
 │  ◉ a1b2c3  Jan 15 14:30  (abc1234)  Act 42 Pln 5 Uns 2 …    │
 │  │                                                           │
 │  ○ f6e7d8  Jan 14 10:15  (def5678)  Act 40 Pln 7 Uns 3 …    │
 │  │                                                           │
 │  ○ 9a0b1c  Jan 13 16:45  (ghi9012)  Act 38 Pln 9 Uns 4 …    │
 │                                                              │
 │  [detail panel for selected]                                 │
 │  ────────────────────────────────                            │
 │  Hash:     a1b2c3d4e5f6                                      │
 │  Created:  Jan 15 14:30:05                                   │
 │  Git:      abc1234 "add auth middleware"                      │
 │  Symbols:  53 total                                          │
 │  Δ:       +2 symbols, +1 todo                               │
 └──────────────────────────────────────────────────────────────┘
```

- Vertical timeline with ◉ = current, ○ = historical
- Timeline connector (│) between entries
- Diff stats: +2 symbols, -1 removed, +1 todo

### TODOs (Tab 5) — File-grouped viewer

```
 ┌──────────────────────────────────────────────────────────────┐
 │  TODOs & FIXMEs                                              │
 │ ──────────────────────────────────────────────────────────── │
 │  ▼ internal/auth/auth.go                                     │
 │      42  TODO   implement OAuth2 refresh                     │
 │     108  FIXME  handle token expiry race                     │
 │  ▼ cmd/root.go                                               │
 │      15  TODO   add --verbose flag                           │
 │                                                              │
 │  Showing 8 todos across 4 files                              │
 └──────────────────────────────────────────────────────────────┘
```

- File headers are collapsible (▼ expanded, ▶ collapsed)
- Line number gutter right-aligned
- Enter opens `$EDITOR file:line`

### Resume (Tab 6) — Context recovery

```
 ┌──────────────────────────────────────────────────────────────┐
 │  Resume — Last worked on 3 days ago                         │
 │ ──────────────────────────────────────────────────────────── │
 │                                                              │
 │    Current branch: main                                      │
 │                                                              │
 │    ⚠ You had 5 planned work items:                          │
 │      • authenticateUser        internal/auth/auth.go         │
 │      • rateLimiter             internal/middleware.go         │
 │      • cacheLayer              internal/cache/store.go       │
 │                                                              │
 │    ✓ 42 symbols working well                                 │
 │    ⚠ 11 items need attention                                │
 │                                                              │
 │    💡 WHAT TO DO NOW                                         │
 │    → Start with authenticateUser — it has a                  │
 │      TODO waiting and the highest churn rate                 │
 │                                                              │
 │    Re-analyze?  [dizz log]  [dizz status]                    │
 └──────────────────────────────────────────────────────────────┘
```

---

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle tabs |
| `1`–`6` | Jump to tab |
| `q` / `Ctrl+C` | Quit |
| `?` | Toggle help overlay |
| `R` | Force refresh |
| `↑`/`↓` | Navigate list |
| `Enter` | Select / drill down |
| `Esc` | Back / close detail |
| `/` | Filter |

### Per-view

| View | Key | Action |
|------|-----|--------|
| Dashboard | `r` | Re-run `dizz log` analysis |
| Symbols | `p`/`u`/`s`/`a` | Filter by state |
| Intents | `i` | Add intent form |
| Intents | `r` | Resolve selected |
| Snapshots | `s` | Create snapshot |
| Snapshots | `d` | Create delta |
| TODOs | `Enter` | Open file at line |

---

## DTOs (Data Transfer Objects)

```go
// dizz/state.go — partial copies of dizz's types
type ProjectState struct {
    UpdatedAt time.Time      `json:"updated_at"`
    Symbols   []Symbol       `json:"symbols"`
    Todos     []Todo         `json:"todos"`
    Metadata  map[string]any `json:"metadata"`
}

type Symbol struct {
    Name       string  `json:"name"`
    File       string  `json:"file"`
    Line       int     `json:"line"`
    Type       string  `json:"type"`
    State      string  `json:"state"`
    Confidence float64 `json:"confidence"`
    ChurnCount int     `json:"churn_count"`
}

type Todo struct {
    File string `json:"file"`
    Line int    `json:"line"`
    Text string `json:"text"`
    Type string `json:"type"`
}
```

```go
// dizz/intents.go
type Intent struct {
    ID       string   `json:"id"`
    Type     string   `json:"type"`
    Message  string   `json:"message"`
    Severity int      `json:"severity"`
    Status   string   `json:"status"`
    Tags     []string `json:"tags"`
}
```

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| `dizz` not on `$PATH` | Show prompt with auto-install option |
| No `.dizz/` directory | Show "run `dizz init`" dialog with inline command execution |
| `state.json.gz` corrupt | Show warning, fall back to `dizz log` CLI output |
| `intent.ton` missing | Create fresh in-memory state, warn on first save |
| dizz command fails | Error toast (3s auto-dismiss via `tea.Tick`), keep UI running |
| Network error during install | Show manual install instructions |
| Terminal resize | Canvas rebuilds at new size (handled by BubbleTea's `tea.WindowSizeMsg`) |

---

## Star / Meteor / Moon Phase Engine

### Star field

- Generated once per session (seeded random)
- Density: ~3.5% of terminal cells
- Each star has a phase offset: `char = f((now + seed) % period)`
- Twinkle cycle: bright (★) → dim (·) → hidden (space) → dim → bright
- Stars near the content viewport are dimmer (visual hierarchy)

### Meteors

- Generated at startup: position, angle, speed, start delay
- Active on left and right edges of the screen
- Trail: head (bright █) → body (dim ░) → fade to space
- Frequency flag `--meteor-frequency 0-5` controls count (0=off, 5=max)

### Moon phases

- Each iteration result gets a moon: 🌕 (success), 🌑 (fail), 🌖 (running)
- Displayed in a horizontal strip of 30 moons
- Running moon cycles through phases at 200ms intervals

---

## Implementation Order

| Phase | What | Why |
|-------|------|-----|
| **1. Scaffold** | `go.mod`, `main.go`, App model, tab routing, tick loop | Foundation |
| **2. Render engine** | `render/canvas.go`, `render/diff.go`, `render/ansi.go`, `render/colors.go` | Core — enables everything visual |
| **3. Procedurals** | `render/procedural.go` — stars, meteors, moons | gnhf-style animation |
| **4. UI components** | `ui/tabs.go`, `ui/statusbar.go`, `ui/help.go` | Navigation shell |
| **5. dizz client** | `dizz/client.go`, `dizz/state.go`, `dizz/intents.go`, `dizz/parser.go` | Data layer |
| **6. Table component** | `ui/table.go` — scrollable, sortable, selectable | Needed for symbols |
| **7. Dashboard** | Views it with procedurals + stats | Shows the animation works |
| **8. Symbols** | Table view, filter, sort, detail panel | Core feature |
| **9. Intents** | List + form overlay + resolve | CRUD interaction |
| **10. TODOs** | File-grouped collapsible list | Quick utility |
| **11. Snapshots** | Timeline + create/diff | Feature parity |
| **12. Resume** | Context recovery view | Wraps the experience |
| **13. Form component** | `ui/form.go` — inline overlays | Intent add / confirm dialogs |
| **14. cmd/ui.go** | `dizz ui` bridge in dizz | Ships the integration |
| **15. Polish** | Help overlay, error toasts, loading spinners, resize handling | Production quality |

---

## File Count Summary

```
tui/
├── go.mod, go.sum                         2
├── main.go                                1
├── app/
│   ├── app.go                             1
│   └── app_test.go                        1
├── views/
│   ├── dashboard.go                       1
│   ├── symbols.go                         1
│   ├── intents.go                         1
│   ├── snapshots.go                       1
│   ├── todos.go                           1
│   └── resume.go                          1
├── render/
│   ├── canvas.go                          1
│   ├── diff.go                            1
│   ├── ansi.go                            1
│   ├── procedural.go                      1
│   └── colors.go                          1
├── ui/
│   ├── help.go                            1
│   ├── tabs.go                            1
│   ├── statusbar.go                       1
│   ├── table.go                           1
│   └── form.go                            1
├── dizz/
│   ├── client.go                          1
│   ├── parser.go                          1
│   ├── state.go                           1
│   └── intents.go                         1
├── Makefile                               1
└── README.md                              1
                                          ─────
                                Total:   25 files
```

Plus **1 file** in the dizz root: `cmd/ui.go` (~60 lines).

### Architecture summary

```
 BubbleTea (tea.Model)     ← input, lifecycle, tick, forms, side effects
    │
    ▼
 Render Engine (Canvas)    ← cell buffer, style + rune per cell, frame diff, ANSI output
    │
    ├── procedural.go      ← stars, meteors, moon phases (gnhf-style)
    ├── ui/*.go            ← tabs, table, statusbar, help, form (Canvas-based components)
    └── views/*.go         ← dashboard, symbols, intents, snapshots, todos, resume
                               │
                               ▼
                         dizz/*.go  ← shells out to `dizz` CLI + reads `.dizz/` files
```
