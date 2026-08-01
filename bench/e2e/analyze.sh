#!/usr/bin/env bash
# Aggregate dizz e2e benchmark results into a comparison summary.
# Shell port of the former analyze.py. Requires: bash 4+, jq, awk, sort, cut.
#
# Reads every results/<task>/<condition>/run_<n>.meta.json plus the exported
# opencode session JSON (run_<n>.json) and produces:
#
#   - a per-task comparison table (medians) with paired deltas
#   - an overall median table
#   - a raw CSV dump at results/all_metrics.csv
#   - a markdown report at results/summary.md
#
# When results/seq/ exists (--sequence mode) it also appends the sequence
# report (per-session, cumulative, first-run vs subsequent) and writes
# results/seq_metrics.csv.
#
# Usage:
#     ./analyze.sh [results-dir]     (default: results)

set -uo pipefail

RESULTS_DIR="${1:-results}"

command -v jq >/dev/null 2>&1 || {
  echo "ERROR: analyze.sh requires jq (install via your package manager)" >&2
  exit 1
}

OUT="$RESULTS_DIR/summary.md"
: > "$OUT"
emit() { printf '%s\n' "$*" | tee -a "$OUT"; }

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# median: reads numbers from stdin (one per line), prints median or "-".
median() {
  local -a vals=()
  local line
  while IFS= read -r line; do
    case "$line" in ""|"-") continue ;; esac
    vals+=("$line")
  done
  local n=${#vals[@]}
  [ "$n" -eq 0 ] && { echo "-"; return; }
  mapfile -t vals < <(printf '%s\n' "${vals[@]}" | sort -n)
  local mid=$((n / 2))
  if (( n % 2 == 1 )); then
    printf '%s\n' "${vals[$mid]}"
  else
    awk -v a="${vals[$((mid - 1))]}" -v b="${vals[$mid]}" \
      'BEGIN { printf "%.1f\n", (a + b) / 2 }'
  fi
}

# fmt: format a number like the python reporter (commas, sign, % or 1-decimal).
fmt() {
  local v="$1" suffix="${2:-}"
  case "$v" in ""|"-") echo "-"; return ;; esac
  if [ "$suffix" = "%" ]; then
    awk -v v="$v" 'BEGIN { printf "%+.1f%%\n", v }'
    return
  fi
  awk -v v="$v" 'BEGIN {
    s = (v < 0) ? "-" : ""
    v = (v < 0) ? -v : v
    if (v >= 1000) printf "%s%s\n", s, group(v)
    else printf "%s%.1f\n", s, v
  }
  function group(x, n, c, str) {
    n = int(x + 0.5); str = ""
    do {
      c = n % 1000; n = int(n / 1000)
      if (n > 0) str = sprintf(",%03d%s", c, str)
      else str = c str
    } while (n > 0)
    return str
  }'
}

# meta_get: value of a JSON field in a meta file (numbers or strings).
meta_get() {
  jq -r --arg k "$2" '.[$k] // empty' "$1" 2>/dev/null
}

