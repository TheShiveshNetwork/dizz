package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Todo struct {
	File		string
	LineNum	int
	Line		string
}

func FindTodos(filepath string) []Todo {
	f, err := os.Open(filepath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var todos []Todo
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.Contains(line, "TODO") ||
				strings.Contains(line, "FIXME") ||
				strings.Contains(line, "@intent:planned") {
			todos = append(
				todos,
				Todo{
					File: filepath,
					LineNum: lineNum,
					Line: strings.TrimSpace(line),
				},
			)
		}
	}

	return todos
}

func FindAllTodos(root string) []Todo {
	var allTodos []Todo

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		// Skip vendor, .prog, and non-code files
		if strings.Contains(path, "vendor/") || 
		   strings.Contains(path, ".prog/") ||
		   strings.Contains(path, "node_modules/") {
			return nil
		}

		// Only scan code files
		// TODO: scan other files including markdown and all that might be relevant
		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".js" || ext == ".ts" || 
		   ext == ".py" || ext == ".java" || ext == ".c" || 
		   ext == ".cpp" || ext == ".rs" {
			todos := FindTodos(path)
			
			// Convert to relative paths
			for i := range todos {
				relPath, _ := filepath.Rel(root, todos[i].File)
				todos[i].File = relPath
			}
			
			allTodos = append(allTodos, todos...)
		}

		return nil
	})

	return allTodos
}

