## SPEED IMPROVEMENTS

## Measured Baselines

Hardware: Intel i5-10300H @ 2.50GHz, 16GB RAM, Linux, SSD
Benchmark target: dizz analyzing itself (245 symbols, ~77 Go files + supporting files)

| Component | Current Performance | Notes |
|-----------|-------------------|-------|
| File discovery (small) | 88 µs | 354 allocations |
| File discovery (large, ~1000 files) | 1.5 ms | 4178 allocations |
| Signal extraction (all files) | 21 ms | Analyzer, regex pattern matching |
| Scorer cold (no previous state) | 552 ms | Dominated by git operations |
| Individual git 175 symbols | 794 ms | 40k+ allocations |
| Batch git 175 symbols (cache miss) | 83 ms | Batch git is 10x faster |
| Each individual git subprocess | ~1 ms + 65 KB allocs | Cost of exec.Command + subprocess |
| Full CLI warm (no changes) | 35 ms | Signal hash short-circuit |

**Critical finding: For a project with just 245 symbols, the cold analysis takes ~1.5s because every symbol triggers individual `git log -L` subprocesses. For 10,000 symbols this would take 60+ seconds purely in git subprocess overhead.**

---

## Bottleneck #1: Per-Symbol Git Subprocess Spawning

### Location: `internal/integrations/git.go:111` (`GetFunctionChurn`), `:41` (`GetFunctionCommits`), `:54` (`GetFileLastModified`)

### Root cause
Every symbol's instability scoring calls `GetFunctionCommits()` which invokes `exec.Command("git", "log", "-L", ...)` as a separate OS subprocess. For 10k symbols, this creates 10k subprocesses. Each subprocess:
- Costs ~1ms to fork + exec
- Allocates ~65KB of memory (benchmarked)
- Waits for the process to initialize, read git objects, and produce output
- The `-L` flag is particularly expensive because git recomputes line-log history

### Proof from benchmarks
```
BenchmarkIndividualGitCalls/symbols_10-8         49 ops,  21.4ms (2.1ms/symbol)
BenchmarkIndividualGitCalls/symbols_50-8         10 ops, 106.7ms (2.1ms/symbol)
BenchmarkIndividualGitCalls/symbols_100-8         5 ops, 214.0ms (2.1ms/symbol)
BenchmarkIndividualGitCalls/symbols_175-8         2 ops, 794.4ms (4.5ms/symbol)
```
Linearly scales with symbol count. 175 symbols = 794ms, 10k symbols = ~45 seconds.

### Fix strategy

**Replace per-symbol `git log -L` with file-level `git log` (single read per file):**
1. For each file, run `git log --numstat -- <file>` once
2. Parse ALL function-level changes from the single file history output in Go
3. Match parsed changes to symbol line ranges in code

This turns `O(symbols)` git calls into `O(files)` git calls. For a project with 10k symbols across 1000 files: 1000 file-level git calls instead of 10k symbol-level calls (potentially 50-200 function ranges per file batched in one git log).

**Stage 2: Use `git diff` between commits for batch change size**
Instead of calling `git show <hash>` per-commit for change size detection, batch it: run `git diff --numstat <hash>^ <hash> -- <file>` once per commit batch and parse all changed lines.

---

## Bottleneck #2: File Discovery Pattern Matching

### Location: `internal/discover/files.go:94` (`matchPattern`), `:171` (`shouldInclude`)

### Root cause
The `matchPattern` function does string-based glob matching with iterative wildcard expansion. For each file, `shouldInclude` iterates over all include patterns until a match is found. With 34 languages, the default pattern list has 34 entries. For 10k files, that's 340k `matchPattern` calls.

### Proof from benchmarks
```
BenchmarkCodeFiles-8               15279 ops,  88µs (small project)
BenchmarkCodeFilesLargeProject-8     837 ops, 1.5ms (~1000 files)
```
Discovery alone takes 1.5ms for 1000 files. For 100k files, this would be ~150ms — not the biggest bottleneck alone, but combined with the other linear scans it adds up.

### Fix strategy

