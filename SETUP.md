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

## Common Commands

### 🔨 Build the Project

```bash
make build
```

What this does:
* Compiles the Go project
* Outputs the binary to `bin/dizz`

---

### 🚀 Run the CLI

```bash
make run
```
Runs `make build` and executes `bin/dizz`.

### ▶️ Run with Arguments

```bash
make run-args ARGS="status"
```
Runs the binary with custom CLI arguments.

---

### 🧪 Testing & Benchmarking

#### Run All Tests
```bash
make test
```
Runs all tests with verbose output.

#### Run Benchmarks
```bash
make bench
```
Runs performance benchmarks with memory statistics.

---

### 📊 Performance Benchmarks (Current)

Based on a codebase with ~40 files and 175 symbols on an Intel i5-10300H CPU:

| Operation | Time | Notes |
|-----------|------|-------|
| **Symbol Extraction** | ~1.3s | AST & Regex analysis |
| **Git Analysis (Batched)** | ~50ms | 175 symbols |
| **Git Analysis (Individual)** | ~370ms | Legacy approach |
| **Total Full Analysis** | ~1.4s | Cold start |

**Optimization Impact:** Batched git operations are **~7x faster** than individual calls.

---

### 🛠 Development Utilities

* `make dev`: Fast rebuild and run from `/tmp`.
* `make fmt`: Format all Go files.
* `make lint`: Run static analysis (requires `golangci-lint`).
* `make clean`: Remove build artifacts.
* `make install`: Install `dizz` to your system path.

---

### 🌐 Website & WASM

* `make build-site`: Builds the documentation site.
* `make wasm`: Compiles the WASM binary for the web.
* `make serve-site`: Starts a local server for the site.

---

## CLI Command Reference

| Command | Description |
|---------|-------------|
| `dizz init` | Initialize dizz in the current directory |
| `dizz status` | Quick project health check and summary |
| `dizz log` | Detailed analysis of planned, unstable, and unused code |
| `dizz snapshot` | Create an immutable state snapshot |
| `dizz list` | List saved snapshots |
| `dizz resume` | Get context after being away from the project |
| `dizz intent` | Manage human-authored intents (todos, etc.) |
| `dizz todo list` | View extracted code markers (TODOs, FIXMEs) |
| `dizz upgrade` | Upgrade to the latest version |
| `dizz version` | Show current version |

---

## Recommended Development Flow

```bash
make fmt      # format code
make test     # run tests
make bench    # check performance
make dev      # quick local run
```

Happy hacking 🚀
