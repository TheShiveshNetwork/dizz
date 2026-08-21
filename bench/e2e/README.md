# dizz e2e benchmark harness

Runs opencode against a fresh copy of the same seed Go project per
(task x condition x run), with and without dizz, and analyzes the exported
sessions for token usage. Both conditions get a fresh git baseline; global
dizz skills are quarantined during the run and restored on exit.

## Requirements

- `opencode` on PATH (authenticated, see `--model`)
- `dizz` on PATH
- `jq`, `timeout`, `git`, `bash`

## Layout

```
bench/e2e/
  run.sh      # orchestrator (checkpointing, resume)
  analyze.sh  # aggregates results/ into summary.md + CSVs
  template/   # seed Go project with planted issues
  tasks/      # per task: prompt.md + verify.sh
  results/    # per-run meta + exports (gitignored)
```

Per-run projects are created in `~/Desktop/projects/dizz-bench-agent-runs/`
(override with `--runs-dir DIR` or `DIZZ_BENCH_RUNS_DIR`).

## Quick start

```bash
cd bench/e2e
./run.sh --selfcheck      # validate harness, no tokens spent
./run.sh --list-tasks     # list available tasks
./run.sh                  # full benchmark: 5 runs x 5 tasks x 2 conditions
./analyze.sh              # writes results/summary.md + all_metrics.csv
```

## run.sh options

```bash
--runs N                       runs per cell (default 5)
--task 01_deadcode             one task only
--condition with_dizz|without_dizz    one condition only
--model PROVIDER/MODEL         model to run
--timeout SEC                  per-run opencode timeout
--max-attempts N               retries for interrupted cells (default 3)
--runs-dir DIR                 where per-run projects are written
--sequence                     sequence mode (all tasks in one project per cell)
--sequence-list a,b            ordered tasks for --sequence (default: all)
--embed-context                inject a dizz context dump into every with_dizz
                               prompt (default: agent uses dizz via its skill)
--fresh                        wipe results/ and start over
--continue-on-error            keep going past an interrupted run
--keep-events                  keep per-run event files
--keep-interrupted             skip interrupted cells on resume
--retry-failed                 redo already-completed cells
--no-global-cleanup            skip global skill quarantine
--selfcheck                    validate harness, spend no tokens
--list-tasks                   list available tasks
-h, --help                     this help
```

## Sequence mode

Runs the tasks in `--sequence-list` as consecutive sessions on the same
project per (condition, run), so later sessions are subsequent runs on
accumulated state:

```bash
./run.sh --sequence
./run.sh --sequence --sequence-list 01_deadcode,02_bugfix
./run.sh --sequence --runs 10
```

## Resuming

Checkpoints are written to `results/state.json` after every run. Re-run the
same command after a crash, token limit, or ctrl-c to pick up where it left
off.

## analyze.sh

```bash
./analyze.sh                    # analyze results/
./analyze.sh [results-dir]      # analyze a different results directory
```

Writes `summary.md` and `all_metrics.csv` (plus `seq_metrics.csv` when
sequence results exist). Requires only `jq` and coreutils - no Python.