# export_metrics: token/tool/summary totals from an opencode session export.
# Prints tsv: input output reasoning cache_read cache_write cost tool_calls
#             files_changed additions deletions
export_metrics() {
  jq -r '[
    ([.messages[] | select(.info.role == "assistant") | .info.tokens.input] | add // 0),
    ([.messages[] | select(.info.role == "assistant") | .info.tokens.output] | add // 0),
    ([.messages[] | select(.info.role == "assistant") | .info.tokens.reasoning] | add // 0),
    ([.messages[] | select(.info.role == "assistant") | .info.tokens.cache.read] | add // 0),
    ([.messages[] | select(.info.role == "assistant") | .info.tokens.cache.write] | add // 0),
    ([.messages[] | select(.info.role == "assistant") | .info.cost] | add // 0),
    ([.messages[] | select(.info.role == "assistant") | .parts[]? | select(.type == "tool")] | length),
    (.info.summary.files // 0),
    (.info.summary.additions // 0),
    (.info.summary.deletions // 0)
  ] | @tsv' "$1" 2>/dev/null
}

# ---------------------------------------------------------------------------
# Discover tasks / conditions
# ---------------------------------------------------------------------------

TASKS=()
for d in "$RESULTS_DIR"/*/; do
  [ -d "$d" ] || continue
  b="$(basename "$d")"
  [ "$b" = "seq" ] && continue
  [ -d "$d/with_dizz" ] || [ -d "$d/without_dizz" ] || continue
  TASKS+=("$b")
done
mapfile -t TASKS < <(printf '%s\n' "${TASKS[@]}" | LC_ALL=C sort)

CONDS=()
for c in with_dizz without_dizz; do
  found=0
  for t in "${TASKS[@]:-}"; do
    if [ -d "$RESULTS_DIR/$t/$c" ]; then found=1; break; fi
  done
  [ "$found" -eq 1 ] && CONDS+=("$c")
done

# ---------------------------------------------------------------------------
# Single-task collection
# ---------------------------------------------------------------------------

# collect_cell: per-run rows for one (task, condition).
# columns: status success duration_s input output reasoning cache_read
#          cache_write tool_calls files_changed additions deletions
collect_cell() {
  local task="$1" cond="$2"
  local dir="$RESULTS_DIR/$task/$cond"
  local meta export_file exp
  for meta in "$dir"/run_*.meta.json; do
    [ -f "$meta" ] || continue
    local success duration_s input output reasoning cache_read cache_write tool_calls files_changed additions deletions
    success="$(meta_get "$meta" success)"; success="${success:-0}"
    duration_s="$(meta_get "$meta" duration_s)"; duration_s="${duration_s:-0}"
    export_file="$(meta_get "$meta" export_file)"
    exp="$dir/$export_file"
    input=""; output=""; reasoning=""; cache_read=""; cache_write=""
    tool_calls=""; files_changed=""; additions=""; deletions=""
    if [ -s "$exp" ]; then
      read -r input output reasoning cache_read cache_write _cost tool_calls files_changed additions deletions \
        <<< "$(export_metrics "$exp")" || true
    fi
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
      "$(meta_get "$meta" status)" "$success" "$duration_s" \
      "$input" "$output" "$reasoning" "$cache_read" "$cache_write" \
      "$tool_calls" "$files_changed" "$additions" "$deletions"
  done
}

# cond_col: all values of a column across a condition's runs.
cond_col() {
  local cond="$1" col="$2" t
  for t in "${TASKS[@]}"; do
    collect_cell "$t" "$cond" | cut -f"$col"
  done
}

cond_med() { cond_col "$1" "$2" | median; }

# ---------------------------------------------------------------------------
# Single-task report
# ---------------------------------------------------------------------------

emit "# dizz e2e benchmark summary"

# overall medians
emit ""
emit "## Overall medians"
emit "| metric | with_dizz | without_dizz | delta% |"
emit "|---|---|---|---|"
for spec in "input_tokens:input tokens:4" "output_tokens:output tokens:5" \
            "cache_read:cache reads:7" "duration_s:duration (s):3" \
            "tool_calls:tool calls:9" "files_changed:files changed:10"; do
  label="${spec%%:*}"; label="${label##*:}"; rest="${spec#*:}"
  label="${rest%%:*}"
  col="${rest##*:}"
  w="$(cond_med with_dizz "$col")"
  o="$(cond_med without_dizz "$col")"
  if [ "$w" != "-" ] && [ "$o" != "-" ] && [ "$o" != "0" ]; then
    delta="$(awk -v w="$w" -v o="$o" 'BEGIN { print ((w - o) / o) * 100 }')"
    emit "| $label | $(fmt "$w") | $(fmt "$o") | $(fmt "$delta" '%') |"
  else
    emit "| $label | $(fmt "$w") | $(fmt "$o") | - |"
  fi
done

# success rate
emit ""
emit "## Success rate"
emit "| task | with_dizz | without_dizz |"
emit "|---|---|---|"
for t in "${TASKS[@]}"; do
  for c in with_dizz without_dizz; do
    eval "$c=$(collect_cell "$t" "$c" | awk -F'\t' '{ if ($2==1) y++; n++ } END { printf "%d/%d", y, n }')"
  done
  emit "| $t | $with_dizz | $without_dizz |"
done

# per-task tables
for spec in "input_tokens:input tokens:4" "duration_s:duration (s):3" "tool_calls:tool calls:9"; do
  label="${spec#*:}"; label="${label%%:*}"
  col="${spec##*:}"
  emit ""
  emit "## Per task - $label"
  emit "| task | with_dizz | without_dizz | delta | delta% |"
  emit "|---|---|---|---|---|"
  for t in "${TASKS[@]}"; do
    w="$(collect_cell "$t" with_dizz | cut -f"$col" | median)"
    o="$(collect_cell "$t" without_dizz | cut -f"$col" | median)"
    if [ "$w" != "-" ] && [ "$o" != "-" ] && [ "$o" != "0" ]; then
      delta="$(awk -v w="$w" -v o="$o" 'BEGIN { print w - o }')"
      pct="$(awk -v w="$w" -v o="$o" 'BEGIN { print ((w - o) / o) * 100 }')"
      emit "| $t | $(fmt "$w") | $(fmt "$o") | $(fmt "$delta") | $(fmt "$pct" '%') |"
    else
      emit "| $t | $(fmt "$w") | $(fmt "$o") | - | - |"
    fi
  done
done

# honesty gate
emit ""
emit "## Honesty gate"
ws="$(awk '{ if ($1==1) y++; n++ } END { if (n>0) printf "%.1f", 100*y/n; else print 0 }' < <(cond_col with_dizz 2))"
os="$(awk '{ if ($1==1) y++; n++ } END { if (n>0) printf "%.1f", 100*y/n; else print 0 }' < <(cond_col without_dizz 2))"
emit "- with_dizz success rate: ${ws}%"
emit "- without_dizz success rate: ${os}%"
if awk -v w="$ws" -v o="$os" 'BEGIN { exit !(o > 0 && w < o - 20) }'; then
  emit ""
  emit "**WARNING: dizz success rate is >20 points below the control. Context savings must not be reported until parity is shown.**"
else
  emit "- OK: dizz success rate is within 20 points of the control."
fi

# ---------------------------------------------------------------------------
# all_metrics.csv
# ---------------------------------------------------------------------------
write_all_csv() {
  local csv="$RESULTS_DIR/all_metrics.csv"
  : > "$csv"
  {
    echo "task,condition,run,status,success,duration_s,rc,dizz_context_bytes,input_tokens,output_tokens,reasoning_tokens,cache_read,cache_write,cost,tool_calls,files_changed,additions,deletions"
    for t in "${TASKS[@]}"; do
      for c in with_dizz without_dizz; do
        for meta in "$RESULTS_DIR/$t/$c"/run_*.meta.json; do
          [ -f "$meta" ] || continue
          local run rc ctx exp input output reasoning cache_read cache_write cost tool_calls files_changed additions deletions
          run="$(meta_get "$meta" run)"; rc="$(meta_get "$meta" rc)"
          ctx="$(meta_get "$meta" dizz_context_bytes)"; ctx="${ctx:-0}"
          export_file="$(meta_get "$meta" export_file)"
          exp="$RESULTS_DIR/$t/$c/$export_file"
          input=""; output=""; reasoning=""; cache_read=""; cache_write=""; cost=""; tool_calls=""; files_changed=""; additions=""; deletions=""
          if [ -s "$exp" ]; then
            read -r input output reasoning cache_read cache_write cost tool_calls files_changed additions deletions \
              <<< "$(export_metrics "$exp")" || true
          fi
          echo "$t,$c,$run,$(meta_get "$meta" status),$(meta_get "$meta" success),$(meta_get "$meta" duration_s),$rc,$ctx,$input,$output,$reasoning,$cache_read,$cache_write,$cost,$tool_calls,$files_changed,$additions,$deletions"
        done
      done
    done
  } > "$csv"
}

# ---------------------------------------------------------------------------
# Sequence report
# ---------------------------------------------------------------------------

# seq_rows: per-session rows for one condition.
# columns: run session_index task success duration_s input output cache_read
seq_rows() {
  local cond="$1"
  local dir="$RESULTS_DIR/seq/$cond"
  [ -d "$dir" ] || return 0
  local meta export_file exp run idx task success duration_s input output cache_read
  for meta in "$dir"/run_*_session_*_*.meta.json; do
    [ -f "$meta" ] || continue
    run="$(meta_get "$meta" run)"; idx="$(meta_get "$meta" session_index)"
    task="$(meta_get "$meta" task)"
    success="$(meta_get "$meta" success)"; success="${success:-0}"
    duration_s="$(meta_get "$meta" duration_s)"; duration_s="${duration_s:-0}"
    export_file="$(meta_get "$meta" export_file)"
    exp="$dir/$export_file"
    input=""; output=""; cache_read=""
    if [ -s "$exp" ]; then
      read -r input output _reasoning cache_read _cache_write _cost _tools _files _adds _dels \
        <<< "$(export_metrics "$exp")" || true
    fi
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
      "$run" "$idx" "$task" "$success" "$duration_s" "$input" "$output" "$cache_read"
  done
}

build_seq_report() {
  local has_seq=0
  [ -d "$RESULTS_DIR/seq" ] && has_seq=1
  [ "$has_seq" -eq 0 ] && return

  local max_idx=0
  for c in with_dizz without_dizz; do
    while IFS=$'\t' read -r _run _idx _task _ok _dur _in _out _cr; do
      [ -n "$_idx" ] && (( _idx > max_idx )) && max_idx=$_idx
    done < <(seq_rows "$c")
  done
  [ "$max_idx" -eq 0 ] && return

  emit ""
  emit "# dizz e2e sequence benchmark"
  emit ""
  emit "One project per (condition, run); tasks run as consecutive opencode sessions on the same project. Session 1 is the first run; later sessions are subsequent runs on accumulated state."

  # per-session tables
  for spec in "input_tokens:input tokens:6" "output_tokens:output tokens:7" \
              "cache_read:cache reads:8" "duration_s:duration (s):5"; do
    label="${spec#*:}"; label="${label%%:*}"
    col="${spec##*:}"
    emit ""
    emit "## Per session - $label"
    emit "| session | with_dizz | without_dizz | delta | delta% |"
    emit "|---|---|---|---|---|"
    for (( i = 1; i <= max_idx; i++ )); do
      w="$(seq_rows with_dizz | awk -F'\t' -v i="$i" -v c="$col" '$2==i { print $c }' | median)"
      o="$(seq_rows without_dizz | awk -F'\t' -v i="$i" -v c="$col" '$2==i { print $c }' | median)"
      if [ "$w" != "-" ] && [ "$o" != "-" ] && [ "$o" != "0" ]; then
        delta="$(awk -v w="$w" -v o="$o" 'BEGIN { print w - o }')"
        pct="$(awk -v w="$w" -v o="$o" 'BEGIN { print ((w - o) / o) * 100 }')"
        emit "| $i | $(fmt "$w") | $(fmt "$o") | $(fmt "$delta") | $(fmt "$pct" '%') |"
      else
        emit "| $i | $(fmt "$w") | $(fmt "$o") | - | - |"
      fi
    done
  done

  # cumulative input tokens after each session
  emit ""
  emit "## Cumulative input tokens after each session"
  emit "| after session | with_dizz | without_dizz | delta% |"
  emit "|---|---|---|---|"
  for (( i = 1; i <= max_idx; i++ )); do
    w="$(seq_cumulative with_dizz "$i" 6 | median)"
    o="$(seq_cumulative without_dizz "$i" 6 | median)"
    if [ "$w" != "-" ] && [ "$o" != "-" ] && [ "$o" != "0" ]; then
      pct="$(awk -v w="$w" -v o="$o" 'BEGIN { print ((w - o) / o) * 100 }')"
      emit "| $i | $(fmt "$w") | $(fmt "$o") | $(fmt "$pct" '%') |"
    else
      emit "| $i | $(fmt "$w") | $(fmt "$o") | - |"
    fi
  done

  # first run vs subsequent runs
  emit ""
  emit "## First run vs subsequent runs"
  emit "| metric | with_dizz | without_dizz | delta% |"
  emit "|---|---|---|---|"
  for spec in "input_tokens:6" "output_tokens:7" "cache_read:8" "duration_s:5"; do
    metric="${spec%%:*}"
    col="${spec##*:}"
    wf="$(seq_first_sub with_dizz "$col" | cut -f1 | median)"
    wsub="$(seq_first_sub with_dizz "$col" | cut -f3 | grep -v '^-$' | median)"
    wtot="$(seq_first_sub with_dizz "$col" | cut -f2 | median)"
    of="$(seq_first_sub without_dizz "$col" | cut -f1 | median)"
    osub="$(seq_first_sub without_dizz "$col" | cut -f3 | grep -v '^-$' | median)"
    otot="$(seq_first_sub without_dizz "$col" | cut -f2 | median)"
    emit "| $metric - first run | $(fmt "$wf") | $(fmt "$of") | $(delta_pct "$wf" "$of") |"
    [ "$wsub" != "-" ] && emit "| $metric - avg per subsequent session | $(fmt "$wsub") | $(fmt "$osub") | $(delta_pct "$wsub" "$osub") |"
    emit "| $metric - total | $(fmt "$wtot") | $(fmt "$otot") | $(delta_pct "$wtot" "$otot") |"
  done

  # success rate by session
  emit ""
  emit "## Success rate by session"
  emit "| session | with_dizz | without_dizz |"
  emit "|---|---|---|"
  for (( i = 1; i <= max_idx; i++ )); do
    w="$(seq_rows with_dizz | awk -F'\t' -v i="$i" '{ if ($2==i) { if ($4==1) y++; n++ } } END { if (n>0) printf "%d/%d", y+0, n; else printf "-" }')"
    o="$(seq_rows without_dizz | awk -F'\t' -v i="$i" '{ if ($2==i) { if ($4==1) y++; n++ } } END { if (n>0) printf "%d/%d", y+0, n; else printf "-" }')"
    emit "| $i | $w | $o |"
  done

  # honesty gate (full-cell completion)
  emit ""
  emit "## Honesty gate (full-cell completion)"
  read -r wok wn <<< "$(seq_cell_success with_dizz)"
  read -r ook on <<< "$(seq_cell_success without_dizz)"
  wr="$(awk -v ok="$wok" -v n="$wn" 'BEGIN { printf "%.1f", (n>0 ? 100*ok/n : 0) }')"
  orr="$(awk -v ok="$ook" -v n="$on" 'BEGIN { printf "%.1f", (n>0 ? 100*ok/n : 0) }')"
  emit "- with_dizz full-cell success rate: ${wr}% (${wok}/${wn})"
  emit "- without_dizz full-cell success rate: ${orr}% (${ook}/${on})"
  if awk -v w="$wr" -v o="$orr" 'BEGIN { exit !(o > 0 && w < o - 20) }'; then
    emit ""
    emit "**WARNING: dizz success rate is >20 points below the control. Context savings must not be reported until parity is shown.**"
  else
    emit "- OK: dizz success rate is within 20 points of the control."
  fi
}

# seq_cumulative: running total of a column per cell, printed when idx reached.
seq_cumulative() {
  local cond="$1" idx="$2" col="$3"
  seq_rows "$cond" | sort -n -k1,1 -k2,2 | awk -F'\t' -v idx="$idx" -v col="$col" '
    $1 != cur { cur = $1; tot = 0 }
    { tot += $col + 0; if ($2 == idx) print tot }'
}

# seq_first_sub: per-cell "first\ttotal\tsubavg" rows.
seq_first_sub() {
  local cond="$1" col="$2"
  seq_rows "$cond" | sort -n -k1,1 -k2,2 | awk -F'\t' -v col="$col" '
    $col == "" { next }
    $1 != cur {
      if (cur != "") emit()
      cur = $1; fv = $col + 0; tot = 0; ssum = 0; scnt = 0
    }
    {
      tot += $col + 0
      if ($2 > 1) { ssum += $col + 0; scnt++ }
    }
    END { if (cur != "") emit() }
    function emit() {
      subv = (scnt > 0) ? (ssum / scnt) : "-"
      printf "%s\t%s\t%s\n", fv, tot, subv
    }'
}

# seq_cell_success: "ok total" where a cell counts as ok if all its sessions
# that have results succeeded.
seq_cell_success() {
  local cond="$1"
  seq_rows "$cond" | sort -n -k1,1 -k2,2 | awk -F'\t' '
    $1 != cur {
      if (cur != "" && have) { if (bad == 0) ok++; cells++ }
      cur = $1; bad = 0; have = 1
    }
    { if ($4 != 1) bad = 1 }
    END {
      if (cur != "" && have) { if (bad == 0) ok++; cells++ }
      printf "%d %d", ok + 0, cells + 0
    }'
}

delta_pct() {
  local w="$1" o="$2"
  if [ "$w" != "-" ] && [ "$o" != "-" ] && [ "$o" != "0" ]; then
    awk -v w="$w" -v o="$o" 'BEGIN { printf "%+.1f%%", ((w - o) / o) * 100 }'
  else
    echo "-"
  fi
}

# ---------------------------------------------------------------------------
# seq_metrics.csv
# ---------------------------------------------------------------------------
write_seq_csv() {
  local csv="$RESULTS_DIR/seq_metrics.csv"
  : > "$csv"
  {
    echo "condition,run,session_index,task,status,success,duration_s,rc,dizz_context_bytes,input_tokens,output_tokens,reasoning_tokens,cache_read,cache_write,cost,tool_calls,files_changed,additions,deletions"
    for c in with_dizz without_dizz; do
      [ -d "$RESULTS_DIR/seq/$c" ] || continue
      for meta in "$RESULTS_DIR/seq/$c"/run_*_session_*_*.meta.json; do
        [ -f "$meta" ] || continue
        local run idx rc ctx exp input output reasoning cache_read cache_write cost tool_calls files_changed additions deletions
        run="$(meta_get "$meta" run)"; idx="$(meta_get "$meta" session_index)"
        rc="$(meta_get "$meta" rc)"; ctx="$(meta_get "$meta" dizz_context_bytes)"; ctx="${ctx:-0}"
        export_file="$(meta_get "$meta" export_file)"
        exp="$RESULTS_DIR/seq/$c/$export_file"
        input=""; output=""; reasoning=""; cache_read=""; cache_write=""; cost=""; tool_calls=""; files_changed=""; additions=""; deletions=""
        if [ -s "$exp" ]; then
          read -r input output reasoning cache_read cache_write cost tool_calls files_changed additions deletions \
            <<< "$(export_metrics "$exp")" || true
        fi
        echo "$c,$run,$idx,$(meta_get "$meta" task),$(meta_get "$meta" status),$(meta_get "$meta" success),$(meta_get "$meta" duration_s),$rc,$ctx,$input,$output,$reasoning,$cache_read,$cache_write,$cost,$tool_calls,$files_changed,$additions,$deletions"
      done
    done
  } > "$csv"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

if [ "${#TASKS[@]}" -eq 0 ] && [ ! -d "$RESULTS_DIR/seq" ]; then
  echo "No benchmark results found under $RESULTS_DIR. Run ./run.sh first."
  exit 1
fi

if [ "${#TASKS[@]}" -gt 0 ]; then
  write_all_csv
  echo "CSV: $RESULTS_DIR/all_metrics.csv"
fi
if [ -d "$RESULTS_DIR/seq" ]; then
  build_seq_report
  write_seq_csv
  echo "CSV: $RESULTS_DIR/seq_metrics.csv"
fi
