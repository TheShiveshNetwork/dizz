package render

import (
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/utils"
)

type SnapshotSaveData struct {
	ShortHash  string
	GitCommit  string
	ObjectPath string
	HasGit     bool
}

func SnapshotSaved(data *SnapshotSaveData) {
	fmt.Printf(ui.Success("✓")+" Snapshot saved: %s\n", ui.Highlight(data.ShortHash))
	if data.HasGit && data.GitCommit != "" {
		fmt.Printf("  %s %s\n", ui.Muted("Git commit:"), ui.Muted(data.GitCommit[:7]))
	}
	fmt.Printf("  %s %s\n", ui.Muted("Object:"), ui.Muted(utils.RelPath(data.ObjectPath)))
	fmt.Println()
	fmt.Println(ui.Muted("💡 Snapshots are immutable. Use them to track progress over time."))
}

type SnapshotDiffData struct {
	ShortHash    string
	IsCheckpoint bool
	Sequence     int
	Delta        *state.SnapshotDelta
}

func SnapshotDiffResult(data *SnapshotDiffData) {
	if data.IsCheckpoint {
		fmt.Printf(ui.Success("✓")+" Checkpoint: %s\n", ui.Highlight(data.ShortHash))
	} else {
		fmt.Printf(ui.Success("✓")+" Delta: %s\n", ui.Highlight(data.ShortHash))
		fmt.Printf("  %s %s\n", ui.Muted("Changes:"), ui.Muted(DeltaSummary(data.Delta)))
	}
	fmt.Printf("  %s %s\n", ui.Muted("Sequence:"), ui.Muted(fmt.Sprintf("%d", data.Sequence)))
}

func SnapshotNoChanges() {
	fmt.Println(ui.Muted("No changes since last snapshot. Skipping delta."))
}

func SnapshotListEmpty() {
	fmt.Println(ui.Muted("No snapshots found."))
}

type SnapshotListItem struct {
	Hash    string
	ModTime string
	Kind    string
	Size    string
}

func SnapshotList(items []SnapshotListItem) {
	fmt.Println(ui.Header("Snapshots:"))
	fmt.Println()
	for _, item := range items {
		fmt.Printf("  %s  %-8s %-7s %s\n",
			ui.Muted(item.ModTime),
			item.Hash,
			item.Kind,
			ui.Muted(item.Size))
	}
	fmt.Println()
}

type SnapshotCheckoutData struct {
	Hash         string
	TotalSymbols int
	ActiveTodos  int
	GitCommit    string
}

func SnapshotCheckout(data *SnapshotCheckoutData) {
	fmt.Printf("  %s %s\n", ui.Muted("Hash:"), ui.Highlight(data.Hash))
	fmt.Printf("  %s %d\n", ui.Muted("Symbols:"), data.TotalSymbols)
	fmt.Printf("  %s %d\n", ui.Muted("Todos:"), data.ActiveTodos)
	if data.GitCommit != "" {
		fmt.Printf("  %s %s\n", ui.Muted("Git:"), ui.Muted(data.GitCommit))
	}
	fmt.Println()
}

func SnapshotPruneResult(count int, totalCheckpoints int, keepCount int) {
	if totalCheckpoints <= keepCount {
		fmt.Println(ui.Muted(fmt.Sprintf("Only %d checkpoints, nothing to prune.", totalCheckpoints)))
		return
	}
	fmt.Printf(ui.Success("✓")+" Pruned %d old snapshots\n", count)
}

func DeltaSummary(d *state.SnapshotDelta) string {
	parts := []string{}
	if len(d.SymbolsAdded) > 0 {
		parts = append(parts, fmt.Sprintf("%d added", len(d.SymbolsAdded)))
	}
	if len(d.SymbolsRemoved) > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", len(d.SymbolsRemoved)))
	}
	if len(d.SymbolsChanged) > 0 {
		parts = append(parts, fmt.Sprintf("%d changed", len(d.SymbolsChanged)))
	}
	if len(d.TodosAdded) > 0 {
		parts = append(parts, fmt.Sprintf("%d todos added", len(d.TodosAdded)))
	}
	if len(d.TodosRemoved) > 0 {
		parts = append(parts, fmt.Sprintf("%d todos removed", len(d.TodosRemoved)))
	}
	return strings.Join(parts, ", ")
}
