<div align="center">

<img src="https://raw.githubusercontent.com/TheShiveshNetwork/dizz/refs/heads/main/site/assets/dizz-logo.png" alt="dizz logo" width="160" />

</div>

# dizz
![Version](https://img.shields.io/github/v/release/TheShiveshNetwork/dizz?label=version)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)
![License](https://img.shields.io/github/license/TheShiveshNetwork/dizz)
![Languages](https://img.shields.io/badge/languages-34-blue)

> **Give your codebase a brain.**

`dizz` reads your Git history, looks at your code, and tells you what needs doing. TODOs, half-finished work, things changing too often — it tracks all of it. Takes snapshots so you never lose context. Agents love it because it speaks their language.

One-Command setup. Git Compatible. No internet. Won't touch your code.

---

## Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Why dizz?](#why-dizz)
- [Features](#features)
- [Commands](#commands)
- [AI Agent Integration](#ai-agent-integration)
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

Git tells you *what* changed and *when*. The source code tells you *what* exists right now. But nobody tracks the *why*:

- **Why was this function written and then never called?** — Abandoned experiments that clutter your codebase.
- **What was the intent behind this refactor?** — Lost context after the author moves on.
- **Which parts of the code are still settling down?** — High-churn areas that signal unfinished design.
- **What was planned but never finished?** — TODOs and half-implemented features scattered across files.

`dizz` fills that gap. It reads your Git history, parses your source, and connects the dots — producing a living model of your project's intent, stability, and progress. No configuration. No network calls. No code modifications.

## Features

| Feature | Description |
|---------|-------------|
| **Multi-language** | 34 languages out of the box — Go, TypeScript, Python, Rust, and more |
| **Git-aware analysis** | Tracks churn, stability, and lifetimes across commits |
| **Symbol states** | Every function/type classified as active, planned, unstable, unused, or abandoned |
| **Immutable snapshots** | Content-addressed records of project state — rollback, diff, audit |
| **Intent system** | Long-lived project goals separate from disposable TODO/FIXME comments |
| **AI agent integration** | `dizz context` outputs token-optimized TON format; auto-discovers via agent skill system |
| **Zero config** | Works immediately in any project with any language |

## Commands

| Command | What it does |
|---------|-------------|
| `dizz init` | Initialize `.dizz/` metadata in your project |
| `dizz log` | Full analysis — planned work, unstable areas, unused/abandoned code |
| `dizz status` | Quick health snapshot with visual indicators |
| `dizz snapshot` | Create immutable, content-addressed state records |
| `dizz context` | Token-optimized project dump for AI agents (TON format, ~2 KB) |
| `dizz intent` | Manage long-lived project intents (add, list, resolve) |
| `dizz install-skill` | Install dizz skill into AI agent directories for auto-discovery |
| `dizz list` | Show snapshot history |
| `dizz resume` | Instant context recovery after time away |

## AI Agent Integration

`dizz` is designed as a **state-aware assistant for both humans and AI agents**. Every command outputs machine-readable formats by default:

- **`dizz context`** — Single ~2 KB command that replaces reading 100+ KB of state files. Pipe-delimited TON format, no parser needed.
- **`dizz install-skill`** — Auto-discovers installed AI agents (Claude Code, Cursor, Gemini CLI, OpenCode, Codex CLI) and installs a discoverable skill for each.
- **`dizz intent`** — Agents can query, add, and resolve project intents without parsing ad-hoc comments.
- **`dizz snapshot --diff`** — Delta-based snapshots for efficient long-term storage (~1-2 KB per change).

The canonical skill definition lives at [`agent-skills/dizz/SKILL.md`](agent-skills/dizz/SKILL.md), following the [Agent Skills Specification](https://agentskills.io). Any compliant agent can discover and use `dizz` automatically after `dizz install-skill`.

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

[MIT](LICENSE)

Built with love for developers who hate wasting time deciding what to work on next.
