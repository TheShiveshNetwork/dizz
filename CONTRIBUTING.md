# Contributing

Thanks for your interest in `dizz`! This document covers everything you need to contribute effectively.

## Development Setup

**Prerequisites**: Go 1.26+

```bash
# Clone and build
git clone https://github.com/TheShiveshNetwork/dizz.git
cd dizz
make build

# Run tests
make test

# Run benchmarks
make bench

# Format code
make fmt

# Lint (requires golangci-lint)
make lint
```

## Project Structure

```
cmd/              # CLI command implementations (cobra commands)
internal/
  analyzer/       # Static analysis (AST for Go, regex for all languages)
  common/         # Shared utilities, project root detection
  config/         # Configuration types and paths
  defaults/       # Sensible defaults, skill templates
  discover/       # File discovery with gitignore support
  integrations/   # Git integration
  language/       # Language registry (34 languages)
  render/         # Context and output rendering
  signals/        # Signal types and collection
  skill/          # Agent skill installation
  state/          # Project state model, scorer, snapshots, TON
  store/          # Persistence (intent store, cache, TON codec)
  ui/             # Terminal output formatting
agent-skills/     # Canonical SKILL.md for AI agent discovery
scripts/          # Install scripts and site build
site/             # WASM documentation site
tests/            # Integration and conformance tests
```

## Code Structure Rules

- All internal packages use short, lowercase, single-word names (`state`, `signals`, `discover`, etc.).
- CLI command implementations go in `cmd/`. Backend/internal code goes under `internal/`.
- Tests go in `tests/` (package `benchmarks`). Benchmark files are named `*_benchmark_test.go`.
- The analysis pipeline is strict: `discover.CodeFiles` -> `analyzer.Registry.AnalyzeFiles` -> `signals.SignalSet` -> `state.Scorer.InterpretSignalsWithIntent` -> `state.ProjectState`. Never bypass layers.
- The language registry (`internal/language/registry.go`) is the single source of truth for extensions, comment styles, and patterns. Never hardcode these elsewhere.

## Coding Conventions

1. **Zero comments in production code** — Intent is captured via `dizz intent` or `AGENTS.md`, not inline. TODO/FIXME comments are acceptable only in test files.

2. **Error handling** — Return errors to the caller. Internal packages never call `log.Fatal` or `os.Exit`. Use `fmt.Errorf` with `%w` for wrapping.

3. **No external dependencies beyond stdlib + cobra** — Currently only depends on `github.com/spf13/cobra`. New dependencies require explicit justification and maintainer review.

4. **Signal metadata** — When adding new signal types or metadata keys, prefix project-specific keys with the package name (e.g., `"regex.source_tier"`). All confidence values are `float64` in [0, 1].

5. **Immutable intents, mutable todos** — TODO/FIXME found in source are ephemeral and re-scanned on every analysis. Intents in `.dizz/intent.ton` are immutable project records — never modify programmatically, only add or resolve.

6. **Config always optional** — dizz must work with zero configuration. Defaults are in `internal/defaults/defaults.go`. Config values override defaults, never require config.

7. **Snapshot immutability** — Objects in `.dizz/objects/` are content-addressed and never modified after creation. Snapshots use compact JSON (no indentation).

## Testing

- Every language gets a conformance test — at minimum, verify `FunctionDefined` extraction. Use the `analyzeContent(t, ext, src)` helper from `tests/language_conformance_test.go`.
- Add or update benchmarks when modifying throughput-sensitive code. Key benchmarks: `BenchmarkScorerInterpretation`, `BenchmarkCodeFiles`, `BenchmarkDetectByExtension`.
- Git-dependent tests must skip gracefully: `if !integrations.IsRepo() { t.Skip(...) }`.
- Always use `t.TempDir()` for temp files. Never create temp files outside the test's temp directory.

## Pull Request Process

1. Run `make fmt && make lint && make test` before submitting.
2. Update or add tests for any new functionality.
3. If adding a new language, add a conformance test.
4. Keep the commit history clean — squash if needed.
5. Link to any related issues.

## AI Agent Development

`dizz` is designed as a state-aware assistant for both humans and AI agents. If you're extending agent-facing features:

- **`dizz context`** is the single entry point for agents — keep it lean (~2 KB).
- All agent data uses TON format (pipe-delimited, line-oriented, no JSON overhead).
- Agent skills follow the [Agent Skills Specification](https://agentskills.io). The canonical skill is at `agent-skills/dizz/SKILL.md`.
- When adding new analysis signals, update `AGENTS.md` so agents know about them.

See [`AGENTS.md`](AGENTS.md) for the full agent protocol and workflow documentation.

## Site Build

The WASM documentation site at `site/` is built by `scripts/build-site.sh`. Source content that needs to appear in the built site lives at the project root, not in `site/` subdirectories:
- Install scripts: `scripts/install.sh`, `scripts/install.ps1`
- Agent skills: `agent-skills/`
- Assets: `site/assets/` (only exception — kept with the site)

Never put source content directly in `site/public/` or create new source directories under `site/`. Update `scripts/build-site.sh` to copy from the root-level location instead.
