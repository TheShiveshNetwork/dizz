# Task: fix the ignored --verbose flag

The taskforge CLI has a `--verbose` flag that is parsed but currently does
nothing useful. When `--verbose` is passed, the program should print extra
debug output.

Investigate how the flag flows through the code. The startup status reporter in
`internal/notify` already supports a verbose mode - the problem is that it is
always called with verbose disabled.

Fix it so that:

1. `go run . --verbose list` prints a `[debug]` line and
2. `go run . list` (without the flag) prints no `[debug]` line.

Keep everything else working: `go build ./...`, `go vet ./...`, and
`go test ./...` must all pass when you are done.
