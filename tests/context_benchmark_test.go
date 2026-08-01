package benchmarks

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/integrations"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/render"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

// buildContextInfo mirrors the git/config gathering done by the `dizz context` command.
func buildContextInfo(tb testing.TB, trackDir string) render.ContextInfo {
	tb.Helper()
	info := render.ContextInfo{
		ProjectName: filepath.Base(trackDir),
		HasGit:      integrations.IsRepo(),
	}
	if cfg, err := store.NewConfigStore(config.TrackDirPath(trackDir)).LoadConfig(); err == nil {
		info.ProjectName = cfg.ProjectName
		info.Description = cfg.Description
		info.Instructions = cfg.Instructions
		info.Guardrails = cfg.Guardrails
		info.Commands = cfg.Commands
		info.AgentDefaults = cfg.AgentDefaults
	}
	if info.HasGit {
		if branch, err := integrations.GetCurrentBranch(); err == nil {
			info.Branch = branch
		}
		if commitInfo, err := integrations.GetCurrentCommitWithMessage(); err == nil {
			info.Commit = commitInfo.Hash
			info.CommitMessage = commitInfo.Message
		}
		info.Dirty = integrations.HasUntrackedOrModifiedChanges()
	}
	return info
}

// BenchmarkAgentContext measures the exact pipeline `dizz context` runs end-to-end:
// state analysis + TON rendering into the token-optimized output agents consume.
func BenchmarkAgentContext(b *testing.B) {
	if !integrations.IsRepo() {
		b.Skip("Not in a git repository")
	}

	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		b.Fatalf("FindProjectRoot: %v", err)
	}
	intentStore := store.NewIntentStore(config.TrackDirPath(trackDir))
	renderer := render.NewContextRenderer()

	b.Run("Warm_Cache_EndToEnd", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
			if err != nil {
				b.Fatalf("analysis: %v", err)
			}
			intentState, err := intentStore.LoadIntentState()
			if err != nil {
				b.Fatalf("intents: %v", err)
			}
			if _, err := renderer.Render(projectState, intentState, buildContextInfo(b, trackDir), nil); err != nil {
				b.Fatalf("render: %v", err)
			}
		}
	})
}

// TestAgentContextEfficiency reports the byte efficiency of producing fresh agent
// context vs the raw state files the agent would otherwise have to parse.
func TestAgentContextEfficiency(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	trackDir, err := commonPkg.FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot: %v", err)
	}
	intentStore := store.NewIntentStore(config.TrackDirPath(trackDir))
	renderer := render.NewContextRenderer()

	intentState, err := intentStore.LoadIntentState()
	if err != nil {
		t.Fatalf("LoadIntentState: %v", err)
	}

	var analysisTime time.Duration
	start := time.Now()
	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	analysisTime = time.Since(start)
	if err != nil {
		t.Fatalf("analysis: %v", err)
	}

	start = time.Now()
	output, err := renderer.Render(projectState, intentState, buildContextInfo(t, trackDir), nil)
	renderTime := time.Since(start)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	trackPath := config.TrackDirPath(trackDir)
	var rawBytes int64
	for _, name := range []string{"state.ton.gz", "intent.ton", "context.ton", "config.json"} {
		if fi, err := os.Stat(trackPath + "/" + name); err == nil {
			rawBytes += fi.Size()
		}
	}

	t.Logf("Symbols analyzed: %d", len(projectState.Symbols))
	t.Logf("Fresh context output: %d bytes (%d tokens ~4ch/tok)", len(output), len(output)/4)
	t.Logf("Raw state files replaced: %d bytes (%dx the context)", rawBytes, rawBytes/int64(len(output)))
	t.Logf("Analysis (warm): %v", analysisTime)
	t.Logf("Render only: %v", renderTime)
}
