# Task: finish the unfinished features

The `internal/plan` package has features that were started but never finished.
There are TODO and FIXME markers in `internal/plan/plan.go` describing exactly
what is missing.

Finish them:

1. `plan --sort due` must work. It should sort tasks by due date, with tasks
   that have no due date placed last.
2. The `Remind` function must return the tasks due on or before a given time.
3. Fix the priority ranking bug noted in the FIXME.

Add tests for the new behavior if appropriate. When you are done,
`go build ./...`, `go vet ./...`, and `go test ./...` must all pass, and the
`plan --sort due` command must produce output instead of an error.
