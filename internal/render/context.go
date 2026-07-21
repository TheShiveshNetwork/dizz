package render

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

type ContextInfo struct {
	ProjectName string
	Branch      string
	Commit      string
	HasGit      bool
}

type ContextRenderer struct{}

func NewContextRenderer() *ContextRenderer {
	return &ContextRenderer{}
}

func (r *ContextRenderer) Render(
	ps *state.ProjectState,
	is *state.IntentState,
	info ContextInfo,
	snapshotHashes []string,
) (string, error) {
	var buf bytes.Buffer

	r.writeProjectInfo(&buf, info)

	if ps.GitCommit != nil {
		r.writeCommitInfo(&buf, ps.GitCommit)
	}

	r.writeIntents(&buf, is)
	r.writePlannedSymbols(&buf, ps)
	r.writeUnstableSymbols(&buf, ps)
	r.writeUnusedSymbols(&buf, ps)
	r.writeTodos(&buf, ps)
	r.writeSnapshots(&buf, snapshotHashes)

	return buf.String(), nil
}

func (r *ContextRenderer) writeProjectInfo(buf *bytes.Buffer, info ContextInfo) {
	gitStatus := "no git"
	if info.HasGit {
		gitStatus = info.Branch + ":" + truncate(info.Commit, 7)
	}
	fmt.Fprintf(buf, "Project: %s | git: %s\n\n", info.ProjectName, gitStatus)
}

func (r *ContextRenderer) writeCommitInfo(buf *bytes.Buffer, c *integrations.Commit) {
	fmt.Fprintf(buf, "# last commit\nhash|msg\ntruncate:%s|%s\n\n", truncate(c.Hash, 7), escapeMsg(c.Message))
}

func (r *ContextRenderer) writeIntents(buf *bytes.Buffer, is *state.IntentState) {
	active := is.GetActiveIntents()
	if len(active) == 0 {
		return
	}

	w := ton.NewWriter(buf)
	fmt.Fprintln(buf, "# intents")
	w.WriteHeader("id", "type", "sev", "status", "msg")
	for _, intent := range active {
		w.WriteRecord(intent.ID, string(intent.Type), fmt.Sprintf("%d", intent.Severity), string(intent.Status), intent.Message)
	}
	fmt.Fprintln(buf)
}

func (r *ContextRenderer) writePlannedSymbols(buf *bytes.Buffer, ps *state.ProjectState) {
	planned := ps.GetSymbolsByState(state.Planned)
	if len(planned) == 0 {
		return
	}

	fmt.Fprintln(buf, "# symbols:planned")
	w := ton.NewWriter(buf)
	w.WriteHeader("name", "file", "line", "confidence")
	for _, sym := range planned {
		w.WriteRecord(sym.Name, sym.File, fmt.Sprintf("%d", sym.Line), fmt.Sprintf("%.2f", sym.Confidence))
	}
	fmt.Fprintln(buf)
}

func (r *ContextRenderer) writeUnstableSymbols(buf *bytes.Buffer, ps *state.ProjectState) {
	unstable := ps.GetSymbolsByState(state.Unstable)
	if len(unstable) == 0 {
		return
	}

	fmt.Fprintln(buf, "# symbols:unstable")
	w := ton.NewWriter(buf)
	w.WriteHeader("name", "file", "line", "churn", "instability")
	sort.Slice(unstable, func(i, j int) bool {
		return unstable[i].InstabilityScore > unstable[j].InstabilityScore
	})
	for _, sym := range unstable {
		w.WriteRecord(sym.Name, sym.File, fmt.Sprintf("%d", sym.Line),
			fmt.Sprintf("%d", sym.ChurnCount), r.InstabilityLabel(sym.InstabilityScore))
	}
	fmt.Fprintln(buf)
}

func (r *ContextRenderer) writeUnusedSymbols(buf *bytes.Buffer, ps *state.ProjectState) {
	unused := ps.GetSymbolsByState(state.Unused)
	abandoned := ps.GetSymbolsByState(state.Abandoned)
	all := append(unused, abandoned...)
	if len(all) == 0 {
		return
	}

	fmt.Fprintln(buf, "# symbols:unused")
	w := ton.NewWriter(buf)
	w.WriteHeader("name", "file", "line", "state", "confidence")
	for _, sym := range all {
		w.WriteRecord(sym.Name, sym.File, fmt.Sprintf("%d", sym.Line),
			string(sym.State), fmt.Sprintf("%.2f", sym.Confidence))
	}
	fmt.Fprintln(buf)
}

func (r *ContextRenderer) writeTodos(buf *bytes.Buffer, ps *state.ProjectState) {
	activeTodos := ps.GetActiveTodos()
	if len(activeTodos) == 0 {
		return
	}

	fmt.Fprintln(buf, "# todos")
	w := ton.NewWriter(buf)
	w.WriteHeader("file", "line", "type", "text")
	for _, todo := range activeTodos {
		w.WriteRecord(todo.File, fmt.Sprintf("%d", todo.Line), todo.Type, todo.Text)
	}
	fmt.Fprintln(buf)
}

func (r *ContextRenderer) writeSnapshots(buf *bytes.Buffer, hashes []string) {
	if len(hashes) == 0 {
		return
	}

	fmt.Fprintln(buf, "# snapshots")
	w := ton.NewWriter(buf)
	w.WriteHeader("hash")
	for _, h := range hashes {
		w.WriteRecord(truncate(h, 8))
	}
	fmt.Fprintln(buf)
}

func (r *ContextRenderer) InstabilityLabel(score float64) string {
	if score < 0 {
		return "?"
	}
	return fmt.Sprintf("%.2f", score)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func escapeMsg(msg string) string {
	msg = strings.ReplaceAll(msg, "|", "/")
	msg = strings.ReplaceAll(msg, "\n", " ")
	return msg
}
