# Dizz Performance Optimization - Next Steps

## Overview
The current `dizz` tool takes over 2 minutes for the first run on larger codebases, which is unacceptable for user experience. This document outlines a comprehensive optimization plan to reduce first-run time to under 60 seconds and subsequent runs to under 5 seconds.

## Priority 1: File Discovery Optimization (Days 1-2)

### Current Performance Issues
- Single-threaded `filepath.WalkDir` processes all files regardless of type
- Pattern matching for every file during inclusion test
- No early filtering for source code files

### Optimizations Needed

#### 1.1 Chunked Directory Walking
- Replace single `filepath.WalkDir` with concurrent chunked walking
- Pre-filter directories using walkFunc with early exit
- Process in chunks for better cache locality

#### 1.2 Early File Extension Filtering
- Replace loop-by-loop extension check with bitmap matching
- Fast-path: precompute extension bitmap for source files
- Skip non-source directories proactively

#### 1.3 WalkFunc Optimization
- Inline extension check before filename processing
- Quick prefix matching for directory exclusion
- Early return for non-matching file extensions

**Files Modified:**
- `internal/discover/files.go` - Lines 30-209

### Expected Gains
- **File discovery time**: 3-5x improvement
- **First run speedup**: ~15-30 seconds

---

## Priority 2: Git Integration Optimization (Days 3-4)

### Current Performance Issues
- `BatchGitAnalysis` still makes multiple git commands
- Git cache invalidated too frequently
- Individual churn calculation per symbol

### Optimizations Needed

#### 2.1 Enhanced Git Batching
- Single `git log --name-status --shortstat` for modified files
- Batch `git blame` for function ranges using `--range`
- Consolidated git commands: `git log --format=%H|%an|%aI --numstat`

#### 2.2 Improved Git Cache
- Implement LRU cache with size limits
- Invalidate on config changes, not just HEAD hash
- Add filesystem modification timestamps

#### 2.3 Parallel Git Operations
- Worker pool for batch processing of function ranges
- Parallel git blame commands with range optimization
- Async processing for non-blocking git operations

**Files Modified:**
- `internal/integrations/git.go` - Lines 288-478

### Expected Gains
- **Git operations**: 5-10x improvement
- **First run speedup**: ~30-60 seconds (depending on git history)

---

## Priority 3: Signal Processing Optimization (Days 5-6)

### Current Performance Issues
- Language detection runs twice per file
- Regex patterns recompiled across files
- No early filtering for changed files

### Optimizations Needed

#### 3.1 Language Detection Caching
- Language registry with efficient caching mechanisms
- LRU cache for recent file detections
- Precomputed extension language mappings

#### 3.2 Regex Pattern Optimization
- Batch regex compilation at startup
- Share compiled patterns across analyzers
- Use sync.Pool for temporary regex matches

#### 3.3 Signal Processing Pipeline
- Early return for unchanged files using signal cache
- Streaming processing for large files
- Skip analysis for already-processed files

**Files Modified:**
- `internal/language/registry.go` - Lines 861-986
- `internal/analyzer/regex/analyzer.go` - Lines 39-367
- `internal/analyzer/analyzer.go` - Lines 75-268

### Expected Gains
- **Signal extraction**: 2-4x improvement
- **First run speedup**: ~20-40 seconds

---

## Priority 4: Memory & Cache Optimization (Days 1-2 of Week 2)

### Current Performance Issues
- Multiple file reads and duplicate data structures
- Signal cache without size limits
- Inefficient memory allocation patterns

### Optimizations Needed

#### 4.1 Signal Cache Compression
- More aggressive gzip compression
- Implement LRU with size limits
- File-based cache eviction for large repositories

#### 4.2 Reduce Data Duplication
- Reuse memory pools instead of new allocations
- Batch processing operations to reduce GC pressure
- Streaming signal processing for memory efficiency

#### 4.3 Improved Cache Invalidation
- Multi-key cache invalidation (content hash + mtime)
- Better cache hit rate tracking
- Pre-warming cache with known file patterns

**Files Modified:**
- `internal/store/cachestore.go` - Lines 252-278
- `internal/state/scorer.go` - Lines 321-505

### Expected Gains
- **Memory usage**: Reduced from ~1GB to <500MB
- **GC pressure**: 50-70% reduction
- **First run speedup**: ~10-20 seconds

---

## Priority 5: Concurrency Optimization (Days 3-4 of Week 2)

### Current Performance Issues
- Limited parallelization opportunities
- Sequential processing bottleneck in analysis
- Underutilized CPU cores

### Optimizations Needed

#### 5.1 Worker Pool for Analysis
- Dynamic worker scaling based on CPU cores
- Batch file processing for better cache locality
- Adaptive parallelization for different project sizes

#### 5.2 Parallel Signal Processing
- Concurrent regex pattern matching
- Async processing for I/O-bound operations
- Load-balanced worker distribution

#### 5.3 Progressive Analysis
- Lazy loading of analysis components
- Background processing for secondary analysis
- Priority-based work distribution

