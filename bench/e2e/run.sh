#!/usr/bin/env bash
# dizz e2e benchmark runner.
#
# Runs opencode against a fresh copy of the seed project for every
# (task, condition, run) cell, with and without dizz, capturing session
# exports for token accounting. Checkpoints every run so an interrupted
# benchmark (token limit, crash, ctrl-c) can be resumed in place.
#
# Usage:
#   ./run.sh [--runs N] [--model PROVIDER/MODEL] [--task 01_deadcode]
#            [--condition with_dizz|without_dizz] [--timeout SEC]
#            [--continue-on-error] [--fresh] [--keep-events]
#            [--keep-interrupted] [--retry-failed] [--max-attempts N]
#            [--no-global-cleanup] [--selfcheck] [--list-tasks]
#            [--runs-dir DIR]
#            [--sequence] [--sequence-list 01_deadcode,02_bugfix]
#
# Results land in results/ (gitignored); per-run projects land in
# ~/Desktop/projects/dizz-bench-agent-runs/ by default (override with
# DIZZ_BENCH_RUNS_DIR or --runs-dir).

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ---------------------------------------------------------------------------
# Bootstrap (defines MODEL defaults, results paths, die/log helpers)
# ---------------------------------------------------------------------------
source "$SCRIPT_DIR/lib/common.sh"
source "$SCRIPT_DIR/lib/dizz_setup.sh"
source "$SCRIPT_DIR/lib/opencode.sh"

# ---------------------------------------------------------------------------
# Defaults / args
# ---------------------------------------------------------------------------
MODEL="$DEFAULT_MODEL"
RUNS="$DEFAULT_RUNS"
TASK_FILTER=""
COND_FILTER=""
RUN_TIMEOUT="$DEFAULT_TIMEOUT"
MAX_ATTEMPTS="$DEFAULT_MAX_ATTEMPTS"
FRESH=0
CONTINUE_ON_ERROR=0
KEEP_EVENTS=0
KEEP_INTERRUPTED=0
RETRY_FAILED=0
NO_GLOBAL_CLEANUP=0
SELFCHECK=0
LIST_TASKS=0
RUNS_DIR=""
SEQUENCE=0
SEQUENCE_LIST=""

usage() {
  sed -n '2,19p' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --runs)          RUNS="$2"; shift 2 ;;
    --model)         MODEL="$2"; shift 2 ;;
    --task)          TASK_FILTER="$2"; shift 2 ;;
    --condition)     COND_FILTER="$2"; shift 2 ;;
    --timeout)       RUN_TIMEOUT="$2"; shift 2 ;;
    --max-attempts)  MAX_ATTEMPTS="$2"; shift 2 ;;
    --runs-dir)      RUNS_DIR="$2"; shift 2 ;;
    --sequence)      SEQUENCE=1; shift ;;
    --sequence-list) SEQUENCE_LIST="$2"; shift 2 ;;
    --fresh)         FRESH=1; shift ;;
    --continue-on-error) CONTINUE_ON_ERROR=1; shift ;;
    --keep-events)   KEEP_EVENTS=1; shift ;;
    --keep-interrupted) KEEP_INTERRUPTED=1; shift ;;
    --retry-failed)  RETRY_FAILED=1; shift ;;
    --no-global-cleanup) NO_GLOBAL_CLEANUP=1; shift ;;
    --selfcheck)     SELFCHECK=1; shift ;;
    --list-tasks)    LIST_TASKS=1; shift ;;
    -h|--help)       usage ;;
    *) die "unknown argument: $1" ;;
  esac
done

if [ -n "$RUNS_DIR" ]; then
  TESTS_DIR="$RUNS_DIR"
fi

# Resolve the ordered task list for --sequence mode.
SEQ_TASKS=()
if [ "$SEQUENCE" -eq 1 ]; then
  if [ -n "$SEQUENCE_LIST" ]; then
    IFS=',' read -r -a SEQ_TASKS <<< "$SEQUENCE_LIST"
  else
    for d in "$TASKS_DIR"/*/; do
      [ -d "$d" ] || continue
      SEQ_TASKS+=("$(basename "$d")")
    done
  fi
  if [ "${#SEQ_TASKS[@]}" -eq 0 ]; then
    die "sequence mode found no tasks under $TASKS_DIR"
  fi
  for t in "${SEQ_TASKS[@]}"; do
    [ -d "$TASKS_DIR/$t" ] || die "unknown sequence task: $t"
  done
  log "sequence mode: tasks = ${SEQ_TASKS[*]}"
fi

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
mkdir -p "$RESULTS_DIR"

if [ "$LIST_TASKS" -eq 1 ]; then
  for d in "$TASKS_DIR"/*/; do
    [ -d "$d" ] || continue
    printf '%s\n' "$(basename "$d")"
  done
  exit 0
