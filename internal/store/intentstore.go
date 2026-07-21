package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheShiveshNetwork/dizz/internal/state"
)

const IntentTONFile = "intent.ton"
const IntentJSONFile = "intent.json"

// IntentStore handles persistence of intent state.
// Primary format is TON (intent.ton) for token efficiency.
// Falls back to JSON (intent.json) for backward compatibility.
type IntentStore struct {
	basePath string
}

// NewIntentStore creates a new intent store
func NewIntentStore(basePath string) *IntentStore {
	return &IntentStore{
		basePath: basePath,
	}
}

// LoadIntentState loads the intent state from disk.
// Prefers intent.ton, falls back to intent.json.
func (s *IntentStore) LoadIntentState() (*state.IntentState, error) {
	tonPath := filepath.Join(s.basePath, IntentTONFile)
	jsonPath := filepath.Join(s.basePath, IntentJSONFile)

	// Try TON first
	if data, err := os.ReadFile(tonPath); err == nil {
		is, err := state.UnmarshalIntentStateTON(data)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal intent.ton: %w", err)
		}
		return is, nil
	}

	// Fall back to JSON
	if data, err := os.ReadFile(jsonPath); err == nil {
		var intentState state.IntentState
		if err := json.Unmarshal(data, &intentState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal intent.json: %w", err)
		}
		return &intentState, nil
	}

	return state.NewIntentState(), nil
}

// SaveIntentState saves the intent state to disk as TON format.
func (s *IntentStore) SaveIntentState(intentState *state.IntentState) error {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create intent directory: %w", err)
	}

	intentPath := filepath.Join(s.basePath, IntentTONFile)

	data, err := intentState.MarshalTON()
	if err != nil {
		return fmt.Errorf("failed to marshal intent state: %w", err)
	}

	tmpPath := intentPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write intent file: %w", err)
	}
	return os.Rename(tmpPath, intentPath)
}
