# dizz - Progress-aware Dev CLI

> Know what to work on next.

## What is this?

> **Git tracks truth. Dizz tracks understanding.**

`dizz` analyzes your codebase to answer the most important development question:

**"What should I work on next?"**

It does this by detecting:
- ✓ Functions that are **used** (called somewhere)
- ⚠ Functions that are **declared but unused** (potential dead code or unconnected)
- ✗ Functions that are **planned** (have TODOs)

# dizz Architecture 

## The Four Dimensions of State

Every project has four dimensions that dizz models:

| Dimension | What it Tracks | How |
|-----------|----------------|-----|
| **Structure** | What exists | AST parsing, regex |
| **Usage** | What's connected | Call graphs, imports |
| **Intent** | What's planned | TODOs, @dizz markers |
| **Time** | What's stable | Git churn, last touched |

---

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Commands                         │
│            (init, whereami, status, snapshot, list, resume)      │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                  Orchestration                          │
│         (Coordinates all other layers)                  │
└───┬───────────┬─────────────┬────────────┬──────────────┘
    │           │             │            │
    ▼           ▼             ▼            ▼
┌─────────┐ ┌──────────┐ ┌─────────┐ ┌─────────┐
│ Discover│ │ Analyzer │ │   Git   │ │  Store  │
│  Files  │ │ Registry │ │ Context │ │ Objects │
└────┬────┘ └─────┬────┘ └────┬────┘ └────┬────┘
     │            │           │           │
     └────────────┴───────────┴───────────┘
                   │
                   ▼
            ┌──────────────┐
            │   Signals    │ (Universal facts)
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ State Engine │ (Interpretation)
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ Project State│ (Understanding)
            └──────────────┘
```

---

### What are Signals?

Signals are **language-agnostic facts** extracted from code.

Instead of:
```go
type Function struct {
    Name string
    File string
}
```

We have:
```go
type Signal struct {
    Type     SignalType  // "function_defined", "function_called", etc.
    Name     string
    File     string
    Language string
    Metadata map[string]interface{}
}
```

### Why Signals?

**Languages are adapters. State is universal.**

A signal from Go:
```json
{"type": "function_defined", "name": "ValidateToken", "file": "auth.go"}
```

A signal from Python:
```json
{"type": "function_defined", "name": "validate_token", "file": "auth.py"}
```

Same format. Same interpretation. Different analyzers.

### Signal Types

```go
const (
    // Structure
    FunctionDefined
    FunctionCalled
    ImportFound
    
    // Intent
    TodoFound
    TodoRemoved
    IntentMarker
    
    // Time
    FileTouched
    FileModified
)
```

---

## Language Analyzers

### Interface

Every analyzer implements:
```go
type Analyzer interface {
    Language() string
    Supports(file string) bool
    Analyze(files []string) (*signals.SignalSet, error)
}
```

### Built-in Analyzers

**Go Analyzer** (`analyzer/golang/`)
- Uses Go's built-in AST parser
- High confidence (0.9+)
- Understands methods, interfaces, packages

**Regex Analyzer** (`analyzer/regex/`)
- Fallback for any language
- Pattern-based extraction
- Lower confidence (0.7)
- Supports: JS, Python, Rust, C/C++, Java, Ruby, PHP

## Installation

```bash
# Clone this repo
cd dizz

# Build
go build -o dizz .

# Install globally (optional)
sudo mv dizz /usr/local/bin/

# Or add to your PATH
export PATH=$PATH:$(pwd)
```

## Quick Start

```bash
# Initialize in your project
cd your-project
dizz init

# See where you are
dizz whereami

# Quick status check
dizz status

# Save a snapshot
dizz snapshot

# List all snapshots
dizz list

# Resume work after time away
dizz resume
```

## Commands

### `dizz init`
Initializes dizz in the current directory.

Creates:
- `.dizz/` directory for state storage
- `.dizz/config.json` with project settings
- Git post-commit hook (if in a git repo)

```bash
$ dizz init
✓ Initialized my-awesome-project
  Created .dizz/
✓ Installed git post-commit hook
  Git integration: enabled
  Project: my-awesome-project

Next: Run 'dizz whereami' to see your project state
```

### `dizz whereami` ⭐ **Most Important**

Analyzes your entire codebase and tells you what needs attention:
- **Planned**: Functions with TODOs (needs implementation)
- **Unstable**: Code changing too frequently (potential issues)
- **Unused**: Functions declared but never called (dead code?)
- **Abandoned**: Old, unused code (consider removal)
- **Active**: Working well (hidden by default)

```bash
$ dizz whereami
🔍 Analyzing project...

  ✓ Active: 15
  ⚠ Planned: 2
  🔥 Unstable: 1
  ⚪ Unused: 3

━━ ⚠ PLANNED ━━ needs implementation
  EnforceRBAC
    internal/auth/rbac.go:45
  AddRateLimiting
    cmd/api.go:12

