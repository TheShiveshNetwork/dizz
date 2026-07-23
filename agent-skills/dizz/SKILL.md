---
name: dizz
description: Project-level dizz skill. Tracks intents, code health, and symbol states for this specific project. Use to understand project state, find work items, or detect dead code.
license: MIT
metadata:
  scope: project
  version: "1.0.0"
---

# dizz Project Skill

dizz is this project's memory layer. It tracks intents, code health, and symbol states - information that survives across sessions and would otherwise be lost between conversations. It is read-only and works in any repo.

## Quick Start

Run `dizz context` to get a token-optimized project summary. This is the primary entry point for agents.

## Project Config

This project's agent rules live in `.dizz/config.json`. Run `dizz config show` to read them.

Key fields:

- `description` - project summary
- `instructions` - agent rules scoped to file patterns (e.g., `rule@scope`)
- `guardrails` - protected paths with actions (e.g., `path|action|reason`)
- `commands` - project-specific commands (build, test, lint, etc.)

To add rules:

```
dizz config add-instruction --rule "Run tests before merge" --scope "internal/**"
dizz config add-guardrail --path "generated/**" --action "read_only" --reason "auto-generated"
dizz config set-description "Project description here"
```

## Commands

- `dizz context` - Token-optimized project context for agents
- `dizz config show` - Read project agent config
- `dizz intent list` - View active intents (what needs doing)
- `dizz status` - Project health overview (unstable, unused, abandoned symbols)
- `dizz log` - Detailed symbol health and todos
- `dizz snapshot --auto` - Record current state before making changes
- `dizz intent add "msg" --type todo` - Add a new intent

## Data Format

All intent data is stored in `.dizz/intent.ton` (Token-Optimized Notation) -
a line-oriented, pipe-delimited format readable by any agent without a parser.
Split on `|` to read. No JSON parser needed.
