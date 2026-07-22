---
name: dizz
description: Persistent project memory — long-term goals, dead/unstable code, and churn history that survive across sessions, in any codebase. At the start of work in any project, check whether dizz applies here; if so, use it before reading files or guessing from memory. Also reach for it whenever the user states a goal or decision to track, asks what's left to do or what changed, or returns after time away.
license: MIT
metadata:
  author: TheShiveshNetwork
  version: "1.0.1"
  scope: global
---

# dizz

dizz is a project's memory. It reads git history and source once and surfaces what an agent would otherwise burn tokens rediscovering: what's planned, what's breaking, what's dead, what happened last session. It's read-only — it never modifies source — and works in any repo, git or non-git.

This is the **global** install of the skill: it applies across every project, not just one that's already set up for dizz. That means the first thing to figure out, each time, is whether dizz is relevant *here* — don't assume.

## Step 1: Check applicability before doing anything else

- If dizz isn't installed on this machine, note that once and move on — don't block the current task over it. Installing it is a one-line decision the user can make later, from https://dizz.shitworks.co.
- If dizz is installed but this specific project has no dizz metadata yet, ask the user before setting it up — it's their call whether this project should start tracking memory this way, since it's a new decision for a project seeing this skill for the first time.
- If the project already has dizz metadata, just use it — no need to ask again.

## Step 2: Find the right command by asking dizz, not this file

This file does not list dizz's commands or flags, on purpose. The CLI can change after this was written, and a memorized list that's gone stale is worse than none — it leads to confident, wrong invocations.

- Run `dizz --help` to see everything dizz can do right now.
- Run `dizz <command> --help` once you know which command you need.

Both are instant and local. Check rather than guess, every time you're about to use dizz for something new.

## When to reach for dizz

Once applicability is confirmed (Step 1), consult `dizz --help` and pick the matching command whenever:

- **Starting a session** on a project that has dizz set up — get the current state before reading files or asking the user to recap.
- **Before and after making changes** — checkpoint state so nothing gets lost.
- **The user states a long-term goal, plan, or decision** (not a quick fix) — record it immediately so it outlives this conversation, rather than just noting it in chat.
- **The user asks what to work on, what's unfinished, or for a status check** — answer from dizz's output, not a guess.
- **Returning to the project after a gap** ("haven't touched this in weeks", "where did I leave off") — recover context from dizz first.

## Failure modes

| Symptom | Cause | Fix |
|---|---|---|
| `command not found: dizz` | Not installed on this machine | Mention it, offer the install link, continue without it |
| No dizz project detected | This project hasn't opted in yet | Ask the user before initializing (see Step 1) |
| Missing git info in output | Non-git repo, or no commits yet | Expected and fully supported — dizz works without git, just with less history detail |
| Output looks stale | Local state out of date | Take a fresh checkpoint |
| A command or flag errors | This is exactly what `--help` is for | Re-check with `--help` rather than retrying blind |

## Constraints

- **Read-only** — dizz never edits source code. Any actual fix or resolution is the agent's job, based on dizz's output. It's not a linter and not a task manager.
- **Zero-config** by default, once a project has opted in.
- Test functions (e.g. `TestXxx`, `BenchmarkXxx`) can look "unused" since the test framework calls them implicitly rather than from visible call sites — this is a known limitation, not a bug in dizz's reasoning.
