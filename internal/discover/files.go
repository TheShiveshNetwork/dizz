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

// @dizz-ignore-unused
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

// matchPattern performs glob-like pattern matching. ** matches any number of
// path segments, * matches within a single segment, and a leading / anchors
// the pattern. Patterns without ** use the legacy wildcard semantics.
func matchPattern(path, pattern string) bool {
	if path == pattern {
		return true
	}
	if strings.Contains(pattern, "**") {
		return matchGlobSegments(strings.Split(path, "/"), strings.Split(pattern, "/"))
	}
	return matchSimplePattern(path, pattern)
}

// matchSimplePattern handles patterns without ** using the legacy behavior:
// a leading * matches a path suffix, a trailing * matches a path prefix.
func matchSimplePattern(path, pattern string) bool {
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

// matchGlobSegments recursively matches path segments against pattern
// segments, where a ** segment matches zero or more path segments.
func matchGlobSegments(wp, pp []string) bool {
	if len(pp) == 0 {
		return len(wp) == 0
	}
	if pp[0] == "**" {
		for i := 0; i <= len(wp); i++ {
			if matchGlobSegments(wp[i:], pp[1:]) {
				return true
			}
		}
		return false
	}
	if len(wp) == 0 {
		return false
	}
	if !matchSegment(wp[0], pp[0]) {
		return false
	}
	return matchGlobSegments(wp[1:], pp[1:])
}

// matchSegment matches one path segment against one pattern segment.
func matchSegment(w, p string) bool {
	if p == "*" {
		return true
	}
	if p == w {
		return true
	}
	if strings.HasPrefix(p, "*") {
		return strings.HasSuffix(w, strings.TrimPrefix(p, "*"))
	}
	if strings.HasSuffix(p, "*") {
		return strings.HasPrefix(w, strings.TrimSuffix(p, "*"))
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
