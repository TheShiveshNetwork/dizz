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
	cfg            language.LanguageConfig
	fnPatterns     []*regexp.Regexp
	callPatterns   []*regexp.Regexp
	importPatterns []*regexp.Regexp
	todoPattern    *regexp.Regexp
	intentState    *regexp.Regexp
	intentFeat     *regexp.Regexp
}

// Analyzer uses the language registry to extract signals from any supported
// source file.  It is intentionally language-agnostic: all behaviour comes from
// the LanguageConfig for the detected language.
type Analyzer struct {
	compiled map[string]*compiledLanguage // keyed by language ID
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
	for _, p := range lc.CallPatterns {
		if re, err := regexp.Compile(p); err == nil {
			cl.callPatterns = append(cl.callPatterns, re)
		}
	}
	for _, p := range lc.ImportPatterns {
		if re, err := regexp.Compile(p); err == nil {
			cl.importPatterns = append(cl.importPatterns, re)
		}
	}

	// Build a TODO pattern that matches any of the language's comment prefixes.
	cl.todoPattern = buildTodoPattern(lc)
	cl.intentState = regexp.MustCompile(`@dizz:state\s+(\w+)`)
	cl.intentFeat = regexp.MustCompile(`@dizz:feature\s+(\w+)`)
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
	}
	if len(prefixes) == 0 {
		// Fallback: accept // or # if no prefix defined
		prefixes = []string{`//`, `#`}
	}
	joined := strings.Join(prefixes, "|")
	// \*+ handles single-star (* TODO:) and double-star (** TODO:) block
	// comment continuation lines.
	return regexp.MustCompile(`(?i)^\s*(?:` + joined + `|\*+|/\*)\s*(TODO|FIXME|XXX|HACK|NOTE):\s*(.+)`)
}

// Language returns a descriptive identifier for this analyzer.
func (a *Analyzer) Language() string { return "regex" }

// Supports returns true for every file whose language is in the registry and is
// not Go (Go is handled by the AST analyzer with higher accuracy).
func (a *Analyzer) Supports(file string) bool {
	lc, ok := language.Detect(file)
	if !ok {
		return false
	}
	return lc.ID != "go"
}

// Analyze extracts signals for all given files.
func (a *Analyzer) Analyze(files []string) (*signals.SignalSet, error) {
	sigSet := &signals.SignalSet{}
	for _, filePath := range files {
		_ = a.analyzeFile(filePath, sigSet) // skip individual file errors
	}
	return sigSet, nil
}

// analyzeFile processes a single file line-by-line.
func (a *Analyzer) analyzeFile(filePath string, sigSet *signals.SignalSet) error {
	lc, ok := language.Detect(filePath)
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

	// Determine accuracy tier label for metadata.
	tier := tierLabel(lc.DefaultTier)

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check whether this line is a comment line (for TODO detection).
		isComment := isCommentLine(line, commentPrefixes(cl.cfg))

		a.extractFunctions(line, filePath, lc.Name, lineNum, tier, cl, sigSet)
		a.extractCalls(line, filePath, lc.Name, lineNum, tier, cl, sigSet)
		a.extractImports(line, filePath, lc.Name, lineNum, cl, sigSet)
		a.extractTodos(line, filePath, lc.Name, lineNum, cl, isComment, sigSet)
		a.extractIntents(line, filePath, lc.Name, lineNum, cl, sigSet)
	}

	return scanner.Err()
}

// commentPrefixes returns the line comment prefixes for a LanguageConfig.
func commentPrefixes(lc language.LanguageConfig) []string {
	var out []string
	for _, cs := range lc.CommentStyles {
		if cs.LinePrefix != "" {
			out = append(out, cs.LinePrefix)
		}
		if cs.BlockStart != "" {
			out = append(out, cs.BlockStart)
		}
	}
	return out
}

// isCommentLine checks if a line starts with a known comment prefix.
// For languages with /* */ comments, it also treats "* " and "*\t"
// block-comment continuation lines as comment lines.
func isCommentLine(line string, prefixes []string) bool {
	trimmed := strings.TrimSpace(line)
	hasBlockCommentPrefix := false
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
		if prefix == "/*" {
			hasBlockCommentPrefix = true
		}
	}

	if !hasBlockCommentPrefix || !strings.HasPrefix(trimmed, "*") {
		return false
	}
	if len(trimmed) == 1 {
		return true
	}
	if len(trimmed) > 2 && trimmed[1] == '*' {
		return trimmed[2] == ' ' || trimmed[2] == '\t'
	}

	return trimmed[1] == ' ' || trimmed[1] == '\t'
}

