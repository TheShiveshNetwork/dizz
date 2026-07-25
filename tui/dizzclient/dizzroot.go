package dizzclient

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	dizzRootCache string
	dizzRootOnce  sync.Once
)

func FindDizzRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		dizzDir := filepath.Join(dir, ".dizz")
		if _, err := os.Stat(dizzDir); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .dizz directory found")
		}
		dir = parent
	}
}

func getDizzRoot() string {
	dizzRootOnce.Do(func() {
		root, err := FindDizzRoot()
		if err == nil {
			dizzRootCache = root
		}
	})
	return dizzRootCache
}

func RelPath(absPath string) string {
	root := getDizzRoot()
	if root == "" {
		return absPath
	}
	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return absPath
	}
	return rel
}
