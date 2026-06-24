# Benchmarks

Hardware: Intel i5-10300H @ 2.50GHz, 16GB RAM, Linux, SSD

## CLI Performance (`dizz status` — dizz itself, 248 symbols, 77 Go files)

| Scenario | Real Time | Notes |
|----------|-----------|-------|
| Cold cache (first run) | **~1.5s** | Full analysis + git per new symbol |
| Warm cache (no changes) | **~35ms** | Signal hash match → short-circuit |
| Content change (existing function) | **~35ms** | Same signal set → short-circuit |
| Content change (new function) | **~100ms** | Scorer carries forward 247/248 InstabilityScores |
| State file size | **8.3 KB** | gzip, 93% reduction from ~109 KB |

## Microbenchmarks

| Benchmark | Before | After | Improvement |
|-----------|--------|-------|-------------|
| CodeFiles (basic) | 231µs | **70µs** | `filepath.WalkDir` + extension map fast path |
| CodeFiles (20 excludes) | 482µs | **171µs** | Same |
| CodeFiles (large, ~1000 files) | 4.3ms | **1.2ms** | Same |
| DetectByExtension | 2.6µs | **2.6µs** | Unchanged |
| GetLanguageByID | 106ns | **106ns** | Already O(1) map |
| ScorerInterpretation (cold) | 445ms | **483ms** | Within noise (cold path unchanged) |
| Individual git 10 syms | 21ms | **21ms** | Unchanged |
| Batch git 175 syms | 51ms | **51ms** | Unchanged |

The cold scorer benchmark is unchanged (no prevState in benchmarks). The real win is in the warm "content change" path: **959ms → ~100ms** because the scorer skips git for 247/248 symbols by carrying forward `InstabilityScore` from the previous state.

## Optimizations Applied

- **Scorer git carry-forward**: `InterpretSignalsWithIntent` accepts `prevState`; unchanged symbols (same name + location) reuse previous `InstabilityScore`, `ChurnCount`, `LastTouched` — no `git log -L` call
- **Content passthrough**: `AnalyzerWithContent` interface avoids triple file reads in the cache slow path
- **Detection cache**: Regex analyzer caches last `language.Detect()` result (avoids redundant detection in `Supports()` → `AnalyzeFile()`)
- **O(1) indexes**: `langIndex` map, `AllExtensions` init-time cache, `nameIndex`/`fileIndex` maps in scorer replace O(n) scans
- **Extension map**: File discovery checks `map[string]bool` for known extensions instead of looping over N patterns
- **fnv cache paths**: Cache store uses `hash/fnv64a` instead of SHA-256 for signal file naming (not security-sensitive)
- **Worker pool**: Ignore marker scanning in sequential path uses 8 workers for ≥20 files
- **Gitignore cache**: `.gitignore` parsing cached with mtime check
- **Skip save on identity**: Phase 4 short-circuit no longer writes state file when nothing changed
- **filepath.WalkDir**: Replaces `Walk` for ~2x fewer syscalls
- **Dead code removed**: Duplicate ignore-signal processing loop in scorer
