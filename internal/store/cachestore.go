package store

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

const cacheVersion = 1

// CacheFileMeta tracks the state of a cached file analysis.
type CacheFileMeta struct {
	ContentHash string    `json:"content_hash"`
	MTime       time.Time `json:"mtime"`
}

// CacheManifest is the on-disk index of cached analyses.
type CacheManifest struct {
	Version int                      `json:"version"`
	Files   map[string]CacheFileMeta `json:"files"`
}

// SignalCache provides a per-file signal cache keyed by content hash and mtime.
// Cache entries are invalidated when the file content or modification time changes.
type SignalCache struct {
	projectRoot  string
	manifestPath string
	signalsDir   string
	manifest     CacheManifest
	mu           sync.RWMutex
	changed      bool
}

func NewSignalCache(projectRoot string, cacheDir string) *SignalCache {
	return &SignalCache{
		projectRoot:  projectRoot,
		manifestPath: filepath.Join(cacheDir, "manifest.json"),
		signalsDir:   filepath.Join(cacheDir, "signals"),
		manifest: CacheManifest{
			Version: cacheVersion,
			Files:   make(map[string]CacheFileMeta),
		},
	}
}

func (sc *SignalCache) LoadManifest() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	data, err := os.ReadFile(sc.manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read cache manifest: %w", err)
	}

	var manifest CacheManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("unmarshal cache manifest: %w", err)
	}

	if manifest.Version != cacheVersion {
		return nil
	}

	if manifest.Files == nil {
		manifest.Files = make(map[string]CacheFileMeta)
	}

	sc.manifest = manifest
	return nil
}

func (sc *SignalCache) SaveManifest() error {
	sc.mu.RLock()
	if !sc.changed {
		sc.mu.RUnlock()
		return nil
	}
	sc.mu.RUnlock()

	sc.mu.Lock()
	if err := os.MkdirAll(filepath.Dir(sc.manifestPath), 0755); err != nil {
		sc.mu.Unlock()
		return fmt.Errorf("create cache dir: %w", err)
	}

	data, err := json.Marshal(sc.manifest)
	if err != nil {
		sc.mu.Unlock()
		return fmt.Errorf("marshal manifest: %w", err)
	}

	tmpPath := sc.manifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		sc.mu.Unlock()
		return fmt.Errorf("write manifest: %w", err)
	}
	if err := os.Rename(tmpPath, sc.manifestPath); err != nil {
		sc.mu.Unlock()
		return fmt.Errorf("rename manifest: %w", err)
	}
	sc.changed = false
	sc.mu.Unlock()
	return nil
}

// GetByMTime checks cache using only mtime (no content read needed).
func (sc *SignalCache) GetByMTime(filePath string, mtime time.Time) ([]signals.Signal, bool) {
	relPath, err := filepath.Rel(sc.projectRoot, filePath)
	if err != nil {
		return nil, false
	}
	return sc.GetByMTimeRel(relPath, mtime)
}

// GetByMTimeRel is like GetByMTime but takes a precomputed relative path.
func (sc *SignalCache) GetByMTimeRel(relPath string, mtime time.Time) ([]signals.Signal, bool) {
	sc.mu.RLock()
	meta, ok := sc.manifest.Files[relPath]
	sc.mu.RUnlock()

	if !ok || !meta.MTime.Equal(mtime) {
		return nil, false
	}
	return sc.loadSignals(relPath)
}

// Get checks cache using content hash + mtime (requires reading file content first).
func (sc *SignalCache) Get(filePath string, contentHash string, mtime time.Time) ([]signals.Signal, bool) {
	relPath, err := filepath.Rel(sc.projectRoot, filePath)
	if err != nil {
		return nil, false
	}
	return sc.GetRel(relPath, contentHash, mtime)
}

// GetRel is like Get but takes a precomputed relative path.
func (sc *SignalCache) GetRel(relPath string, contentHash string, mtime time.Time) ([]signals.Signal, bool) {
	sc.mu.RLock()
	meta, ok := sc.manifest.Files[relPath]
	sc.mu.RUnlock()

	if !ok {
		return nil, false
	}
	if meta.ContentHash != contentHash || !meta.MTime.Equal(mtime) {
		return nil, false
	}
	return sc.loadSignals(relPath)
}

