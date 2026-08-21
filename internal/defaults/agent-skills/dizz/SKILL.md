---
name: dizz
description: This project's persistent memory - intents, conventions, guardrails, code health, symbol relationships. Use this whenever you need project context (conventions, past decisions, TODOs, guardrails), before making non-trivial changes (check blast radius/coupling/tests via graph queries), or after learning something durable about the project (write it back via config/intent commands instead of letting it live only in chat). Query state through the CLI instead of reading files or code cold.
license: MIT
metadata:
  scope: project
  version: "1.1.0"
---

# Project memory

This project's history, conventions, guardrails, and open work live in a CLI-queryable store. Treat CLI output as ground truth over what you infer from reading code. This replaces re-deriving conventions and history from scratch, not reading the file you're actually about to edit.

If `dizz` is not found, tell the user and point them to https://dizz.shitworks.co for install instructions; don't fall back to reading `.dizz/` directly. If it's installed but this project isn't initialized, suggest `dizz init`.

## Session start (once per session)

```bash
dizz context
```

Run this once at the start of a session and hold the result for the rest of it. Do not re-run `dizz context` or list commands again mid-session unless the project state has actually changed (new commits, new files, a resolved intent) or a fresh session has started. Re-querying unchanged state wastes tokens for no new information.

## Guardrails (enforce this on every run)

```bash
dizz config show --guardrails
```

Run once per session and hold the result; don't re-fetch mid-session.

Each guardrail is `id|scope|action|reason`. Scope is `global`, `file`, or `group`. Action:

- `read_only` - never modify matching files
- `require_review` - needs explicit user approval before changing
- `warn` - proceed, but flag it
- `skip` - ignore in analysis
- `forbid` - hard block

Check every file you're about to touch against the loaded guardrails first. Guardrails override your own defaults, no exceptions.

## Query before reading code cold

| Need | Command |
|---|---|
| Conventions / rules | `dizz config show -i` |
| Guardrails / protected paths | `dizz config show -g` |
| Project commands / scripts | `dizz config show -c` |
| Everything | `dizz config show` |
| Per-symbol detail | `dizz log` |
| Discovered TODOs | `dizz todo` |

Flags combine, e.g. `dizz config show -i -g`. Add `--json` to any command for compact output.

## Write durable knowledge back (propose first, always)

Any convention, rule, guardrail, decision, or intent you discover while working gets written back through these commands, never left only in chat.

| Learned | Command |
|---|---|
| Coding convention | `dizz config add-instruction --rule "..."` |
| Rule scoped to files | `dizz config add-instruction --rule "..." --scope "*.tsx"` |
| Hard constraint / protected path | `dizz config add-guardrail --action forbid --reason "..."` |
| Review-gated paths | `dizz config add-guardrail --paths "..." --require-all --action require_review --reason "..."` |
| Project description | `dizz config set-description "..."` |
| New work item | `dizz intent add "..." --type <TYPE> --severity <0-3>` (TYPE: todo, fixme, refactor, question, hack, temporary) |
| Resolved work item | `dizz intent list` then `dizz intent resolve INTENT_ID` |

Guardrail actions: `read_only | require_review | warn | skip | forbid`.

## Graph queries (use instead of tracing imports by hand)

| Need | Command |
|---|---|
| Blast radius of a change | `dizz graph query <entity> --depth 3` |
| Full context on an entity | `dizz graph trace <entity>` |
| Hidden coupling | `dizz graph cochanges <file>` |
| Test coverage | `dizz graph tests <entity>` |
| Path between two entities | `dizz graph path <from> <to>` |

Entity forms: `symbol:Name@file`, `file:path`, `intent:id`, or a bare unique name. All accept `--json`; add `--cochange` for git-history coupling.

## Before large changes

```bash
dizz snapshot --auto
```

## Gotchas

- Do not re-run `dizz context` or a help/list command more than once per session out of habit. Reuse what you already fetched.
- Never read or write `.dizz/` files directly, even to "just check." Always go through the CLI. If a command fails, surface the failure rather than falling back to reading the files.
- Every write command (config, guardrail, intent add/resolve) requires explicit user approval first, regardless of how confident or small the change seems.
