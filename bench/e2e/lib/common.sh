#!/usr/bin/env bash
# Shared helpers for the dizz e2e benchmark harness.
# Sourced by run.sh - do not execute directly.

BENCH_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROJECT_ROOT="$(cd "$BENCH_DIR/../.." && pwd)"
RESULTS_DIR="$BENCH_DIR/results"
STATE_FILE="$RESULTS_DIR/state.json"
# Run projects live OUTSIDE the repo so the benchmark operates on a clean tree
# and never pollutes the dizz working directory. Defaults to a sibling of the
# project in ~/Desktop/projects; override with DIZZ_BENCH_RUNS_DIR or --runs-dir.
TESTS_DIR="${DIZZ_BENCH_RUNS_DIR:-$(dirname "$PROJECT_ROOT")/dizz-bench-agent-runs}"
TEMPLATE_DIR="$BENCH_DIR/template"
TASKS_DIR="$BENCH_DIR/tasks"
GLOBAL_BACKUP_DIR="$RESULTS_DIR/.global_skills_backup"
GLOBAL_BACKUP_MANIFEST="$GLOBAL_BACKUP_DIR/manifest.tsv"
SKILL_SRC="$BENCH_DIR/../internal/defaults/agent-skills/dizz/SKILL.md"

DEFAULT_MODEL="opencode/deepseek-v4-flash-free"
DEFAULT_RUNS=5
DEFAULT_TIMEOUT=1800
DEFAULT_MAX_ATTEMPTS=3

GIT_USER_NAME="dizzbench"
GIT_USER_EMAIL="dizzbench@local"

log()  { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*" | tee -a "$RESULTS_DIR/bench.log"; }
die()  { log "ERROR: $*"; exit 1; }

# Global dizz skills installed by `dizz install-skill` would leak into the
# "without_dizz" condition. Move them aside for the whole run, restore on exit.
GLOBAL_SKILL_DIRS=(
  "$HOME/.agents/skills/dizz"
  "$HOME/.config/opencode/skills/dizz"
  "$HOME/.claude/skills/dizz"
)

move_global_dizz_skills() {
  [ "$NO_GLOBAL_CLEANUP" -eq 1 ] && { log "global skill cleanup skipped"; return; }
  mkdir -p "$GLOBAL_BACKUP_DIR"
  : > "$GLOBAL_BACKUP_MANIFEST"
  local backup base
  for d in "${GLOBAL_SKILL_DIRS[@]}"; do
    [ -e "$d" ] || continue
    base="$(printf '%s' "$d" | tr '/' '_')"
    backup="$GLOBAL_BACKUP_DIR/${base}"
    [ -e "$backup" ] && rm -rf "$backup"
    mv "$d" "$backup" && log "moved global skill aside: $d -> $backup"
    printf '%s\t%s\n' "$backup" "$d" >> "$GLOBAL_BACKUP_MANIFEST"
  done
}

restore_global_dizz_skills() {
  [ "$NO_GLOBAL_CLEANUP" -eq 1 ] && return 0
  [ -f "$GLOBAL_BACKUP_MANIFEST" ] || return 0
  local backup target
  while IFS=$'\t' read -r backup target; do
    [ -n "$backup" ] || continue
    if [ -e "$backup" ] && [ ! -e "$target" ]; then
      mkdir -p "$(dirname "$target")"
      mv "$backup" "$target" && log "restored global skill: $target"
    fi
  done < "$GLOBAL_BACKUP_MANIFEST"
  return 0
}
