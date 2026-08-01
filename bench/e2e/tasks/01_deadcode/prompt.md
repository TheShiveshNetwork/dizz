# Task: remove dead code

The taskforge codebase contains dead code - functions and types that are
declared but never used anywhere. Dead code is confusing to maintainers and
inflates the project.

Find all unused code and remove it. When you are done:

1. The project must still build: `go build ./...`
2. All tests must still pass: `go test ./...`
3. No function or type that is never called or referenced may remain.

Do not remove or change anything that is actually used by the CLI or by
another function that is itself used. Do not remove tests.
