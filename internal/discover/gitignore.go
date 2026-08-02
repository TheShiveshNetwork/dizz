package discover

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ParseGitignore parses the .gitignore file at the project root and returns
// root-relative glob patterns.
func ParseGitignore(projectRoot string) ([]string, error) {
	return readGitignorePatterns(filepath.Join(projectRoot, ".gitignore"), convertGitignorePattern)
}

// ParseGitignoreTree collects patterns from the root .gitignore plus every
// nested .gitignore under projectRoot. Patterns from nested files are anchored
// to their containing directory, matching git's behavior where a pattern in
// subdir/.gitignore applies only below subdir.
func ParseGitignoreTree(projectRoot string) ([]string, error) {
	rootPatterns, err := ParseGitignore(projectRoot)
	if err != nil {
		rootPatterns = nil // missing root gitignore is fine; nested may exist
	}
	patterns := append([]string(nil), rootPatterns...)

	_ = filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldExclude(path, projectRoot, patterns) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != ".gitignore" {
			return nil
		}
		dirRel, err := filepath.Rel(projectRoot, filepath.Dir(path))
		if err != nil || dirRel == "." {
			return nil
		}
		nested, err := parseNestedGitignore(projectRoot, dirRel)
		if err != nil {
			return nil
		}
		patterns = append(patterns, nested...)
		return nil
	})

	return patterns, nil
}

// parseNestedGitignore parses the .gitignore in dirRel (project-root-relative)
// and returns root-relative patterns.
func parseNestedGitignore(projectRoot, dirRel string) ([]string, error) {
	return readGitignorePatterns(filepath.Join(projectRoot, dirRel, ".gitignore"),
		func(line string) string { return convertNestedGitignorePattern(dirRel, line) })
}

func readGitignorePatterns(path string, convert func(string) string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "!") {
			continue
		}
		if p := convert(line); p != "" {
			patterns = append(patterns, p)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

// convertNestedGitignorePattern converts a single pattern line from a nested
// .gitignore (located at dirRel, project-root-relative) into a
// project-root-relative glob pattern.
func convertNestedGitignorePattern(dirRel, pattern string) string {
	pattern = strings.TrimRight(pattern, " \t")
	if pattern == "" {
		return ""
	}

	isDir := strings.HasSuffix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")
	anchored := strings.HasPrefix(pattern, "/")
	pattern = strings.TrimPrefix(pattern, "/")
	if pattern == "" {
		return ""
	}

	// A bare name (no slash) matches at any depth below dirRel.
	if !anchored && !strings.Contains(pattern, "/") {
		if isDir {
			return dirRel + "/**/" + pattern + "/**"
		}
		return dirRel + "/**/" + pattern
	}

	// Anchored (leading /) or slash-containing patterns are relative to dirRel.
	if isDir {
		return dirRel + "/" + pattern + "/**"
	}
	return dirRel + "/" + pattern
}

func convertGitignorePattern(pattern string) string {
	pattern = strings.TrimRight(pattern, " \t")
	if pattern == "" {
		return ""
	}
	pattern = strings.TrimPrefix(pattern, "/")
	if pattern == "" {
		return ""
	}

	isDir := strings.HasSuffix(pattern, "/")
	if isDir {
		pattern = pattern[:len(pattern)-1]
	}

	if !strings.Contains(pattern, "/") {
		pattern = "**/" + pattern
	}

	if isDir {
		pattern += "/**"
	}

	return pattern
}
