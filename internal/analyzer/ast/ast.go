package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// Analyzer implements the analyzer interface for Go
type Analyzer struct{}

// Language returns "go"
func (a *Analyzer) Language() string {
	return "go"
}

// Supports checks if file is a Go file
func (a *Analyzer) Supports(file string) bool {
	return filepath.Ext(file) == ".go" && !strings.HasSuffix(file, "_test.go")
}

func (a *Analyzer) Analyze(files []string) (*signals.SignalSet, error) {
	sigSet := &signals.SignalSet{}
	fset := token.NewFileSet()

	for _, filePath := range files {
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Extract signals
		a.extractDefinitions(file, filePath, fset, sigSet)
		a.extractCalls(file, filePath, fset, sigSet)
		a.extractImports(file, filePath, fset, sigSet)
		a.extractTodos(file, filePath, fset, sigSet)
		a.extractIntents(file, filePath, sigSet)
	}

	return sigSet, nil
}

func (a *Analyzer) extractDefinitions(file *ast.File, filePath string, fset *token.FileSet, sigSet *signals.SignalSet) {
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name != nil {
			start := fset.Position(fn.Pos())
			end := fset.Position(fn.End())

			sig := signals.NewSignal(signals.FunctionDefined, filePath).
				WithName(fn.Name.Name).
				WithLanguage("go").
				WithRange(
					start.Line,
					start.Column,
					end.Line,
					end.Column,
				)

			// Add metadata
			if fn.Recv != nil {
				sig.WithMeta("is_method", true)
			}
			if fn.Body == nil {
				sig.WithMeta("is_interface", true)
			}

			sigSet.Add(*sig)
		}
		return true
	})
}

// extractCalls finds function calls
func (a *Analyzer) extractCalls(file *ast.File, filePath string, fset *token.FileSet, sigSet *signals.SignalSet) {
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			var name string

			switch fun := call.Fun.(type) {
			case *ast.Ident:
				name = fun.Name
			case *ast.SelectorExpr:
				name = fun.Sel.Name
			}

			start := fset.Position(call.Pos())
			end := fset.Position(call.End())

			if name != "" {
				sig := signals.NewSignal(signals.FunctionCalled, filePath).
					WithName(name).
					WithLanguage("go").
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
				WithLanguage("go").
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

func (a *Analyzer) extractTodos(file *ast.File, filePath string, fset *token.FileSet, sigSet *signals.SignalSet) {
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text
			if strings.Contains(text, "TODO") || strings.Contains(text, "FIXME") {
				pos := fset.Position(comment.Pos())
				sig := signals.NewSignal(signals.TodoFound, filePath).
					WithLine(pos.Line).
					WithLanguage("go").
					WithMeta("text", strings.TrimSpace(text))
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
					WithLanguage("go").
					WithMeta("marker_type", "state").
					WithMeta("value", state)
				sigSet.Add(*sig)
			}

			if strings.Contains(text, "@dizz:feature") {
				feature := extractMarkerValue(text, "@dizz:feature")
				sig := signals.NewSignal(signals.IntentMarker, filePath).
					WithLanguage("go").
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

	return filepath.Ext(path) == ".go" && !strings.HasSuffix(path, "_test.go")
}
