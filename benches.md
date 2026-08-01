# Benchmarks

Hardware: Intel i5-10300H @ 2.50GHz, 8GB RAM, Linux, SSD

## CLI Performance (`dizz status` — dizz itself, 251 symbols, 77 Go files)

| Scenario | Real Time | Notes |
|----------|-----------|-------|
| Cold cache (first run) | **~900ms** | Full analysis + batch git |
| Warm cache (no changes) | **~35ms** | Signal hash match → short-circuit |
| Content change (existing function) | **~35ms** | Same signal set → short-circuit |
| Content change (new function) | **~80ms** | Scorer carries forward git data for unchanged symbols |
| State file size | **8.3 KB** | gzip, 93% reduction from ~109 KB |

## Agent Context (`dizz context` — the agent-facing path, dizz repo, 742 symbols)

Measured with `BenchmarkAgentContext` and `TestAgentContextEfficiency` in
`tests/context_benchmark_test.go`.

| Metric | Value | Notes |
|--------|-------|-------|
| Warm end-to-end (`dizz context`) | **~70ms** | Full analysis + TON render (CLI-measured 71-77ms, in-process 66ms) |
| Cold first run | **~740ms** | After cache wipe |
| Fresh context output | **1.8KB / ~450 tokens** | TON format, pipe-delimited |
| Raw state files replaced | **141.8KB** | state.ton.gz + intent.ton + context.ton + config.json |
| Byte reduction | **78x** | 141,795 bytes vs 1,816 bytes of context |

## Microbenchmarks

| Benchmark | Before | After | Improvement |
|-----------|--------|-------|-------------|
| CodeFiles (basic) | 231µs | **117µs** | `filepath.WalkDir` + extension map fast path |
| CodeFiles (20 excludes) | 482µs | **253µs** | Same |
| CodeFiles (large, ~1000 files) | 4.3ms | **2.0ms** | Same |
| DetectByExtension | 2.6µs | **2.4µs** | Unchanged |
| GetLanguageByID | 106ns | **98ns** | Already O(1) map |
| ScorerInterpretation (cold, 175 syms) | 445ms | **460ms** | Within noise (cold path unchanged) |
| Individual git 10 syms | 21ms | **23ms** | Unchanged |
| Batch git 175 syms | 51ms | **27ms** | Per-file batching + parallel workers (2x) |
| Batch git 100 syms | n/a | **27ms** | Good scaling (same cost as 175) |
| Extraction only (251 syms) | n/a | **25ms** | Combined line passes + O(1) analyzer lookup |
| Full analysis (cached) | n/a | **31ms** | Per-file signal cache hit path |
| GetSymbolsByState (500 syms) | n/a | **22µs** | Preallocated result slice |

### Batch git scaling

| Symbols | Individual | Batch | Speedup |
|---------|-----------|-------|---------|
| 10 | 23ms | **19ms** | 1.2x |
| 50 | 114ms | **27ms** | 4.2x |
| 100 | 227ms | **27ms** | 8.4x |
| 175 | 401ms | **27ms** | **15x** |

## Optimizations Applied

### Past optimizations
- **Scorer git carry-forward**: `InterpretSignalsWithIntent` accepts `prevState`; unchanged symbols (same name + location) reuse previous `InstabilityScore`, `ChurnCount`, `LastTouched` — no `git log -L` call
- **Content passthrough**: `AnalyzerWithContent` interface avoids triple file reads in the cache slow path
- **Detection cache**: Regex analyzer caches last `language.Detect()` result (avoids redundant detection in `Supports()` → `AnalyzeFile()`)
- **O(1) indexes**: `langIndex` map, `AllExtensions` init-time cache, `nameIndex`/`fileIndex` maps in scorer replace O(n) scans
- **Extension map**: File discovery checks `map[string]bool` for known extensions instead of looping over N patterns
- **fnv cache paths**: Cache store uses `hash/fnv64a` instead of SHA-256 for signal file naming (not security-sensitive)
- **Gitignore cache**: `.gitignore` parsing cached with mtime check
- **Skip save on identity**: Phase 4 short-circuit no longer writes state file when nothing changed
- **filepath.WalkDir**: Replaces `Walk` for ~2x fewer syscalls
- **Dead code removed**: Duplicate ignore-signal processing loop in scorer

### This round (v2 optimizations)
- **Batch git analysis (typed API)**: `BatchGitAnalysis` now takes `[]SymbolRange` instead of `[]interface{}`. Uses per-file `git log --name-only` (one call) + parallel `git log -L` workers for function churn. 15x faster at 175 symbols.
- **Rolling hash (FNV-1a)**: SignalSet uses incremental FNV-1a XOR hash instead of JSON marshal + SHA-256. Zero-allocation identity check.
- **Extension-only fast path**: `shouldInclude` checks `map[string]bool` for known extensions instead of looping over N `**/*.ext` patterns. O(1) per file.
- **gzip BestSpeed**: Cache store uses `gzip.BestSpeed` instead of `gzip.DefaultCompression`. ~same ratio at 2x faster compression.
- **Precomputed relative paths**: `SignalCache` now has `*Rel` methods accepting precomputed relative paths. Caller computes once per file → avoids N× `filepath.Rel` calls.
- **Compute only needed percentiles**: Scorer computes 75th and 90th percentiles only (was computing all from 0-100). Saves ~40% on sort costs.
- **Combined regex extraction passes**: `scanFile` makes 3 passes per line instead of 5. `extractDefinitions` merges functions+types; `extractAnnotations` merges todos+intents. Redundant `fnPatterns` re-check in `extractCalls` removed.
- **Extension-to-analyzer O(1) lookup**: Registry builds `extToAnalyzer` map at registration time. `FindAnalyzer` is now a direct extension map lookup instead of iterating analyzers calling `Supports()`.
- **Preallocated slices**: 15+ slice declarations across state, analyzer, common, signals changed from `var x []T` to `make([]T, 0, capacity)` where upper bound is known.
- **Adaptive worker pool**: Ignore marker workers capped at `min(files, 8)` to avoid spawning 8 goroutines for 2 files.
- **Buffer pooling**: `sync.Pool` for `bytes.Buffer` in gzip compression/decompression paths in cachestore and statestore.
- **nthLine helper**: `extractIgnoreTypeFromSignal` uses index-based line extraction instead of `strings.Split(source, "\n")` to find a single line.
