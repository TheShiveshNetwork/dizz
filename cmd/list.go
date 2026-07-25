package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/config"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
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

// @ignore-unused
func init() {
	rootCmd.AddCommand(listCmd)
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
		render.ListEmpty()
		return
	}

	snapshots := []SnapshotInfo{}

	filepath.WalkDir(objectsDir, func(path string, info fs.DirEntry, err error) error {
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

			dir := filepath.Base(filepath.Dir(path))
			file := filepath.Base(path)
			hash := dir + file[:len(file)-5]

			snapshots = append(snapshots, SnapshotInfo{
				Hash:      hash,
				Timestamp: ps.UpdatedAt,
				GitCommit: func() string {
					if ps.GitCommit != nil {
						return ps.GitCommit.Hash
					}
					return ""
				}(),
				Summary: ps.GetSummary(),
			})
		}

		return nil
	})

	if len(snapshots) == 0 {
		render.ListEmpty()
		return
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	items := make([]render.ListSnapshotItem, 0, len(snapshots))
	for _, snap := range snapshots {
		items = append(items, render.ListSnapshotItem{
			Hash:      snap.Hash,
			Timestamp: snap.Timestamp,
			GitCommit: snap.GitCommit,
			Summary:   snap.Summary,
		})
	}

	render.ListSnapshots(items)
}
