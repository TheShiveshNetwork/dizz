package render

import (
	"fmt"

	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
	"time"
)

type ListSnapshotItem struct {
	Hash      string
	Timestamp time.Time
	GitCommit string
	Summary   state.Summary
}

func ListEmpty() {
	fmt.Println(ui.Warning("⚠ No snapshots found"))
	fmt.Println(ui.Muted("  Run 'dizz snapshot' to create your first snapshot"))
}

func ListSnapshots(snapshots []ListSnapshotItem) {
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

		if snap.GitCommit != "" {
			shortCommit := snap.GitCommit
			if len(shortCommit) > 7 {
				shortCommit = shortCommit[:7]
			}
			fmt.Printf("%s", ui.Muted("("+shortCommit+")"))
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
