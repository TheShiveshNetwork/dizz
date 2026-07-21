package discover

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/TheShiveshNetwork/dizz/internal/language"
)

var (
	knownExtsOnce sync.Once
	knownExts     map[string]bool
)

func extensionSet() map[string]bool {
	knownExtsOnce.Do(func() {
		exts := language.AllExtensions()
		knownExts = make(map[string]bool, len(exts))
		for _, ext := range exts {
			knownExts[ext] = true
		}
	})
	return knownExts
}

// Discover all relevant files in a directory tree
func Files(root string, include, exclude []string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
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

// CodeFiles returns all source files under root that belong to any language
// registered in the language registry, honoring the given exclude patterns.
//
// The include patterns in the project config are respected: if the caller
// passes non-empty customIncludes they are used instead of the registry-derived
// extension list.  Pass nil or an empty slice to use the auto-generated list.
func CodeFiles(root string, exclude []string) ([]string, error) {
	return CodeFilesWithIncludes(root, nil, exclude)
}

// CodeFilesWithIncludes is like CodeFiles but also accepts explicit include
// glob patterns (e.g. from config.Include).  When customIncludes is non-empty
// it is used verbatim; otherwise the extension list is derived from the
// language registry.
func CodeFilesWithIncludes(root string, customIncludes, exclude []string) ([]string, error) {
	var includePatterns []string

	if len(customIncludes) > 0 {
		// Use caller-supplied patterns directly (e.g. "**/*", "src/**/*.ts")
		includePatterns = customIncludes
	} else {
		// Build from the language registry so new language entries are picked
		// up automatically without touching this file.
		exts := language.AllExtensions()
		for _, ext := range exts {
			includePatterns = append(includePatterns, "**/*"+ext)
		}
	}

	return Files(root, includePatterns, exclude)
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

		// Handle **/middle/** pattern (e.g. **/node_modules/**)
		if len(parts) >= 3 && parts[0] == "" && parts[len(parts)-1] == "" {
			slashed := "/" + path + "/"
			for i := 1; i < len(parts)-1; i++ {
				middle := strings.Trim(parts[i], "/")
				if middle != "" && !strings.Contains(slashed, "/"+middle+"/") {
					return false
				}
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

// patternsAreExtensionBased returns true when ALL patterns look like the
// auto-generated set from the language registry (**/*.go, **/*.ts, etc.)
// AND every referenced extension is a known language extension.
func patternsAreExtensionBased(patterns []string) bool {
	if len(patterns) < 2 {
		return false
	}
	exts := extensionSet()
	for _, p := range patterns {
		if !strings.HasPrefix(p, "**/*.") {
			return false
		}
		ext := p[4:] // strip "**/*."
		if !exts[ext] {
			return false
		}
	}
	return true
}

func shouldInclude(path, root string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	relPath, _ := filepath.Rel(root, path)

	// Fast path: auto-generated patterns are all **/*.ext. Check the file
	// extension against known language extensions in O(1) instead of
	// looping over all N patterns.
	if patternsAreExtensionBased(patterns) {
		if ext := filepath.Ext(relPath); ext != "" && extensionSet()[ext] {
			return true
		}
		// No need to loop through patterns — if the extension doesn't match
		// any known language extension, no **/*.ext pattern can match.
		return false
	}

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
