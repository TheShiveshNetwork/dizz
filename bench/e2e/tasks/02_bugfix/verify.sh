#!/usr/bin/env bash
# Verify task 02: --verbose wired through to notify.
set -uo pipefail

if ! go build ./... ; then echo "FAIL: build"; exit 1; fi
if ! go vet ./... ; then echo "FAIL: vet"; exit 1; fi
if ! go test ./... >/dev/null 2>&1; then echo "FAIL: tests"; exit 1; fi

if ! go run . --verbose list 2>&1 | grep -q '\[debug\]'; then
  echo "FAIL: --verbose does not print [debug]"
  exit 1
fi

if go run . list 2>&1 | grep -q '\[debug\]'; then
  echo "FAIL: [debug] printed without --verbose"
  exit 1
fi

echo "PASS: verbose flag wired through"
