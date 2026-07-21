package state

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/integrations"
	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

func escapeTON(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '|':
			b.WriteString(`\|`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func unescapeTON(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
			case '|':
				b.WriteByte('|')
				i++
			case '\\':
				b.WriteByte('\\')
				i++
			default:
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func (ps *ProjectState) MarshalTON() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "# state\n")
	w := ton.NewWriter(&buf)
	w.WriteHeader("version", "updated_at")
	w.WriteRecord("1.0", ps.UpdatedAt.Format(time.RFC3339))

	if ps.GitCommit != nil {
		fmt.Fprintf(&buf, "\n# git\n")
		w2 := ton.NewWriter(&buf)
		w2.WriteHeader("hash", "message", "time")
		hash := ps.GitCommit.Hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		w2.WriteRecord(hash, escapeTON(ps.GitCommit.Message), ps.GitCommit.Time.Format(time.RFC3339))
	}

	if len(ps.Symbols) > 0 {
		fmt.Fprintf(&buf, "\n# symbols\n")
		w3 := ton.NewWriter(&buf)
		w3.WriteHeader("name", "file", "line", "col", "end_line", "end_col", "type", "lang", "state", "confidence", "defined", "called", "todo", "churn", "instability", "marker", "source", "touched")
		for _, sym := range ps.Symbols {
			lastTouched := ""
			if sym.LastTouched != nil {
				lastTouched = sym.LastTouched.Format(time.RFC3339)
			}
			w3.WriteRecord(
				sym.Name, sym.File,
				strconv.Itoa(sym.Line), strconv.Itoa(sym.Column),
				strconv.Itoa(sym.EndLine), strconv.Itoa(sym.EndColumn),
				sym.Type, sym.Language,
				string(sym.State), strconv.FormatFloat(sym.Confidence, 'f', 2, 64),
				strconv.FormatBool(sym.IsDefined), strconv.FormatBool(sym.IsCalled), strconv.FormatBool(sym.HasTodo),
				strconv.Itoa(sym.ChurnCount), strconv.FormatFloat(sym.InstabilityScore, 'f', 4, 64),
				sym.IntentMarker, sym.SignalSource,
				lastTouched,
			)
		}
	}

	if len(ps.Todos) > 0 {
		fmt.Fprintf(&buf, "\n# todos\n")
		w4 := ton.NewWriter(&buf)
		w4.WriteHeader("file", "line", "text", "type", "lang", "resolved")
		for _, todo := range ps.Todos {
			w4.WriteRecord(todo.File, strconv.Itoa(todo.Line), escapeTON(todo.Text), todo.Type, todo.Language, strconv.FormatBool(todo.Resolved))
		}
	}

	if len(ps.Files) > 0 {
		fmt.Fprintf(&buf, "\n# files\n")
		w5 := ton.NewWriter(&buf)
		w5.WriteHeader("path", "lang", "modified", "churn", "symbols")
		for _, f := range ps.Files {
			w5.WriteRecord(f.Path, f.Language, f.LastModified.Format(time.RFC3339), strconv.Itoa(f.ChurnCount), escapeTON(strings.Join(f.Symbols, ",")))
		}
	}

	return buf.Bytes(), nil
}

func UnmarshalProjectStateTON(data []byte) (*ProjectState, error) {
	ps := NewProjectState()

	rawLines := strings.Split(string(data), "\n")

	var section string
	sectionHeader := ""
	var sectionLines []string

	flushSection := func() {
		if len(sectionLines) == 0 {
			return
		}
		switch section {
		case "state":
			parseStateSection(ps, sectionHeader, sectionLines)
		case "git":
			parseGitSection(ps, sectionHeader, sectionLines)
		case "symbols":
			parseSymbolsSection(ps, sectionHeader, sectionLines)
		case "todos":
			parseTodosSection(ps, sectionHeader, sectionLines)
		case "files":
			parseFilesSection(ps, sectionHeader, sectionLines)
		}
		sectionLines = nil
	}

	for _, line := range rawLines {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "# ") {
			flushSection()
			section = strings.TrimPrefix(trimmed, "# ")
			sectionHeader = ""
			sectionLines = nil
			continue
		}

		if section == "" {
			continue
		}

		if sectionHeader == "" {
			sectionHeader = trimmed
		} else {
			sectionLines = append(sectionLines, trimmed)
		}
	}
	flushSection()

	return ps, nil
}

func parseStateSection(ps *ProjectState, header string, lines []string) {
	cols := strings.Split(header, "|")
	colMap := make(map[string]int)
	for i, c := range cols {
		colMap[strings.TrimSpace(c)] = i
	}

	for _, line := range lines {
		rec := parseLine(line)
		if len(rec) < 2 {
			continue
		}
		if getCol(rec, colMap, "updated_at") != "" {
			if t, err := time.Parse(time.RFC3339, getCol(rec, colMap, "updated_at")); err == nil {
				ps.UpdatedAt = t
			}
		}
	}
}

