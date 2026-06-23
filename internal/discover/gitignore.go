package discover

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func ParseGitignore(projectRoot string) ([]string, error) {
	path := filepath.Join(projectRoot, ".gitignore")
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
		patterns = append(patterns, convertGitignorePattern(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
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
