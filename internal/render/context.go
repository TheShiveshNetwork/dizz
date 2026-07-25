package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

type ContextInfo struct {
	ProjectName   string
	Branch        string
	Commit        string
	HasGit        bool
	CommitMessage string
	Dirty         bool
	Description   string
	Instructions  []config.Instruction
	Guardrails    []config.Guardrail
	Commands      map[string]string
	AgentDefaults config.AgentDefaults
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

	r.writeProjectConfig(&buf, info)

	if ps.GitCommit != nil {
		r.writeCommitInfo(&buf, ps.GitCommit)
	}

	r.writeIntents(&buf, is)
	r.writeSymbolSummary(&buf, ps)
	r.writeTodos(&buf, ps)
	r.writeSnapshots(&buf, snapshotHashes)

	return buf.String(), nil
}

func (r *ContextRenderer) writeProjectConfig(buf *bytes.Buffer, info ContextInfo) {
	gitStatus := "no git"
	if info.HasGit {
		gitStatus = info.Branch + ":" + truncate(info.Commit, 7)
	}

	fmt.Fprintln(buf, "# project")
	w := ton.NewWriter(buf)
	w.WriteHeader("field", "value", "description")

	w.WriteRecord("name", info.ProjectName, "Project identifier from .dizz/config.json")
	w.WriteRecord("description", info.Description, "Human-readable project summary from config")
	w.WriteRecord("git", gitStatus, "Current branch and latest commit hash")

	if info.CommitMessage != "" {
		w.WriteRecord("last_commit_msg", escapeMsg(info.CommitMessage), "Message of the latest commit")
	}
	if info.HasGit {
		dirtyStatus := "clean"
		if info.Dirty {
			dirtyStatus = "dirty"
		}
		w.WriteRecord("dirty", dirtyStatus, "Whether working tree has uncommitted changes")
	}

	if len(info.Instructions) > 0 {
		for _, inst := range info.Instructions {
			w.WriteRecord("instruction", inst.Rule+"@"+inst.Scope, "Agent instruction from config (rule@scope)")
		}
	}
	if len(info.Guardrails) > 0 {
		for _, g := range info.Guardrails {
			if g.Reason != "" {
				w.WriteRecord("guardrail", string(g.Action), g.Reason)
			} else {
				w.WriteRecord("guardrail", string(g.Action))
			}
		}
	}
	if len(info.Commands) > 0 {
		for name, cmd := range info.Commands {
			w.WriteRecord("command", name+"="+cmd, "Project command from config (name=command)")
		}
	}
	if info.AgentDefaults != (config.AgentDefaults{}) {
		if info.AgentDefaults.DefaultLens != "" {
			w.WriteRecord("agent_lens", info.AgentDefaults.DefaultLens, "Default analysis lens from config")
		}
		if info.AgentDefaults.MinSeverity > 0 {
			w.WriteRecord("agent_min_severity", fmt.Sprintf("%d", info.AgentDefaults.MinSeverity), "Minimum severity for agent alerts from config")
		}
	}

	fmt.Fprintln(buf)
}

func (r *ContextRenderer) writeCommitInfo(buf *bytes.Buffer, c *integrations.Commit) {
	fmt.Fprintf(buf, "# last commit\nhash|msg\n%s|%s\n\n", truncate(c.Hash, 7), escapeMsg(c.Message))
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

func (r *ContextRenderer) writeSymbolSummary(buf *bytes.Buffer, ps *state.ProjectState) {
	summary := ps.GetSummary()
	if summary.TotalSymbols == 0 {
		return
	}

	fmt.Fprintln(buf, "# symbols")
	w := ton.NewWriter(buf)
	w.WriteHeader("state", "count")
	states := []state.SymbolState{state.Planned, state.Unstable, state.Unused, state.Abandoned}
	for _, st := range states {
		count := summary.ByState[st]
		if count == 0 {
			continue
		}
		w.WriteRecord(string(st), fmt.Sprintf("%d", count))
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