**Replace pattern matching with `filepath.Ext` map lookup in the common case:**
The function `patternsAreExtensionBased` already detects extension-only patterns but `shouldInclude` still falls through to looping. Fix: when all patterns are `**/*.ext`, use a single `map[string]bool` lookup (already computed as `extensionSet()`) and skip the loop entirely. This is partially done but the fast path doesn't early-return — it still falls through to the loop.

**Use `filepath.WalkDir` properly:**
Current code uses `WalkDir` but the callback does multiple `shouldExclude` calls and string operations. Refactor to pre-filter before the callback using a trie-based exclusion structure for node_modules and other common exclusion patterns.

---

## Bottleneck #3: Signal Hash Computation (JSON + SHA-256)

### Location: `internal/cmd/state.go:208-210`

### Root cause
To detect identical analysis results, the entire signal set is serialized to JSON and SHA-256 hashed. For large projects with millions of signals, this is:
- Memory: creating a full JSON blob of all signals
- CPU: SHA-256 of potentially megabytes of JSON
- GC pressure: bytes.Buffer + JSON marshal buffers + hash buffers

### Fix strategy

**Replace JSON marshal + SHA-256 with an incremental hash:**
Track a rolling hash (e.g., FNV-1a or xxHash) during signal collection. Each time a signal is added to the SignalSet, update the rolling hash with `hash.Write(signalBytes)`. This eliminates the need to re-serialize everything just for comparison.

**Or: Use content-addressed hashing of individual files:**
Instead of hashing the aggregate signal set, track `map[filename]contentHash`. If no file has changed content, no analysis results can differ. This avoids signal serialization entirely.

---

## Bottleneck #4: Ignore Marker Extraction String Splitting

### Location: `internal/signals/flags.go:94` (`ExtractIgnoreMarkers`)

### Root cause
Each call to `ExtractIgnoreMarkers` splits the entire file content into lines with `strings.Split(source, "\n")`. This allocates a full slice of strings for every file. Combined with the regex analyzer's line-by-line scanning, every file is split twice.

The `findNextSymbol` and `findSymbolEnd` functions also allocate new line slices and iterate character-by-character.

### Fix strategy

**Pass a scanner (shared interface) instead of splitting:**
The regex analyzer already uses `bufio.Scanner`. Share this scanner interface or pass a `[]string` of lines to avoid the double split. Better yet, use `strings.IndexByte` and `strings.Index` on the raw content with offset tracking instead of splitting.

**Replace `findSymbolEnd` brace counting with line range heuristics:**
Brace counting by scanning entire file for `{` and `}` per ignore marker is extremely expensive. Use the existing signal data: symbol end lines are already known from function definition extraction.

---

## Bottleneck #5: Scorer Percentile Computation

### Location: `internal/state/scorer.go:141` (`calculatePercentiles`)

### Root cause
`calculatePercentiles` computes 101 percentile points (0-100 inclusive) by iterating and interpolating. This is called once for instability scores and once for age scores. For projects with thousands of symbols, sorting O(n log n) is fine, but computing 101 interpolated values is wasteful when only the 75th and 90th percentiles are actually used.

### Fix strategy

**Compute only needed percentiles on demand:**
Replace the `map[float64]float64` full computation with a function `percentile(sorted []float64, p float64) float64`. The 101-point map is never used — only 75.0 and 90.0 are accessed.

```go
func percentile(sorted []float64, p float64) float64 {
    if len(sorted) == 0 {
        return 0
    }
    rank := (p / 100.0) * float64(len(sorted)-1)
    lower := int(math.Floor(rank))
    upper := int(math.Ceil(rank))
    if lower == upper {
        return sorted[lower]
    }
    weight := rank - float64(lower)
    return sorted[lower]*(1-weight) + sorted[upper]*weight
}
```

---

## Bottleneck #6: Redundant `filepath.Rel` Calls in Caching

### Location: `internal/store/cachestore.go`

### Root cause
Every cache operation (`GetByMTime`, `Get`, `Set`, `Evict`, `EvictStale`, `MTimeInCache`) calls `filepath.Rel(sc.projectRoot, filePath)`. This is called multiple times per file per analysis pass. `filepath.Rel` is relatively expensive — it does path cleaning, comparison, and relative path computation.

### Fix strategy

