package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/ui/render"
	"github.com/spf13/cobra"
)

var (
	autoSnapshot bool
	diffSnapshot bool
	pruneKeep    int
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Save current state snapshot",
	Long: `Creates an immutable snapshot of current project state.

Snapshots are content-addressed (like git objects) and stored in .dizz/objects/
If in a git repo, creates a ref linking the commit to the snapshot.

With --diff, stores only the delta from the previous snapshot instead of a full copy.
Checkpoints (full snapshots) are created every 10 deltas.`,
	Run: func(cmd *cobra.Command, args []string) {
		if diffSnapshot {
			runSnapshotDiff()
		} else {
			runSnapshot()
		}
	},
}

// @dizz-ignore-unused
func init() {
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.Flags().BoolVar(&autoSnapshot, "auto", false, "Automatic snapshot (called by git hook)")
	snapshotCmd.Flags().BoolVar(&diffSnapshot, "diff", false, "Store delta from previous snapshot")

	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotCheckoutCmd)
	snapshotCmd.AddCommand(snapshotPruneCmd)
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new snapshot (same as 'dizz snapshot')",
	Run: func(cmd *cobra.Command, args []string) {
		runSnapshot()
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all snapshots and deltas",
	Run: func(cmd *cobra.Command, args []string) {
		runSnapshotList()
	},
}

var snapshotCheckoutCmd = &cobra.Command{
	Use:   "checkout <hash>",
	Short: "Reconstruct full state from snapshot hash",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSnapshotCheckout(args[0])
	},
}

var snapshotPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old snapshots, keeping the last N checkpoints",
	Run: func(cmd *cobra.Command, args []string) {
		runSnapshotPrune()
	},
}

func init() {
	snapshotPruneCmd.Flags().IntVar(&pruneKeep, "keep", 10, "Number of checkpoints to keep")
}

func runSnapshot() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		if !autoSnapshot {
			fmt.Fprintln(os.Stderr, ui.Error("✗")+" "+err.Error())
		}
		os.Exit(1)
	}

	stateJSON, err := json.Marshal(projectState)
	if err != nil {
		if !autoSnapshot {
			fmt.Fprintf(os.Stderr, ui.Error("Error serializing state: %v\n"), err)
		}
		os.Exit(1)
	}

	hash := sha256.Sum256(stateJSON)
	hashStr := hex.EncodeToString(hash[:])
	shortHash := hashStr[:6]

	objectsDir := config.ObjectsDirPath(trackDir)
	objectSubdir := filepath.Join(objectsDir, hashStr[:2])
	objectPath := filepath.Join(objectSubdir, hashStr[2:]+".json")

	if err := os.MkdirAll(objectSubdir, 0755); err != nil {
		if !autoSnapshot {
			fmt.Fprintf(os.Stderr, ui.Error("Error creating object directory: %v\n"), err)
		}
		os.Exit(1)
	}

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		if err := os.WriteFile(objectPath, stateJSON, 0644); err != nil {
			if !autoSnapshot {
				fmt.Fprintf(os.Stderr, ui.Error("Error saving snapshot: %v\n"), err)
			}
			os.Exit(1)
		}
	}

	setLatestRef(trackDir, hashStr)

	if !autoSnapshot {
		commit := ""
		hasGit := integrations.IsRepo()
		if hasGit {
			if c, err := integrations.GetCurrentCommit(); err == nil {
				commit = c
				refsDir := config.RefsDirPath(trackDir)
				gitRefDir := filepath.Join(refsDir, "git")
				os.MkdirAll(gitRefDir, 0755)
				refPath := filepath.Join(gitRefDir, commit)
				os.WriteFile(refPath, []byte(hashStr), 0644)
			}
		}
		render.SnapshotSaved(&render.SnapshotSaveData{
			ShortHash:  shortHash,
			GitCommit:  commit,
			ObjectPath: objectPath,
			HasGit:     hasGit,
		})
	}
}

