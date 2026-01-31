package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/spf13/cobra"
)

var (
	autoSnapshot bool
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Save current state snapshot",
	Long: `Creates an immutable snapshot of current project state.

Snapshots are content-addressed (like git objects) and stored in .dizz/objects/
If in a git repo, creates a ref linking the commit to the snapshot.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSnapshot()
	},
}

func init() {
	snapshotCmd.Flags().BoolVar(&autoSnapshot, "auto", false, "Automatic snapshot (called by git hook)")
}

func runSnapshot() {
	cwd, _ := os.Getwd()
	trackDir := config.TrackDirPath(cwd)

	// Load current state
	var projectState state.ProjectState
	statePath := config.StateFilePath(trackDir)
	if err := store.Load(statePath, &projectState); err != nil {
		if !autoSnapshot {
			fmt.Fprintln(os.Stderr, ui.Error("✗")+" No state found. Run 'dizz whereami' first.")
		}
		os.Exit(1)
	}

	// Serialize state to JSON
	stateJSON, err := json.MarshalIndent(projectState, "", "  ")
	if err != nil {
		if !autoSnapshot {
			fmt.Fprintf(os.Stderr, ui.Error("Error serializing state: %v\n"), err)
		}
		os.Exit(1)
	}

	// Calculate hash (content-addressed)
	hash := sha256.Sum256(stateJSON)
	hashStr := hex.EncodeToString(hash[:])
	shortHash := hashStr[:6]

	// Create object path: objects/ab/cdef...
	objectsDir := config.ObjectsDirPath(trackDir)
	objectSubdir := filepath.Join(objectsDir, hashStr[:2])
	objectPath := filepath.Join(objectSubdir, hashStr[2:]+".json")

	// Create subdirectory if needed
	if err := os.MkdirAll(objectSubdir, 0755); err != nil {
		if !autoSnapshot {
			fmt.Fprintf(os.Stderr, ui.Error("Error creating object directory: %v\n"), err)
		}
		os.Exit(1)
	}

	// Save snapshot (if not already exists)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		if err := os.WriteFile(objectPath, stateJSON, 0644); err != nil {
			if !autoSnapshot {
				fmt.Fprintf(os.Stderr, ui.Error("Error saving snapshot: %v\n"), err)
			}
			os.Exit(1)
		}
	}

	// Create git ref if in git repo
	if integrations.IsRepo() {
		if commit, err := integrations.GetCurrentCommit(); err == nil {
			refsDir := config.RefsDirPath(trackDir)
			gitRefDir := filepath.Join(refsDir, "git")
			os.MkdirAll(gitRefDir, 0755)

			refPath := filepath.Join(gitRefDir, commit)
			os.WriteFile(refPath, []byte(hashStr), 0644)

			if !autoSnapshot {
				fmt.Printf(ui.Success("✓")+" Snapshot saved: %s\n", ui.Highlight(shortHash))
				fmt.Printf("  %s %s\n", ui.Muted("Git commit:"), ui.Muted(commit[:7]))
				fmt.Printf("  %s %s\n", ui.Muted("Object:"), ui.Muted(objectPath))
			}
		}
	} else {
		if !autoSnapshot {
			fmt.Printf(ui.Success("✓")+" Snapshot saved: %s\n", ui.Highlight(shortHash))
			fmt.Printf("  %s %s\n", ui.Muted("Object:"), ui.Muted(objectPath))
		}
	}

	if !autoSnapshot {
		fmt.Println()
		fmt.Println(ui.Muted("💡 Snapshots are immutable. Use them to track progress over time."))
	}
}
