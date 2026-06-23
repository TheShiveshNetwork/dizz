# AGENTS.md

## Purpose

`dizz` is designed as a **state‑aware assistant for both humans and AI agents**.  
It continuously models the project’s progress by combining static analysis, Git history, and intent markers, enabling agents and developers to:

- Remember what has been done, what is planned, and what has been forgotten.
- Efficiently clean up unused, abandoned, or unstable code.
- Maintain a clear backlog of TODOs, intent‑driven goals, and leftover work.
- Make informed decisions about what to work on next without relying on external task trackers.

---

## How dizz Helps Remember Project State

| Feature | What It Does | How It Helps Humans/Agents |
|---------|--------------|----------------------------|
| **Git‑aware analysis** | Tracks file changes, churn, and symbol lifetimes across commits. | Shows which parts of the code are evolving, stable, or neglected. |
| **Symbol states** (`active`, `planned`, `unstable`, `unused`, `abandoned`) | Derived from usage signals, intent markers, and Git churn. | Provides a quick health map of every function, type, or variable. |
| **Snapshots** (`dizz snapshot`) | Content‑addressed, immutable records of the whole‑project state stored in `.dizz/objects/`. | Allows agents to compare before/after states, roll back to a known clean point, or audit progress over time. |
| **Intent system** (`dizz intent`) | Separates disposable TODO/FIXME comments from long‑lived, immutable project goals. | Agents can query, add, or resolve intents without parsing ad‑hoc comments. |
| **Log & status** (`dizz log`, `dizz status`) | Summarizes planned work, unstable areas, unused/abandoned code. | Gives a concise, actionable “what‑to‑work‑on‑next” view. |

---

## Efficient Cleaning of Unused & Leftover Code

1. **Detect dead code**  
   Run `dizz log` (or `dizz log -a` to include healthy symbols) and look for `unused` or `abandoned` states.

2. **Review unstable areas**  
   High‑churn (`unstable`) symbols may indicate experimental or half‑finished logic that should be either completed or removed.

3. **Remove or refactor**  
   - If a symbol is truly unused and not referenced anywhere, delete it.  
   - If it’s abandoned (old + unused), consider whether it represents a dead feature.  
   - If it’s unstable but intentional, add an intent to track the ongoing work.

4. **Verify impact**  
   After deletions, run `dizz snapshot --auto` (or manually) to create a new baseline, then `dizz log` to confirm the problematic states are gone.

---

## Maintaining TODOs, Intent, and Leftover Work

### TODO/FIXME Comments
- Treated as **temporary, code‑level** signals.  
- Appear as `planned` state until the associated symbol becomes `active` or is removed.  
- Agents should periodically scan for these comments and either resolve them or promote them to formal intents.

### Intent Management
- **Add**: `dizz intent add "Description" --type todo --severity 2`  
- **List**: `dizz intent list`  
- **Resolve**: `dizz intent resolve <intent‑id>` (marks as completed, removing its influence on symbol states).  
- Intents are immutable; once created they persist until explicitly resolved, providing a reliable long‑term roadmap.

### Periodic Housekeeping
- Add a lightweight cron or CI step that runs `dizz log --all` and flags any new `unused`/`abandoned` symbols for review.  
- Use `dizz snapshot` before major refactors to have a rollback point.

---

## Guidelines for AI Agents Working in This Repository

1. **Initialize state**  
   ```bash
   dizz init
   ```
   (Ensures the `.dizz/` metadata exists.)

2. **Snapshot before changes**  
   ```bash
   dizz snapshot --auto
   ```
   Records the pre‑change state.

3. **Assess current state**  
   ```bash
   dizz status
   dizz log   # or `dizz log -a` for full detail
   ```
   Identify what to work on (e.g., high‑severity intents, unstable symbols, obvious unused code).

4. **Implement changes**  
   - Prefer addressing `planned` symbols (TODOs/intent) before touching stable `active` code.  
   - When removing code, verify it is truly `unused`/`abandoned` via the log.

5. **Validate after changes**  
   ```bash
   dizz snapshot --auto   # new post‑change snapshot
   dizz log               # confirm unintended states did not appear
   ```

6. **Document intent**  
   If your work introduces a new goal that should outlive the current change, add an intent:
   ```bash
   dizz intent add "Explain why this refactor was needed" --type question --severity 1
   ```

7. **Keep the metadata clean**  
   - Avoid committing large `.dizz/` objects unless using snapshots for CI caching; the directory is intended to be ephemeral and rebuildable via `dizz init` + `dizz log`.  
   - Add `.dizz/` to `.gitignore` if you prefer not to store snapshots in the repo (the tool works fine without them).

---

## Example Workflow for an AI Agent

```bash
# 1. Set up
dizz init

# 2. Record current state
dizz snapshot --auto   # creates .dizz/objects/<hash>

# 3. See what needs attention
dizz status
# → shows 2 unintents, 1 unused symbol, 3 unstable areas

# 4. Focus on the highest‑priority intent
dizz intent list
# → pick intent id int_1122334455: "Refactor auth middleware"

# 5. Implement the refactor, removing stale code discovered via `dizz log -a`

# 6. After code is ready, snapshot again
dizz snapshot --auto

# 7. Verify the refactor resolved the unintents and cleared the unused symbol
dizz log
# → all relevant symbols now `active` or `planned` with new intents if any follow‑up work

# 8. Resolve the completed intent
dizz intent resolve int_1122334455
```

---

## Extending the Agent‑Friendly Behavior

- **Hooks**: Project maintainers can add scripts to `.githooks/` or CI that run `dizz status` and fail on new `abandoned` symbols, guaranteeing a clean baseline.  
- **Integration**: AI agents can call `dizz` programmatically (it outputs plain text; `--json` flag can be added in future versions for easier parsing).  
- **Custom Signals**: If the project adds new analysis plugins, update this document to reflect additional states or intent types.

---

**Remember:** `dizz` is a *read‑only* assistant—it never modifies source code. All clean‑up, intent resolution, and state changes are performed by you (human or agent) based on the information `dizz` provides. Use it as a compass, not as an autopilot.