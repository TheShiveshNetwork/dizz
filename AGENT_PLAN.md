# Agent Skill Architecture Plan

## The Problem

dizz stores project state in human-friendly formats that waste tokens when read by LLM agents:

| Current Format | Size | Problem |
|---|---|---|
| `intent.json` | 4.6 KB (9 intents) | Pretty-printed JSON: repeated keys, whitespace, ISO timestamps |
| `state.json.gz` | 97 KB uncompressed | Full symbol table included; agents rarely need all of it |
| `objects/*.json` | 34-82 KB each | Pretty-printed full state snapshots; stored per-commit |
| `config.json` | 209 B | Small, but verbose keys |

An agent reading these files spends 60-70% of tokens on structural overhead (quotes, braces, key names, padding) rather than actual data.

---

## Design

### Token-Optimized Notation (TON)

A line-oriented compact format for all persisted dizz data. Each record lives on one line, fields are pipe-delimited, and the first line is a header declaring the schema.

**Intent file (`intent.ton`)** - 80% smaller:

```ton
id|type|sev|status|msg|scope|tags|created_by|resolution
int_1770020361|refactor|2|active|Refactor scoring system to use weights|project|performance,architecture|user|
int_1770047724|fixme|3|active|improve the speed of the cli|project||user|
```

Versus JSON equivalent (~1,200 bytes -> ~250 bytes). No repeated keys, no quotes, no braces, timestamps stripped (redundant when git stores them).

**Snapshot format** - compact JSON, no indentation:

```json
{"updated_at":"2026-01-01T00:00:00Z","symbols":[{"name":"Foo","file":"src/main.go","line":10,"state":"active","confidence":0.9}],"todos":[],"files":[],"metadata":{}}
```

Compact JSON (via `json.Marshal` instead of `json.MarshalIndent`) cuts object size by ~50-60%.

### Agent Interface: `dizz context`

A single command that dumps everything an agent needs in one shot, token-optimized.

Output (TON format):

```
Project: myproject | git: main:a1b2c3d

# intents
id|type|sev|status|msg
int_1770020361|refactor|2|active|Refactor scoring system to use weights

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

Everything is positional, pipe-delimited, and excludes data agents can infer. This reduces a typical context dump from ~100 KB to ~2-3 KB.

---

## Implementation

### Phase 1: Compact Storage Backend

1. TON codec in `internal/store/ton/` - Writer, Reader, BytesWriter with proper escaping
2. IntentStore uses `intent.ton` by default, falls back to `intent.json`
3. StateStore left as-is (already gzipped, compact JSON used at rest)
4. Snapshots use `json.Marshal` (compact, no indentation) instead of `json.MarshalIndent`
5. Signal cache - no changes needed (already gzipped compact JSON)

### Phase 2: Agent Context Command

1. `dizz context` command in `cmd/context.go`
2. Context renderer in `internal/render/context.go` with TON output
3. Sections: project info, intents, unstable symbols, unused symbols, todos
4. Flags: `--intents`, `--symbols`, `--todos` for filtered output

### Phase 3: Agent Skill System

`dizz init` is now agent-optimized by default:
1. Creates `.agents/skills/dizz/` in project root with `skill.json` + `SKILL.md`
2. Project-level skill tells agents how to use dizz for context
3. Old `--agent` flag removed - agent optimization is the default
4. No AGENTS.md generated on init (skill directory is the mechanism)

Install scripts create a global skill at `~/.agents/skills/dizz/`:
- `site/scripts/install.sh` - Linux/macOS global skill
- `site/scripts/install.ps1` - Windows global skill (`%LOCALAPPDATA%\.agents\skills\dizz\`)

### Phase 4: Snapshot Delta Format -- NOT STARTED

See below.

### Phase 5: Agent Protocol File -- NOT STARTED

See below.

---

## Remaining Work

### Phase 4: Snapshot Delta Format

**Goal**: Store snapshot diffs instead of full copies to reduce long-term storage bloat.

1. **`dizz snapshot --diff`**:
   - Compare current state against last snapshot
   - Store only the delta (added/removed/changed symbols, intent changes)
   - Full snapshot every N commits as a checkpoint

2. **Reconstruction**:
   - `dizz snapshot checkout <hash>` walks delta chain to reconstruct full state
   - Each delta is tiny (~1-2 KB) instead of 40-80 KB

3. **Auto-pruning**:
   - `dizz snapshot prune --keep 10` keeps last 10 checkpoints + recent deltas
   - Older deltas can be squashed into periodic full snapshots

### Phase 5: Agent Protocol File

**Goal**: A single `.dizz/state.ton` that is the canonical agent-readable project state.

1. Generated on every analysis - always up to date, always compact
2. Contains: project info, active intents, symbol health map, active todos, recent git context, snapshot pointers
3. Consumed by: `dizz context` (just dumps this file), or agents read it directly
4. Signed with content hash so agents can detect staleness

---

## Token Savings Analysis

| Data | Before | After | Savings |
|---|---|---|---|
| Intent file | 4.6 KB | ~400 B | ~91% |
| Snapshot object | 60 KB avg | ~25 KB avg | ~58% |
| Context dump | ~97 KB (full state) | ~2-3 KB | ~97% |
| Config file | 209 B | unchanged | 0% |

For an agent session reading context once and storing intents once, this saves **~100 KB per session**. At typical LLM token rates, that is hundreds of tokens saved per interaction.

---

## Agent Skill Structure

### Project-Level (created by `dizz init`)

```
.agents/skills/dizz/
  skill.json    - Metadata (name, version, commands, project name)
  SKILL.md      - Instructions for agents on using dizz in this project
```

### Global (created by install scripts)

```
~/.agents/skills/dizz/       (Linux/macOS)
%LOCALAPPDATA%\.agents\skills\dizz\   (Windows)
  skill.json    - Metadata (global=true, commands)
  SKILL.md      - Instructions for agents
```

The skill system makes dizz discoverable by AI agents. Agents scan `.agents/skills/` to find available tools and their usage instructions.

---

## Files Changed

| File | Change |
|---|---|
| `internal/store/ton/ton.go` | TON codec (Writer, Reader, escaping) |
| `internal/state/model_ton.go` | `IntentState.MarshalTON()`, `UnmarshalIntentStateTON()` |
| `internal/store/intentstore.go` | Default to `intent.ton`, fall back to `intent.json` |
| `cmd/snapshot.go` | Compact JSON (`json.Marshal` not `json.MarshalIndent`) |
| `cmd/context.go` | New `dizz context` command |
| `internal/render/context.go` | Context renderer with TON sections |
| `cmd/init.go` | Create `.agents/skills/dizz/` on init (no flag needed) |
| `internal/defaults/defaults.go` | Skill templates, gitignore updates |
| `site/scripts/install.sh` | Global skill dir `~/.agents/skills/dizz/` |
| `site/scripts/install.ps1` | Global skill dir `%LOCALAPPDATA%\.agents\skills\dizz\` |
| `AGENTS.md` | Updated with context workflow, TON docs |

---

## Branch

```
feat/agent-skill
```

---

## Principles

1. **Backward compatible at read** - old JSON files must still load
2. **Agent-first by default** - no flags needed for agent optimization
3. **Token efficiency first** - optimize for LLM context window
4. **Single entrypoint** - `dizz context` is the one command agents need
5. **Discoverable** - `.agents/skills/` makes the skill findable by any agent
