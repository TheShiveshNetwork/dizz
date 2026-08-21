# Dizz

> **Give your codebase a brain.**

Every AI coding agent starts cold. It can't tell abandoned experiments from production code, misses TODOs scattered across your files, and trusts docs that went stale the day after they were written.

`dizz` gives agents a live model of your project instead. It re-derives context directly from your codebase and Git history on every run - surfacing abandoned code, unstable areas, intents, and open TODOs - and distills 141KB of raw state into a 1.8KB read. Your agent gets current, signal-rich orientation in ~450 tokens, so it spends its budget on the task, not on re-discovering the codebase.

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

* `--all, -a` - include healthy symbols
* `--verbose, -v` - detailed reasoning
* `--dump, -d` - dump every symbol including active ones, with full details
* `--filter <state>` - filter by state (repeatable): `active`, `planned`, `unused`, `unstable`, `abandoned`

```bash
dizz log                    # what needs attention
dizz log --all              # include healthy symbols
dizz log --filter=unused    # only unused code
dizz log --dump             # every symbol, full detail
```

### `dizz status`

Quick health snapshot with visual indicators. Shows how many symbols are in each state (active, planned, unstable, unused, abandoned) and active TODO count.

### `dizz context`

Token-optimized project dump designed for AI agents. Outputs everything an agent needs in ~2 KB using TON format (pipe-delimited, one record per line). No parser needed - split on `|`.

**Sections:**

* Project info (name, git branch, commit)
* Active intents (goals, tasks, questions)
* Unstable symbols (high churn)
* Unused / abandoned symbols (dead code candidates)
* Active TODOs from source code
* Recent snapshot hashes

**Flags:**

* `--intents` - show intents only
* `--symbols` - show symbols only
* `--todos` - show todos only

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

* `--auto` - for Git hooks (non-interactive)
* `--diff` - store only the delta from the last snapshot

**Subcommands**

* `dizz snapshot list` - show all snapshots and deltas
* `dizz snapshot checkout <hash>` - reconstruct and inspect a past snapshot
* `dizz snapshot prune [--keep N]` - remove old snapshots, keep the N most recent checkpoints (default 10)
* `dizz snapshot create` - same as `dizz snapshot`

```bash
dizz snapshot                   # full snapshot
dizz snapshot --diff            # delta snapshot
dizz snapshot --auto            # non-interactive (git hook)
dizz snapshot list
dizz snapshot checkout a1b2c3d4
dizz snapshot prune --keep 5
```

### `dizz list`

Shows snapshot history and project evolution.

### `dizz resume`

Instant context recovery after time away.

Optimized for:

> "I haven't touched this project in weeks."

### `dizz intent`

Manage human-authored intent separate from source-level TODO/FIXME comments.

Intents are immutable project goals. Unlike TODOs (which are re-scanned on every analysis), intents persist in `.dizz/intent.ton` until explicitly resolved. They carry a severity score from 0-3.

**Subcommands**

* `dizz intent add "message"` - add a new intent
  * `--type <type>` - intent type (default `todo`)
  * `--severity <0-3>` - severity level (default 1)
  * `--tags <tags>` - comma-separated tags
* `dizz intent list` - list active intents
  * `--all` - include resolved and closed intents
  * `--severity <0,1,2,3>` - filter by severity
  * `--type <types>` - filter by type
* `dizz intent resolve <id>` - mark an intent as fixed/resolved
  * `--note "text"` - optional resolution note
* `dizz intent close <id>` - close an intent (wontfix, duplicate, no longer relevant)
  * `--note "text"` - optional closing note

```bash
dizz intent add "Refactor auth layer" --severity 2 --type refactor
dizz intent add "Fix flaky test" --tags "test,flaky" --type fixme
dizz intent list
dizz intent list --all --type todo
dizz intent resolve int_1770020361 --note "done in PR #42"
dizz intent close int_1770020361 --note "superseded"
```

Types: `todo`, `fixme`, `refactor`, `question`, `hack`, `temporary`

### `dizz install-skill`

Scans your system for installed AI agents and installs the dizz skill into their skill directories. After this, agents like Claude Code, Cursor, Gemini CLI, and OpenCode can discover and invoke dizz automatically.

```bash
dizz install-skill                  # install to all detected agents
dizz install-skill --provider opencode   # install to one agent only
```

**Flags**

* `--provider, -p <provider>` - install to a specific provider only: `agents`, `claude-code`, `claude-desktop`, `copilot`, `cursor`, `gemini-cli`, `gemini-config`, `antigravity`, `opencode`

Supported agents: Claude Code, Cursor, Gemini CLI, OpenCode, Codex CLI / Copilot.