// extractFunctions emits FunctionDefined signals.
func (a *Analyzer) extractFunctions(
	line, filePath, langName string, lineNum int, tier string,
	cl *compiledLanguage, sigSet *signals.SignalSet,
) {
	for _, re := range cl.fnPatterns {
		if m := re.FindStringSubmatch(line); m != nil && len(m) > 1 && m[1] != "" {
			if !cl.cfg.Keywords[m[1]] {
				sig := signals.NewSignal(signals.FunctionDefined, filePath).
					WithName(m[1]).
					WithLine(lineNum).
					WithLanguage(langName).
					WithConfidence(confidenceFor(cl.cfg.DefaultTier)).
					WithMeta("source", "regex").
					WithMeta("source_tier", tier)
				sigSet.Add(*sig)
				return // one definition per line
			}
		}
	}
}

// extractCalls emits FunctionCalled signals.
// To avoid treating definition headers as calls, we skip lines that already
// matched a function definition pattern.
func (a *Analyzer) extractCalls(
	line, filePath, langName string, lineNum int, tier string,
	cl *compiledLanguage, sigSet *signals.SignalSet,
) {
	// Skip blank lines and pure comment lines — no call sites there.
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}
	for _, cs := range cl.cfg.CommentStyles {
		if cs.LinePrefix != "" && strings.HasPrefix(trimmed, cs.LinePrefix) {
			return
		}
	}

	// If the line is a function definition, skip it so we do not emit a
	// spurious FunctionCalled for the definition itself.
	for _, re := range cl.fnPatterns {
		if re.MatchString(line) {
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
				WithLanguage(langName).
				WithConfidence(confidenceFor(cl.cfg.DefaultTier)).
				WithMeta("source", "regex").
				WithMeta("source_tier", tier)
			sigSet.Add(*sig)
		}
	}
}

// extractImports emits ImportFound signals.
func (a *Analyzer) extractImports(
	line, filePath, langName string, lineNum int,
	cl *compiledLanguage, sigSet *signals.SignalSet,
) {
	for _, re := range cl.importPatterns {
		if m := re.FindStringSubmatch(line); m != nil && len(m) > 1 {
			sig := signals.NewSignal(signals.ImportFound, filePath).
				WithName(m[1]).
				WithLine(lineNum).
				WithLanguage(langName)
			sigSet.Add(*sig)
			return // one import per line
		}
	}
}

// extractTodos emits TodoFound signals — only for comment lines.
func (a *Analyzer) extractTodos(
	line, filePath, langName string, lineNum int,
	cl *compiledLanguage, isComment bool, sigSet *signals.SignalSet,
) {
	if !isComment || cl.todoPattern == nil {
		return
	}
	m := cl.todoPattern.FindStringSubmatch(line)
	if m == nil || len(m) < 3 {
		return
	}
	sig := signals.NewSignal(signals.TodoFound, filePath).
		WithLine(lineNum).
		WithLanguage(langName).
		WithMeta("type", strings.ToUpper(m[1])).
		WithMeta("text", strings.TrimSpace(m[2]))
	sigSet.Add(*sig)
}

// extractIntents emits IntentMarker signals for @dizz: annotations.
func (a *Analyzer) extractIntents(
	line, filePath, langName string, lineNum int,
	cl *compiledLanguage, sigSet *signals.SignalSet,
) {
	if m := cl.intentState.FindStringSubmatch(line); m != nil && len(m) > 1 {
		sig := signals.NewSignal(signals.IntentMarker, filePath).
			WithLine(lineNum).
			WithLanguage(langName).
			WithMeta("marker_type", "state").
			WithMeta("value", m[1])
		sigSet.Add(*sig)
	}
	if m := cl.intentFeat.FindStringSubmatch(line); m != nil && len(m) > 1 {
		sig := signals.NewSignal(signals.IntentMarker, filePath).
			WithLine(lineNum).
			WithLanguage(langName).
			WithMeta("marker_type", "feature").
			WithMeta("value", m[1])
		sigSet.Add(*sig)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

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
