package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

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

// LoadProjectState loads the project state from disk
func (s *StateStore) LoadProjectState() (*state.ProjectState, error) {
	statePath := config.StateFilePath(s.basePath)

	// If file doesn't exist, return error
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("state file does not exist")
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var projectState state.ProjectState
	if err := json.Unmarshal(data, &projectState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project state: %w", err)
	}

	return &projectState, nil
}

// SaveProjectState saves the project state to disk
func (s *StateStore) SaveProjectState(projectState *state.ProjectState) error {
	statePath := filepath.Join(s.basePath, "state.json")

	// Ensure directory exists
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(projectState, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
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

