package regex

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// LanguageDef defines regex patterns for a specific language
type LanguageDef struct {
	Name            string
	Extensions      []string
	FuncPattern     *regexp.Regexp
	ImportPattern   *regexp.Regexp
	CommentPrefixes []string
}

// Analyzer implements a robust regex-based analyzer for any language
type Analyzer struct {
	languages map[string]LanguageDef
	extToLang map[string]string
	
	// Global patterns
	todoPattern   *regexp.Regexp
	statePattern  *regexp.Regexp
	featurePattern *regexp.Regexp
}

// NewAnalyzer creates a new language-agnostic regex analyzer
func NewAnalyzer() *Analyzer {
	// Define supported languages
	langs := []LanguageDef{
		{
			Name:       "javascript",
			Extensions: []string{".js", ".jsx"},
			FuncPattern: regexp.MustCompile(`(?:async\s+)?function\s+([a-zA-Z0-9_]+)\s*\(`),
			ImportPattern: regexp.MustCompile(`import\s+.*\s+from\s+['"](.*)['"]`),
			CommentPrefixes: []string{"//", "/*"},
		},
		{
			Name:       "typescript",
			Extensions: []string{".ts", ".tsx"},
			FuncPattern: regexp.MustCompile(`(?:async\s+)?function\s+([a-zA-Z0-9_]+)\s*\(`),
			ImportPattern: regexp.MustCompile(`import\s+.*\s+from\s+['"](.*)['"]`),
			CommentPrefixes: []string{"//", "/*"},
		},
		{
			Name:       "python",
			Extensions: []string{".py"},
			FuncPattern: regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\s*\(`),
			ImportPattern: regexp.MustCompile(`(?:import|from)\s+([a-zA-Z0-9_.]+)`),
			CommentPrefixes: []string{"#"},
		},
		{
			Name:       "rust",
			Extensions: []string{".rs"},
			FuncPattern: regexp.MustCompile(`fn\s+([a-zA-Z0-9_]+)\s*(?:<.*>)?\s*\(`),
			ImportPattern: regexp.MustCompile(`use\s+([a-zA-Z0-9_:]+)`),
			CommentPrefixes: []string{"//"},
		},
		{
			Name:       "ruby",
			Extensions: []string{".rb"},
			FuncPattern: regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)`),
			ImportPattern: regexp.MustCompile(`require\s+['"](.*)['"]`),
			CommentPrefixes: []string{"#"},
		},
		{
			Name:       "php",
			Extensions: []string{".php"},
			FuncPattern: regexp.MustCompile(`function\s+([a-zA-Z0-9_]+)\s*\(`),
			ImportPattern: regexp.MustCompile(`(?:include|require)(?:_once)?\s*['"](.*)['"]`),
			CommentPrefixes: []string{"//", "#", "/*"},
		},
		{
			Name:       "java",
			Extensions: []string{".java"},
			FuncPattern: regexp.MustCompile(`(?:public|protected|private|static|\s) +[\w\<\>\[\]]+\s+([a-zA-Z0-9_]+)\s*\(.*\)\s*\{`),
			ImportPattern: regexp.MustCompile(`import\s+([a-zA-Z0-9_.]+)`),
			CommentPrefixes: []string{"//", "/*"},
		},
		{
			Name:       "cpp",
			Extensions: []string{".cpp", ".hpp", ".c", ".h"},
			FuncPattern: regexp.MustCompile(`(?:\w+\s+)+([a-zA-Z0-9_]+)\s*\([^)]*\)\s*\{`),
			ImportPattern: regexp.MustCompile(`#include\s+[<"](.*)[>"]`),
			CommentPrefixes: []string{"//", "/*"},
		},
	}

	langMap := make(map[string]LanguageDef)
	extMap := make(map[string]string)
	for _, l := range langs {
		langMap[l.Name] = l
		for _, ext := range l.Extensions {
			extMap[ext] = l.Name
		}
	}

	return &Analyzer{
		languages: langMap,
		extToLang: extMap,
		todoPattern:    regexp.MustCompile(`(?i)\b(TODO|FIXME|XXX|HACK|NOTE):\s*(.*)`),
		statePattern:   regexp.MustCompile(`@dizz:state\s+(\w+)`),
		featurePattern: regexp.MustCompile(`@dizz:feature\s+(\w+)`),
	}
}

// Language returns "regex"
func (a *Analyzer) Language() string {
	return "regex"
}

// Supports checks if we have a definition for this file's extension
func (a *Analyzer) Supports(file string) bool {
	ext := filepath.Ext(file)
	_, ok := a.extToLang[ext]
	return ok
}

// Analyze extracts signals using language-specific patterns
func (a *Analyzer) Analyze(files []string) (*signals.SignalSet, error) {
	sigSet := &signals.SignalSet{}

	for _, filePath := range files {
		ext := filepath.Ext(filePath)
		langName, ok := a.extToLang[ext]
		if !ok {
			continue
		}
		
		langDef := a.languages[langName]
		if err := a.analyzeFile(filePath, langDef, sigSet); err != nil {
			continue
		}
	}

	return sigSet, nil
}

// analyzeFile analyzes a single file using a specific language definition
func (a *Analyzer) analyzeFile(filePath string, lang LanguageDef, sigSet *signals.SignalSet) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Determine if this line contains a comment
		isComment := false
		for _, prefix := range lang.CommentPrefixes {
			if strings.Contains(line, prefix) {
				isComment = true
				break
			}
		}

		// 1. Extract Functions
		if matches := lang.FuncPattern.FindStringSubmatch(line); matches != nil && len(matches) > 1 {
			sig := signals.NewSignal(signals.FunctionDefined, filePath).
				WithName(matches[1]).
				WithLine(lineNum).
				WithLanguage(lang.Name).
				WithConfidence(0.7)
			sigSet.Add(*sig)
		}

		// 2. Extract Imports
		if matches := lang.ImportPattern.FindStringSubmatch(line); matches != nil && len(matches) > 1 {
			sig := signals.NewSignal(signals.ImportFound, filePath).
				WithName(matches[1]).
				WithLine(lineNum).
				WithLanguage(lang.Name)
			sigSet.Add(*sig)
		}

		// 3. Extract TODOs (Only in comments)
		if isComment {
			if matches := a.todoPattern.FindStringSubmatch(line); matches != nil && len(matches) > 2 {
				sig := signals.NewSignal(signals.TodoFound, filePath).
					WithLine(lineNum).
					WithLanguage(lang.Name).
					WithMeta("type", strings.ToUpper(matches[1])).
					WithMeta("text", strings.TrimSpace(matches[2]))
				sigSet.Add(*sig)
			}
		}

		// 4. Extract Intents
		if matches := a.statePattern.FindStringSubmatch(line); matches != nil && len(matches) > 1 {
			sig := signals.NewSignal(signals.IntentMarker, filePath).
				WithLine(lineNum).
				WithLanguage(lang.Name).
				WithMeta("marker_type", "state").
				WithMeta("value", matches[1])
			sigSet.Add(*sig)
		}

		if matches := a.featurePattern.FindStringSubmatch(line); matches != nil && len(matches) > 1 {
			sig := signals.NewSignal(signals.IntentMarker, filePath).
				WithLine(lineNum).
				WithLanguage(lang.Name).
				WithMeta("marker_type", "feature").
				WithMeta("value", matches[1])
			sigSet.Add(*sig)
		}
	}

	return scanner.Err()
}
