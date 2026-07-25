package cmd

import (
	"fmt"
	"time"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
	"github.com/spf13/cobra"
)

var (
	intentType     string
	intentSeverity int
	intentTags     []string
	intentNote     string
)

var intentCmd = &cobra.Command{
	Use:   "intent",
	Short: "Manage human-authored intents",
	Long: `Track and manage human-authored intents separate from computed state.
Intents capture planned work, known issues, and strategic decisions.`,
}

var intentAddCmd = &cobra.Command{
	Use:   "add \"message\"",
	Short: "Add a new intent",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntentAdd,
}

var intentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List intents",
	RunE:  runIntentList,
}

var intentResolveCmd = &cobra.Command{
	Use:   "resolve <id>",
	Short: "Resolve an intent",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntentResolve,
}

var intentCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close an intent (wontfix, duplicate, no longer relevant)",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntentClose,
}

// @ignore-unused
func init() {
	rootCmd.AddCommand(intentCmd)
	intentCmd.AddCommand(intentAddCmd)
	intentCmd.AddCommand(intentListCmd)
	intentCmd.AddCommand(intentResolveCmd)
	intentCmd.AddCommand(intentCloseCmd)

	// Add flags
	intentAddCmd.Flags().StringVar(&intentType, "type", "todo", "Intent type (todo, fixme, refactor, question, hack, temporary)")
	intentAddCmd.Flags().IntVar(&intentSeverity, "severity", 1, "Severity level (0-3)")
	intentAddCmd.Flags().StringSliceVar(&intentTags, "tags", []string{}, "Tags for the intent")
	intentResolveCmd.Flags().StringVar(&intentNote, "note", "", "Optional note for the resolution")
	intentCloseCmd.Flags().StringVar(&intentNote, "note", "", "Optional note for closing")
}

// @ignore-unused
func runIntentAdd(cmd *cobra.Command, args []string) error {
	message := args[0]

	intentType := state.IntentType(intentType)
	validTypes := map[state.IntentType]bool{
		state.IntentTodo: true,
		state.Fixme:      true,
		state.Refactor:   true,
		state.Question:   true,
		state.Hack:       true,
		state.Temporary:  true,
	}

	if !validTypes[intentType] {
		return fmt.Errorf("invalid intent type: %s", intentType)
	}

	// Validate severity
	if intentSeverity < 0 || intentSeverity > 3 {
		return fmt.Errorf("severity must be between 0 and 3")
	}

	intentState, err := loadIntentState()
	if err != nil {
		return fmt.Errorf("failed to load intent state: %w", err)
	}

	now := time.Now()
	intent := state.Intent{
		ID:         generateIntentID(),
		Type:       intentType,
		Message:    message,
		Scope:      determineScope(), // could be enhanced to accept --scope flag
		CreatedAt:  now,
		UpdatedAt:  now,
		CreatedBy:  getCurrentUser(),
		Severity:   intentSeverity,
		Confidence: 0.8, // Default confidence
		Status:     state.IntentActive,
		Tags:       intentTags,
	}

	intentState.AddIntent(intent)

	// Save intent state
	if err := saveIntentState(intentState); err != nil {
		return fmt.Errorf("failed to save intent state: %w", err)
	}

	fmt.Println("✓ Added new intent")
	return nil
}

// @ignore-unused
func runIntentList(cmd *cobra.Command, args []string) error {
	intentState, err := loadIntentState()
	if err != nil {
		return fmt.Errorf("failed to load intent state: %w", err)
	}

	intents := intentState.GetActiveIntents()
	if len(intents) == 0 {
		fmt.Println("No active intents found.")
		return nil
	}

	render.RenderIntents(intents)
	return nil
}

// @ignore-unused
func runIntentResolve(cmd *cobra.Command, args []string) error {
	intentID := args[0]

	intentState, err := loadIntentState()
	if err != nil {
		return fmt.Errorf("failed to load intent state: %w", err)
	}

	description := "Resolved via CLI"
	if intentNote != "" {
		description = intentNote
	}

	resolution := state.Resolution{
		Method:      "fixed",
		Description: description,
		ResolvedAt:  time.Now(),
		ResolvedBy:  getCurrentUser(),
	}

	if err := intentState.ResolveIntent(intentID, resolution); err != nil {
		return fmt.Errorf("failed to resolve intent: %w", err)
	}

	if err := saveIntentState(intentState); err != nil {
		return fmt.Errorf("failed to save intent state: %w", err)
	}

	fmt.Printf("✓ Resolved intent %s\n", intentID)
	return nil
}

// @ignore-unused
func runIntentClose(cmd *cobra.Command, args []string) error {
	intentID := args[0]

	intentState, err := loadIntentState()
	if err != nil {
		return fmt.Errorf("failed to load intent state: %w", err)
	}

	description := "Closed via CLI"
	if intentNote != "" {
		description = intentNote
	}

	resolution := state.Resolution{
		Method:      "closed",
		Description: description,
		ResolvedAt:  time.Now(),
		ResolvedBy:  getCurrentUser(),
	}

	if err := intentState.CloseIntent(intentID, resolution); err != nil {
		return fmt.Errorf("failed to close intent: %w", err)
	}

	if err := saveIntentState(intentState); err != nil {
		return fmt.Errorf("failed to save intent state: %w", err)
	}

	fmt.Printf("✓ Closed intent %s\n", intentID)
	return nil
}

func loadIntentState() (*state.IntentState, error) {
	projectRoot, err := commonPkg.FindProjectRoot()
	if err != nil {
		return nil, err
	}
	trackDir := config.TrackDirPath(projectRoot)
	intentStore := store.NewIntentStore(trackDir)
	return intentStore.LoadIntentState()
}

func saveIntentState(intentState *state.IntentState) error {
	projectRoot, err := commonPkg.FindProjectRoot()
	if err != nil {
		return err
	}
	trackDir := config.TrackDirPath(projectRoot)
	intentStore := store.NewIntentStore(trackDir)
	return intentStore.SaveIntentState(intentState)
}

func generateIntentID() string {
	// Simple ID generation - could be enhanced
	return fmt.Sprintf("int_%d", time.Now().Unix())
}

func determineScope() string {
	return "project"
}

func getCurrentUser() string {
	return integrations.DetectUser()
}
