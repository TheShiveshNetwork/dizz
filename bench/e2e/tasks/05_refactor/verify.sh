#!/usr/bin/env bash
# Verify task 05: internal/store refactored, tests pass, REFACTOR.md written.
set -uo pipefail

if ! go build ./... ; then echo "FAIL: build"; exit 1; fi
if ! go vet ./... ; then echo "FAIL: vet"; exit 1; fi
if ! go test ./... >/dev/null 2>&1; then echo "FAIL: tests"; exit 1; fi

if [ ! -f REFACTOR.md ]; then echo "FAIL: REFACTOR.md missing"; exit 1; fi
if ! grep -q 'internal/store' REFACTOR.md; then
  echo "FAIL: REFACTOR.md does not mention internal/store"
  exit 1
fi

changed="$(git diff --name-only -- 'internal/store/*.go' | wc -l)"
if [ "$changed" -lt 1 ]; then echo "FAIL: no internal/store file changed"; exit 1; fi

echo "PASS: store refactored and documented"
