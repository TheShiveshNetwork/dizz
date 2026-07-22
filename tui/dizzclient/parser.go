package dizzclient

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func ParseStatusOutput(raw string) *Summary {
	summary := &Summary{}

	numberRe := regexp.MustCompile(`Active\s+(\d+)`)
	plannedRe := regexp.MustCompile(`Planned\s+(\d+)`)
	unusedRe := regexp.MustCompile(`Unused\s+(\d+)`)
	unstableRe := regexp.MustCompile(`Unstable\s+(\d+)`)
	abandonedRe := regexp.MustCompile(`Abandoned\s+(\d+)`)
	todoRe := regexp.MustCompile(`TODOs:\s*(\d+)`)
	intentRe := regexp.MustCompile(`Intents:\s*(\d+)`)

	for _, line := range strings.Split(raw, "\n") {
		clean := stripANSI(line)

		if m := numberRe.FindStringSubmatch(clean); len(m) >= 2 {
			summary.Active = parseIntSafe(m[1])
		} else if m := plannedRe.FindStringSubmatch(clean); len(m) >= 2 {
			summary.Planned = parseIntSafe(m[1])
		} else if m := unusedRe.FindStringSubmatch(clean); len(m) >= 2 {
			summary.Unused = parseIntSafe(m[1])
		} else if m := unstableRe.FindStringSubmatch(clean); len(m) >= 2 {
			summary.Unstable = parseIntSafe(m[1])
		} else if m := abandonedRe.FindStringSubmatch(clean); len(m) >= 2 {
			summary.Abandoned = parseIntSafe(m[1])
		} else if m := todoRe.FindStringSubmatch(clean); len(m) >= 2 {
			summary.ActiveTodos = parseIntSafe(m[1])
		} else if m := intentRe.FindStringSubmatch(clean); len(m) >= 2 {
			summary.Intents = parseIntSafe(m[1])
		} else if strings.Contains(clean, "Project:") {
			parts := strings.SplitN(clean, ":", 2)
			if len(parts) == 2 {
				summary.ProjectName = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(clean, "Branch:") {
			parts := strings.SplitN(clean, ":", 2)
			if len(parts) == 2 {
				summary.Branch = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(clean, "Last Commit:") {
			re := regexp.MustCompile(`\(([a-f0-9]+)\)`)
			matches := re.FindStringSubmatch(clean)
			if len(matches) >= 2 {
				summary.Commit = matches[1]
			}
		}
	}

	summary.TotalSymbols = summary.Active + summary.Planned +
		summary.Unstable + summary.Unused + summary.Abandoned

	return summary
}

func ParseLogOutput(raw string) []Symbol {
	lines := strings.Split(raw, "\n")
	var symbols []Symbol

	var currentState string
	var currentFile string

	sectionRe := regexp.MustCompile(`^\s*(PLANNED|UNUSED|UNSTABLE|ABANDONED)(\s|$)`)
	activeSectionRe := regexp.MustCompile(`ACTIVE\s`)
	fileRe := regexp.MustCompile(`^\s{2}(/\S+)\s+\((\d+)\s+items?\)`)
	symbolRe := regexp.MustCompile(`^\s{5}•\s+(\S+)\s+(\d+):(\d+)`)
	activeSymbolRe := regexp.MustCompile(`^\s{2}(\S[\S\s]*)$`)

	inActiveSection := false

	for _, line := range lines {
		clean := stripANSI(line)
		if clean == "" && !inActiveSection {
			continue
		}

		if m := sectionRe.FindStringSubmatch(clean); len(m) >= 2 {
			s := strings.ToUpper(m[1])
			switch s {
			case "PLANNED":
				currentState = "planned"
			case "UNUSED":
				currentState = "unused"
			case "UNSTABLE":
				currentState = "unstable"
			case "ABANDONED":
				currentState = "abandoned"
			}
			currentFile = ""
			inActiveSection = false
			continue
		}

		if activeSectionRe.MatchString(clean) {
			currentState = "active"
			currentFile = ""
			inActiveSection = true
			continue
		}

		if inActiveSection {
			if clean == "" || clean == " " {
				inActiveSection = false
				continue
			}
			if m := fileRe.FindStringSubmatch(clean); len(m) >= 3 {
				currentFile = m[1]
				continue
			}
			if m := symbolRe.FindStringSubmatch(clean); len(m) >= 4 && currentFile != "" {
				symbols = append(symbols, Symbol{
					State: "active",
					Name:  m[1],
					File:  currentFile,
					Line:  parseIntSafe(m[2]),
				})
				continue
			}
			if m := activeSymbolRe.FindStringSubmatch(clean); len(m) >= 2 {
				name := strings.TrimSpace(m[1])
				if name != "" && !strings.HasPrefix(name, "...") {
					symbols = append(symbols, Symbol{
						State: "active",
						Name:  name,
						File:  "(active)",
						Line:  0,
					})
				}
			}
			continue
		}

		if m := fileRe.FindStringSubmatch(clean); len(m) >= 3 {
			currentFile = m[1]
			continue
		}

		if m := symbolRe.FindStringSubmatch(clean); len(m) >= 4 && currentFile != "" && currentState != "" {
			symbols = append(symbols, Symbol{
				State: currentState,
				Name:  m[1],
				File:  currentFile,
				Line:  parseIntSafe(m[2]),
			})
		}
	}

	return symbols
}

func ParseTodosOutput(data []byte) ([]Todo, error) {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var todos []Todo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		todos = append(todos, Todo{
			File: parts[0],
			Line: parseIntSafe(parts[1]),
			Type: parts[2],
			Text: parts[3],
		})
	}

	return todos, nil
}

func ParseIntentsTON(data []byte) ([]Intent, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty intent file")
	}

	header := scanner.Text()
	cols := strings.Split(header, "|")
	colMap := make(map[string]int)
	for i, col := range cols {
		colMap[strings.TrimSpace(col)] = i
	}

	required := []string{"id", "type", "msg", "sev", "status"}
	for _, r := range required {
		if _, ok := colMap[r]; !ok {
			return nil, fmt.Errorf("missing required column: %s", r)
		}
	}

	var intents []Intent
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "|")
		intent := Intent{
			ID:       getField(fields, colMap, "id"),
			Type:     getField(fields, colMap, "type"),
			Message:  getField(fields, colMap, "msg"),
			Scope:    getField(fields, colMap, "scope"),
			Severity: parseInt(getField(fields, colMap, "sev")),
			Status:   getField(fields, colMap, "status"),
			Tags:     parseTags(getField(fields, colMap, "tags")),
		}
		intents = append(intents, intent)
	}

	return intents, nil
}

