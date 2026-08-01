# AGENTS.md

## Project overview

taskforge is a small command-line task manager written in Go (stdlib only).
The CLI entry point is `main.go`; the `internal/` packages are `store`
(persistence), `plan` (sorting and due dates), `render` (output), and `notify`
(startup status).

## Commands

- `go build ./...` - build
- `go test ./...` - run tests
- `go vet ./...` - vet
- `go run . list [--all]` - list tasks
- `go run . plan --sort <priority|due>` - show a sorted plan
- `go run . add "<title>" [--due YYYY-MM-DD]` - add a task

## Conventions

- Stdlib only - never add external dependencies.
- Return errors to callers; log only at the CLI layer in `main.go`.
- Keep tests passing at all times.