// MTimeInCache returns true if the file is cached with the given mtime (fast check).
func (sc *SignalCache) MTimeInCache(filePath string, mtime time.Time) bool {
	relPath, err := filepath.Rel(sc.projectRoot, filePath)
	if err != nil {
		return false
	}
	return sc.MTimeInCacheRel(relPath, mtime)
}

// MTimeInCacheRel is like MTimeInCache but takes a precomputed relative path.
func (sc *SignalCache) MTimeInCacheRel(relPath string, mtime time.Time) bool {
	sc.mu.RLock()
	meta, ok := sc.manifest.Files[relPath]
	sc.mu.RUnlock()
	return ok && meta.MTime.Equal(mtime)
}

func (sc *SignalCache) Set(filePath string, contentHash string, mtime time.Time, sigs []signals.Signal) error {
	relPath, err := filepath.Rel(sc.projectRoot, filePath)
	if err != nil {
		return err
	}
	return sc.SetRel(relPath, contentHash, mtime, sigs)
}

// SetRel is like Set but takes a precomputed relative path.
func (sc *SignalCache) SetRel(relPath string, contentHash string, mtime time.Time, sigs []signals.Signal) error {
	if err := sc.storeSignals(relPath, sigs); err != nil {
		return err
	}

	sc.mu.Lock()
	sc.manifest.Files[relPath] = CacheFileMeta{ContentHash: contentHash, MTime: mtime}
	sc.changed = true
	sc.mu.Unlock()
	return nil
}

func (sc *SignalCache) Evict(filePath string) {
	relPath, err := filepath.Rel(sc.projectRoot, filePath)
	if err != nil {
		return
	}
	sc.EvictRel(relPath)
}

// EvictRel is like Evict but takes a precomputed relative path.
func (sc *SignalCache) EvictRel(relPath string) {
	sc.mu.Lock()
	delete(sc.manifest.Files, relPath)
	sc.changed = true
	sc.mu.Unlock()

	cachePath := sc.signalFilePath(relPath)
	os.Remove(cachePath)
}

func (sc *SignalCache) EvictStale(discoveredFiles map[string]struct{}) {
	sc.mu.Lock()
	for relPath := range sc.manifest.Files {
		if _, exists := discoveredFiles[relPath]; !exists {
			delete(sc.manifest.Files, relPath)
			sc.changed = true
			cachePath := sc.signalFilePath(relPath)
			os.Remove(cachePath)
		}
	}
	sc.mu.Unlock()
}

func (sc *SignalCache) signalFilePath(relPath string) string {
	h := fnv.New64a()
	h.Write([]byte(relPath))
	return filepath.Join(sc.signalsDir, fmt.Sprintf("%016x", h.Sum64()))
}

func (sc *SignalCache) loadSignals(relPath string) ([]signals.Signal, bool) {
	cachePath := sc.signalFilePath(relPath)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	// Auto-detect gzip
	var raw []byte
	if len(data) >= 2 && data[0] == gzipMagic[0] && data[1] == gzipMagic[1] {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, false
		}
		defer reader.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(reader); err != nil {
			return nil, false
		}
		raw = buf.Bytes()
	} else {
		raw = data
	}

	var sigs []signals.Signal
	if err := json.Unmarshal(raw, &sigs); err != nil {
		return nil, false
	}
	return sigs, true
}

func (sc *SignalCache) storeSignals(relPath string, sigs []signals.Signal) error {
	if err := os.MkdirAll(sc.signalsDir, 0755); err != nil {
		return fmt.Errorf("create signals cache dir: %w", err)
	}

	jsonData, err := json.Marshal(sigs)
	if err != nil {
		return fmt.Errorf("marshal cached signals: %w", err)
	}

	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("create gzip: %w", err)
	}
	if _, err := gzWriter.Write(jsonData); err != nil {
		return fmt.Errorf("compress cached signals: %w", err)
	}
	gzWriter.Close()

	cachePath := sc.signalFilePath(relPath)
	tmpPath := cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write signal cache: %w", err)
	}
	return os.Rename(tmpPath, cachePath)
}
