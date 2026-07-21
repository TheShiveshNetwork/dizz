# Dizz

> **The only brain your codebase will ever need.**

`dizz` reads your Git history, looks at your code, and tells you what needs doing. TODOs, half-finished work, things changing too often — it tracks all of it. Takes snapshots so you never lose context. Agents love it because it speaks their language.

No setup. No internet. Won't touch your code.

## What dizz is not

- Not a linter
- Not a task manager
- Not an AI agent that edits your code or executes changes autonomously

## Quick Start

```bash
cd your-project
dizz init
dizz status
dizz log
```

> **Important**: After installing dizz, run `dizz install-skill` to enable AI agent discovery. This detects Claude Code, Cursor, Gemini CLI, OpenCode, and other agents, then installs a discoverable skill so they can use dizz automatically.

## Commands

### `dizz init`

Initializes `.dizz/` metadata in your project. Creates the tracking directory, config, gitignore, and a project-level agent skill at `.agents/skills/dizz/`.

### `dizz log`

Full project analysis. Shows planned work, unstable areas, unused code, and abandoned code.

**Flags**

* `--all, -a` — include healthy symbols
* `--verbose, -v` — detailed reasoning

### `dizz status`

Quick health snapshot with visual indicators. Shows how many symbols are in each state (active, planned, unstable, unused, abandoned) and active TODO count.

### `dizz context`

Token-optimized project dump designed for AI agents. Outputs everything an agent needs in ~2 KB using TON format (pipe-delimited, one record per line). No parser needed — split on `|`.

**Sections:**

* Project info (name, git branch, commit)
* Active intents (goals, tasks, questions)
* Unstable symbols (high churn)
* Unused / abandoned symbols (dead code candidates)
* Active TODOs from source code
* Recent snapshot hashes

**Flags:**

* `--intents` — show intents only
* `--symbols` — show symbols only
* `--todos` — show todos only

**Output example:**

```
Project: myproject | git: main:a1b2c3d

# intents
id|type|sev|status|msg
int_001|refactor|2|active|Refactor scoring system

# symbols:unstable
name|file|line|churn
Scorer.InterpretSignalsWithIntent|internal/state/scorer.go|12|12

# symbols:unused
name|file|line|state
oldHelper|internal/util/helpers.go|45|unused

# todos
file|line|type|text
cmd/main.go|42|TODO|handle edge case
```

### `dizz snapshot`

Creates an immutable snapshot of project state.

* Content-addressed (Git-like)
* Stored in `.dizz/objects/`
* Supports deltas (`--diff`) for efficient storage
* Checkpoints every 10 deltas

**Flags**

* `--auto` — for Git hooks (non-interactive)
* `--diff` — store only the delta from the last snapshot
* `list` — show all snapshots
* `checkout <hash>` — inspect a past snapshot
* `prune --keep N` — remove old snapshots, keep the N most recent

### `dizz list`

Shows snapshot history and project evolution.

### `dizz resume`

Instant context recovery after time away.

Optimized for:

> "I haven't touched this project in weeks."

### `dizz intent`

Manage human-authored intent separate from source-level TODO/FIXME comments.

Intents are immutable project goals. Unlike TODOs (which are re-scanned on every analysis), intents persist in `.dizz/intent.ton` until explicitly resolved. They carry a severity score from 0-3.

```bash
dizz intent add "Refactor auth layer" --severity 2 --type todo
dizz intent list
dizz intent resolve int_1770020361
```

Types: `todo`, `fixme`, `refactor`, `question`, `hack`, `temporary`

### `dizz install-skill`

Scans your system for installed AI agents and installs the dizz skill into their skill directories. After this, agents like Claude Code, Cursor, Gemini CLI, and OpenCode can discover and invoke dizz automatically.

```bash
dizz install-skill
```

Supported agents: Claude Code, Cursor, Gemini CLI, OpenCode, Codex CLI / Copilot.

The canonical skill definition lives in the repository at `agent-skills/dizz/SKILL.md` and follows the [Agent Skills Specification](https://agentskills.io).

## AI Agent Integration

dizz is built for AI agents from the ground up. Key features:

**`dizz context`** — A single ~2 KB command that replaces reading 100+ KB of state files. Outputs in TON format (Token-Optimized Notation): pipe-delimited, line-oriented, no JSON overhead. Any agent can parse it by splitting on `|`.

**`dizz install-skill`** — Auto-discovers installed AI agents and installs a skill for each. The skill tells agents what dizz does and how to use it. Agents discover skills automatically via the `.agents/skills/` convention.

**`dizz intent`** — Agents can query, add, and resolve project intents programmatically. No need to parse ad-hoc TODO comments in source files.

**TON Format** — All intent data uses TON:
```ton
id|type|sev|status|msg|scope|tags|created_by
int_001|fixme|3|active|Fix critical bug|project|urgent|user
```

~90% fewer tokens than equivalent JSON. No quotes, braces, or indentation.

## Symbol States

Every symbol in your codebase is classified into one of five states, derived from usage signals, intent markers, and Git history:

| State       | Meaning              |
| ----------- | -------------------- |
| `active`    | Used and stable      |
| `planned`   | Has TODO / intent    |
| `unstable`  | High churn           |
| `unused`    | Declared, never used |
| `abandoned` | Old + unused         |

## Architecture Overview

### The Four Dimensions of Project State

| Dimension     | What It Represents | How It's Derived      |
| ------------- | ------------------ | --------------------- |
| **Structure** | What exists        | AST parsing, regex    |
| **Usage**     | What's connected   | Call graphs, imports  |
| **Intent**    | What's planned     | TODOs, intent markers |
| **Time**      | What's stable      | Git history & churn   |

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    CLI Commands                           │
│   (init, log, status, context, snapshot, intent, list)    │
└───────────────────────┬──────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│                  Orchestration                            │
│         (Coordinates all other layers)                    │
└───┬───────────┬─────────────┬────────────┬───────────────┘
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
            │   Signals    │
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ State Engine │
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ Project State│
            └──────────────┘
```

## License

[MIT](https://github.com/TheShiveshNetwork/dizz/blob/main/LICENSE)

Built with love for developers who hate wasting time deciding what to work on next.
