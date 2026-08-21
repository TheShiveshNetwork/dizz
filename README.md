<div align="center">
<img src="https://raw.githubusercontent.com/TheShiveshNetwork/dizz/refs/heads/main/site/assets/dizz-logo.png" alt="dizz logo" width="160" />
</div>

# dizz

![Version](https://img.shields.io/github/v/release/TheShiveshNetwork/dizz?label=version)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)
![License](https://img.shields.io/github/license/TheShiveshNetwork/dizz)
![Languages](https://img.shields.io/badge/languages-34-blue)

> **Give your codebase a brain.**

Every AI coding agent starts cold. It can't tell abandoned experiments from production code, misses TODOs scattered across your files, and trusts `AGENTS.md` files that went stale the day after they were written.

`dizz` gives agents a live model of your project instead. It re-derives context directly from your codebase and Git history on every run - surfacing abandoned code, unstable areas, intents, and open TODOs - and distills 141KB of raw state into a 1.8KB read. Your agent gets current, signal-rich orientation in ~450 tokens, so it spends its budget on the task, not on re-discovering a codebase from scratch.

One-command setup. Git compatible. No internet. Won't touch your code.

---

## Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Why dizz?](#why-dizz)
- [Features](#features)
- [Commands](#commands)
- [Knowledge Graph](#knowledge-graph)
- [AI Agent Integration](#ai-agent-integration)
- [What dizz is not](#what-dizz-is-not)
- [Symbol States](#symbol-states)
- [Supported Languages](#supported-languages)
- [Documentation](#documentation)
- [License](#license)

---

## Installation

```bash
# Linux & macOS
curl -fsSL https://dizz.shitworks.co/install.sh | bash

# Windows (PowerShell)
powershell -c "irm https://dizz.shitworks.co/install.ps1 | iex"
```

Then enable AI agent discovery:

```bash
dizz install-skill
```

Verify:

```bash
dizz version
```

## Quick Start

```bash
cd your-project
dizz init      # set up .dizz/ metadata
dizz status    # project health overview
dizz log       # what needs attention
```

## Why dizz?

> **Git tracks every change. dizz tracks the story behind them.**

Git tells you *what* changed and *when*. The source code tells you *what* exists right now. But nobody tracks the *why* — and that's exactly the part an agent (or a human coming back after a break) actually needs:

- **Why was this function written and then never called?** — Abandoned experiments that clutter your codebase.
- **What was the intent behind this refactor?** — Lost context after the author moves on.
- **Which parts of the code are still settling down?** — High-churn areas that signal unfinished design.
- **What was planned but never finished?** — TODOs and half-implemented features scattered across files.

`dizz` fills that gap. It analyzes your source code, scores it using Git history, and connects the dots — producing a living model of your project's intent, stability, and progress that stays current every time you run it, instead of decaying the way a hand-written context file does. No configuration. No network calls. No code modifications.

## Features

| Feature | Description |
|---------|-------------|
| **Multi-language** | 34 languages out of the box — Go, TypeScript, Python, Rust, and more |
| **Git-aware analysis** | Tracks churn, stability, and lifetimes across commits |
| **Symbol states** | Every function/type classified as active, planned, unstable, unused, or abandoned |
| **Immutable snapshots** | Content-addressed records of project state — rollback, diff, audit |
| **Intent system** | Long-lived project goals, severity-scored, kept separate from disposable TODO/FIXME comments — an intent can't quietly vanish the way a comment does when someone deletes a line, so an agent can trust it's still accurate |
| **Knowledge graph** | Derives a live project graph from persisted state (symbols, files, imports, test coverage, intent similarity), queryable via `dizz graph` and explorable in an interactive 3D web view |
| **AI agent integration** | `dizz context` outputs a compact, token-optimized format (TON — Token-Optimized Notation); auto-discovers via agent skill system |
| **Zero config** | Works immediately in any project with any language |

## Commands

| Command | What it does |
|---------|-------------|
| `dizz init` | Initialize `.dizz/` metadata in your project |
| `dizz log` | Full analysis — planned work, unstable areas, unused/abandoned code |
| `dizz status` | Quick health snapshot with visual indicators |
| `dizz snapshot` | Create immutable, content-addressed state records |
| `dizz context` | Token-optimized project dump for AI agents (TON format, ~2 KB) |
| `dizz config` | Manage and inspect `.dizz/config.json` (supports filtering flags) |
| `dizz intent` | Manage long-lived project intents (add, list, resolve) |
| `dizz graph` | Query the derived project knowledge graph (see [Knowledge Graph](#knowledge-graph)) |
| `dizz visualize` | Serve an interactive 3D web view of the graph in your browser |
| `dizz install-skill` | Install dizz skill into AI agent directories for auto-discovery |
| `dizz list` | Show snapshot history |
| `dizz resume` | Instant context recovery after time away |

## Knowledge Graph

`dizz graph` derives a project knowledge graph purely from persisted dizz state
(`state.ton.gz`, `intent.ton`, the per-file signal cache, snapshots, and config),
plus opt-in Git history. No code analysis happens here and the graph is never
materialized, so every invocation re-derives it in-memory and it is always live.
Run `dizz context` or `dizz log` at least once so state exists.

```bash
dizz graph build         # summary of graph shape (nodes, edges, node/edge types)
dizz graph query main    # blast radius of changing an entity
dizz graph trace main    # callers, callees, tests, intents, co-changes of an entity
dizz graph tests main    # which tests cover a symbol or file
dizz graph path main Foo # shortest path between two entities
dizz graph scope "internal/auth/**"   # all files matching a path glob
dizz graph cochanges main             # files that historically change together (requires git)
```

Subcommands | What it does
------------|-------------
`graph build` / `graph stats` | Summarize the graph shape
`graph query <entity>` | Blast radius: everything affected by changing an entity, scored by depth
`graph trace <entity>` | Full trace: callers, callees, imports, tests, intents, TODOs, co-changes
`graph tests <entity>` | Which tests cover a symbol or file (with match confidence)
`graph cochanges <entity>` | Hidden coupling: files that historically change together (requires git, use `--cochange`)
`graph path <from> <to>` | Shortest path between two entities
`graph scope <glob>` | All files matching a path glob
`graph ton` / `graph dump` | Dump the full graph in TON or JSON format
`graph visualize` | Same as `dizz visualize`

Entities resolve by name (`main`, `ParseConfig`), or unambiguously with
`symbol:Name@path` / `file:path` / `intent:id` prefixes. All graph subcommands
accept `--json` for machine-readable output.

### Interactive visualization

`dizz visualize` (or `dizz graph visualize`) serves an interactive 3D
force-directed view of the graph on `127.0.0.1` and opens it in your default
browser. Explore the project as a living map: files, symbols, imports, test
coverage, and intent similarity links (intents connected by text similarity,
computed with IDF-weighted cosine over stemmed tokens).

```bash
dizz visualize                        # open in browser on a free port
dizz visualize --open=false           # just print the URL, don't launch a browser
dizz visualize --port 8080            # pick a specific port
dizz visualize --cochange             # include git co-change coupling analysis
dizz visualize --similarity-threshold 0.3 --similarity-topk 4   # tune intent links
```

The view renders client-side with no server round-trips after load: nodes group
by type (file, symbol, intent, test), colored by symbol state, and you can click
a node to focus on its neighbors.

## AI Agent Integration

`dizz` is designed as a **state-aware assistant for both humans and AI agents** — the layer an agent reads from instead of a static context file. Every command outputs machine-readable formats by default:

- **`dizz context`** — Single command optimized for agents first runs to follow up analysis. Pipe-delimited TON format, no parser needed.
- **`dizz config`** — Single source of truth for persistent agentic project guidance. Supports filtering flags (`--only-description`, `--only-instructions`, `--only-guardrails`, `--only-commands`, `--json`).
- **`dizz install-skill`** — Installs a global skill for all agent harnesses. Auto-discovers installed AI agents (Claude Code, Cursor, Gemini CLI, OpenCode, Codex CLI) and installs a discoverable skill for each.
- **`dizz intent`** — Agents can query, add, and resolve project intents directly, without parsing ad-hoc comments or relying on a doc someone forgot to update.
- **`dizz snapshot --diff`** — Delta-based snapshots for efficient long-term storage, so an agent picking the project back up gets the *change* since it last looked, not the whole state again.

## Symbol States

Every symbol in your codebase is classified into one of five states:

| State | Meaning |
|-------|---------|
| `active` | Used and stable |
| `planned` | Has TODO / intent marker |
| `unstable` | High churn (frequent changes) |
| `unused` | Declared but never called |
| `abandoned` | Old and unused |

Derived from usage signals, intent markers, and Git history.

## Supported Languages

`dizz` supports 34 languages through a tiered analysis pipeline:

| Tier | Accuracy | Languages |
|------|----------|-----------|
| AST | Highest | Go |
| Lexical | Good | JavaScript, TypeScript, Python, Rust, Java, Kotlin, Swift, C#, Ruby, PHP, Scala, Lua, Elixir, Julia, Dart, Nim, Zig, Clojure, Erlang |
| Regex | Fair | C, C++, Shell, Haskell, R, Perl, OCaml, F#, SQL, MATLAB, Terraform |

New languages are added by registering a `LanguageConfig` — no core changes needed.

## Documentation

- [Full documentation](https://dizz.shitworks.co/docs)
- [Contributing](CONTRIBUTING.md)
- [Agent guidelines](AGENTS.md)
- [Performance benchmarks](benches.md)
- [Setup guide](SETUP.md)

## License

[Apache-2.0](LICENSE)

Built with love for developers — and their agents — who hate wasting time deciding what to work on next.
