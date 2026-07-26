package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// Analyzer implements the analyzer interface for Go
type Analyzer struct{}

// Language returns "go"
func (a *Analyzer) Language() string {
	return "go"
}

// SupportedExtensions returns Go file extensions.
func (a *Analyzer) SupportedExtensions() []string {
	return []string{".go"}
}

// @dizz-ignore-unused
// Supports checks if file is a Go file
func (a *Analyzer) Supports(file string) bool {
	return IsGoFile(file)
}

func (a *Analyzer) Analyze(files []string) (*signals.SignalSet, error) {
	sigSet := &signals.SignalSet{}
	fset := token.NewFileSet()

	for _, filePath := range files {
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Extract signals in a single pass
		a.inspectFile(file, filePath, fset, sigSet)

		// These are relatively fast as they iterate over fixed lists
		a.extractImports(file, filePath, fset, sigSet)
		a.extractTodos(file, filePath, fset, sigSet)
		a.extractIntents(file, filePath, sigSet)
	}

	return sigSet, nil
}

// AnalyzeFile extracts signals from a single Go file.
func (a *Analyzer) AnalyzeFile(file string) ([]signals.Signal, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, nil
	}
	return a.AnalyzeFileContent(file, content)
}

// AnalyzeFileContent extracts signals from a single Go file using pre-read content.
func (a *Analyzer) AnalyzeFileContent(file string, content []byte) ([]signals.Signal, error) {
	sigSet := &signals.SignalSet{}
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, content, parser.ParseComments)
	if err != nil {
		return nil, nil
	}

	a.inspectFile(f, file, fset, sigSet)
	a.extractImports(f, file, fset, sigSet)
	a.extractTodos(f, file, fset, sigSet)
	a.extractIntents(f, file, sigSet)

	return sigSet.Signals, nil
}

// inspectFile performs a single pass over the AST to extract definitions and calls
func (a *Analyzer) inspectFile(file *ast.File, filePath string, fset *token.FileSet, sigSet *signals.SignalSet) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				start := fset.Position(node.Pos())
				end := fset.Position(node.End())

				sig := signals.NewSignal(signals.FunctionDefined, filePath).
					WithName(node.Name.Name).
					WithLanguage(a.Language()).
					WithRange(
						start.Line,
						start.Column,
						end.Line,
						end.Column,
					)

				// Add metadata
				if node.Recv != nil {
					sig.WithMeta("is_method", true)
				}
				if node.Body == nil {
					sig.WithMeta("is_interface", true)
				}

				sigSet.Add(*sig)
			}
		case *ast.CallExpr:
			var name string
			switch fun := node.Fun.(type) {
			case *ast.Ident:
				name = fun.Name
			case *ast.SelectorExpr:
				name = fun.Sel.Name
			}

			if name != "" {
				start := fset.Position(node.Pos())
				end := fset.Position(node.End())

				sig := signals.NewSignal(signals.FunctionCalled, filePath).
					WithName(name).
					WithLanguage(a.Language()).
					WithRange(
						start.Line,
						start.Column,
						end.Line,
						end.Column,
					)

				sigSet.Add(*sig)
			}
		}
		return true
	})
}

// extractImports finds import statements
func (a *Analyzer) extractImports(file *ast.File, filePath string, fset *token.FileSet, sigSet *signals.SignalSet) {
	for _, imp := range file.Imports {
		if imp.Path != nil {
			start := fset.Position(imp.Pos())
			end := fset.Position(imp.End())
			path := strings.Trim(imp.Path.Value, `"`)
			sig := signals.NewSignal(signals.ImportFound, filePath).
				WithName(path).
				WithLanguage(a.Language()).
				WithRange(
					start.Line,
					start.Column,
					end.Line,
					end.Column,
				)

			sigSet.Add(*sig)
		}
	}
}

var todoRegex = regexp.MustCompile(`(?i)(?://|/\*|\*)\s*(TODO|FIXME|XXX|HACK|NOTE):\s*(.*)`)

func (a *Analyzer) extractTodos(file *ast.File, filePath string, fset *token.FileSet, sigSet *signals.SignalSet) {
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text
			if matches := todoRegex.FindStringSubmatch(text); len(matches) > 1 {
				pos := fset.Position(comment.Pos())
				sig := signals.NewSignal(signals.TodoFound, filePath).
					WithLine(pos.Line).
					WithLanguage(a.Language()).
					WithMeta("type", strings.ToUpper(matches[1])).
					WithMeta("text", strings.TrimSpace(matches[2]))
				sigSet.Add(*sig)
			}
		}
	}
}

// extractIntents finds @dizz intent markers
func (a *Analyzer) extractIntents(file *ast.File, filePath string, sigSet *signals.SignalSet) {
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text

			// Look for @dizz:state or @dizz:feature markers
			if strings.Contains(text, "@dizz:state") {
				state := extractMarkerValue(text, "@dizz:state")
				sig := signals.NewSignal(signals.IntentMarker, filePath).
					WithLanguage(a.Language()).
					WithMeta("marker_type", "state").
					WithMeta("value", state)
				sigSet.Add(*sig)
			}

			if strings.Contains(text, "@dizz:feature") {
				feature := extractMarkerValue(text, "@dizz:feature")
				sig := signals.NewSignal(signals.IntentMarker, filePath).
					WithLanguage(a.Language()).
					WithMeta("marker_type", "feature").
					WithMeta("value", feature)
				sigSet.Add(*sig)
			}
		}
	}
}

// extractMarkerValue extracts the value from a marker comment
// e.g., "// @dizz:state planned" -> "planned"
func extractMarkerValue(text, marker string) string {
	idx := strings.Index(text, marker)
	if idx == -1 {
		return ""
	}

	after := strings.TrimSpace(text[idx+len(marker):])
	parts := strings.Fields(after)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// IsGoFile checks if a path is a Go source file
func IsGoFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return filepath.Ext(path) == ".go"
}
