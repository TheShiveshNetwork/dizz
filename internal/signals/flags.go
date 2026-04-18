package signals

import (
	"regexp"
	"strings"
)

type IgnoreSignal struct {
	Type      string
	File      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
	Language  string
	Metadata  map[string]interface{}
}

func NewIgnoreFlag(file, name string, line, column, endLine, endColumn int, language string) IgnoreSignal {
	return IgnoreSignal{
		Type:      string(IgnoreFlag),
		File:      file,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
		Language:  language,
		Metadata: map[string]interface{}{
			"symbol_name": name,
		},
	}
}

var (
	// TODO: make it compatible to most of the languages. currently just extracts // comments
	ignorePattern = regexp.MustCompile(`^\s*//\s*@ignore-(unstable|unused|abandoned)\b`)
	ignoreTypeRe  = regexp.MustCompile(`@ignore-(unstable|unused|abandoned)`)
	funcPattern   = regexp.MustCompile(`func\s+(\w+)\s*\(`)
)

// ExtractIgnoreMarkers finds @ignore-* patterns in code
func ExtractIgnoreMarkers(source, file, language string) []IgnoreSignal {
	var signals []IgnoreSignal

	lines := strings.Split(source, "\n")

	for lineNum, line := range lines {
		if ignorePattern.MatchString(line) {
			ignoreType := extractIgnoreTypeFromComment(line)

			// Find the symbol this comment applies to by looking for the next function/struct/class definition
			symbolInfo := findNextSymbol(lines, lineNum+1)

			if symbolInfo.name != "" {
				// Try to find the symbol end by looking for closing braces
				endLine := findSymbolEnd(lines, symbolInfo.line, symbolInfo.name)
				if endLine > symbolInfo.line {
					signal := NewIgnoreFlag(file, symbolInfo.name, lineNum+1, 0, endLine, len(line), language)
					// Store the ignore type in metadata
					signal.Metadata["ignore_type"] = ignoreType
					signals = append(signals, signal)
				}
			}
		}
	}

	return signals
}

// extractIgnoreTypeFromComment extracts ignore type from @ignore-* comment
func extractIgnoreTypeFromComment(comment string) string {
	matches := ignoreTypeRe.FindStringSubmatch(comment)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// symbolInfo contains information about a found symbol
type symbolInfo struct {
	name string
	line int
}

// findNextSymbol finds the next function/struct/class definition after the given line
// startLine is 1-based (line number, not index)
func findNextSymbol(lines []string, startLine int) symbolInfo {
	// Convert to 0-based index and look for the next symbol definition
	for i := startLine; i < len(lines) && i < startLine+10; i++ { // startLine-1 would be the comment line
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}

		// Look for function definitions (Go) - more flexible matching
		if strings.HasPrefix(line, "func ") {
			// Handle both "func name()" and "func name() returnType {" patterns
			matches := funcPattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				return symbolInfo{name: matches[1], line: i + 1}
			}
		}

		// Look for other definitions as fallback
		if strings.HasPrefix(line, "var ") || strings.HasPrefix(line, "const ") {
			words := strings.Fields(line)
			if len(words) >= 2 {
				// Remove parentheses or other characters from the name
				name := strings.TrimSuffix(words[1], ")")
				name = strings.TrimSuffix(name, "(")
				return symbolInfo{name: name, line: i + 1}
			}
		}

		// Look for type definitions
		if strings.HasPrefix(line, "type ") {
			words := strings.Fields(line)
			if len(words) >= 2 {
				name := strings.TrimSuffix(words[1], ")")
				name = strings.TrimSuffix(name, "(")
				return symbolInfo{name: name, line: i + 1}
			}
		}
	}

	return symbolInfo{}
}

// findSymbolEnd finds the end line of a symbol by looking for closing braces
func findSymbolEnd(lines []string, startLine int, symbolName string) int {
	// Simple heuristic: find the matching closing brace for this function/class
	for i := startLine; i < len(lines) && i < startLine+50; i++ {
		line := strings.TrimSpace(lines[i])

		// Look for the symbol name followed by opening patterns
		if strings.Contains(line, symbolName) &&
			(strings.Contains(line, "{") || strings.Contains(line, "(")) {
			braceCount := 0
			for j := i; j < len(lines); j++ {
				scanLine := strings.TrimSpace(lines[j])
				braceCount += strings.Count(scanLine, "{") - strings.Count(scanLine, "}")

				// Check if we've closed the opening braces
				if braceCount <= 0 &&
					(strings.Contains(scanLine, "}") ||
						strings.HasSuffix(scanLine, ")") ||
						strings.HasSuffix(scanLine, ";")) {
					return j + 1
				}
			}
		}
	}
	return startLine + 10 // Default fallback
}
