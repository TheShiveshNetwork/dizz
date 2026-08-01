# taskforge

A small command-line task manager used as the seed project for the dizz
benchmark suite (`bench/e2e`). It tracks a handful of tasks with due dates and
priorities.

## Usage

```bash
go run . list --all
go run . plan --sort priority
go run . add "write docs" --due 2026-09-01
```

## Layout

- `main.go` - CLI entry point
- `internal/store` - task persistence (JSON file in the working directory)
- `internal/plan` - sorting and due-date logic
- `internal/render` - output formatting
- `internal/notify` - startup status reporting