func ParseSnapshotOutput(raw string) string {
	clean := stripANSI(raw)
	re := regexp.MustCompile(`Snapshot saved:\s+(\S+)`)
	matches := re.FindStringSubmatch(clean)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func ParseVersionOutput(raw string) string {
	return strings.TrimSpace(stripANSI(raw))
}

func ParseSnapshotListOutput(raw string) []SnapshotInfo {
	lines := strings.Split(raw, "\n")
	var entries []SnapshotInfo
	re := regexp.MustCompile(`^\s{2}(\w{3}\s+\d{2}\s+\d{2}:\d{2})\s+(\S+)\s+(\S+)\s+(\S+)`)
	for _, line := range lines {
		clean := stripANSI(line)
		m := re.FindStringSubmatch(clean)
		if len(m) >= 5 {
			entries = append(entries, SnapshotInfo{
				Timestamp: m[1],
				Hash:      m[2],
				Kind:      m[3],
				Size:      m[4],
			})
		}
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}

func parseIntSafe(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func IsDizzInitialized() bool {
	_, err := FindDizzRoot()
	return err == nil
}

func WriteStateTON(data []byte) error {
	root, err := FindDizzRoot()
	if err != nil {
		return err
	}
	path := filepath.Join(root, ".dizz", "context.ton")
	return os.WriteFile(path, data, 0644)
}

func getField(fields []string, colMap map[string]int, key string) string {
	idx, ok := colMap[key]
	if !ok || idx >= len(fields) {
		return ""
	}
	return strings.TrimSpace(fields[idx])
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
