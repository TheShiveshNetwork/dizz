package store

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

// gzipMagic is the gzip magic number used to detect gzipped files.
var gzipMagic = []byte{0x1f, 0x8b}

// StateStore handles persistence of project state
type StateStore struct {
	basePath string
}

// NewStateStore creates a new state store
func NewStateStore(basePath string) *StateStore {
	return &StateStore{
		basePath: basePath,
	}
}

// LoadProjectState loads the project state from disk.
// It auto-detects gzip-compressed files (by magic number) and falls back
// to uncompressed JSON for backward compatibility.
func (s *StateStore) LoadProjectState() (*state.ProjectState, error) {
	statePath := config.StateFilePath(s.basePath)

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		// Fallback: try the old uncompressed path for backward compatibility
		oldPath := filepath.Join(s.basePath, "state.json")
		if _, err := os.Stat(oldPath); err == nil {
			statePath = oldPath
		} else {
			return nil, fmt.Errorf("state file does not exist")
		}
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Auto-detect gzip by magic number
	if len(data) >= 2 && data[0] == gzipMagic[0] && data[1] == gzipMagic[1] {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(reader); err != nil {
			return nil, fmt.Errorf("failed to decompress state file: %w", err)
		}
		data = buf.Bytes()
	}

	var projectState state.ProjectState
	if err := json.Unmarshal(data, &projectState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project state: %w", err)
	}

	return &projectState, nil
}

// SaveProjectState saves the project state to disk as gzip-compressed JSON.
func (s *StateStore) SaveProjectState(projectState *state.ProjectState) error {
	statePath := config.StateFilePath(s.basePath)

	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	jsonData, err := json.Marshal(projectState)
	if err != nil {
		return fmt.Errorf("failed to marshal project state: %w", err)
	}

	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := gzWriter.Write(jsonData); err != nil {
		return fmt.Errorf("failed to compress state data: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to finalize gzip: %w", err)
	}

	tmpPath := statePath + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return os.Rename(tmpPath, statePath)
}

// ConfigStore handles persistence of configuration
type ConfigStore struct {
	basePath string
}

// NewConfigStore creates a new config store
func NewConfigStore(basePath string) *ConfigStore {
	return &ConfigStore{
		basePath: basePath,
	}
}

// LoadConfig loads the configuration from disk
func (s *ConfigStore) LoadConfig() (*config.Config, error) {
	configPath := filepath.Join(s.basePath, "config.json")

	// If file doesn't exist, return error
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