fi

# ---------------------------------------------------------------------------
# State helpers (checkpointing / resume)
# ---------------------------------------------------------------------------
load_state() {
  [ -f "$STATE_FILE" ] && cat "$STATE_FILE" || printf '{}'
}

save_state() {
  local tmp="$STATE_FILE.tmp"
  printf '%s\n' "$1" > "$tmp"
  mv "$tmp" "$STATE_FILE"
}

state_get() {
  printf '%s' "$1" | jq -r --arg k "$2" --arg f "$3" '.runs[$k][$f] // ""'
}

state_update() {
  printf '%s' "$1" | jq -c --arg k "$2" --arg s "$3" --arg a "$4" --arg m "$MODEL" --arg r "$RUNS" \
    '.runs[$k] = {"status": $s, "attempts": ($a | tonumber)} | .model = $m | .run_count = ($r | tonumber)'
}

# ---------------------------------------------------------------------------
# One benchmark cell
# ---------------------------------------------------------------------------
RUN_STATUS=""
run_one() {
  local task="$1" cond="$2" run="$3"
  local key="${task}/${cond}/${run}"
  local run_dir="$TESTS_DIR/${task}_${cond}_run${run}"
  local out_dir="$RESULTS_DIR/$task/$cond"
  local tag="db-${task}-${cond%%_*}-${run}"
  local started duration_s session_id export_file rc verify_rc ctx_bytes

  mkdir -p "$out_dir"

  rm -rf "$run_dir"
  mkdir -p "$(dirname "$run_dir")"
  if ! cp -a "$TEMPLATE_DIR" "$run_dir"; then
    RUN_STATUS="failed"
    log "FAIL: could not copy template to $run_dir"
    return 1
  fi
  log "== $key: project copied =="

  if ! setup_run_dir "$run_dir" "$cond"; then
    RUN_STATUS="failed"
    return 1
  fi

  ctx_bytes=0
  if [ "$cond" = "with_dizz" ]; then
    ( cd "$run_dir" && dizz context > "$out_dir/run_${run}.context.ton" 2>/dev/null ) || true
    ctx_bytes="$(wc -c < "$out_dir/run_${run}.context.ton" 2>/dev/null || echo 0)"
  fi

  local events_file="$out_dir/run_${run}.events.json"
  local log_file="$out_dir/run_${run}.log"
  local prompt_file="$TASKS_DIR/$task/prompt.md"
  local built_prompt="$out_dir/run_${run}.prompt.md"
  local ctx_for_prompt=""
  if [ "$cond" = "with_dizz" ]; then
    ctx_for_prompt="$out_dir/run_${run}.context.ton"
  fi
  build_prompt "$task" "$prompt_file" "$ctx_for_prompt" "$built_prompt"

  started="$(date +%s)"
  run_opencode "$run_dir" "$built_prompt" "$events_file" "$log_file" "$MODEL" "$tag" "$RUN_TIMEOUT"
  rc="$RUN_OPENCODE_RC"
  duration_s="$(( $(date +%s) - started ))"

  session_id=""
  if [ -f "$events_file" ] && [ -s "$events_file" ]; then
    session_id="$(resolve_session_id "$tag" "$events_file")" || session_id=""
  fi

  export_file="run_${run}.json"
  if [ -n "$session_id" ]; then
    ( cd "$run_dir" && export_session "$session_id" "$out_dir/$export_file" ) || true
  fi

  success=0
  verify_rc=1
  if [ "$rc" -eq 0 ] && [ -f "$TASKS_DIR/$task/verify.sh" ]; then
    ( cd "$run_dir" && bash "$TASKS_DIR/$task/verify.sh" >/dev/null 2>&1 )
    verify_rc=$?
    [ "$verify_rc" -eq 0 ] && success=1
  fi

  if [ "$rc" -eq 0 ]; then
    RUN_STATUS="done"
  else
    RUN_STATUS="interrupted"
    log "  $key: opencode exited rc=$rc (timeout=${RUN_TIMEOUT}s). Saving partial state."
  fi

  cat > "$out_dir/run_${run}.meta.json" <<EOF
{
  "task": "$task",
  "condition": "$cond",
  "run": $run,
  "status": "$RUN_STATUS",
  "success": $success,
  "session_id": "$session_id",
  "model": "$MODEL",
  "started_at": "$(date -u -Iseconds)",
  "duration_s": $duration_s,
  "rc": $rc,
  "verify_rc": $verify_rc,
  "export_file": "$export_file",
  "dizz_context_bytes": $ctx_bytes
}
EOF

  if [ "$KEEP_EVENTS" -ne 1 ] && [ -f "$events_file" ]; then
    rm -f "$events_file"
  fi

  log "  $key -> status=$RUN_STATUS success=$success duration=${duration_s}s rc=$rc session=$session_id"
}

