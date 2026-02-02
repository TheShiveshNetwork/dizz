package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/state"
)

// IntentStore handles persistence of intent state
type IntentStore struct {
	basePath string
}

// NewIntentStore creates a new intent store
func NewIntentStore(basePath string) *IntentStore {
	return &IntentStore{
		basePath: basePath,
	}
}

// LoadIntentState loads the intent state from disk
func (s *IntentStore) LoadIntentState() (*state.IntentState, error) {
	intentPath := config.IntentFilePath(s.basePath)

	// If file doesn't exist, return new state
	if _, err := os.Stat(intentPath); os.IsNotExist(err) {
		return state.NewIntentState(), nil
	}

	data, err := os.ReadFile(intentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read intent file: %w", err)
	}

	var intentState state.IntentState
	if err := json.Unmarshal(data, &intentState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal intent state: %w", err)
	}

	return &intentState, nil
}

// SaveIntentState saves the intent state to disk
func (s *IntentStore) SaveIntentState(intentState *state.IntentState) error {
	intentPath := filepath.Join(s.basePath, "intent.json")

	// Ensure directory exists
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create intent directory: %w", err)
	}

	data, err := json.MarshalIndent(intentState, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal intent state: %w", err)
	}

	if err := os.WriteFile(intentPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write intent file: %w", err)
	}

	return nil
}
