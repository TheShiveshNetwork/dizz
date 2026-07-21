---
name: dizz
description: Always invoke at the start of any coding session on a dizz-initialized project, and before/after making changes. Maintains the project's memory — active intents (goals/TODOs), symbol health (unstable/unused/abandoned code), git-aware churn where available, and immutable snapshots. Use to answer "what should I work on next", "what's unfinished", "what changed since last time", "is this code dead", or to record a decision so future sessions remember it. Read-only, zero-config, works in any repo — git or non-git.
license: MIT
compatibility: Claude Code, Cursor, Gemini CLI, OpenCode, Codex CLI / Copilot
metadata:
  author: TheShiveshNetwork
  version: "1.0.0"
---

# dizz

dizz is this project's memory. It reads git history and source once and returns the state an agent would otherwise burn tokens rediscovering: what's planned, what's breaking, what's dead, what happened last session. `dizz context` is the default first move on any dizz-initialized project — cheaper and more reliable than re-reading files or asking the user to recap.

## Rule: context first, snapshot around every change

1. **Start of session** → `dizz context` (always — ~2KB, replaces re-reading state files)
2. **Before edits** → `dizz snapshot --auto`
3. **After edits** → `dizz snapshot --auto`, then `dizz log` to confirm no new unstable/unused symbols
4. **New goal or task surfaces** → `dizz intent add "<msg>" --type <todo|fixme|refactor|question> --severity <1-3>`
5. **`No dizz project found`** → `dizz init`, then go to step 1

## Commands

| Command | Use for |
|---|---|
| `dizz init` | One-time setup, creates `.dizz/` |
| `dizz context` | Full state dump for agents (TON format) — **default first call** |
| `dizz status` | Quick human-readable health check |
| `dizz log` / `dizz log -a` | Symbol-level detail (unstable/unused/abandoned); `-a` includes healthy symbols too |
| `dizz intent list` | Active goals/tasks |
| `dizz intent add "<msg>" --type <t> --severity <1-3>` | Record a goal (types: `todo`, `fixme`, `refactor`, `question`) |
| `dizz intent resolve <id>` | Close a completed intent |
| `dizz snapshot --auto` | Immutable state checkpoint |
| `dizz snapshot --diff` | Same, delta-only (compact) |
| `dizz snapshot list` / `checkout <hash>` / `prune --keep N` | History / rollback / cleanup |
| `dizz install-skill` | Re-run agent discovery for this skill |

## `dizz context` output — TON (Token-Optimized Notation)

Pipe-delimited, one record per line, split on `|`, no parser needed:

```
Project: myproject | git: main:a1b2c3d

# intents
id|type|sev|status|msg
int_1770020361|refactor|2|active|Refactor scoring system

# symbols:unstable
name|file|line|churn|instability
Scorer.InterpretSignalsWithIntent|internal/state/scorer.go|12|12|0.87

# symbols:unused
name|file|line|state|confidence
oldHelper|internal/util/helpers.go|45|unused|0.65

# todos
file|line|type|text
cmd/snapshot.go|42|TODO|handle git-less repos

# snapshots
hash
a1b2c3d
```

Symbol states: `active` (used, stable) · `planned` (has TODO/intent) · `unstable` (high churn) · `unused` (never called) · `abandoned` (old + unused).

Intent storage (`.dizz/intent.ton`) uses the same style:

```
id|type|sev|status|msg|scope|tags|created_by|resolution
int_001|fixme|3|active|Fix critical bug|project|urgent|user|
```

## Failure modes

| Symptom | Cause | Fix |
|---|---|---|
| `No dizz project found` | Project not initialized | `dizz init` |
| Context missing git info | Non-git repo, or no commits yet | Expected and fully supported — dizz works without git, just with less churn/history detail |
| Data looks stale | `.dizz/` out of date | `dizz snapshot --auto` |
| `install-skill` finds no agents | No supported agent installed | Install one, or manually link `.agents/skills/dizz/` |

## Constraints

- **Read-only** — dizz never modifies source code. Fixes and intent resolution are the agent's job, based on dizz's output.
- **Zero-config** by default.
- Test functions (`TestXxx`, `BenchmarkXxx`) appear "unused" because the test framework calls them implicitly. Exclude them via `"exclude": ["**/*_test.go"]` in `.dizz/config.json`, or mark individually with `// @ignore-unused`.