# Build the prompt file given to opencode. with_dizz prompts append the
# current `dizz context` TON dump so the agent sees compact project state
# instead of having to re-read the whole tree.
build_prompt() {
  local task="$1" prompt_file="$2" context_file="$3" out_prompt="$4"
  if [ -n "$context_file" ] && [ -s "$context_file" ]; then
    {
      cat "$prompt_file"
      printf '\n\n# Project state (dizz context)\n\n'
      cat "$context_file"
    } > "$out_prompt"
  else
    cp "$prompt_file" "$out_prompt"
  fi
}

# One --sequence cell: all tasks in SEQ_TASKS run as consecutive opencode
# sessions against ONE project copy, so later sessions are "subsequent runs"
# on the accumulated project state. Each session is exported separately and
# checkpoints per-session so an interrupted cell can resume where it left off
# (the project dir persists in TESTS_DIR).
run_sequence_cell() {
  local cond="$1" run="$2"
  local cell_key="${cond}/seq${run}"
  local run_dir="$TESTS_DIR/${cond}_seq_run${run}"
  local out_dir="$RESULTS_DIR/seq/$cond"
  local n="${#SEQ_TASKS[@]}"
  mkdir -p "$out_dir"

  log "== $cell_key: sequence cell (${n} sessions) =="

  local first_st
  first_st="$(state_get "$STATE" "${cell_key}/1" "status")"
  if [ "$first_st" != "done" ]; then
    rm -rf "$run_dir"
    mkdir -p "$(dirname "$run_dir")"
    if ! cp -a "$TEMPLATE_DIR" "$run_dir"; then
      RUN_STATUS="failed"
      log "FAIL: could not copy template to $run_dir"
      return 1
    fi
    if ! setup_run_dir "$run_dir" "$cond"; then
      RUN_STATUS="failed"
      return 1
    fi
    log "== $cell_key: project copied =="
  elif [ ! -d "$run_dir" ]; then
    log "WARN: $cell_key project dir missing but session 1 done; re-copying template"
    cp -a "$TEMPLATE_DIR" "$run_dir" || { RUN_STATUS="failed"; return 1; }
    setup_run_dir "$run_dir" "$cond" || { RUN_STATUS="failed"; return 1; }
  fi

  local i task key st at attempts
  for (( i=1; i<=n; i++ )); do
    task="${SEQ_TASKS[$((i-1))]}"
    key="${cell_key}/${i}"
    st="$(state_get "$STATE" "$key" "status")"
    at="$(state_get "$STATE" "$key" "attempts")"
    at="${at:-0}"
    case "$st" in
      done)
        log "skip $key (already done)"
        continue
        ;;
      gave_up)
        log "skip $key (gave up)"
        continue
        ;;
    esac
    attempts=$(( at + 1 ))
    if [ "$attempts" -gt "$MAX_ATTEMPTS" ]; then
      log "skip $key (attempt limit $MAX_ATTEMPTS reached)"
      STATE="$(state_update "$STATE" "$key" "gave_up" "$attempts")"
      save_state "$STATE"
      continue
    fi

    local tag="dbseq-${cond%%_*}-${run}-${i}"
    local prefix="run_${run}_session_${i}_${task}"
    local events_file="$out_dir/${prefix}.events.json"
    local log_file="$out_dir/${prefix}.log"
    local prompt_file="$TASKS_DIR/$task/prompt.md"
    local built_prompt="$out_dir/${prefix}.prompt.md"
    local ctx_file="" ctx_bytes=0
    local started duration_s session_id export_file rc verify_rc

    log "== running $key (attempt $attempts) task=$task =="
    if [ "$cond" = "with_dizz" ]; then
      ctx_file="$out_dir/${prefix}.context.ton"
      ( cd "$run_dir" && dizz context > "$ctx_file" 2>/dev/null ) || true
      ctx_bytes="$(wc -c < "$ctx_file" 2>/dev/null || echo 0)"
    fi
    build_prompt "$task" "$prompt_file" "$ctx_file" "$built_prompt"

    started="$(date +%s)"
    run_opencode "$run_dir" "$built_prompt" "$events_file" "$log_file" "$MODEL" "$tag" "$RUN_TIMEOUT"
    rc="$RUN_OPENCODE_RC"
    duration_s="$(( $(date +%s) - started ))"

    session_id=""
    if [ -f "$events_file" ] && [ -s "$events_file" ]; then
      session_id="$(resolve_session_id "$tag" "$events_file")" || session_id=""
    fi
    export_file="${prefix}.json"
    if [ -n "$session_id" ]; then
      ( cd "$run_dir" && export_session "$session_id" "$out_dir/$export_file" ) || true
    fi

    success=0
    verify_rc=1
    if [ "$rc" -eq 0 ] && [ -f "$TASKS_DIR/$task/verify.sh" ]; then
      ( cd "$run_dir" && bash "$TASKS_DIR/$task/verify.sh" >/dev/null 2>&1 )
      verify_rc=$?
      [ "$verify_rc" -eq 0 ] && success=1
    fi

    if [ "$rc" -eq 0 ]; then
      RUN_STATUS="done"
    else
      RUN_STATUS="interrupted"
      log "  $key: opencode exited rc=$rc (timeout=${RUN_TIMEOUT}s). Saving partial state."
    fi

    cat > "$out_dir/${prefix}.meta.json" <<EOF
{
  "mode": "sequence",
  "condition": "$cond",
  "run": $run,
  "session_index": $i,
  "task": "$task",
  "status": "$RUN_STATUS",
  "success": $success,
  "session_id": "$session_id",
  "model": "$MODEL",
  "started_at": "$(date -u -Iseconds)",
  "duration_s": $duration_s,
  "rc": $rc,
  "verify_rc": $verify_rc,
  "export_file": "$export_file",
  "dizz_context_bytes": $ctx_bytes
}
EOF

    if [ "$KEEP_EVENTS" -ne 1 ] && [ -f "$events_file" ]; then
      rm -f "$events_file"
    fi

    STATE="$(state_update "$STATE" "$key" "$RUN_STATUS" "$attempts")"
    save_state "$STATE"
    log "  $key -> status=$RUN_STATUS success=$success duration=${duration_s}s rc=$rc session=$session_id"

    if [ "$INTERRUPTED" -eq 1 ]; then
      log "ctrl-c received: state saved, stopping."
      stopped=1
      return
    fi
    if [ "$RUN_STATUS" = "interrupted" ] && [ "$CONTINUE_ON_ERROR" -eq 0 ]; then
      log "run $key was interrupted. Resume safely: re-run ./run.sh with the same args."
      stopped=1
      return
    fi
  done
  log "== $cell_key: cell complete =="
}

