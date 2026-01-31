package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// find all the function definitions and calls
// TODO: make this analyzer language agnostic
func AnalyzeFunctions(root string) (defined map[string]string, called map[string]bool) {
	defined = make(map[string]string)
	called = make(map[string]bool)

	fset := token.NewFileSet()

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and **non-go** files
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		// Parse the file
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		// Inspect AST nodes
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				// Found a function definition
				if x.Name != nil {
					funcName := x.Name.Name
					// Store with relative path
					relPath, _ := filepath.Rel(root, path)
					defined[funcName] = relPath
				}

			case *ast.CallExpr:
				// Found a function call
				switch fun := x.Fun.(type) {
				case *ast.Ident:
					// Direct call: foo()
					called[fun.Name] = true
				case *ast.SelectorExpr:
					// Method call: obj.Method()
					if ident, ok := fun.X.(*ast.Ident); ok {
						called[ident.Name] = true
					}
					called[fun.Sel.Name] = true
				}
			}
			return true
		})

		return nil
	})

	return
}

// @returns: the source code of a specific function
// TODO: Inplement for deeper analysis
func GetFunctionBody(file string, funcName string) (string, error) {
	return "", nil
}

