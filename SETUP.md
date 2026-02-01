# Project Setup & Development Guide (Dizz)

This document explains how to **set up, build, run, and develop** the `dizz` CLI using the provided `Makefile`. It is intended for contributors and maintainers so everyone follows the same workflow.

---

## Prerequisites

Before you begin, make sure you have the following installed:

* **Go (1.21+)** – required to build and run the project
* **GNU Make** – used to run the provided commands
* *(Optional)* **golangci-lint** – required only if you want to run `make lint`

Verify installation:

```bash
go version
make --version
```

---

## Project Build Configuration

The Makefile defines a few important variables:

| Variable      | Description                                            |
| ------------- | ------------------------------------------------------ |
| `BINARY_NAME` | Name of the compiled CLI binary (`dizz`)               |
| `BUILD_DIR`   | Directory where the compiled binary is placed (`bin/`) |
| `MAIN_PKG`    | Entry point package for the application (`.`)          |
| `GO`          | Go compiler command                                    |
| `GOFLAGS`     | Go build flags (`-trimpath` for reproducible builds)   |

These variables keep the Makefile clean and easy to change later.

---

## Common Commands

### 🔨 Build the Project

```bash
make build
```

What this does:

* Creates a `bin/` directory if it doesn’t exist
* Compiles the Go project
* Outputs the binary to `bin/dizz`

Use this when you want a reusable binary.

---

### 🚀 Run the CLI

```bash
make run
```

What this does:

* Runs `make build`
* Executes the compiled binary (`bin/dizz`)

This is the standard way to run the tool locally.

---

### ▶️ Run with Arguments

```bash
make run-args ARGS="status"
```

What this does:

* Builds the binary
* Runs it with custom CLI arguments

Example:

```bash
make run-args ARGS="scan --all"
```

This is useful for testing specific commands.

---

### ⚡ Fast Development Mode

```bash
make dev
```

What this does:

* Builds the binary into `/tmp/dizz`
* Immediately runs it

Why this exists:

* Faster than full builds
* Avoids polluting the project directory
* Ideal for rapid iteration

---

### 🧪 Run Tests

```bash
make test
```

What this does:

* Runs all Go tests recursively (`./...`)

Use this before committing or opening a PR.

---

### 🎨 Format Code

```bash
make fmt
```

What this does:

* Formats all Go files using `go fmt`

Always run this before committing code.

---

### 🔍 Lint Code (Optional)

```bash
make lint
```

Requirements:

* `golangci-lint` must be installed

What this does:

* Runs static analysis and style checks

Install lint tool:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

### 🧹 Clean Build Artifacts

```bash
make clean
```

What this does:

* Removes `bin/`
* Removes temporary dev binaries

Use this if builds behave strangely or before a fresh build.

---

### 📦 Install Locally

```bash
make install
```

What this does:

* Installs `dizz` into your `$GOPATH/bin` or `$GOBIN`

After this, you can run:

```bash
dizz
```

from anywhere on your system.

---

## Recommended Development Flow

```text
make fmt      # format code
make test     # run tests
make dev      # quick local run
make build    # final build
```

---

## Notes

* The `Makefile` avoids Git hooks to keep history clean
* Generated files (e.g. `.dizz` snapshots) should be gitignored
* All commands are safe to run on Linux, macOS, and WSL

---

Happy hacking 🚀