━━ 🔥 UNSTABLE ━━ changing too much
  ValidateToken
    internal/auth/token.go (24 changes)

━━ ⚪ UNUSED ━━ not called anywhere
  CheckUserRole
    internal/auth/rbac.go

💡 NEXT ACTION
→ Implement EnforceRBAC

15 symbols · 6 need attention · 15 active
```

**Flags:**
- `--all, -a`: Show all symbols including active ones
- `--verbose, -v`: Show detailed analysis info

### `dizz status`

Quick project health check with visual indicators.

```bash
$ dizz status
  Project: my-awesome-project
  Branch: main
  Commit: a1b2c3d
  Updated: 2 hours ago

  Health: ● 75%

  Symbols:
    ✓ Active         15 ████████████████████
    ⚠ Planned         2 ██░░░░░░░░░░░░░░░░░
    🔥 Unstable        1 █░░░░░░░░░░░░░░░░░░░
    ⚪ Unused          3 ███░░░░░░░░░░░░░░░░
    ──────────────────────
    Total          21

  📝 TODOs: 7

  6 items need attention
  Run 'dizz whereami' for details
```

### `dizz snapshot`

Creates an immutable snapshot of current project state.
Snapshots are content-addressed (like git objects) and stored in `.dizz/objects/`.

```bash
$ dizz snapshot
✓ Snapshot saved: a1b2c3
  Git commit: a1b2c3d
  Object: .dizz/objects/a1/b2c3d4.json

💡 Snapshots are immutable. Use them to track progress over time.
```

**Flags:**
- `--auto`: Automatic snapshot (called by git hook)

### `dizz list`

Shows all saved snapshots with timestamps and project history.

```bash
$ dizz list

SNAPSHOT HISTORY

  a1b2c3 2h ago (a1b2c3d)
     15✓ 2⚠ 1🔥 3⚪

  d4e5f6 3d ago (d4e5f67)
     13✓ 3⚠ 2🔥 4⚪

  g7h8i9 1w ago (g7h8i9j)
     11✓ 5⚠ 1🔥 6⚪

Total: 3 snapshots
```

### `dizz resume`

Quick context after being away from the project.
Optimized for the "I haven't touched this in weeks" scenario.

```bash
$ dizz resume

  Last worked on: 2 weeks ago
  Current branch: main

  ⚠ You had 2 planned work:
    • EnforceRBAC
      internal/auth/rbac.go:45
    • AddRateLimiting
      cmd/api.go:12

  🔥 Code with high churn:
    • ValidateToken (24 changes)

  QUICK SUMMARY

  ✓ 15 symbols working well
  ⚠ 6 items need attention

  💡 WHAT TO DO NOW

  1. Re-analyze to get current state
     dizz whereami

  → Implement EnforceRBAC

💡 Run 'dizz whereami' to refresh the analysis
```

## How It Works

```
Your Code ──▶ Static Analysis ──▶ Facts ──▶ State ──▶ Output
              (AST, TODOs)           ↓
                                 Confidence
                                  Scoring
```

1. **AST Analysis**: Parses Go code to find function definitions and calls
2. **TODO Detection**: Scans for TODO/FIXME comments
3. **State Scoring**: Combines signals to determine function state
4. **Smart Output**: Shows you exactly what needs attention

- Uses Go's built-in AST parser
- Reads your actual code comments
- No network calls
- No tokens consumed
- Runs in milliseconds

## Configuration

`.dizz/config.json`:
```json
{
  "project_name": "my-project",
  "root_path": ".",
  "exclude": ["vendor/**", "node_modules/**", ".git/**", ".dizz/**"]
}
```

## State Storage

`.dizz/state.json` contains the current project state:
```json
{
  "updated_at": "2026-01-31T10:30:00Z",
  "git_commit": "a1b2c3d4e5f6...",
  "symbols": [
    {
      "name": "ValidateToken",
      "file": "internal/auth/token.go",
      "state": "active",
      "confidence": 0.9,
      "churn_count": 12,
      "last_touched": "2026-01-30T15:20:00Z"
    }
  ],
  "todos": [
    "internal/auth/rbac.go:45: // TODO: implement role checking"
  ]
}
```

### Symbol States

- **`active`**: Used and stable
- **`planned`**: Has TODO markers, needs implementation  
- **`unstable`**: High churn, changing too frequently
- **`unused`**: Declared but never called
- **`abandoned`**: Old, unused, and stale

## Snapshot Storage

Snapshots are stored in `.dizz/objects/` using content-addressed storage:
```
.dizz/objects/
├── a1/
│   └── b2c3d4.json      # Full state snapshot
├── d4/
│   └── e5f67a.json      # Another snapshot
└── refs/
    └── git/
        └── a1b2c3d      # Links git commit to snapshot
```

Each snapshot is immutable and contains the complete project state at that moment.


---

Built with ❤️ for developers who hate wasting time deciding what to work on.

