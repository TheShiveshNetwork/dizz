#!/usr/bin/env bash
# Set up one benchmark run directory.
# Usage: setup_run_dir <run_dir> <condition>
#   with_dizz  -> git baseline + dizz init + primed context + project skill
#   without_dizz -> git baseline only
# Sourced by run.sh - does not exit on error by itself (returns codes).

setup_run_dir() {
  local run_dir="$1" condition="$2"

  mkdir -p "$run_dir"
  if [ -z "$(ls -A "$run_dir" | grep -v '^\.git$')" ]; then
    log "FAIL: run dir is empty after copy: $run_dir"
    return 1
  fi
  ( cd "$run_dir" && git init -q && \
    git -c user.name="$GIT_USER_NAME" -c user.email="$GIT_USER_EMAIL" add -A && \
    git -c user.name="$GIT_USER_NAME" -c user.email="$GIT_USER_EMAIL" commit -qm "baseline" )
  if [ "$?" -ne 0 ]; then
    log "FAIL: could not create git baseline in $run_dir"
    return 1
  fi

  if [ "$condition" = "with_dizz" ]; then
    # Start every dizz project from `dizz init` so the metadata and the agent
    # skill (.agents/skills/dizz/SKILL.md) are created together. Skip only if
    # the copied template was already initialized.
    if [ -f "$run_dir/.dizz/config.json" ]; then
      log "with_dizz: $run_dir already initialized, skipping dizz init"
    else
      # Non-interactive because the git repo exists.
      ( cd "$run_dir" && dizz init >/dev/null 2>&1 )
      if [ "$?" -ne 0 ]; then
        log "FAIL: dizz init failed in $run_dir"
        return 1
      fi
      log "with_dizz: dizz init complete in $run_dir"
    fi

    # Prime the analysis so the agent's first read is warm and state exists.
    ( cd "$run_dir" && dizz context >/dev/null 2>&1 ) || true

    # Ensure opencode can discover the project skill. dizz init writes
    # .agents/skills/dizz/SKILL.md; mirror the generated skill (falling back
    # to the repo's canonical copy) into .opencode/skill too. Note: opencode
    # expects the singular "skill" directory, not "skills".
    local skill_md=""
    if [ -f "$run_dir/.agents/skills/dizz/SKILL.md" ]; then
      skill_md="$run_dir/.agents/skills/dizz/SKILL.md"
    elif [ -f "$SKILL_SRC" ]; then
      skill_md="$SKILL_SRC"
    fi
    if [ -n "$skill_md" ]; then
      mkdir -p "$run_dir/.opencode/skill/dizz"
      cp "$skill_md" "$run_dir/.opencode/skill/dizz/SKILL.md"
    fi
    log "with_dizz setup complete: $run_dir"
  else
    log "without_dizz setup complete: $run_dir"
  fi

  return 0
}
