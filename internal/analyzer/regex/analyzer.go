package regex

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// Analyzer implements a simple regex-based analyzer for any language
type Analyzer struct {
	patterns map[string]*regexp.Regexp
}

// NewAnalyzer creates a new regex analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		patterns: map[string]*regexp.Regexp{
			"function_js":    regexp.MustCompile(`function\s+(\w+)\s*\(`),
			"function_py":    regexp.MustCompile(`def\s+(\w+)\s*\(`),
			"function_rust":  regexp.MustCompile(`fn\s+(\w+)\s*\(`),
			"function_c":     regexp.MustCompile(`\w+\s+(\w+)\s*\([^)]*\)\s*\{`),
			"todo":           regexp.MustCompile(`(?i)^\s*(//|#|/\*|\*)\s*(TODO|FIXME|XXX|HACK|NOTE):\s*(.+)`),
			"intent_state":   regexp.MustCompile(`@dizz:state\s+(\w+)`),
			"intent_feature": regexp.MustCompile(`@dizz:feature\s+(\w+)`),
		},
	}
}

// Language returns "regex"
func (a *Analyzer) Language() string {
	return "regex"
}

// Supports checks if we should use regex analyzer
// Returns true for common code files
func (a *Analyzer) Supports(file string) bool {
	ext := filepath.Ext(file)
	supportedExts := map[string]bool{
		".js":   true,
		".ts":   true,
		".jsx":  true,
		".tsx":  true,
		".py":   true,
		".rs":   true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
		".java": true,
		".rb":   true,
		".php":  true,
	}
	return supportedExts[ext]
}

// Analyze extracts signals using regex patterns
func (a *Analyzer) Analyze(files []string) (*signals.SignalSet, error) {
	sigSet := &signals.SignalSet{}

	for _, filePath := range files {
		if err := a.analyzeFile(filePath, sigSet); err != nil {
			continue // Skip files with errors
		}
	}

	return sigSet, nil
}

// analyzeFile analyzes a single file
func (a *Analyzer) analyzeFile(filePath string, sigSet *signals.SignalSet) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	lang := a.detectLanguage(filePath)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		a.extractFunctions(line, filePath, lang, lineNum, sigSet)

		a.extractTodos(line, filePath, lang, lineNum, sigSet)

		a.extractIntents(line, filePath, lang, lineNum, sigSet)
	}

	return scanner.Err()
}

// extractFunctions finds function definitions
func (a *Analyzer) extractFunctions(line, filePath, lang string, lineNum int, sigSet *signals.SignalSet) {
	patterns := []string{"function_js", "function_py", "function_rust", "function_c"}

	for _, patternName := range patterns {
		if matches := a.patterns[patternName].FindStringSubmatch(line); matches != nil && len(matches) > 1 {
			sig := signals.NewSignal(signals.FunctionDefined, filePath).
				WithName(matches[1]).
				WithLine(lineNum).
				WithLanguage(lang).
				WithConfidence(0.7) // Lower confidence for regex
			sigSet.Add(*sig)
			break // Only match one pattern per line
		}
	}
}

func (a *Analyzer) extractTodos(line, filePath, lang string, lineNum int, sigSet *signals.SignalSet) {
	if matches := a.patterns["todo"].FindStringSubmatch(line); matches != nil && len(matches) > 2 {
		sig := signals.NewSignal(signals.TodoFound, filePath).
			WithLine(lineNum).
			WithLanguage(lang).
			WithMeta("type", strings.ToUpper(matches[1])).
			WithMeta("text", strings.TrimSpace(matches[2]))
		sigSet.Add(*sig)
	}
}

// extractIntents finds @dizz markers
func (a *Analyzer) extractIntents(line, filePath, lang string, lineNum int, sigSet *signals.SignalSet) {
	if matches := a.patterns["intent_state"].FindStringSubmatch(line); matches != nil && len(matches) > 1 {
		sig := signals.NewSignal(signals.IntentMarker, filePath).
			WithLine(lineNum).
			WithLanguage(lang).
			WithMeta("marker_type", "state").
			WithMeta("value", matches[1])
		sigSet.Add(*sig)
	}

	if matches := a.patterns["intent_feature"].FindStringSubmatch(line); matches != nil && len(matches) > 1 {
		sig := signals.NewSignal(signals.IntentMarker, filePath).
			WithLine(lineNum).
			WithLanguage(lang).
			WithMeta("marker_type", "feature").
			WithMeta("value", matches[1])
		sigSet.Add(*sig)
	}
}

// detectLanguage infers language from file extension
func (a *Analyzer) detectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	langMap := map[string]string{
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "javascript",
		".tsx":  "typescript",
		".py":   "python",
		".rs":   "rust",
		".c":    "c",
		".cpp":  "cpp",
		".h":    "c",
		".hpp":  "cpp",
		".java": "java",
		".rb":   "ruby",
		".php":  "php",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}
