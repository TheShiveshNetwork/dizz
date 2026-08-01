# dizz e2e benchmark harness

Measures how much context dizz saves for a real AI coding agent
(opencode). A fresh copy of the same seed Go project is created for every
cell of (task x condition x run), the agent is asked the same prompt, and
the exported session is analyzed for token usage.

- condition `with_dizz`: repo has `.dizz/` metadata, the agent prompt
  includes `dizz context` output, and the dizz skill is available to
  opencode.
- condition `without_dizz`: identical project, no dizz metadata, no skill.

Both conditions get a fresh `git init` baseline commit so the agent can
`git diff` its own work. Global dizz skills installed with
`dizz install-skill` are moved aside for the whole run so they cannot leak
into the control condition, and are restored on exit.

The harness itself is validated without spending tokens using a fake
`opencode` binary (`/tmp/fakebin/opencode`).

## Layout

```
bench/e2e/
  run.sh          # orchestrator, checkpointing, resume
  analyze.sh      # aggregates results/ into summary.md + all_metrics.csv
  lib/
    common.sh     # shared paths + global-skill backup/restore
    dizz_setup.sh # git baseline, dizz init, primed context, skill copy
    opencode.sh   # run / resolve session id / export session
  template/       # taskforge Go CLI seed project with planted issues
  tasks/
    01_deadcode/  # prompt.md + verify.sh, per task
    02_bugfix/
    03_todos/
    04_plan/
    05_refactor/
  results/        # per-run meta + exported sessions (gitignored)
  results/seq/    # --sequence mode cells (condition/run + per-session meta)
```

Test projects are created in `~/Desktop/projects/dizz-bench-agent-runs/` -
outside the repo so the benchmark never touches the dizz working tree.
Override the location with `--runs-dir DIR` or `DIZZ_BENCH_RUNS_DIR`.
Every `with_dizz` project starts from `dizz init` (skipped only if the copied
template already has `.dizz/config.json`), so the metadata and the project
agent skill (`.agents/skills/dizz/SKILL.md`) are created together, then the
benchmark runs against that initialized project.

## Requirements

- `opencode` on PATH with an authenticated provider (see `--model`)
- `dizz` on PATH
- `jq` (analysis only)
- `timeout` (coreutils), `git`, `bash`

## Usage

Full benchmark (default: 5 runs, all 5 tasks, both conditions):

```bash
cd bench/e2e
./run.sh
```

Common options:

```bash
./run.sh --runs 10                 # more runs per cell
./run.sh --task 01_deadcode        # one task only
./run.sh --condition with_dizz     # one condition only
./run.sh --model opencode/other/model
./run.sh --timeout 3600            # per-run opencode timeout (seconds)
./run.sh --runs-dir /tmp/bench-runs  # where per-run projects are written
./run.sh --fresh                   # wipe results/ and start over
./run.sh --list-tasks
./run.sh --selfcheck               # validate harness, no tokens spent
```

Sequence mode (all tasks in one project per condition/run, so later
sessions are "subsequent runs" on accumulated state):

```bash
./run.sh --sequence                          # all 5 tasks in order, one project per cell
./run.sh --sequence --runs 10
./run.sh --sequence --sequence-list 01_deadcode,02_bugfix   # ordered subset
```

Resuming:

- The benchmark checkpoints to `results/state.json` after every run.
- If a run dies (token limit, crash, ctrl-c), re-run the same command to
  pick up where it left off. Interrupted cells are retried up to
  `--max-attempts` (default 3), then marked `gave_up`.
- `--continue-on-error`: keep going past an interrupted run instead of pausing.
- `--keep-interrupted`: skip interrupted cells on resume (keep partial exports).
- `--retry-failed`: redo already-completed cells.

Analyzing:

```bash
cd bench/e2e
./analyze.sh
# writes results/summary.md and results/all_metrics.csv
# when sequence results exist, also appends the sequence report
# and writes results/seq_metrics.csv
```

`analyze.sh` takes an optional results directory argument
(`./analyze.sh [results-dir]`, default `results`) and needs only `jq`
plus coreutils - no Python.

The sequence report answers the "first run vs subsequent runs" question:
per-session medians (session 1 = first run, later sessions = subsequent runs),
cumulative input tokens after each session, and a first-run vs
avg-per-subsequent-session vs total comparison, all with_dizz vs without_dizz.

## Experiment design

### Single-task mode (default)

- Each run starts from the exact same `template/` tree, copied fresh.
- The agent is invoked non-interactively:
  `opencode run --format json --title db-<task>-<cond>-<run> -m <model> <prompt>`
- The session export (when available) is used for token accounting.
  Input/output/cache tokens come from the export; a missing export (failed
  or interrupted run) contributes no tokens and lowers the success rate.
- Metrics: medians across runs of input tokens, output tokens, cache reads,
  duration, tool calls, files changed; plus success rate per cell.
- Honesty gate: if the `with_dizz` success rate is >20 points below the
  `without_dizz` control, `analyze.sh` refuses to report savings - the
  benchmark only counts as evidence when dizz does not hurt correctness.

### Sequence mode (`--sequence`)

Tests the hypothesis that dizz context costs a little on the first session
but saves tokens on follow-up sessions (because the agent gets a compact
`dizz context` dump instead of re-reading the tree).

- One project copy per `(condition, run)`; the tasks in `--sequence-list`
  (default: all 5, in order) run as consecutive opencode sessions against
  that same project.
- Each session is a separate export, recorded as
  `  results/seq/<condition>/run_<n>_session_<i>_<task>.meta.json`.
- Both analyzers (`analyze.sh` and `analyze.py`) exclude `seq/` from the
  single-task tables, so sequence cells never leak into single-task counts.
- For `with_dizz`, a fresh `dizz context` is generated before *every* session
  and embedded in that session's prompt, so later sessions reflect the state
  accumulated by earlier ones.
- Per-session checkpointing: an interrupted cell resumes from the next
  unfinished session on the existing project dir (no re-copy, no re-run of
  finished sessions).

The seed project (`template/`) is a small Go CLI called `taskforge` with
planted issues:

- dead code in `internal/store/dead.go` (4 unused symbols)
- a `--verbose` flag that is parsed but never forwarded
- missing due-date sorting and an unimplemented `Remind` (TODOs)
- a priority bug (high == medium == 1) and incomplete planning docs (FIXME)
- undocumented/unexported packages that need a refactor + README

Every task's `verify.sh` must fail on the untouched baseline and pass after
the intended fix; this was checked for all 5 tasks.

## Notes

- `dizz context` is run once during setup and its output is embedded in the
  `with_dizz` prompt; the byte size is recorded in each run's
  `run_N.meta.json` as `dizz_context_bytes`.
- Project skill discovery is mirrored into `.opencode/skill/dizz` (note the
  singular `skill` - opencode rejects `.opencode/skills`) so the dizz skill is
  available to opencode while global installs are quarantined during the run.
- Results are content per-run; files changed come from the session export
  summary when present.
