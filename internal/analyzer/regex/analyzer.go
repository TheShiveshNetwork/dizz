// Package regex provides a language-agnostic regex/lexical analyzer that works
// for any language registered in the language registry.  It is the fallback
// analyzer for all non-Go files; the AST analyzer handles Go.
package regex

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/language"
	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// compiledLanguage caches compiled regexps for a LanguageConfig so we only pay
// the compilation cost once per language.
type compiledLanguage struct {
	cfg          language.LanguageConfig
	fnPatterns   []*regexp.Regexp
	typePatterns []*regexp.Regexp
	callPatterns []*regexp.Regexp
	todoPattern  *regexp.Regexp
}

// Analyzer uses the language registry to extract signals from any supported
// source file.  It is intentionally language-agnostic: all behaviour comes from
// the LanguageConfig for the detected language.
type Analyzer struct {
	compiled map[string]*compiledLanguage // keyed by language ID
	lastFile string                       // single-entry detection cache
	lastLang string                       // language ID of lastFile
}

// NewAnalyzer builds the analyzer and pre-compiles patterns for all registered
// languages.
func NewAnalyzer() *Analyzer {
	a := &Analyzer{compiled: make(map[string]*compiledLanguage)}
	for _, lc := range language.All() {
		a.compiled[lc.ID] = compile(lc)
	}
	return a
}

func compile(lc language.LanguageConfig) *compiledLanguage {
	cl := &compiledLanguage{cfg: lc}

	for _, p := range lc.FunctionPatterns {
		if re, err := regexp.Compile(p); err == nil {
			cl.fnPatterns = append(cl.fnPatterns, re)
		}
	}
	for _, p := range lc.TypePatterns {
		if re, err := regexp.Compile(p); err == nil {
			cl.typePatterns = append(cl.typePatterns, re)
		}
	}
	for _, p := range lc.CallPatterns {
		if re, err := regexp.Compile(p); err == nil {
			cl.callPatterns = append(cl.callPatterns, re)
		}
	}

	// Build a TODO pattern that matches any of the language's comment prefixes.
	cl.todoPattern = buildTodoPattern(lc)
	return cl
}

// buildTodoPattern creates a single regexp that matches TODO/FIXME/etc. in any
// comment style defined for the language.
func buildTodoPattern(lc language.LanguageConfig) *regexp.Regexp {
	var prefixes []string
	for _, cs := range lc.CommentStyles {
		if cs.LinePrefix != "" {
			prefixes = append(prefixes, regexp.QuoteMeta(cs.LinePrefix))
		}
		if cs.BlockStart != "" {
			prefixes = append(prefixes, regexp.QuoteMeta(cs.BlockStart))
		}
	}
	if len(prefixes) == 0 {
		// Fallback: accept // or # if no prefix defined
		prefixes = []string{`//`, `#`}
	}
	joined := strings.Join(prefixes, "|")
	return regexp.MustCompile(`(?i)^\s*(?:` + joined + `|\*+|/\*)\s*(TODO|FIXME|XXX|HACK|NOTE):\s*(.+)`)
}

// Language returns a descriptive identifier for this analyzer.
func (a *Analyzer) Language() string { return "regex" }

// SupportedExtensions returns all registered language extensions except Go.
func (a *Analyzer) SupportedExtensions() []string {
	all := language.AllExtensions()
	nonGo := make([]string, 0, len(all)-1)
	for _, ext := range all {
		if ext != ".go" {
			nonGo = append(nonGo, ext)
		}
	}
	return nonGo
}

// @dizz-ignore-unused
// Supports returns true for every file whose language is in the registry and is
// not Go (Go is handled by the AST analyzer with higher accuracy).
func (a *Analyzer) Supports(file string) bool {
	lc, ok := a.detectCached(file)
	if !ok {
		return false
	}
	return lc.ID != "go"
}

// detectCached returns the detected language for file, reusing the previous
// result if the same file is queried again.  This avoids redundant detection
// in the common pattern: Supports(file) → AnalyzeFile(file).
func (a *Analyzer) detectCached(file string) (language.LanguageConfig, bool) {
	if a.lastFile == file && a.lastLang != "" {
		return language.Get(a.lastLang)
	}
	lc, ok := language.Detect(file)
	if ok {
		a.lastFile = file
		a.lastLang = lc.ID
	}
	return lc, ok
}

// Analyze extracts signals for all given files.
func (a *Analyzer) Analyze(files []string) (*signals.SignalSet, error) {
	sigSet := &signals.SignalSet{}
	for _, filePath := range files {
		_ = a.analyzeFile(filePath, sigSet) // skip individual file errors
	}
	return sigSet, nil
}

// AnalyzeFile extracts signals from a single file.
func (a *Analyzer) AnalyzeFile(file string) ([]signals.Signal, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return a.AnalyzeFileContent(file, content)
}

// AnalyzeFileContent extracts signals from a single file using pre-read content.
func (a *Analyzer) AnalyzeFileContent(file string, content []byte) ([]signals.Signal, error) {
	sigSet := &signals.SignalSet{}
	if err := a.analyzeFileFromContent(file, string(content), sigSet); err != nil {
		return nil, err
	}
	return sigSet.Signals, nil
}