# ---------------------------------------------------------------------------
# Selfcheck: validate the harness without spending tokens on opencode.
# ---------------------------------------------------------------------------
if [ "$SELFCHECK" -eq 1 ]; then
  log "== selfcheck: validating harness without running opencode =="
  for t in 01_deadcode 02_bugfix 03_todos; do
    tmp_dir="$(mktemp -d)"
    cp -a "$TEMPLATE_DIR/." "$tmp_dir/"
    if ! setup_run_dir "$tmp_dir" "with_dizz"; then
      die "selfcheck: with_dizz setup failed for $t"
    fi
    if ( cd "$tmp_dir" && bash "$TASKS_DIR/$t/verify.sh" >/dev/null 2>&1 ); then
      log "selfcheck WARN: $t verify unexpectedly PASSES on untouched baseline"
    else
      log "selfcheck OK: $t verify fails on baseline (as expected)"
    fi
    rm -rf "$tmp_dir"
  done
  log "selfcheck done"
  exit 0
fi

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
if [ "$FRESH" -eq 1 ]; then
  log "--fresh: removing previous results"
  rm -rf "$RESULTS_DIR"
  mkdir -p "$RESULTS_DIR"
fi

log "dizz e2e benchmark starting"
log "  model=$MODEL runs=$RUNS timeout=${RUN_TIMEOUT}s max_attempts=$MAX_ATTEMPTS"
log "  task=$TASK_FILTER condition=$COND_FILTER continue_on_error=$CONTINUE_ON_ERROR"
log "  runs_dir=$TESTS_DIR"

