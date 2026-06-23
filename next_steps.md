## Summary

Incremental analysis pipeline and misc optimizations cutting warm `dizz status` from ~2.25s to **~35ms** (99% reduction) and content-change runs from ~960ms to **~100ms** (90% reduction). Also fixes `.gitignore` integration for nested dirs.

---

## Before vs After

| Scenario | Before | After |
|----------|--------|-------|
| Cold cache (first run) | ~2.25s | ~1.5s |
| Warm cache, no changes | ~2.25s | **~35ms** |
| Content change (new function) | ~960ms | **~100ms** |
| `state.json` size | ~109 KB | **~8.3 KB** (gzip) |

## Key Changes

- **Signal cache**: Per-file two-tier cache (mtime fast path → content hash) eliminates redundant re-analysis. Cache manifest + gzipped signal blobs at `.dizz/cache/`.
- **Scorer git carry-forward**: `InterpretSignalsWithIntent` accepts `prevState`; unchanged symbols reuse previous `InstabilityScore`, `ChurnCount`, `LastTouched` — no `git log -L` call. Biggest single win.
- **Signal hash short-circuit**: SHA256 of merged SignalSet stored in metadata; match → return cached state without entering scorer at all.
- **gzip state file**: `state.json.gz` with magic-number detection for backward compat. Atomic writes everywhere.
- **Git data carry-forward**: `ChurnCount`/`LastTouched` from previous state for unchanged files; `BatchGitAnalysis` only for changed files.
- **Content passthrough**: `AnalyzerWithContent` interface avoids triple file reads in the cache slow path.
- **O(1) indexes**: `langIndex` map, `AllExtensions` init-time cache, `nameIndex`/`fileIndex` maps in scorer replace O(n) scans.
- **Extension map**: File discovery checks `map[string]bool` for known extensions instead of looping over N patterns.
- **Worker pool**: Ignore marker scanning uses 8 workers for ≥20 files.
- **Gitignore cache**: `.gitignore` parsing cached with mtime check.
- **Skip identity save**: No state file write when signal hash matches.
- **fnv cache paths**: `hash/fnv64a` instead of SHA-256 for cache filenames.
- **WalkDir**: `filepath.WalkDir` replaces `Walk` for 2× fewer syscalls.

## Bug Fixes

- **Gitignore nested dirs**: Directory patterns (e.g. `node_modules/`) now produce `**/node_modules/**` instead of root-only `node_modules/**`. `matchPattern` handles `**/middle/**` with path-boundary checks to avoid substring false matches.
- **Atomic state save**: `SaveProjectState` uses temp-file + rename (was direct write → crash-corruptible).
- **Slice copy safety**: `excludePatterns` copies `cfg.Exclude` before appending gitignore patterns.
- **Dead code removed**: Duplicate ignore-signal processing loop in scorer.

| Benchmark | Before | After |
|-----------|--------|-------|
| CodeFiles (basic) | 231µs | **70µs** |
| CodeFiles (20 excludes) | 482µs | **171µs** |
| CodeFiles (large) | 4.3ms | **1.2ms** |
| Scorer cold | 445ms | ~483ms (within noise) |
| File discovery | ~5ms | **~0.1ms** |
