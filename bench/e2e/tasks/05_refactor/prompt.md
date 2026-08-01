# Task: refactor internal/store without changing behavior

The `internal/store` package is messy: it mixes persistence concerns, seeding
logic, and leftover helper code. Refactor it to improve code quality without
changing any behavior.

Constraints:

1. The CLI behavior must stay identical. The public functions `Load` and `Add`
   must keep the same signatures and semantics.
2. `go build ./...`, `go vet ./...`, and `go test ./...` must all pass.
3. Do not delete tests.

After you finish, write `REFACTOR.md` at the repository root listing every file
you changed, with one sentence each on why.