func runSnapshotDiff() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		if !autoSnapshot {
			fmt.Fprintln(os.Stderr, ui.Error("✗")+" "+err.Error())
		}
		os.Exit(1)
	}

	prevHash := getLatestRef(trackDir)
	var prevState *state.ProjectState
	if prevHash != "" {
		prevState = loadSnapshotObject(trackDir, prevHash)
	}

	seq := 0
	isCheckpoint := false
	var delta *state.SnapshotDelta

	if prevState == nil {
		delta = &state.SnapshotDelta{
			IsCheckpoint: true,
			Sequence:     0,
		}
		isCheckpoint = true
	} else {
		prevSeq := getDeltaSequence(trackDir, prevHash)
		seq = prevSeq + 1
		isCheckpoint = seq%state.CheckpointInterval == 0

		if isCheckpoint {
			delta = &state.SnapshotDelta{
				PrevHash:     prevHash,
				IsCheckpoint: true,
				Sequence:     seq,
			}
		} else {
			delta = state.DiffSnapshots(prevState, projectState)
			delta.PrevHash = prevHash
			delta.IsCheckpoint = false
			delta.Sequence = seq

			if delta.IsEmpty() {
				if !autoSnapshot {
					render.SnapshotNoChanges()
				}
				return
			}
		}
	}

	var data []byte
	if isCheckpoint {
		data, err = json.Marshal(projectState)
	} else {
		data, err = json.Marshal(delta)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error serializing: %v\n"), err)
		os.Exit(1)
	}

	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])
	shortHash := hashStr[:6]

	objectsDir := config.ObjectsDirPath(trackDir)
	objectSubdir := filepath.Join(objectsDir, hashStr[:2])
	var objectPath string
	if isCheckpoint {
		objectPath = filepath.Join(objectSubdir, hashStr[2:]+".json")
	} else {
		objectPath = filepath.Join(objectSubdir, hashStr[2:]+".delta")
	}

	if err := os.MkdirAll(objectSubdir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error creating object directory: %v\n"), err)
		os.Exit(1)
	}

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		if err := os.WriteFile(objectPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, ui.Error("Error saving: %v\n"), err)
			os.Exit(1)
		}
	}

	setLatestRef(trackDir, hashStr)

	if integrations.IsRepo() {
		if commit, err := integrations.GetCurrentCommit(); err == nil {
			refsDir := config.RefsDirPath(trackDir)
			gitRefDir := filepath.Join(refsDir, "git")
			os.MkdirAll(gitRefDir, 0755)
			refPath := filepath.Join(gitRefDir, commit)
			os.WriteFile(refPath, []byte(hashStr), 0644)
		}
	}

	if !autoSnapshot {
		render.SnapshotDiffResult(&render.SnapshotDiffData{
			ShortHash:    shortHash,
			IsCheckpoint: isCheckpoint,
			Sequence:     seq,
			Delta:        delta,
		})
	}
}

func runSnapshotList() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	entries := listSnapshotEntries(trackDir)
	if len(entries) == 0 {
		render.SnapshotListEmpty()
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].modTime.Before(entries[j].modTime)
	})

	items := make([]render.SnapshotListItem, 0, len(entries))
	for _, e := range entries {
		kind := "snapshot"
		if strings.HasSuffix(e.path, ".delta") {
			kind = "delta"
		}
		size := fmt.Sprintf("%.1f KB", float64(e.size)/1024.0)
		shortHash := e.hash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}
		items = append(items, render.SnapshotListItem{
			Hash:    shortHash,
			ModTime: e.modTime.Format("Jan 02 15:04"),
			Kind:    kind,
			Size:    size,
		})
	}
	render.SnapshotList(items)
}

func runSnapshotCheckout(hash string) {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	fullHash := resolveHash(trackDir, hash)
	if fullHash == "" {
		fmt.Fprintf(os.Stderr, ui.Error("Snapshot not found: %s\n"), hash)
		os.Exit(1)
	}

	stateJSON, err := reconstructState(trackDir, fullHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error reconstructing state: %v\n"), err)
		os.Exit(1)
	}

	var ps state.ProjectState
	if err := json.Unmarshal(stateJSON, &ps); err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("Error parsing state: %v\n"), err)
		os.Exit(1)
	}

	summary := ps.GetSummary()
	gitCommit := ""
	if ps.GitCommit != nil {
		gitCommit = ps.GitCommit.Hash
	}

	render.SnapshotCheckout(&render.SnapshotCheckoutData{
		Hash:         fullHash[:8],
		TotalSymbols: summary.TotalSymbols,
		ActiveTodos:  summary.ActiveTodos,
		GitCommit:    gitCommit,
	})
}

func runSnapshotPrune() {
	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.Error("✗ %v\n"), err)
		os.Exit(1)
	}

	entries := listSnapshotEntries(trackDir)

	var checkpoints []snapshotEntry
	for _, e := range entries {
		if strings.HasSuffix(e.path, ".json") {
			checkpoints = append(checkpoints, e)
		}
	}

	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].modTime.After(checkpoints[j].modTime)
	})

	if len(checkpoints) <= pruneKeep {
		render.SnapshotPruneResult(0, len(checkpoints), pruneKeep)
		return
	}

	toRemove := checkpoints[pruneKeep:]
	removedCount := 0

	removeSet := make(map[string]bool)
	for _, cr := range toRemove {
		removeSet[cr.hash] = true
	}

	for _, e := range entries {
		if removeSet[e.hash] {
			os.Remove(e.path)
			removedCount++
			continue
		}
		if strings.HasSuffix(e.path, ".delta") {
			for crHash := range removeSet {
				if strings.HasPrefix(e.hash, crHash) || strings.HasPrefix(crHash, e.hash) {
					continue
				}
				if hasDeltaParent(e.path, crHash) {
					os.Remove(e.path)
					removedCount++
					break
				}
			}
		}
	}

	render.SnapshotPruneResult(removedCount, len(checkpoints), pruneKeep)
}