**Precompute relative paths in the main analysis loop:**
Compute `relPath` once per file in `runCurrentAnalysisAtRoot` and pass it down the chain. Add a `SetWithRelPath` / `GetByRelPath` variant to SignalCache that accepts precomputed relative paths.

---

## Bottleneck #7: Analysis Pipeline Serialization with Unnecessary Phase Breaks

### Location: `internal/cmd/state.go:149-198`

### Root cause
The analysis pipeline has two phases for signal collection:
- Phase 1: Calls `registry.AnalyzeSingleFile(file, content)` which reads the file, detects language, runs regex/ast analysis, AND runs ignore marker extraction
- Phase 2: Merges all signals into a SignalSet
- Phase 3: Serializes the entire SignalSet to JSON for hash computation

The issue is that each file is fully read and analyzed, then ALL results are merged, then ALL results are serialized for hashing. This means gigabytes of memory for large projects.

### Fix strategy

**Stream the pipeline:**
1. Analyze file → emit signals directly to scoring pipeline
2. Update incremental hash as signals are emitted
3. No intermediate serialization of the entire signal set

**Use an iterator/visitor pattern:**
```go
type SignalVisitor func(signal signals.Signal) error
```

Pass this through the analysis chain so signals are processed immediately without storage.

---

## Bottleneck #8: Gzip DefaultCompression in State Storage

### Location: `internal/store/statestore.go:88-91`, `cachestore.go:263`

### Root cause
Both `SaveProjectState` and `storeSignals` use `gzip.DefaultCompression` which prioritizes compression ratio over speed. For state files that are read/written frequently, this adds unnecessary CPU time. The benchmark says the state file is 8.3 KB gzipped — at this size, `gzip.BestSpeed` would be nearly as compact but 3-5x faster.

### Fix strategy

**Switch to `gzip.BestSpeed` (level 1):**
```go
gzWriter, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
```

For the signal cache files (which are per-file and even smaller), consider switching to `gzip.NoCompression` or just JSON with snappy compression, or skip compression entirely and rely on the filesystem-level compression.

---

## Bottleneck #9: Memory Allocation Pattern - Repeated New + Append

### Location: Throughout the codebase

### Root cause
The code uses Go slices with default initialization and `append` extensively:
- `symbolSlice` in scorer always starts nil and grows via append
- `nameIndex` and `fileIndex` grow per-symbol
- `mergedSigSet` stores ALL signals in a flat slice
- Each `AnalyzeFileContent` call creates new `SignalSet` structs
- The `analyzeSequentially` path creates per-file channels and goroutines for as few as 20 files

### Fix strategy

**Preallocate all slices where capacity is known:**
```go
symbolSlice := make([]*Symbol, 0, len(symbolIndex))
symbolSlice = append(symbolSlice, symbolSlice...)
```

**Pool intermediate buffers:**
Use `sync.Pool` for frequently allocated types like `[]signals.Signal`, `bytes.Buffer`, and JSON encoder buffers.

---

## Bottleneck #10: Regex Engine Overhead in Signal Extraction

### Location: `internal/analyzer/regex/analyzer.go:85-196`

### Root cause
The `scanFile` function runs 5 separate regex extractors per line (functions, types, calls, todos, intents). Each extractor compiles and matches against the language-specific regex patterns. For a file with 1000 lines, this is 5000 regex operations.

Additionally, `extractCalls` checks if the line is a function definition by re-matching against `cl.fnPatterns` — this is redundant since `extractFunctions` already ran.

### Fix strategy

**Combine extraction into a single pass per line:**
Track whether the current line was a function definition during the first pass. Skip call extraction for definition lines without re-matching patterns.

