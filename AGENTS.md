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
| **Agent context** (`dizz context`) | Token-optimized TON-format dump of active intents, symbol health, todos, and git state. | Single ~2 KB command that replaces reading 100+ KB of state files. Primary entry point for agents. |

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
   (Ensures the `.dizz/` metadata exists. Use `dizz init --agent` for agent-optimized setup with TON format.)

2. **Assess current state**  
   ```bash
   dizz context
   ```
   Token-optimized dump of active intents, symbol health, todos, and git state (~2 KB).

3. **Snapshot before changes**  
   ```bash
   dizz snapshot --auto
   ```
   Records the pre‑change state (compact JSON, no indentation).

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

## Language Registry

The project supports **34 languages** through `internal/language/registry.go`. Every language is defined as a `LanguageConfig` with extensions, comment styles, regex patterns for function/type/call extraction, and keyword sets. To add a new language, append a `LanguageConfig` to the `languages` slice — no other file needs changing (extension detection and file discovery derive from the registry automatically).

Each language is assigned an **analysis tier** that determines signal accuracy:
- **AST** (Tier 1): Full parser-backed — only Go currently. Highest confidence.
- **Lexical** (Tier 2): Structural regex patterns — most mainstream languages. Good accuracy.
- **Regex** (Tier 3): Regex fallback — C, C++, Shell, Haskell, etc. Lower accuracy; signals are weighted accordingly.

All language tests live in `tests/language_conformance_test.go` and `tests/language_conformance_extended_test.go`. When adding a language, add a conformance test that verifies function definition and call extraction.

---

## Agent Rulesets & Patterns

### Code Structure Rules

1. **Package naming**: All internal packages use short, lowercase, single-word names (`state`, `signals`, `discover`, `store`, `config`, `analyzer`, `integrations`, `language`). Never use underscores or mixedCase.

2. **File organization**: Place all backend/internal code under `internal/`. CLI command implementations go in `cmd/`. Tests go in `tests/` (package `benchmarks`). Benchmark files are named `*_benchmark_test.go`; test-only files use `*_test.go`.

3. **Signal pipeline**: The analysis architecture follows a strict pipeline: `discover.CodeFiles` → `analyzer.Registry.AnalyzeFiles` → `signals.SignalSet` → `state.Scorer.InterpretSignalsWithIntent` → `state.ProjectState`. Never bypass layers — the scorer must interpret signals, not raw analysis output.

4. **Language registry is the single source of truth**: Never hardcode extensions, comment styles, or analysis patterns outside `internal/language/registry.go`. The `discover` package derives file discovery from `language.AllExtensions()`. The regex analyzer derives all behaviour from `LanguageConfig` entries.

### Testing Rules

5. **Every language gets a conformance test**: At minimum, test `FunctionDefined` extraction. Use the `analyzeContent(t, ext, src)` helper from `tests/language_conformance_test.go`. New language entries without a corresponding test will be rejected.

6. **Benchmarks for throughput-sensitive code**: `BenchmarkScorerInterpretation`, `BenchmarkCodeFiles`, and `BenchmarkDetectByExtension` cover the hot path. When adding new analysis passes or modifying the scorer, add or update benchmarks to prevent regressions.

7. **Git-dependent tests must skip gracefully**: Use `if !integrations.IsRepo() { b.Skip("Not in a git repository") }` at the top of any test or benchmark that requires git. Never assume the test is running inside the dizz repo.

8. **Temp files use `t.TempDir()`**: Never create temp files outside the test's temp directory. Use `writeTemp(t, ext, content)` for single-file tests.

### Code Style & Conventions

9. **Zero comments in source code**: Production code must not contain explanatory comments. Intent is captured via `dizz intent` or AGENTS.md, not inline. TODO/FIXME comments are acceptable only in test files.

10. **Error handling**: Return errors to the caller; log only at the CLI layer (cmd/). Internal packages never call `log.Fatal` or `os.Exit`. Use `fmt.Errorf` with `%w` for error wrapping.

11. **No external dependencies beyond stdlib + cobra**: The project currently only depends on `github.com/spf13/cobra` for CLI. New external dependencies require explicit justification and a maintainer review.

12. **Signal metadata conventions**: When adding new `SignalType`s or `Metadata` keys, prefix project-specific keys with the package name (e.g., `"regex.source_tier"`). All confidence values are float64 in [0, 1].

13. **Immutable intents, mutable todos**: TODO/FIXME found in source are ephemeral and re-scanned on every analysis. Intents persisted in `.dizz/intent.ton` (token-optimized format) are immutable project records — never delete or modify them programmatically; only add or resolve.

14. **Config is always optional**: dizz must work with zero configuration. Defaults are defined in `internal/defaults/defaults.go`. Config values in `.dizz/config.json` override defaults — never require config for basic operation.

15. **Snapshot immutability**: Snapshot objects in `.dizz/objects/` are content-addressed and must never be modified after creation. Use `dizz snapshot --auto` to create new snapshots; never overwrite or delete existing ones. Snapshots use compact JSON (no indentation) for reduced size.

---

---

## Token-Optimized Notation (TON)

The project uses **TON** (Token-Optimized Notation) for storing intent data. TON is a line-oriented, pipe-delimited format that eliminates JSON structural overhead.

**Intent file (`.dizz/intent.ton`) example:**
```ton
id|type|sev|status|msg|scope|tags
int_001|fixme|3|active|Fix critical bug|project|urgent
```

Key differences from JSON:
- No quotes, braces, or indentation (~90% fewer tokens)
- One line per record, pipe-delimited fields
- First line is the header declaring column names
- Agents can read it directly with minimal token cost

**Reading TON**: Any agent can parse it by splitting on `|` — no JSON parser needed.

---

**Remember:** `dizz` is a *read‑only* assistant—it never modifies source code. All clean‑up, intent resolution, and state changes are performed by you (human or agent) based on the information `dizz` provides. Use it as a compass, not as an autopilot.