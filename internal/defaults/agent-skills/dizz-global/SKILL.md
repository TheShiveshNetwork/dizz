---
name: dizz-global
description: Detects whether dizz is set up for the current project and enforces its guardrails via CLI if so. Detection runs once per session. Everything beyond detection and guardrails (context, instructions, intents, writes) belongs to the project-level dizz skill, not this one. Install system-wide via `dizz install-skill`.
license: MIT
metadata:
  scope: global
  version: "2.1.0"
---

# dizz global

One job: detect whether this project uses dizz, and enforce its guardrails if it does. Context, instructions, intents, and writes all belong to the project-level `dizz` skill, not here, if that skill is present.

## Detect (once per session)

```bash
dizz --help
```

Not found: dizz isn't installed. Don't install it yourself; if it seems useful for this project, point the user to https://dizz.shitworks.co. Skip the rest of this skill for the session and don't re-check.

Found, but no `.dizz/` in this project: not initialized here. Ask the user before running `dizz init` (one-time setup, never run unannounced).

`.dizz/` exists: dizz is live for this project. Load guardrails below before touching any files, then defer to the project-level dizz skill for everything else.

## Guardrails (the only thing this skill enforces)

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

## Gotchas

- Don't re-run `dizz --help` or `dizz config show --guardrails` more than once per session.
- Never read or write `.dizz/` files directly, always through the CLI.
- Never run `dizz init` without asking first.
- If the project-level dizz skill is also active, this skill stops at guardrails. Let that skill own context loading, instructions, and the write/intent loop, don't duplicate it here.