**Files Modified:**
- `internal/analyzer/analyzer.go` - Lines 109-268

### Expected Gains
- **CPU utilization**: 60-80% on multi-core systems
- **First run speedup**: ~15-30 seconds

---

## Testing & Validation

### Performance Benchmarks
1. **Baseline Measurement**: Track current performance on target repository
2. **Unit Tests**: Ensure correctness of optimizations
3. **Integration Tests**: Validate end-to-end functionality
4. **Performance Tests**: Automated regression testing
5. **Load Testing**: Validate with large codebases (>5000 files)

### Test Framework
```go
test/testmain.go
├── test/performance/
│   ├── file_discovery_test.go
│   ├── git_integration_test.go
│   ├── signal_processing_test.go
│   ├── memory_usage_test.go
│   └── cpu_performance_test.go
```

### Validation Scripts
```bash
test/performance/run_benchmarks.sh
# Measures: startup_time, memory_usage, cpu_time, signal_count
test/performance/validate_optimization.sh
# Validates: cache_hit_rates, speedup_metrics, correctness
```

---

## Monitoring & Metrics

### Key Performance Indicators
1. **Cache Effectiveness**: Hit/miss ratios for all cache layers
2. **Git Command Counts**: Number of git commands executed
3. **File Processing Metrics**: Files per second, average processing time
4. **Memory Footprint**: Peak usage, GC pressure metrics
5. **CPU Utilization**: Core usage during analysis phase

### Monitoring Implementation
```go
package monitoring

type Metrics struct {
    GitCommandsExecuted    int
    CacheHits             int
    CacheMisses           int
    FilesProcessed        int
    ProcessingTimeSeconds float64
    MemoryPeakBytes      int64
    CPUCoreUtilization    float64
}

// Collect metrics during analysis phases
// Export to Prometheus for alerting
// Generate performance reports
```

---

## Configuration & Tuning

### Performance Profiles
1. **Fast Mode**: Aggressive caching, minimal I/O
2. **Balanced Mode**: Standard caching, good performance
3. **Thorough Mode**: Maximum accuracy, slower performance

### Adaptive Tuning
```go
// config/performance.go
type PerformanceProfile int

const (
    ProfileFast PerformanceProfile = iota
    ProfileBalanced
    ProfileThorough
    ProfileCustom
)

// Auto-tuning based on available resources
// Project size detection
// Dynamic batch sizing
```

---

## Rollout Strategy

### Phased Deployment
1. **Phase 1** (Week 1): Core optimizations (File discovery, Git integration)
2. **Phase 2** (Week 2): Advanced optimizations (Caching, Concurrency)
3. **Phase 3** (Week 3): Testing, validation, and tuning
4. **Phase 4** (Week 4): Production rollout with monitoring

### Backward Compatibility
- Maintain existing CLI interface
- Preserve all existing functionality
- Ensure no breaking changes to output format
- Validate with existing test suite

---

## Timeline & Dependencies

### Week 1
- Day 1-2: File discovery optimization
- Day 3-4: Git integration optimization
- Day 5-6: Signal processing optimization
- Day 7: Integration testing

### Week 2
- Day 1-2: Memory & cache optimization
- Day 3-4: Concurrency optimization
- Day 5-6: Performance tuning and testing
- Day 7: Production readiness

### Week 3-4
- Final validation
- Performance regression testing
- Documentation updates
- User training materials

---

## Expected Impact

### Performance Goals
| Metric | Current Target | Optimized Target | Improvement |
|--------|----------------|------------------|-------------|
| First Run Time | 120+ seconds | <60 seconds | 2-3x speedup |
| Memory Usage | 1GB+ | <500MB | 2x reduction |
| CPU Utilization | 20-30% | 60-80% | 2-3x improvement |
| Subsequent Runs | 30+ seconds | <5 seconds | 6-10x speedup |

### Business Impact
- **Developer Experience**: Instant feedback on project analysis
- **CI/CD Integration**: Faster build times and feedback loops
- **Scalability**: Support for larger monorepos (>10k files)
- **Resource Efficiency**: Lower cloud infrastructure costs

---

## Risk Mitigation

### Technical Risks
1. **Cache Invalidation**: Implement robust invalidation strategy
2. **Memory Leaks**: Comprehensive memory profiling and validation
3. **Performance Regressions**: Automated benchmarking on every commit
4. **Testing Challenges**: Increased test coverage for new code paths

### Mitigation Strategies
- **Incremental Rollout**: Gradually roll out optimizations
- **A/B Testing**: Compare performance before/after changes
- **Canary Releases**: Deploy to subset of users first
- **Rollback Plan**: Quick rollback mechanism for failed deployments

---

## Success Criteria

1. **Performance**: 60-second first run on repositories with 2000+ files
2. **Reliability**: 99.9% uptime after deployment
3. **Compatibility**: All existing functionality preserved
4. **Maintainability**: Code quality and documentation maintained
5. **Scalability**: Support for repositories >50,000 files

---

*This plan is a living document. Performance metrics and optimizations will be continuously refined based on real-world usage and testing results.*