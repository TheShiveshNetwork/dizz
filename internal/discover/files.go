package discover

import (
	"os"
	"path/filepath"
	"strings"
)

// Discover all relevant files in a directory tree
func Files(root string, include, exclude []string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// exclude dirs
		if info.IsDir() {
			if shouldExclude(path, root, exclude) {
				return filepath.SkipDir
			}
			return nil
		}

		// exclude files
		if shouldExclude(path, root, exclude) {
			return nil
		}

		// check file matches using include patterns
		if shouldInclude(path, root, include) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// return common code file extensions
func CodeFiles(root string, exclude []string) ([]string, error) {
	codeExtensions := []string{
		"**/*.go",
		"**/*.js", "**/*.ts", "**/*.jsx", "**/*.tsx",
		"**/*.py",
		"**/*.rs",
		"**/*.c", "**/*.cpp", "**/*.h", "**/*.hpp",
		"**/*.java",
		"**/*.rb",
		"**/*.php",
	}
	
	return Files(root, codeExtensions, exclude)
}

// matchPattern performs simple glob-like pattern matching
func matchPattern(path, pattern string) bool {
	if path == pattern {
		return true
	}

	// Convert glob pattern to simple matching
	// ** means any directory depth
	// * means any characters in segment
	
	// Handle ** (recursive)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")
			
			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}
			
			if suffix != "" && !strings.HasSuffix(path, suffix) {
				// Check if suffix is an extension pattern
				if strings.HasPrefix(suffix, "*") {
					ext := strings.TrimPrefix(suffix, "*")
					return strings.HasSuffix(path, ext)
				}
				return false
			}
			
			return true
		}
	}
	
	// Handle simple wildcard
	if strings.Contains(pattern, "*") {
		// Simple extension matching
		if strings.HasPrefix(pattern, "*") {
			ext := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(path, ext)
		}
		
		// Simple prefix matching
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			return strings.HasPrefix(path, prefix)
		}
	}
	
	return false
}

func shouldInclude(path, root string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	relPath, _ := filepath.Rel(root, path)

	for _, pattern := range patterns {
		if matchPattern(relPath, pattern) {
			return true
		}
	}
	
	return false
}

func shouldExclude(path, root string, patterns []string) bool {
	relPath, _ := filepath.Rel(root, path)
	
	for _, pattern := range patterns {
		if matchPattern(relPath, pattern) {
			return true
		}
	}
	
	return false
}

