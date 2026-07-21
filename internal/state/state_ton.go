package state

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

// MarshalStateTON produces a compact TON representation of the full
// project state suitable for agent consumption. Includes content hash
// for staleness detection.
func MarshalStateTON(ps *ProjectState, is *IntentState, snapshotHashes []string) ([]byte, string) {
	var buf bytes.Buffer

	projectName := "project"
	branch := ""
	commitHash := ""
	if ps.GitCommit != nil {
		commitHash = ps.GitCommit.Hash
		if len(commitHash) > 7 {
			commitHash = commitHash[:7]
		}
	}
	gitStatus := "no git"
	if commitHash != "" {
		gitStatus = branch + ":" + commitHash
	}
	fmt.Fprintf(&buf, "project|%s|git:%s\n", escapeTON(projectName), escapeTON(gitStatus))

	if is != nil {
		active := is.GetActiveIntents()
		if len(active) > 0 {
			fmt.Fprintf(&buf, "\n# intents\n")
			w := ton.NewWriter(&buf)
			w.WriteHeader("id", "type", "sev", "status", "msg")
			for _, intent := range active {
				w.WriteRecord(intent.ID, string(intent.Type), fmt.Sprintf("%d", intent.Severity), string(intent.Status), intent.Message)
			}
		}
	}

	symbolStates := []SymbolState{Unstable, Unused, Abandoned, Planned, Active}
	for _, st := range symbolStates {
		syms := ps.GetSymbolsByState(st)
		if len(syms) == 0 {
			continue
		}
		section := "# symbols:" + string(st)
		fmt.Fprintf(&buf, "\n%s\n", section)
		w2 := ton.NewWriter(&buf)
		w2.WriteHeader("name", "file", "line", "state", "confidence")
		for _, sym := range syms {
			w2.WriteRecord(sym.Name, sym.File, fmt.Sprintf("%d", sym.Line), string(sym.State), fmt.Sprintf("%.2f", sym.Confidence))
		}
	}

	activeTodos := ps.GetActiveTodos()
	if len(activeTodos) > 0 {
		fmt.Fprintf(&buf, "\n# todos\n")
		w3 := ton.NewWriter(&buf)
		w3.WriteHeader("file", "line", "type", "text")
		for _, todo := range activeTodos {
			w3.WriteRecord(todo.File, fmt.Sprintf("%d", todo.Line), todo.Type, todo.Text)
		}
	}

	if len(snapshotHashes) > 0 {
		fmt.Fprintf(&buf, "\n# snapshots\n")
		w4 := ton.NewWriter(&buf)
		w4.WriteHeader("hash")
		for _, h := range snapshotHashes {
			w4.WriteRecord(h)
		}
	}

	h := sha256.Sum256(buf.Bytes())
	contentHash := hex.EncodeToString(h[:8])

	fmt.Fprintf(&buf, "\nhash|%s\n", contentHash)

	return buf.Bytes(), contentHash
}

// VerifyStateTONHash checks if a state.ton blob has a valid content hash.
func VerifyStateTONHash(data []byte) bool {
	s := string(data)
	idx := strings.LastIndex(s, "\nhash|")
	if idx < 0 {
		return false
	}
	content := data[:idx]
	storedHash := strings.TrimSpace(s[idx+6:])
	h := sha256.Sum256(content)
	computedHash := hex.EncodeToString(h[:8])
	return storedHash == computedHash
}

func escapeTON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
