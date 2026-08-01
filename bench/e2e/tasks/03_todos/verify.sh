#!/usr/bin/env bash
# Verify task 03: due sorting + Remind implemented, priority bug fixed.
set -uo pipefail

if ! go build ./... ; then echo "FAIL: build"; exit 1; fi
if ! go vet ./... ; then echo "FAIL: vet"; exit 1; fi
if ! go test ./... >/dev/null 2>&1; then echo "FAIL: tests"; exit 1; fi

# "not implemented" placeholders must be gone from the plan package.
if grep -rq 'not implemented' internal/plan/; then
  echo "FAIL: plan package still contains 'not implemented'"
  exit 1
fi

# Due sorting must produce ordered output (no error).
# Filter to plan rows (they contain ' | ') to ignore the notify header line.
out="$(go run . plan --sort due 2>&1 | grep ' | ')"
if [ "$?" -ne 0 ]; then
  echo "FAIL: plan --sort due errors"
  exit 1
fi
if ! echo "$out" | grep -q '2026-08-01'; then
  echo "FAIL: due-sorted output missing earliest due date"
  exit 1
fi
# Earliest due date (2026-08-01) must appear before the later one.
first="$(echo "$out" | head -1)"
case "$first" in
  2026-08-01*) ;;
  *) echo "FAIL: first plan row is not the earliest due date: $first"; exit 1 ;;
esac

echo "PASS: due sorting and Remind implemented"
