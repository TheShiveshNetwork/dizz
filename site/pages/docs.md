# Dizz

`dizz` is a local, Git-aware developer CLI that analyzes your codebase and answers one core question:

> **“What should I work on next?”**

## Overview

`dizz` is a local, Git-aware developer CLI that analyzes your codebase to understand **progress**, not just correctness.  
It helps developers answer the question **"What should I work on next?"** by detecting unused code, planned work (TODOs), unstable areas, and forgotten or abandoned logic using static analysis and Git context.

Unlike linters or task managers, `dizz` models **developer intent and code evolution**. It runs fully offline, requires no configuration to start, and works across multiple languages through a unified signal-based architecture.

## Why dizz?

> **Git tracks truth.
> dizz tracks understanding.**

Modern projects fail not because of bugs, but because of:

* Lost context
* Forgotten TODOs  
* Unused or half-connected code
* Unclear priorities after time away

`dizz` continuously models your project's **state of progress** and surfaces what actually deserves your attention.

## Quick Start

```bash
cd your-project
dizz init
dizz status
dizz log
```

## Commands

### `dizz init`

Initializes `.dizz/` metadata in your project.

### `dizz log`

Full project analysis.

Shows:

* Planned work
* Unstable areas  
* Unused code
* Abandoned code

**Flags**

* `--all, -a` — include healthy symbols
* `--verbose, -v` — detailed reasoning

### `dizz status`

Quick health snapshot with visual indicators.

### `dizz snapshot`

Creates an immutable snapshot of project state.

* Content-addressed (Git-like)
* Stored in `.dizz/objects/`

Flags:

* `--auto` — for Git hooks

### `dizz list`

Shows snapshot history and project evolution.

### `dizz resume`

Instant context recovery after time away.

Optimized for:

> "I haven't touched this project in weeks."

### `dizz intent`

Manage human-authored intent.

```bash
dizz intent add "Refactor auth layer" --severity 2 --type todo
dizz intent list
dizz intent resolve int_1770020361
```

## Symbol States

| State       | Meaning              |
| ----------- | -------------------- |
| `active`    | Used and stable      |
| `planned`   | Has TODO / intent    |
| `unstable`  | High churn           |
| `unused`    | Declared, never used |
| `abandoned` | Old + unused         |

## Architecture Overview

### The Four Dimensions of Project State

Every project has four dimensions that `dizz` models explicitly:

| Dimension     | What It Represents | How It's Derived      |
| ------------- | ------------------ | --------------------- |
| **Structure** | What exists        | AST parsing, regex    |
| **Usage**     | What's connected   | Call graphs, imports  |
| **Intent**    | What's planned     | TODOs, intent markers |
| **Time**      | What's stable      | Git history & churn   |

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Commands                         │
│      (init, log, status, snapshot, list, resume)        │
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
            │   Signals    │  ← universal facts
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ State Engine │  ← interpretation
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ Project State│  ← understanding
            └──────────────┘
```

## License

[LICENSE](https://github.com/TheShiveshNetwork/dizz/blob/main/LICENSE)

Built with ❤️ for developers who hate wasting time deciding what to work on next.
