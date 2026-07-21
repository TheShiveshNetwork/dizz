---
name: dizz
description: State-aware project assistant for AI agents. Tracks project intents, code health, symbol states, and git context. Use when you need to understand project state, find what to work on next, detect dead or unstable code, or record project decisions.
license: MIT
compatibility: Designed for AI agents (Claude Code, OpenClaw, Cursor, Gemini CLI, OpenCode, Codex CLI / Copilot)
metadata:
  author: TheShiveshNetwork
  version: "1.0.0"
---

# dizz Skill

dizz continuously models the project's progress by combining static analysis, Git history, and intent markers. It helps agents remember what has been done, what is planned, and what has been forgotten.

## When to Invoke

- **On first interaction** with a project: run `dizz context` to get a full state overview
- **Before making changes**: run `dizz snapshot --auto` to record the pre-change baseline
- **After making changes**: run `dizz snapshot --auto` again, then `dizz log` to verify no unintended states appeared
- **To find work items**: run `dizz intent list` to see active project intents
- **To detect dead code**: run `dizz log` and look for `unused` or `abandoned` symbols
- **To check project health**: run `dizz status` for a quick summary
- **To record a goal**: run `dizz intent add "description" --type todo --severity 2`

## Procedure

### 1. Initialize (first time)

```bash
dizz init
```

Creates `.dizz/` metadata directory. Already initialized projects show current state.

### 2. Get context

```bash
dizz context
```

Returns a ~2 KB token-optimized dump of:
- Active intents (goals, tasks, questions)
- Unstable symbols (high churn areas needing attention)
- Unused/abandoned symbols (dead code candidates)
- Active TODOs from source code
- Git branch and commit hash

**Output format** (TON - Token-Optimized Notation):

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

Pipe-delimited, one record per line. Split on `|` to parse. No JSON parser needed.

### 3. Record baseline (before changes)

```bash
dizz snapshot --auto
```

Creates a content-addressed, immutable record in `.dizz/objects/`.

### 4. View intents

```bash
dizz intent list          # List all active intents
dizz intent add "msg" --type todo --severity 2   # Add new intent
dizz intent resolve <id>  # Mark intent as completed
```

### 5. Check health

```bash
dizz status     # Quick overview: planned work, unstable areas, unused code
dizz log        # Detailed symbol-level health (unused/abandoned focus)
dizz log -a     # All symbols including healthy ones
```

### 6. Snapshot management

```bash
dizz snapshot --auto            # Create new snapshot
dizz snapshot list              # List all snapshots
dizz snapshot checkout <hash>   # View a past snapshot state
dizz snapshot prune --keep 10   # Clean up old snapshots
dizz snapshot --diff            # Store only delta (compact)
```

## Data Format

All intent data is stored in `.dizz/intent.ton` (Token-Optimized Notation):

```
id|type|sev|status|msg|scope|tags|created_by|resolution
int_001|fixme|3|active|Fix critical bug|project|urgent|user|
```

Line-oriented, pipe-delimited. First line is the header. No quotes, braces, or indentation. ~90% fewer tokens than equivalent JSON.

## Failure Modes

| Symptom | Cause | Resolution |
|---------|-------|------------|
| `No dizz project found` | Project not initialized | Run `dizz init` |
| Missing git info in context | Not a git repo, or no commits | dizz works without git, but context will show less info |
| Old snapshot data | Stale `.dizz/` state | Run `dizz snapshot --auto` to refresh |
| `install-skill` no agent dirs found | No supported AI agent installed | Install any supported agent first, or manually link `.agents/skills/dizz/` |

## Notes

- dizz is **read-only** — it never modifies source code
- All cleanup, intent resolution, and state changes are performed by the agent based on dizz's output
- Config is always optional — dizz works with zero configuration
