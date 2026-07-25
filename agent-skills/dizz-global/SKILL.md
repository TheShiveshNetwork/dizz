---
name: dizz-global
description: Detects whether dizz is initialized in the current project and, if so, enforces its guardrails via CLI. Detection runs once per session. Defers instructions, intent workflow, and full context handling to the project-level dizz skill. Install system-wide via `dizz install-skill`.
license: MIT
metadata:
  scope: global
  version: "1.0.0"
---

# dizz Global Skill

dizz is a project's persistent memory — intents, code health, symbol states — built up across sessions instead of a static file that goes stale. If a project has dizz initialized, treat its state as that project's primary source of truth over your own assumptions.

This skill has exactly one job: detect dizz, and enforce its guardrails. Everything else — loading context, querying instructions, managing intents, the self-growing memory loop — lives in the project-level dizz skill, if the skill is not there but dizz is initialized, ask the user to create the skill [read Links#3]. Keep the two separate; don't duplicate work across them.

## Detection — once per session

At the start of a session, check once:

```bash
dizz --help
```

If the command isn't found, dizz isn't installed. Skip it for the rest of this session and don't check again — don't re-run this on every message.

If dizz is available but `.dizz/` doesn't exist in this project, it hasn't been initialized here. Ask the user before running:

```bash
dizz init
```

Never initialize without asking first — it's a one-time setup.

If `.dizz/` exists, dizz is live for this project. Load its guardrails before making any file changes.

## Guardrails — this skill's only responsibility once dizz is detected

```bash
dizz config show --guardrails
```

Guardrails are non-negotiable and override your own defaults. Each guardrail has: `id|scope|action|reason`.

- **id** — stable identifier (e.g. `gr-generated-code`)
- **scope** — `global` (repo-wide), `file` (single path/glob), or `group` (set of paths)
- **action** — what you must do:
  - `read_only` — never modify files matching this path
  - `require_review` — changes need explicit user approval
  - `warn` — proceed, but flag it
  - `skip` — ignore this path entirely in analysis
  - `forbid` — hard block (e.g. deletions)
- **reason** — why this guardrail exists

Optional fields: `paths` (globs), `match` (any|all for groups), `ops` (restrict to write/delete), `severity` (0-3), `exceptions` (exempt globs).

Check every file you're about to touch against the loaded guardrails before making changes.

## Rules

1. **Never read `.dizz/` files directly** — always go through CLI commands.
2. **Never write to `.dizz/` yourself** — it's dizz's internal state; the project-level skill owns the ask-before-write workflow for intents.
3. **Never initialize dizz without asking the user first.**
4. **Enforce guardrails strictly**, every time, before any file change.
5. **If a project-level dizz skill is also present, leave everything beyond detection and guardrails to it** — don't re-explain instructions, commands, or the intent workflow here.

## Links

1. https://github.com/TheShiveshNetwork/dizz
2. https://github.com/TheShiveshNetwork/dizz/tree/main/agent-skills
3. https://github.com/TheShiveshNetwork/dizz/tree/main/agent-skills/SKILL.md
