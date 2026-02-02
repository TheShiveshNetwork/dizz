package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
	
	"github.com/spf13/cobra"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved snapshots",
	Long: `Shows all saved snapshots with their timestamps and summary.
Use this to see project history over time.`,
	Run: func(cmd *cobra.Command, args []string) {
		runList()
	},
}

type SnapshotInfo struct {
	Hash      string
	Timestamp time.Time
	GitCommit string
	Summary   state.Summary
}

func runList() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	objectsDir := config.ObjectsDirPath(trackDir)

	if _, err := os.Stat(objectsDir); os.IsNotExist(err) {
		fmt.Println(ui.Warning("⚠ No snapshots found"))
		fmt.Println(ui.Muted("  Run 'dizz snapshot' to create your first snapshot"))
		return
	}

	// Scan for snapshots
	snapshots := []SnapshotInfo{}

	// Walk objects directory
	filepath.Walk(objectsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".json" {
			var ps state.ProjectState
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			if err := json.Unmarshal(data, &ps); err != nil {
				return nil
			}

			// Get hash from path
			dir := filepath.Base(filepath.Dir(path))
			file := filepath.Base(path)
			hash := dir + file[:len(file)-5] // Remove .json

			snapshots = append(snapshots, SnapshotInfo{
				Hash:      hash,
				Timestamp: ps.UpdatedAt,
				GitCommit: ps.GitCommit,
				Summary:   ps.GetSummary(),
			})
		}

		return nil
	})

	if len(snapshots) == 0 {
		fmt.Println(ui.Warning("⚠ No snapshots found"))
		return
	}

	// Sort by timestamp (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	fmt.Println()
	fmt.Println()
	fmt.Println(ui.Header("SNAPSHOT HISTORY"))
	fmt.Println()

	for i, snap := range snapshots {
		if i >= 10 {
			fmt.Printf(ui.Muted("  ... and %d more snapshots\n"), len(snapshots)-10)
			break
		}

		timeSince := time.Since(snap.Timestamp)
		timeDisplay := utils.FormatTime(timeSince)

		shortHash := snap.Hash[:6]
		fmt.Printf("  %s ", ui.Highlight(shortHash))
		fmt.Printf("%s ", ui.Muted(timeDisplay.Text))

		// Git commit
		if snap.GitCommit != "" {
			shortCommit := snap.GitCommit
			if len(shortCommit) > 7 {
				shortCommit = shortCommit[:7]
			}
			fmt.Printf(ui.Muted("(" + shortCommit + ")"))
		}
		fmt.Println()

		active := snap.Summary.ByState[state.Active]
		planned := snap.Summary.ByState[state.Planned]
		unstable := snap.Summary.ByState[state.Unstable]
		unused := snap.Summary.ByState[state.Unused]

		summary := ""
		if active > 0 {
			summary += ui.Success(fmt.Sprintf("Act %d", active))
		}
		if planned > 0 {
			if summary != "" {
				summary += " "
			}
			summary += ui.Warning(fmt.Sprintf("Pln %d", planned))
		}
		if unstable > 0 {
			if summary != "" {
				summary += " "
			}
			summary += ui.Error(fmt.Sprintf("Uns %d", unstable))
		}
		if unused > 0 {
			if summary != "" {
				summary += " "
			}
			summary += ui.Info(fmt.Sprintf("Use %d", unused))
		}

		fmt.Printf("     %s\n", summary)
		fmt.Println()
	}

	fmt.Printf(ui.Muted("  Total: %d snapshots\n"), len(snapshots))
	fmt.Println()
}
