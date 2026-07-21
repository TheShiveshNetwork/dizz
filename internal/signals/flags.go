package signals

import (
	"regexp"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/language"
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

func NewIgnoreFlag(file, name string, line, column, endLine, endColumn int, lang string) IgnoreSignal {
	return IgnoreSignal{
		Type:      string(IgnoreFlag),
		File:      file,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
		Language:  lang,
		Metadata: map[string]interface{}{
			"symbol_name": name,
		},
	}
}

// ignoreRe matches @ignore-* and @dizz-ignore-* in any text fragment
// that has already been stripped of its comment prefix.
var ignoreRe = regexp.MustCompile(`@(?:dizz-)?ignore-(unstable|unused|abandoned)\b`)

// ignoreMarkerRe is used to recognise whether a line even contains an ignore
// marker before the more expensive processing begins.
var ignoreMarkerRe = regexp.MustCompile(`@(?:dizz-)?ignore-(unstable|unused|abandoned)\b`)

// fileIgnoreRe matches @dizz-ignore-file for whole-file exclusion.
var fileIgnoreRe = regexp.MustCompile(`@dizz-ignore-file\b`)

// HasFileIgnoreFlag checks whether the first 5 lines of source contain
// @dizz-ignore-file.  When present, the entire file is skipped during analysis.
func HasFileIgnoreFlag(source string) bool {
	end := len(source)
	if end > 200 {
		end = 200
	}
	return fileIgnoreRe.MatchString(source[:end])
}

// Universal fallback patterns that cover the most common definition styles
// across languages so even unconfigured languages work reasonably.
// Compiled once at package init.
var universalFunctionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`func\s+(\w+)\s*\(`),
	regexp.MustCompile(`(?:async\s+)?def\s+(\w+[?!]?)`),
	regexp.MustCompile(`(?:async\s+)?function\s+(\w+)\s*\(`),
	regexp.MustCompile(`(?:pub\s+)?fn\s+(\w+)\s*[<(]`),
	regexp.MustCompile(`sub\s+(\w+)\s*[{(]`),
	regexp.MustCompile(`defp?\s+(\w+[?!]?)`),
}

// compiledFunctionPatterns caches compiled function patterns for a LanguageConfig
// so we only pay the compilation cost once per language.
var compiledFunctionPatterns = make(map[string][]*regexp.Regexp)

// init compiles function patterns for all registered languages at startup.
func init() {
	for _, lc := range language.All() {
		var patterns []*regexp.Regexp
		for _, p := range lc.FunctionPatterns {
			if re, err := regexp.Compile(p); err == nil {
				patterns = append(patterns, re)
			}
		}
		compiledFunctionPatterns[lc.ID] = patterns
	}
}

// ExtractIgnoreMarkers finds @ignore-* patterns in code for any registered
// language.  It uses the language's CommentStyles so it works correctly with
// //, #, --, ;, %, and other comment syntaxes.
func ExtractIgnoreMarkers(source, file, langID string) []IgnoreSignal {
	lc, ok := language.Get(langID)
	if !ok {
		// Unknown language — fall back to scanning for the bare marker regardless
		// of comment syntax.
		lc = language.LanguageConfig{
			ID: langID,
			CommentStyles: []language.CommentStyle{
				{LinePrefix: "//"},
				{LinePrefix: "#"},
				{LinePrefix: "--"},
				{LinePrefix: "%"},
				{LinePrefix: ";"},
			},
		}
	}

	lines := strings.Split(source, "\n")
	result := make([]IgnoreSignal, 0, len(lines))

	for lineIdx, line := range lines {
		if !ignoreMarkerRe.MatchString(line) {
			continue
		}

		// Check the line contains the marker inside a comment.
		if !lineContainsIgnoreInComment(line, lc) {
			continue
		}

		ignoreType := extractIgnoreTypeFromLine(line)
		symbolInfo := findNextSymbol(lines, lineIdx+1, lc)
		if symbolInfo.name == "" {
			continue
		}

		endLine := findSymbolEnd(lines, symbolInfo.line, symbolInfo.name)
		if endLine <= symbolInfo.line {
			endLine = symbolInfo.line + 1
		}

		sig := NewIgnoreFlag(file, symbolInfo.name, lineIdx+1, 0, endLine, len(line), langID)
		sig.Metadata["ignore_type"] = ignoreType
		result = append(result, sig)
	}

	return result
}

// lineContainsIgnoreInComment returns true when the @ignore-* marker appears
// inside a comment on this line, according to the language's comment styles.
func lineContainsIgnoreInComment(line string, lc language.LanguageConfig) bool {
	trimmed := strings.TrimSpace(line)

	for _, cs := range lc.CommentStyles {
		if cs.LinePrefix != "" {
			idx := strings.Index(trimmed, cs.LinePrefix)
			if idx >= 0 {
				after := trimmed[idx+len(cs.LinePrefix):]
				if ignoreRe.MatchString(after) {
					return true
				}
			}
		}
		// Block comment start on same line (e.g. /* @ignore-unused */ or (* ... *))
		if cs.BlockStart != "" {
			idx := strings.Index(trimmed, cs.BlockStart)
			if idx >= 0 {
				after := trimmed[idx+len(cs.BlockStart):]
				if ignoreRe.MatchString(after) {
					return true
				}
			}
		}
	}

	// If no comment style matched but the marker is present, accept it as a
	// best-effort match so unknown languages are not silently broken.
	if len(lc.CommentStyles) == 0 && ignoreRe.MatchString(line) {
		return true
	}

	return false
}

// extractIgnoreTypeFromLine extracts the ignore type keyword from a line.
func extractIgnoreTypeFromLine(line string) string {
	m := ignoreRe.FindStringSubmatch(line)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// symbolInfo contains information about a found symbol.
type symbolInfo struct {
	name string
	line int // 1-based
}

// findNextSymbol searches for the next symbol definition after startLine.
// It uses language-specific function patterns when available, with a set of
// universal fallback patterns for common languages.
func findNextSymbol(lines []string, startLine int, lc language.LanguageConfig) symbolInfo {
	// Get cached language function patterns
	patterns := compiledFunctionPatterns[lc.ID]

	allPatterns := append(patterns, universalFunctionPatterns...)

	for i := startLine; i < len(lines) && i < startLine+10; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		for _, re := range allPatterns {
			if m := re.FindStringSubmatch(line); len(m) > 1 && m[1] != "" {
				return symbolInfo{name: m[1], line: i + 1}
			}
		}
	}

	return symbolInfo{}
}

// findSymbolEnd estimates the line where the symbol body ends.
func findSymbolEnd(lines []string, startLine int, symbolName string) int {
	for i := startLine; i < len(lines) && i < startLine+50; i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, symbolName) &&
			(strings.Contains(line, "{") || strings.Contains(line, "(")) {
			braceCount := 0
			for j := i; j < len(lines); j++ {
				scan := strings.TrimSpace(lines[j])
				braceCount += strings.Count(scan, "{") - strings.Count(scan, "}")
				if braceCount <= 0 &&
					(strings.Contains(scan, "}") ||
						strings.HasSuffix(scan, ")") ||
						strings.HasSuffix(scan, ";")) {
					return j + 1
				}
			}
		}
	}
	return startLine + 10
}