type snapshotEntry struct {
	path    string
	hash    string
	modTime time.Time
	size    int64
}

func listSnapshotEntries(trackDir string) []snapshotEntry {
	objectsDir := config.ObjectsDirPath(trackDir)
	entries := []snapshotEntry{}

	subdirs, err := os.ReadDir(objectsDir)
	if err != nil {
		return entries
	}

	for _, sub := range subdirs {
		if !sub.IsDir() {
			continue
		}
		subPath := filepath.Join(objectsDir, sub.Name())
		files, err := os.ReadDir(subPath)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".delta") {
				continue
			}
			info, err := f.Info()
			if err != nil {
				continue
			}
			ext := filepath.Ext(name)
			hash := sub.Name() + strings.TrimSuffix(name, ext)
			entries = append(entries, snapshotEntry{
				path:    filepath.Join(subPath, name),
				hash:    hash,
				modTime: info.ModTime(),
				size:    info.Size(),
			})
		}
	}
	return entries
}

func loadSnapshotObject(trackDir, hash string) *state.ProjectState {
	objectsDir := config.ObjectsDirPath(trackDir)
	objectPath := filepath.Join(objectsDir, hash[:2], hash[2:]+".json")
	data, err := os.ReadFile(objectPath)
	if err != nil {
		return nil
	}
	var ps state.ProjectState
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil
	}
	return &ps
}

func loadDeltaObject(trackDir, hash string) *state.SnapshotDelta {
	objectsDir := config.ObjectsDirPath(trackDir)
	objectPath := filepath.Join(objectsDir, hash[:2], hash[2:]+".delta")
	data, err := os.ReadFile(objectPath)
	if err != nil {
		return nil
	}
	var d state.SnapshotDelta
	if err := json.Unmarshal(data, &d); err != nil {
		return nil
	}
	return &d
}

func reconstructState(trackDir, hash string) ([]byte, error) {
	delta := loadDeltaObject(trackDir, hash)
	if delta == nil {
		objectsDir := config.ObjectsDirPath(trackDir)
		objectPath := filepath.Join(objectsDir, hash[:2], hash[2:]+".json")
		return os.ReadFile(objectPath)
	}

	chain := []*state.SnapshotDelta{delta}
	current := delta
	for current.PrevHash != "" {
		prevDelta := loadDeltaObject(trackDir, current.PrevHash)
		if prevDelta != nil {
			chain = append(chain, prevDelta)
			current = prevDelta
		} else {
			break
		}
	}

	last := chain[len(chain)-1]
	checkpointHash := last.PrevHash

	var baseData []byte
	if checkpointHash == "" {
		return nil, fmt.Errorf("no checkpoint found in chain")
	}
	objectsDir := config.ObjectsDirPath(trackDir)
	basePath := filepath.Join(objectsDir, checkpointHash[:2], checkpointHash[2:]+".json")
	var err error
	baseData, err = os.ReadFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("base checkpoint %s not found: %w", checkpointHash, err)
	}

	var baseState state.ProjectState
	if err := json.Unmarshal(baseData, &baseState); err != nil {
		return nil, err
	}

	for i := len(chain) - 1; i >= 0; i-- {
		nextState := state.ApplyDelta(&baseState, chain[i])
		baseState = *nextState
	}

	return json.Marshal(baseState)
}

func getLatestRef(trackDir string) string {
	refsDir := config.RefsDirPath(trackDir)
	latestPath := filepath.Join(refsDir, "snapshots", "latest")
	data, err := os.ReadFile(latestPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func setLatestRef(trackDir, hash string) {
	refsDir := config.RefsDirPath(trackDir)
	latestDir := filepath.Join(refsDir, "snapshots")
	os.MkdirAll(latestDir, 0755)
	os.WriteFile(filepath.Join(latestDir, "latest"), []byte(hash), 0644)
}

func getDeltaSequence(trackDir, hash string) int {
	d := loadDeltaObject(trackDir, hash)
	if d != nil {
		return d.Sequence
	}
	return 0
}

func resolveHash(trackDir, partial string) string {
	entries := listSnapshotEntries(trackDir)
	for _, e := range entries {
		if strings.HasPrefix(e.hash, partial) {
			return e.hash
		}
	}
	return ""
}

func hasDeltaParent(path, parentHash string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var d state.SnapshotDelta
	if err := json.Unmarshal(data, &d); err != nil {
		return false
	}
	return d.PrevHash == parentHash || strings.HasPrefix(d.PrevHash, parentHash)
}
