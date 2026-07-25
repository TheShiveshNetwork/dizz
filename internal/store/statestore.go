package store

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

var gzipMagic = []byte{0x1f, 0x8b}

type StateStore struct {
	basePath string
}

func NewStateStore(basePath string) *StateStore {
	return &StateStore{
		basePath: basePath,
	}
}

func (s *StateStore) LoadProjectState() (*state.ProjectState, error) {
	// Try state.ton first (may be gzipped or uncompressed for backward compat)
	tonPath := config.StateTONFilePath(s.basePath)
	if data, err := os.ReadFile(tonPath); err == nil {
		// Auto-detect gzip by magic number
		if len(data) >= 2 && data[0] == gzipMagic[0] && data[1] == gzipMagic[1] {
			reader, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("failed to create gzip reader for state.ton: %w", err)
			}
			defer reader.Close()

			buf := bufferPool.Get().(*bytes.Buffer)
			buf.Reset()
			if _, err := buf.ReadFrom(reader); err != nil {
				bufferPool.Put(buf)
				return nil, fmt.Errorf("failed to decompress state.ton: %w", err)
			}
			data = buf.Bytes()
			bufferPool.Put(buf)
		}

		ps, err := state.UnmarshalProjectStateTON(data)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal state.ton: %w", err)
		}
		return ps, nil
	}

	// Fall back to state.json.gz
	gzPath := config.StateFilePath(s.basePath)
	if data, err := os.ReadFile(gzPath); err == nil {
		// Auto-detect gzip by magic number
		if len(data) >= 2 && data[0] == gzipMagic[0] && data[1] == gzipMagic[1] {
			reader, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer reader.Close()

			buf := bufferPool.Get().(*bytes.Buffer)
			buf.Reset()
			if _, err := buf.ReadFrom(reader); err != nil {
				bufferPool.Put(buf)
				return nil, fmt.Errorf("failed to decompress state file: %w", err)
			}
			data = buf.Bytes()
			bufferPool.Put(buf)
		}

		var projectState state.ProjectState
		if err := json.Unmarshal(data, &projectState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal project state: %w", err)
		}
		return &projectState, nil
	}

	// Fall back to uncompressed state.json (legacy)
	oldPath := filepath.Join(s.basePath, "state.json")
	if data, err := os.ReadFile(oldPath); err == nil {
		var projectState state.ProjectState
		if err := json.Unmarshal(data, &projectState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal state.json: %w", err)
		}
		return &projectState, nil
	}

	return nil, fmt.Errorf("state file does not exist")
}

func (s *StateStore) SaveProjectState(projectState *state.ProjectState) error {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	tonPath := config.StateTONFilePath(s.basePath)

	data, err := projectState.MarshalTON()
	if err != nil {
		return fmt.Errorf("failed to marshal state.ton: %w", err)
	}

	// Compress with gzip for compact on-disk storage (internal-only file, not LLM-facing)
	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	if _, err := gzWriter.Write(data); err != nil {
		return fmt.Errorf("failed to compress state.ton data: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to finalize gzip: %w", err)
	}

	tmpPath := tonPath + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write state.ton: %w", err)
	}
	return os.Rename(tmpPath, tonPath)
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

// oldConfig is the old configuration structure used for migration.
type oldConfig struct {
	ProjectName string   `json:"project_name"`
	Include     []string `json:"include"`
	Exclude     []string `json:"exclude"`
	Agentic     struct {
		Description  string   `json:"description"`
		Rules        []string `json:"rules"`
		Standards    []string `json:"standards"`
		Instructions []string `json:"instructions"`
	} `json:"agentic"`
}

// defaultConfig returns a minimal configuration with only required fields set.
func defaultConfig(projectName string) *config.Config {
	return &config.Config{
		Version:     config.ConfigVersion,
		ProjectName: projectName,
		Description: "",
		Include:     []string{"**/*"},
		Exclude:     []string{"**/*_test.go", "vendor/**", "node_modules/**", ".git/**", ".dizz/**"},
	}
}

// LoadConfig loads the configuration from disk
func (s *ConfigStore) LoadConfig() (*config.Config, error) {
	configPath := filepath.Join(s.basePath, config.ConfigFile)

	data, err := os.ReadFile(configPath)
	if err != nil {
		// If config file does not exist, return the default config.
		projectRoot, _ := os.Getwd()
		return defaultConfig(projectRoot), nil
	}

	// First, check if it's the new config format by looking for version.
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		// If the config file is not valid JSON, return the default config.
		projectRoot, _ := os.Getwd()
		return defaultConfig(projectRoot), nil
	}

	if _, ok := temp["version"]; ok {
		// Try to unmarshal into the new config struct.
		var cfg config.Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			// If the new config struct fails, return the default config.
			projectRoot, _ := os.Getwd()
			return defaultConfig(projectRoot), nil
		}
		return &cfg, nil
	}

	// Otherwise, try to unmarshal into the old config struct.
	var oldCfg oldConfig
	if err := json.Unmarshal(data, &oldCfg); err != nil {
		// If the old config struct fails, return the default config.
		projectRoot, _ := os.Getwd()
		return defaultConfig(projectRoot), nil
	}

	// Convert the old config to the new config.
	newCfg := defaultConfig(oldCfg.ProjectName)
	newCfg.Version = config.ConfigVersion
	newCfg.Description = oldCfg.Agentic.Description
	newCfg.Include = oldCfg.Include
	newCfg.Exclude = oldCfg.Exclude
	// Note: The old config did not have Commands, Instructions, Guardrails, etc.
	// We leave them as empty/zero values (the struct's zero value).

	return newCfg, nil
}

// SaveConfig saves the configuration to disk
func (s *ConfigStore) SaveConfig(cfg *config.Config) error {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(s.basePath, config.ConfigFile)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return os.Rename(tmpPath, configPath)
}
