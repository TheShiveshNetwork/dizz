# Task: assess the project and fix the most important problem

Review the taskforge codebase and identify what most needs work. Look for
unfinished features, broken behavior, and risky code.

Write your analysis to a file named `REPORT.md` at the repository root. It must
contain exactly three numbered items (`1.`, `2.`, `3.`) ordered by impact, each
with a one-sentence reason.

Then implement the most important item you identified. When you are done:

1. `go build ./...`, `go vet ./...`, and `go test ./...` must pass.
2. Your change must be a real code change (not only documentation).
3. `REPORT.md` must exist with the three numbered items.
