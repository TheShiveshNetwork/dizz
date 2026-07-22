---
name: dizz
description: State-aware project assistant. Tracks intents, code health, and symbol states. Use to understand project state, find work items, or detect dead code.
license: MIT
metadata:
  project: dizz
  version: "1.0.0"
---

# dizz Skill for dizz

dizz tracks project intents, code health, and symbol states.

## Quick Start

Run dizz context to get a token-optimized project summary.

## Commands

- dizz context - Token-optimized project context for agents
- dizz intent list - View active intents (what needs doing)
- dizz status - Project health overview (unstable, unused, abandoned symbols)
- dizz log - Detailed symbol health and todos
- dizz snapshot --auto - Record current state before making changes
- dizz intent add "msg" --type todo - Add a new intent

## Data Format

All intent data is stored in .dizz/intent.ton (Token-Optimized Notation) -
a line-oriented, pipe-delimited format readable by any agent without a parser.
Split on | to read. No JSON parser needed.
