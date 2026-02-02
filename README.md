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
│            (init, log, status, snapshot, list, resume)      │
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
dizz log

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

### `dizz log`

Analyzes your entire codebase and tells you what needs attention:
- **Planned**: Functions with TODOs (needs implementation)
- **Unstable**: Code changing too frequently (potential issues)
- **Unused**: Functions declared but never called (dead code?)
- **Abandoned**: Old, unused code (consider removal)
- **Active**: Working well (hidden by default)

**Flags:**
- `--all, -a`: Show all symbols including active ones
- `--verbose, -v`: Show detailed analysis info

### `dizz status`

Quick project health check with visual indicators.

### `dizz snapshot`

Creates an immutable snapshot of current project state.
Snapshots are content-addressed (like git objects) and stored in `.dizz/objects/`.

**Flags:**
- `--auto`: Automatic snapshot (called by git hook)

### `dizz list`

Shows all saved snapshots with timestamps and project history.

### `dizz resume`

Quick context after being away from project.
Optimized for "I haven't touched this in weeks" scenario.

### `dizz intent`

Manage human-authored intents and TODOs for better project planning.

**Subcommands:**
- `dizz intent add <message>` - Add a new intent
- `dizz intent list` - List all intents
- `dizz intent resolve <id>` - Resolve/remove an intent by ID

**Flags:**
- `--severity <0-3>`: Set intent severity (default: 1)
- `--tags <tag1,tag2>`: Add tags to intent
- `--type <todo|fixme|refactor|question|hack|temporary>`: Set intent type

**Examples:**
```bash
# Add a high-priority intent
dizz intent add "Fix critical security issue" --severity 3 --type fixme --tags security,urgent

# Add a refactor intent with tags
dizz intent add "Refactor authentication system" --severity 2 --type refactor --tags architecture,performance

# Add a simple TODO
dizz intent add "Add unit tests for user module" --type todo

# List all intents
dizz intent list

# Resolve an intent
dizz intent resolve int_1770020361
```

### `dizz version`

Show current dizz version information.

### `dizz upgrade`

Upgrade dizz to the latest version.

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

### Symbol States

- **`active`**: Used and stable
- **`planned`**: Has TODO markers, needs implementation  
- **`unstable`**: High churn, changing too frequently
- **`unused`**: Declared but never called
- **`abandoned`**: Old, unused, and stale

## Snapshot Storage

Snapshots are stored in `.dizz/objects/` using content-addressed storage.

Each snapshot is immutable and contains the complete project state at that moment.


---

Built with ❤️ for developers who hate wasting time deciding what to work on.

