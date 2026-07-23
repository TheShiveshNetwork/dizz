---
name: dizz-global
description: Global dizz skill for AI agents. Detects dizz projects, reads config-backed agent guidance, and uses dizz commands for project state. Install system-wide via `dizz install-skill`.
license: MIT
metadata:
  scope: global
  version: "1.0.0"
---

# dizz Global Skill

dizz is a state-aware assistant for AI agents. It keeps track of project intents, code health, and symbol states - information that survives across sessions and would otherwise be lost between conversations. It is read-only and works in any repo.

This skill applies system-wide across every project. Each time you enter a project, determine whether dizz is relevant there before acting.

## Step 1: Check applicability

- If dizz is not installed on this machine, note that and move on - do not block the current task over it.
- If dizz is installed but this project has no `.dizz/` directory yet, ask the user before initializing it.
- If `.dizz/` already exists, proceed directly - no need to ask.

## Step 2: Read project agent rules from config

Project-level agent standards live in `.dizz/config.json`. Treat this as the source of truth for agent runs in this repository.

Run `dizz config show` to read the current config. Key fields:

- `project_name` - the project identifier
- `description` - human-readable project summary
- `instructions` - agent rules scoped to file patterns
- `guardrails` - protected paths with actions and reasons
- `commands` - project-specific commands (build, test, lint, etc.)
- `agent_defaults` - default analysis lens and severity thresholds

To update config:

```
dizz config add-instruction --rule "..." --scope "..."
dizz config add-guardrail --path "..." --action "..." --reason "..."
dizz config set-description "..."
```

## Step 3: Use dizz commands, not this file

This file does not list dizz commands or flags. The CLI changes after this was written, and a stale list leads to confident wrong invocations.

- Run `dizz --help` to see everything available right now.
- Run `dizz <command> --help` for specific flags.

Both are instant and local. Check rather than guess.

## When to reach for dizz

Once this project has dizz set up (Step 1), use dizz whenever:

- **Starting a session** - get current state before reading files or asking the user to recap.
- **Before and after making changes** - checkpoint state so nothing gets lost.
- **The user states a long-term goal or decision** - record it immediately so it outlives this conversation.
- **The user asks what to work on or what changed** - answer from dizz output, not a guess.
- **Returning after a gap** - recover context from dizz first.

## Failure modes

| Symptom | Cause | Fix |
|---|---|---|
| `command not found: dizz` | Not installed on this machine | Mention it, offer the install link, continue without it |
| No dizz project detected | This project has not been initialized yet | Ask the user before running `dizz init` |
| Output looks stale | Local state out of date | Take a fresh checkpoint with `dizz snapshot --auto` |

## Constraints

- **Read-only** - dizz never edits source code. Any actual fix or resolution is the agent's job.
- **Zero-config** by default, once a project has been initialized.
