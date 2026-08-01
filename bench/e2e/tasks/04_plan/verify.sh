#!/usr/bin/env bash
# Verify task 04: REPORT.md with 3 priorities + a real code change.
set -uo pipefail

if ! go build ./... ; then echo "FAIL: build"; exit 1; fi
if ! go vet ./... ; then echo "FAIL: vet"; exit 1; fi
if ! go test ./... >/dev/null 2>&1; then echo "FAIL: tests"; exit 1; fi

if [ ! -f REPORT.md ]; then echo "FAIL: REPORT.md missing"; exit 1; fi
for n in 1 2 3; do
  if ! grep -qE "^${n}\." REPORT.md; then
    echo "FAIL: REPORT.md missing item $n."
    exit 1
  fi
done

changed="$(git diff --name-only -- '*.go' | wc -l)"
if [ "$changed" -lt 1 ]; then echo "FAIL: no Go file changed"; exit 1; fi

echo "PASS: report present with 3 priorities and a code change"
