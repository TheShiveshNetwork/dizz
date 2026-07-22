package benchmarks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/render"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

func TestContextTokenReductionAgainstVerboseSymbols(t *testing.T) {
	if !integrations.IsRepo() {
		t.Skip("Not in a git repository")
	}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origWD)

	projectRoot, err := commonPkg.FindProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	projectState, err := commonPkg.EnsureCurrentStateWithAnalysis(nil)
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}
	if len(projectState.Symbols) == 0 {
		t.Skip("analysis returned no symbols")
	}

	contextInfo := render.ContextInfo{
		ProjectName: filepath.Base(projectRoot),
		HasGit:      integrations.IsRepo(),
		ConfigRoot:  ".",
	}
	configStore := store.NewConfigStore(config.TrackDirPath(projectRoot))
	if cfg, err := configStore.LoadConfig(); err == nil {
		contextInfo.ProjectName = cfg.ProjectName
		contextInfo.ConfigRoot = cfg.RootPath
		contextInfo.ConfigIncludeCount = len(cfg.Include)
		contextInfo.ConfigExcludeCount = len(cfg.Exclude)
	}

	compactOutput, err := render.NewContextRenderer().Render(projectState, state.NewIntentState(), contextInfo, nil)
	if err != nil {
		t.Fatalf("failed to render compact context: %v", err)
	}

	verboseOutput := compactOutput + legacyVerboseSymbolDetails(projectState)

	compactTokens := estimateTokenCount(compactOutput)
	verboseTokens := estimateTokenCount(verboseOutput)
	if compactTokens >= verboseTokens {
		t.Fatalf("expected compact context to use fewer tokens, compact=%d verbose=%d", compactTokens, verboseTokens)
	}

	t.Logf("context token comparison on dizz repo: compact=%d verbose=%d reduction=%.2f%%",
		compactTokens,
		verboseTokens,
		(1-float64(compactTokens)/float64(verboseTokens))*100)
}

func legacyVerboseSymbolDetails(ps *state.ProjectState) string {
	var b strings.Builder

	planned := ps.GetSymbolsByState(state.Planned)
	if len(planned) > 0 {
		b.WriteString("# symbols:planned\n")
		w := ton.NewWriter(&b)
		w.WriteHeader("name", "file", "line", "confidence")
		for _, sym := range planned {
			w.WriteRecord(sym.Name, sym.File, fmt.Sprintf("%d", sym.Line), fmt.Sprintf("%.2f", sym.Confidence))
		}
		b.WriteString("\n")
	}

	unstable := ps.GetSymbolsByState(state.Unstable)
	if len(unstable) > 0 {
		b.WriteString("# symbols:unstable\n")
		w := ton.NewWriter(&b)
		w.WriteHeader("name", "file", "line", "churn", "instability")
		for _, sym := range unstable {
			w.WriteRecord(sym.Name, sym.File, fmt.Sprintf("%d", sym.Line), fmt.Sprintf("%d", sym.ChurnCount), fmt.Sprintf("%.2f", sym.InstabilityScore))
		}
		b.WriteString("\n")
	}

	unused := ps.GetSymbolsByState(state.Unused)
	abandoned := ps.GetSymbolsByState(state.Abandoned)
	if len(unused)+len(abandoned) > 0 {
		b.WriteString("# symbols:unused\n")
		w := ton.NewWriter(&b)
		w.WriteHeader("name", "file", "line", "state", "confidence")
		for _, sym := range append(unused, abandoned...) {
			w.WriteRecord(sym.Name, sym.File, fmt.Sprintf("%d", sym.Line), string(sym.State), fmt.Sprintf("%.2f", sym.Confidence))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func estimateTokenCount(text string) int {
	return len(strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || r == '|'
	}))
}