The canonical skill definition lives in the repository at `agent-skills/dizz/SKILL.md` and follows the [Agent Skills Specification](https://agentskills.io).

### `dizz config`

Inspect and manage the project configuration in `.dizz/config.json`.

**Subcommands**

* `dizz config show` - show the current config
  * `--json` - compact JSON output (optimized for agents)
  * `--name, -n` - project name only
  * `--description, -d` - description only
  * `--instructions, -i` - instructions only
  * `--guardrails, -g` - guardrails only
  * `--commands, -c` - commands only
  * `--severity-scale` - severity scale only
  * `--agent-defaults` - agent defaults only
  * `--links` - links only
  * `--version` - config version only
  * `--include` - include patterns only
  * `--exclude` - exclude patterns only
* `dizz config add-instruction --rule "<rule>" [--scope "<glob>"]` - add a coding instruction (glob-scoped or global)
* `dizz config add-guardrail` - add an enforceable guardrail
  * `--action <action>` (required) - `read_only` | `require_review` | `warn` | `skip` | `forbid`
  * `--reason "<reason>"` (required) - human-readable reason
  * `--id <id>` - stable identifier (e.g. `gr-generated-code`)
  * `--paths "<glob>"` - glob patterns (repeatable; omit for global)
  * `--require-all` - fire only when ALL paths are touched together
* `dizz config set-description "<text>"` - set the agentic description

```bash
dizz config show
dizz config show --json --guardrails
dizz config add-instruction --rule "Run tests before merge"
dizz config add-instruction --rule "No class components" --scope "*.tsx"
dizz config add-guardrail --action forbid --reason "no force-push"
dizz config add-guardrail --id gr-gen --paths "generated/**" --action read_only --reason "auto-generated"
dizz config set-description "Payment platform"
```

### `dizz graph`

Derives a project knowledge graph purely from persisted dizz state (`state.ton.gz`, `intent.ton`, the per-file signal cache, snapshots, and config) plus opt-in Git history. No code analysis is performed and the graph is never materialized - every invocation re-derives it in-memory, so it is always live. Run `dizz context` or `dizz log` at least once so state exists.

**Flags** (apply to every graph subcommand)

* `--cochange` - include co-change coupling analysis (requires git)
* `--min-jaccard <float>` - minimum Jaccard similarity for co-change edges (default 0.3)
* `--min-commits <n>` - minimum commits a file must appear in for co-change analysis (default 3)
* `--max-commits <n>` - maximum git history depth for co-change analysis (default 1000)
* `--depth <n>` - traversal depth for blast radius (default 3)
* `--file <path>` - disambiguate a symbol by its file path
* `--json` - machine-readable JSON output
* `--similarity-threshold <float>` - minimum text similarity for RELATED_TO intent edges (default 0.2)
* `--similarity-topk <n>` - max RELATED_TO edges per intent (default 6)

**Subcommands**

* `dizz graph build` - derive the graph and print its shape
* `dizz graph stats` - summarize graph shape (nodes, edges, node/edge types)
* `dizz graph query <entity>` - blast radius: everything affected by changing an entity, scored by depth
* `dizz graph trace <entity>` - full trace: callers, callees, imports, tests, intents, TODOs, co-changes
* `dizz graph tests <entity>` - which tests cover a symbol or file (with match confidence)
* `dizz graph cochanges <entity>` - files that historically change together with the entity (requires git)
* `dizz graph path <from> <to>` - shortest path between two entities
* `dizz graph scope <glob>` - files matching a path glob (e.g. `internal/auth/**`)
* `dizz graph ton` - dump the full graph in TON format
* `dizz graph dump` - dump the graph as JSON for visualizers
* `dizz graph visualize` - serve the interactive 3D web view (same as `dizz visualize`)

Entities resolve by name (`main`, `ParseConfig`) or unambiguously with `symbol:<name>@<file>`, `file:<path>`, `intent:<id>`, `test:<name>`, `dep:<import>`, `commit:<hash>`, `guardrail:<id>`, `snapshot:<hash>`, and `module:<name>` prefixes.

```bash
dizz graph build
dizz graph query main --depth 3
dizz graph trace ParseConfig
dizz graph tests AuthMiddleware
dizz graph cochanges main --cochange --min-jaccard 0.4
dizz graph path "file:cmd/root.go" "file:internal/state/scorer.go"
dizz graph scope "internal/state/**" --json
```

### `dizz visualize`

Serves an interactive 3D force-directed web view of the knowledge graph on `127.0.0.1` and opens it in the default browser. Explore the project as a living map: files, symbols, imports, test coverage, and intent similarity links (intents connected by text similarity, computed with IDF-weighted cosine over stemmed tokens).

**Flags**

* `--port <n>` - port for the web visualizer (0 picks a free port)
* `--open` - open the browser automatically (default true)
* `--cochange` - include co-change coupling analysis (requires git)
* `--similarity-threshold <float>` - minimum text similarity for RELATED_TO edges (default 0.2)
* `--similarity-topk <n>` - max RELATED_TO edges per intent (default 6)

```bash
dizz visualize                    # open in browser on a free port
dizz visualize --port 8080       # pick a port
dizz visualize --open=false      # print the URL only, don't launch a browser
dizz visualize --cochange        # include git co-change coupling
dizz graph visualize             # identical command under the graph family
```

The view renders client-side with no server round-trips after load: nodes group by type (file, symbol, intent, test), colored by symbol state, and clicking a node focuses its neighbors.

### `dizz todo`

Lists TODO, FIXME, and other markers found in your source code during analysis.

```bash
dizz todo list
```

`dizz todos` is an alias.

### `dizz version`

Shows the current dizz version.

```bash
dizz version
```

### `dizz upgrade`

Downloads and installs the latest dizz release binary in place.

```bash
dizz upgrade
```

### `dizz ui`

Launches the interactive terminal UI (`dizzie`). If dizzie is not installed, it is downloaded and installed automatically first.

```bash
dizz ui
```

## AI Agent Integration

dizz is built for AI agents from the ground up. Key features:

**`dizz context`** - A single ~2 KB command that replaces reading 100+ KB of state files. Outputs in TON format (Token-Optimized Notation): pipe-delimited, line-oriented, no JSON overhead. Any agent can parse it by splitting on `|`.

**`dizz install-skill`** - Auto-discovers installed AI agents and installs a skill for each. The skill tells agents what dizz does and how to use it. Agents discover skills automatically via the `.agents/skills/` convention.

**`dizz intent`** - Agents can query, add, and resolve project intents programmatically. No need to parse ad-hoc TODO comments in source files.

**TON Format** - All intent data uses TON:
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
