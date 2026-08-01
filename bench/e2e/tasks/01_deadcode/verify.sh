#!/usr/bin/env bash
# Verify task 01: dead code removed, build + tests still pass.
set -uo pipefail

if ! go build ./... ; then echo "FAIL: build"; exit 1; fi
if ! go vet ./... ; then echo "FAIL: vet"; exit 1; fi
if ! go test ./... >/dev/null 2>&1; then echo "FAIL: tests"; exit 1; fi

# These symbols are planted dead code and must be gone.
if grep -rEq 'deduplicateTasks|legacyMigrate|compactArchive|TaskArchiver' --include='*.go' .; then
  echo "FAIL: dead code still present"
  exit 1
fi

echo "PASS: dead code removed, tests green"
