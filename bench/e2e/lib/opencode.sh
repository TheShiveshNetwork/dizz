#!/usr/bin/env bash
# Run one opencode session for the benchmark and capture its session id.
# Sourced by run.sh.
#
# Usage: run_opencode <run_dir> <prompt_file> <events_file> <log_file> <model> <title> <timeout_sec>
# Prints the opencode exit code to stdout (via RUN_OPENCODE_RC global).

RUN_OPENCODE_RC=0

run_opencode() {
  local run_dir="$1" prompt_file="$2" events_file="$3" log_file="$4" model="$5" title="$6" timeout_sec="$7"

  local prompt
  prompt="$(cat "$prompt_file")"

  ( cd "$run_dir" && \
    timeout --foreground "$timeout_sec" \
      opencode run --format json --title "$title" -m "$model" "$prompt" ) \
    > "$events_file" 2> "$log_file"
  RUN_OPENCODE_RC=$?
}

# Resolve an opencode session id. Returns non-zero if not found.
resolve_session_id() {
  local title="$1" events_file="$2" sid=""

  # 1. Look for a session id in the JSON event stream.
  sid="$(grep -oE '"sessionID"[[:space:]]*:[[:space:]]*"ses_[A-Za-z0-9]+"' "$events_file" 2>/dev/null | head -1 | sed -E 's/.*"(ses_[A-Za-z0-9]+)".*/\1/')"
  if [ -n "$sid" ]; then
    printf '%s' "$sid"
    return 0
  fi

  # 2. Any bare ses_ token in the events.
  sid="$(grep -oE 'ses_[A-Za-z0-9]+' "$events_file" 2>/dev/null | head -1)"
  if [ -n "$sid" ]; then
    printf '%s' "$sid"
    return 0
  fi

  # 3. Fallback: recent session list matching the run title.
  sid="$(opencode session list 2>/dev/null | grep -F "$title" | awk '{print $1}' | head -1)"
  if [ -n "$sid" ]; then
    printf '%s' "$sid"
    return 0
  fi

  return 1
}

# Export a session to a file. Returns non-zero on failure.
export_session() {
  local session_id="$1" out_file="$2"
  opencode export "$session_id" > "$out_file" 2>/dev/null
}