// analyzeFile processes a single file line-by-line.
func (a *Analyzer) analyzeFile(filePath string, sigSet *signals.SignalSet) error {
	lc, ok := a.detectCached(filePath)
	if !ok {
		return nil
	}
	cl := a.compiled[lc.ID]
	if cl == nil {
		return nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	tier := tierLabel(lc.DefaultTier)
	return a.scanFile(filePath, bufio.NewScanner(f), lc.ID, tier, cl, sigSet)
}

// analyzeFileFromContent is like analyzeFile but uses pre-read content.
func (a *Analyzer) analyzeFileFromContent(filePath string, content string, sigSet *signals.SignalSet) error {
	lc, ok := a.detectCached(filePath)
	if !ok {
		return nil
	}
	cl := a.compiled[lc.ID]
	if cl == nil {
		return nil
	}
	tier := tierLabel(lc.DefaultTier)
	return a.scanFile(filePath, bufio.NewScanner(strings.NewReader(content)), lc.ID, tier, cl, sigSet)
}

// scanFile is the shared line-by-line scanning core used by both
// analyzeFile and analyzeFileFromContent.
func (a *Analyzer) scanFile(filePath string, scanner *bufio.Scanner, langID string, tier string, cl *compiledLanguage, sigSet *signals.SignalSet) error {
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		isDef := a.extractDefinitions(line, filePath, langID, lineNum, tier, cl, sigSet)
		if !isDef {
			a.extractCalls(line, filePath, langID, lineNum, tier, cl, sigSet)
		}
		a.extractAnnotations(line, filePath, langID, lineNum, cl, sigSet)
	}

	return scanner.Err()
}

// extractDefinitions combines function and type definition extraction
// into a single pass. Returns true if the line matched a definition.
func (a *Analyzer) extractDefinitions(
	line, filePath, langID string, lineNum int, tier string,
	cl *compiledLanguage, sigSet *signals.SignalSet,
) bool {
	for _, re := range cl.fnPatterns {
		if m := re.FindStringSubmatch(line); m != nil && len(m) > 1 && m[1] != "" {
			if !cl.cfg.Keywords[m[1]] {
				sig := signals.NewSignal(signals.FunctionDefined, filePath).
					WithName(m[1]).
					WithLine(lineNum).
					WithLanguage(langID).
					WithConfidence(confidenceFor(cl.cfg.DefaultTier)).
					WithMeta("source", "regex").
					WithMeta("source_tier", tier)
				sigSet.Add(*sig)
				return true
			}
		}
	}
	for _, re := range cl.typePatterns {
		if m := re.FindStringSubmatch(line); m != nil && len(m) > 1 && m[1] != "" {
			if !cl.cfg.Keywords[m[1]] {
				sig := signals.NewSignal(signals.FunctionDefined, filePath).
					WithName(m[1]).
					WithLine(lineNum).
					WithLanguage(langID).
					WithConfidence(confidenceFor(cl.cfg.DefaultTier)).
					WithMeta("source", "regex").
					WithMeta("source_tier", tier)
				sigSet.Add(*sig)
				return true
			}
		}
	}
	return false
}

// extractCalls emits FunctionCalled signals.
func (a *Analyzer) extractCalls(
	line, filePath, langID string, lineNum int, tier string,
	cl *compiledLanguage, sigSet *signals.SignalSet,
) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}
	for _, cs := range cl.cfg.CommentStyles {
		if cs.LinePrefix != "" && strings.HasPrefix(trimmed, cs.LinePrefix) {
			return
		}
	}

	seen := make(map[string]bool)
	for _, re := range cl.callPatterns {
		all := re.FindAllStringSubmatch(line, -1)
		for _, m := range all {
			if len(m) < 2 || m[1] == "" {
				continue
			}
			name := m[1]
			if cl.cfg.Keywords[name] || seen[name] {
				continue
			}
			seen[name] = true
			sig := signals.NewSignal(signals.FunctionCalled, filePath).
				WithName(name).
				WithLine(lineNum).
				WithLanguage(langID).
				WithConfidence(confidenceFor(cl.cfg.DefaultTier)).
				WithMeta("source", "regex").
				WithMeta("source_tier", tier)
			sigSet.Add(*sig)
		}
	}
}

// extractAnnotations combines todo and intent marker extraction
// into a single pass.
func (a *Analyzer) extractAnnotations(
	line, filePath, langID string, lineNum int,
	cl *compiledLanguage, sigSet *signals.SignalSet,
) {
	if cl.todoPattern != nil {
		if m := cl.todoPattern.FindStringSubmatch(line); m != nil && len(m) >= 3 {
			tier := tierLabel(cl.cfg.DefaultTier)
			sig := signals.NewSignal(signals.TodoFound, filePath).
				WithLine(lineNum).
				WithLanguage(langID).
				WithMeta("source", "regex").
				WithMeta("source_tier", tier).
				WithMeta("type", strings.ToUpper(m[1])).
				WithMeta("text", strings.TrimSpace(m[2]))
			sigSet.Add(*sig)
		}
	}

}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func tierLabel(t language.Tier) string {
	switch t {
	case language.TierAST:
		return "ast"
	case language.TierLexical:
		return "lexical"
	default:
		return "regex"
	}
}

func confidenceFor(t language.Tier) float64 {
	switch t {
	case language.TierAST:
		return 1.0
	case language.TierLexical:
		return 0.8
	default:
		return 0.6
	}
}