move_global_dizz_skills
trap restore_global_dizz_skills EXIT INT TERM

INTERRUPTED=0
on_signal() { INTERRUPTED=1; }
trap on_signal INT

declare -a TASKS
if [ -n "$TASK_FILTER" ]; then
  [ -d "$TASKS_DIR/$TASK_FILTER" ] || die "unknown task: $TASK_FILTER (see --list-tasks)"
  TASKS=("$TASK_FILTER")
else
  for d in "$TASKS_DIR"/*/; do
    [ -d "$d" ] || continue
    TASKS+=("$(basename "$d")")
  done
fi

declare -a CONDITIONS
if [ -n "$COND_FILTER" ]; then
  CONDITIONS=("$COND_FILTER")
else
  CONDITIONS=(with_dizz without_dizz)
fi

STATE="$(load_state)"
if [ "$FRESH" -eq 1 ] || ! echo "$STATE" | grep -q '"runs"'; then
  STATE="$(printf '%s' "$STATE" | jq -c --arg m "$MODEL" --arg r "$RUNS" '.model = $m | .run_count = ($r | tonumber) | .runs = (.runs // {})')"
  save_state "$STATE"
fi

stopped=0
if [ "$SEQUENCE" -eq 1 ]; then
  log "sequence mode: ${#SEQ_TASKS[@]} tasks per cell, ${RUNS} runs, ${#CONDITIONS[@]} conditions"
  for cond in "${CONDITIONS[@]}"; do
    for (( run=1; run<=RUNS; run++ )); do
      [ "$stopped" -eq 1 ] && break
      run_sequence_cell "$cond" "$run"
    done
    [ "$stopped" -eq 1 ] && break
  done
else
for task in "${TASKS[@]}"; do
  for cond in "${CONDITIONS[@]}"; do
    for (( run=1; run<=RUNS; run++ )); do
      key="${task}/${cond}/${run}"

      st="$(state_get "$STATE" "$key" "status")"
      at="$(state_get "$STATE" "$key" "attempts")"
      at="${at:-0}"

      case "$st" in
        done)
          if [ "$RETRY_FAILED" -eq 1 ]; then
            log "redo $key (--retry-failed)"
          else
            log "skip $key (already done)"
            continue
          fi
          ;;
        interrupted)
          if [ "$KEEP_INTERRUPTED" -eq 1 ]; then
            log "skip $key (interrupted, --keep-interrupted)"
            continue
          fi
          log "redo $key (interrupted; previous partial export preserved)"
          ;;
        gave_up)
          log "skip $key (gave up after $MAX_ATTEMPTS attempts)"
          continue
          ;;
      esac

      attempts=$(( at + 1 ))
      if [ "$attempts" -gt "$MAX_ATTEMPTS" ]; then
        log "skip $key (attempt limit $MAX_ATTEMPTS reached)"
        STATE="$(state_update "$STATE" "$key" "gave_up" "$attempts")"
        save_state "$STATE"
        continue
      fi

      log "== running $key (attempt $attempts) =="
      run_one "$task" "$cond" "$run"

      STATE="$(state_update "$STATE" "$key" "$RUN_STATUS" "$attempts")"
      save_state "$STATE"

      if [ "$INTERRUPTED" -eq 1 ]; then
        log "ctrl-c received: state saved, stopping."
        stopped=1
        break
      fi

      if [ "$RUN_STATUS" = "interrupted" ] && [ "$CONTINUE_ON_ERROR" -eq 0 ]; then
        log "run $key was interrupted. Resume safely: re-run ./run.sh with the same args."
        stopped=1
        break
      fi
    done
    [ "$stopped" -eq 1 ] && break
  done
  [ "$stopped" -eq 1 ] && break
done
fi

restore_global_dizz_skills
trap - EXIT INT TERM

if [ "$stopped" -eq 1 ]; then
  log "benchmark paused. Resume with the same command."
else
  log "benchmark complete. Analyze with: $SCRIPT_DIR/analyze.sh"
fi
