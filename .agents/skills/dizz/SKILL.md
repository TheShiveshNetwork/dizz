---
name: dizz
description: Project-level dizz skill. The project's growing memory of intents, code health, and symbol state across sessions. Loads context on every run, answers targeted queries via config flags, and proposes (never silently makes) changes to the project's brain.
license: MIT
metadata:
  project: dizz
  version: "1.0.0"
---

# dizz Project Skill

dizz is this project's memory — not a snapshot, a running one. It tracks what needs doing, what's breaking, what's dead, and what's already been fixed, and it gets more accurate every time it's used. Treat its state as the source of truth for project state ahead of your own inference from reading code cold.

The global dizz skill handles detection and guardrails. This skill handles everything specific to working in this project: context, config, and the intent memory itself. Don't repeat the global skill's job here — read `.dizz-global`'s guardrail output when you need it, and stay focused on the project-level work below.

## Core rules

1. **Never read `.dizz/` files directly** — always use CLI commands.
2. **Never modify `.dizz/` files yourself** — dizz manages its own state.
3. **Use CLI for all data** — if a command fails, do not fall back to reading files directly; surface the failure instead.
4. **Never add, resolve, or otherwise write an intent without asking the user first.** This applies every time, with no exceptions — even when a fix looks obviously correct or a pattern is unambiguous. Propose it, wait for a yes.

## First session — load context once

Run this **only the first time** a new agent session starts in this project:

```bash
dizz context
```

This returns a token-optimized summary of active intents, symbol health, and git state. Use it to orient yourself before answering project-specific questions. Do not re-run `dizz context` on subsequent prompts within the same session — it stays loaded as baseline context. For follow-up queries, use the targeted commands in the table below instead.

## Config — pull only what the question needs

Use targeted flags after `dizz context` to query specific config sections. Add `--json` to any command for compact JSON output:

| User is asking about... | Command |
|---|---|
| Conventions / coding rules | `dizz config show --instructions` |
| Protected paths / guardrails | `dizz config show --guardrails` |
| Build / test / lint commands | `dizz config show --commands` |
| What severity levels mean | `dizz config show --severity` |
| Everything | `dizz config show` |

Flags combine: `dizz config show -i -g` loads instructions and guardrails together. Append `--json` to any of the above for machine-readable TON/JSON output.

## The self-growing memory loop

This is what makes dizz more than a static file: every fix that gets made can feed back into the project's own memory, so it gets more accurate over time instead of going stale the way a hand-written doc does. This loop is the actual "brain" — more than any single command.

**What dizz detects automatically:** `TODO:`, `FIXME:`, `REFACTOR:`, `HACK:` comments, unfinished features (`wip`), and missing tests. These appear as `planned` symbols until resolved or removed.

**Severity is yours to assign.** dizz does not set severities. When adding an intent, you pick 1-3 via `--severity SEV` based on how critical it is.

**The loop:**

1. **Detect** — notice a pattern above, or notice you've just fixed something, while working.
2. **Propose** — tell the user what you found (or fixed), and what you'd add or resolve, with type and severity attached. Never skip this step.
3. **Act only on approval:**
   - New intent: `dizz intent add "message" --type TYPE --severity SEV`
   - Resolved intent: check `dizz intent list` first to confirm it exists, then `dizz intent resolve INTENT_ID`
4. If a fix you just made doesn't match any existing intent, propose adding one that documents it — that's how the memory grows to reflect what actually happened in the project, not just what was originally planned.

**Hard rule, worth repeating because it matters most: propose, don't act.** Adding or resolving an intent, or writing anything into dizz's state, always requires the user's explicit go-ahead first — no exceptions for confidence, obviousness, or how small the change seems.

## Core commands

| Command | When to use |
|---|---|
| `dizz context` | First run of a new session only |
| `dizz status` | Quick health check before changes |
| `dizz log` | Deeper per-symbol detail |
| `dizz snapshot --auto` | Before/after a significant change |

Other commands: `intent`, `config`, `todo`, `init`, `list`, `resume`, `upgrade`, `version`, `install-skill`. Use `dizz <command> --help` to learn their usage.

## Guardrails

Guardrails are enforced by the global dizz skill on every file change. If you need the specifics for a task, load them directly:

```bash
dizz config show --guardrails
```