func parseGitSection(ps *ProjectState, header string, lines []string) {
	cols := strings.Split(header, "|")
	colMap := make(map[string]int)
	for i, c := range cols {
		colMap[strings.TrimSpace(c)] = i
	}

	for _, line := range lines {
		rec := parseLine(line)
		if len(rec) < 1 {
			continue
		}
		hash := getCol(rec, colMap, "hash")
		msg := unescapeTON(getCol(rec, colMap, "message"))
		timeStr := getCol(rec, colMap, "time")

		var t time.Time
		if timeStr != "" {
			t, _ = time.Parse(time.RFC3339, timeStr)
		}

		ps.GitCommit = &integrations.Commit{
			Hash:    hash,
			Message: msg,
			Time:    t,
		}
	}
}

func parseSymbolsSection(ps *ProjectState, header string, lines []string) {
	cols := strings.Split(header, "|")
	colMap := make(map[string]int)
	for i, c := range cols {
		colMap[strings.TrimSpace(c)] = i
	}

	for _, line := range lines {
		rec := parseLine(line)
		if len(rec) < 1 {
			continue
		}

		sym := Symbol{
			Name:         getCol(rec, colMap, "name"),
			File:         getCol(rec, colMap, "file"),
			Type:         getCol(rec, colMap, "type"),
			Language:     getCol(rec, colMap, "lang"),
			IntentMarker: getCol(rec, colMap, "marker"),
			SignalSource: getCol(rec, colMap, "source"),
		}

		sym.Line, _ = strconv.Atoi(getCol(rec, colMap, "line"))
		sym.Column, _ = strconv.Atoi(getCol(rec, colMap, "col"))
		sym.EndLine, _ = strconv.Atoi(getCol(rec, colMap, "end_line"))
		sym.EndColumn, _ = strconv.Atoi(getCol(rec, colMap, "end_col"))
		sym.ChurnCount, _ = strconv.Atoi(getCol(rec, colMap, "churn"))

		if v := getCol(rec, colMap, "confidence"); v != "" {
			sym.Confidence, _ = strconv.ParseFloat(v, 64)
		}
		if v := getCol(rec, colMap, "instability"); v != "" {
			sym.InstabilityScore, _ = strconv.ParseFloat(v, 64)
		}
		if v := getCol(rec, colMap, "defined"); v != "" {
			sym.IsDefined, _ = strconv.ParseBool(v)
		}
		if v := getCol(rec, colMap, "called"); v != "" {
			sym.IsCalled, _ = strconv.ParseBool(v)
		}
		if v := getCol(rec, colMap, "todo"); v != "" {
			sym.HasTodo, _ = strconv.ParseBool(v)
		}
		if v := getCol(rec, colMap, "touched"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				sym.LastTouched = &t
			}
		}

		sym.State = SymbolState(getCol(rec, colMap, "state"))

		ps.Symbols = append(ps.Symbols, sym)
	}
}

func parseTodosSection(ps *ProjectState, header string, lines []string) {
	cols := strings.Split(header, "|")
	colMap := make(map[string]int)
	for i, c := range cols {
		colMap[strings.TrimSpace(c)] = i
	}

	for _, line := range lines {
		rec := parseLine(line)
		if len(rec) < 1 {
			continue
		}

		todo := Todo{
			File:     getCol(rec, colMap, "file"),
			Text:     unescapeTON(getCol(rec, colMap, "text")),
			Type:     getCol(rec, colMap, "type"),
			Language: getCol(rec, colMap, "lang"),
		}
		todo.Line, _ = strconv.Atoi(getCol(rec, colMap, "line"))
		if v := getCol(rec, colMap, "resolved"); v != "" {
			todo.Resolved, _ = strconv.ParseBool(v)
		}

		ps.Todos = append(ps.Todos, todo)
	}
}

func parseFilesSection(ps *ProjectState, header string, lines []string) {
	cols := strings.Split(header, "|")
	colMap := make(map[string]int)
	for i, c := range cols {
		colMap[strings.TrimSpace(c)] = i
	}

	for _, line := range lines {
		rec := parseLine(line)
		if len(rec) < 1 {
			continue
		}

		fc := FileContext{
			Path:     getCol(rec, colMap, "path"),
			Language: getCol(rec, colMap, "lang"),
		}
		fc.ChurnCount, _ = strconv.Atoi(getCol(rec, colMap, "churn"))
		if v := getCol(rec, colMap, "modified"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				fc.LastModified = t
			}
		}
		if v := getCol(rec, colMap, "symbols"); v != "" {
			fc.Symbols = strings.Split(v, ",")
		}

		ps.Files = append(ps.Files, fc)
	}
}

func parseLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return []string{""}
	}
	var fields []string
	var current strings.Builder
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '\\' && i+1 < len(line) {
			next := line[i+1]
			switch next {
			case 'n':
				current.WriteByte('\n')
				i++
			case '|':
				current.WriteByte('|')
				i++
			case '\\':
				current.WriteByte('\\')
				i++
			default:
				current.WriteByte(ch)
			}
		} else if ch == '|' {
			fields = append(fields, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	fields = append(fields, strings.TrimSpace(current.String()))
	return fields
}

func MarshalStateTON(ps *ProjectState, is *IntentState, snapshotHashes []string) ([]byte, string) {
	var buf bytes.Buffer

	writeSection := func(name string, header []string, rows [][]string) {
		if len(rows) == 0 {
			return
		}
		fmt.Fprintf(&buf, "# %s\n", name)
		w := ton.NewWriter(&buf)
		w.WriteHeader(header...)
		for _, row := range rows {
			w.WriteRecord(row...)
		}
	}

	gitInfo := "no git"
	if ps.GitCommit != nil {
		gitInfo = fmt.Sprintf("git:%s", ps.GitCommit.Hash)
		if len(gitInfo) > 11 {
			gitInfo = gitInfo[:11]
		}
	}

	buf.WriteString(fmt.Sprintf("project|%s|%s|%d symbols|%d todos\n",
		ps.UpdatedAt.Format(time.RFC3339),
		gitInfo,
		len(ps.Symbols),
		len(ps.Todos),
	))

	if len(snapshotHashes) > 0 {
		buf.WriteString(fmt.Sprintf("snapshots|%s\n", strings.Join(snapshotHashes, ",")))
	}

	if is != nil {
		activeIntents := is.GetActiveIntents()
		if len(activeIntents) > 0 {
			fmt.Fprintf(&buf, "\n# intents\n")
			w := ton.NewWriter(&buf)
			w.WriteHeader("id", "type", "sev", "status", "msg", "scope", "tags", "created_by")
			for _, intent := range activeIntents {
				tags := strings.Join(intent.Tags, ",")
				w.WriteRecord(
					intent.ID,
					string(intent.Type),
					strconv.Itoa(intent.Severity),
					string(intent.Status),
					intent.Message,
					intent.Scope,
					tags,
					intent.CreatedBy,
				)
			}
		}
	}

	symbolsByState := make(map[SymbolState][]Symbol)
	for _, sym := range ps.Symbols {
		symbolsByState[sym.State] = append(symbolsByState[sym.State], sym)
	}
	stateOrder := []SymbolState{Unstable, Unused, Abandoned, Planned, Active}
	for _, st := range stateOrder {
		syms := symbolsByState[st]
		if len(syms) == 0 {
			continue
		}
		sort.Slice(syms, func(i, j int) bool {
			if syms[i].File != syms[j].File {
				return syms[i].File < syms[j].File
			}
			return syms[i].Name < syms[j].Name
		})
		var rows [][]string
		for _, sym := range syms {
			lastTouched := ""
			if sym.LastTouched != nil {
				lastTouched = sym.LastTouched.Format(time.RFC3339)
			}
			rows = append(rows, []string{
				sym.Name, sym.File,
				strconv.Itoa(sym.Line), strconv.Itoa(sym.Column),
				strconv.Itoa(sym.EndLine), strconv.Itoa(sym.EndColumn),
				sym.Type, sym.Language,
				string(sym.State), strconv.FormatFloat(sym.Confidence, 'f', 2, 64),
				strconv.FormatBool(sym.IsDefined), strconv.FormatBool(sym.IsCalled), strconv.FormatBool(sym.HasTodo),
				strconv.Itoa(sym.ChurnCount), strconv.FormatFloat(sym.InstabilityScore, 'f', 4, 64),
				sym.IntentMarker, sym.SignalSource,
				lastTouched,
			})
		}
		writeSection("symbols:"+string(st),
			[]string{"name", "file", "line", "col", "end_line", "end_col", "type", "lang", "state", "confidence", "defined", "called", "todo", "churn", "instability", "marker", "source", "touched"},
			rows,
		)
	}

	activeTodos := ps.GetActiveTodos()
	if len(activeTodos) > 0 {
		var rows [][]string
		for _, todo := range activeTodos {
			rows = append(rows, []string{
				todo.File, strconv.Itoa(todo.Line), escapeTON(todo.Text), todo.Type, todo.Language,
			})
		}
		writeSection("todos", []string{"file", "line", "text", "type", "lang"}, rows)
	}

	h := sha256.Sum256(buf.Bytes())
	hash := hex.EncodeToString(h[:])
	fmt.Fprintf(&buf, "hash|%s\n", hash)

	return buf.Bytes(), hash
}

func VerifyStateTONHash(data []byte) bool {
	s := string(data)
	idx := strings.LastIndex(s, "\nhash|")
	if idx < 0 {
		return false
	}
	content := data[:idx+1]
	rest := strings.TrimPrefix(s[idx:], "\nhash|")
	storedHash := strings.TrimSpace(rest)
	if storedHash == "" {
		return false
	}
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:]) == storedHash
}