**Consolidate regex patterns:**
Combine the per-language function patterns into a single combined regex with alternation for faster matching (Go's regex engine optimizes alternation via Aho-Corasick internally).

**Use compiled pattern groups:**
Pre-group patterns by regex type for each language so the same pattern is not compiled multiple times.

---

## Bottleneck #11: Language Registry Detection Calls

### Location: `internal/analyzer/regex/analyzer.go:98-119`

### Root cause
`detectCached` caches by file, but `Supports` and `AnalyzeFile` both call `detectCached`, and the registry's `FindAnalyzer` calls `Supports` which ALSO calls `detectCached`. This means a single file through `AnalyzeSingleFile` triggers detection at least 3 times.

### Fix strategy

**Use extension-to-analyzer lookup table:**
Build a `map[string]Analyzer` at initialization (extension → analyzer). The registry can look up which analyzer handles a file in O(1) without iteration or detection:

```go
type Registry struct {
    analyzers    []Analyzer
    extIndex     map[string]Analyzer // precomputed
}

func (r *Registry) FindAnalyzer(file string) Analyzer {
    return r.extIndex[filepath.Ext(file)]
}
```

---

## Bottleneck #12: Goroutine Overhead for Small Batches

### Location: `internal/analyzer/analyzer.go:15-19`

### Root cause
The 100-file threshold for parallelization is too high. On modern SSDs, parallel I/O is beneficial well below 100 files. But for very large projects (1000+ files), the goroutine-per-analyzer approach creates unnecessary contention. The current architecture creates:
- 1 goroutine per analyzer (typically 2: Go AST + generic regex)
- 8 goroutines for ignore marker extraction
- 2 goroutines for channel management

### Fix strategy

**Use a dynamic worker pool with adaptive parallelism:**
```go
workers := runtime.NumCPU()
if len(files) < workers {
    workers = len(files)
}
```

This scales parallelism to the actual available CPU cores instead of fixed numbers. For very large projects, this prevents goroutine thrashing while maximizing throughput.

**Remove the goroutine-per-analyzer pattern:**
Instead of parallelizing by analyzer (which only has 2 entries), parallelize by file batch in a single worker pool.

---

## Summary: Prioritized Fix List

| Priority | Fix | Expected Speedup | Complexity | Risk |
|----------|-----|-----------------|------------|------|
| P0 | Per-symbol git → per-file git batching | 10-20x on cold analysis | Medium | Low |
| P0 | Rolling hash instead of JSON+SHA256 | 2x on warm analysis | Low | Low |
| P1 | Extension-only pattern fast-path return | 1.5x on discovery | Low | Low |
| P1 | gzip BestSpeed instead of DefaultCompression | 3-5x on save/load | Low | Low |
| P1 | Precompute relative paths | 1.1x on caching | Low | Low |
| P1 | Compute only needed percentiles | 1.2x on scoring | Low | Low |
| P2 | Combined line extraction pass | 2x on analysis | Medium | Medium |
| P2 | Extension-to-analyzer lookup table | 1.3x on analysis | Low | Low |
| P2 | Preallocated slices with known capacity | 1.1x on memory | Low | Low |
| P3 | Adaptive worker pool | 1.2x on analysis | Medium | Low |
| P3 | Buffer pooling with sync.Pool | 1.1x on memory | Medium | Low |
| P3 | Avoid strings.Split in ignore markers | 1.1x on analysis | Medium | Medium |

---

## Appendix A: How to Reproduce the Benchmarks

```bash
# Extract-only benchmark
go test ./tests/ -bench=BenchmarkExtractionOnly -benchmem -count=3

# Full analysis benchmark (requires git repo)
go test ./tests/ -bench=BenchmarkFullAnalysis -benchmem -count=3

# Git bottleneck comparison
go test ./tests/ -bench=BenchmarkGitCommandsComparison -benchmem -count=3

# Microbenchmarks for each git operation
go test ./tests/ -bench=BenchmarkMicroGitOperations -benchmem -count=3 -benchtime=5s

# Scorer benchmarks
go test ./tests/ -bench=BenchmarkScorer -benchmem -count=3

# All benchmarks
go test ./tests/ -bench=Benchmark -benchmem -count=1 -timeout 120s
```

## Appendix B: Expected Performance After All Fixes

| Scenario | Current | After P0 Fixes | After All Fixes |
|----------|---------|----------------|-----------------|
| Cold analysis (245 symbols) | ~1.5s | ~300ms | ~150ms |
| Cold analysis (10k symbols) | ~60s | ~5s | ~2s |
| Warm analysis (no changes) | ~35ms | ~35ms | ~10ms |
| Warm analysis (1 file changed) | ~100ms | ~80ms | ~30ms |
| Memory (10k symbols) | ~1.5GB | ~500MB | ~200MB |
