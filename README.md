# dizz - Progress-aware Dev CLI

> Know what to work on next. No AI. No tokens. Just facts from your code.

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
│              (init, whereami, resume)                   │
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
dizz commit
```

## Commands

### `dizz init`
Initializes dizz in the current directory.

Creates:
- `.dizz/` directory for state storage
- `.dizz/config.json` with project settings

```bash
$ dizz init
✓ Initialized dizz
  Created .dizz/
  Project: my-awesome-project

Next: Run 'dizz whereami' to see your project state
```

### `dizz whereami` ⭐ **Most Important**

Analyzes your entire codebase and tells you:
1. Which functions are actively used
2. Which are declared but never called (dead code?)
3. What work is planned (TODOs)
4. **What you should work on next**

```bash
$ dizz whereami
🔍 Analyzing project...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📍 WHERE YOU ARE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ USED FUNCTIONS
  ValidateToken
    internal/auth/token.go
  ProcessRequest
    internal/api/handler.go

⚠ UNUSED FUNCTIONS (declared but not called)
  CheckUserRole
    internal/auth/rbac.go

✗ PLANNED (TODOs found)
  EnforceRBAC
    internal/auth/rbac.go

📝 TODOs: 3 found
  • internal/auth/rbac.go:45: // TODO: implement role checking
  • cmd/api.go:12: // TODO: add rate limiting
  • README.md:89: // TODO: add deployment docs

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💡 NEXT SUGGESTED ACTION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
→ Implement EnforceRBAC

Summary: 4 functions analyzed (2 used, 1 unused, 1 planned)
```

### `dizz status`

Quick summary of your project state.

```bash
$ dizz status
📊 Project Status

Last updated: 2026-01-31T10:30:00Z

Functions:
  ✓ Used:    15
  ⚠ Unused:  3
  ✗ Planned: 2
  ─────────────
  Total:     20

TODOs: 7

💡 Tip: Run 'dizz whereami' for detailed analysis
```

### `dizz commit`

Saves a timestamped snapshot of your current state to `.dizz/history/`.

Useful for:
- Tracking progress over time
- Comparing states before/after refactoring
- Generating metrics

```bash
$ dizz commit
✓ Saved snapshot: .dizz/history/state_2026-01-31_10-30-00.json

💡 Use this to track progress over time
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
  "include": ["**/*.go"],
  "exclude": ["vendor/**", "node_modules/**", ".git/**"]
}
```

## State Storage

`.dizz/state.json` contains:
```json
{
  "updated_at": "2026-01-31T10:30:00Z",
  "functions": [
    {
      "name": "ValidateToken",
      "file": "internal/auth/token.go",
      "state": "used",
      "confidence": 0.9
    }
  ],
  "todos": [
    "internal/auth/rbac.go:45: // TODO: implement role checking"
  ]
}
```


---

Built with ❤️ for developers who hate wasting time deciding what to work on.

