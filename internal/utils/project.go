package utils

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	projectRoot     string
	projectRootOnce sync.Once
)

func getProjectRoot() string {
	projectRootOnce.Do(func() {
		cwd, err := os.Getwd()
		if err != nil {
			return
		}
		dir := cwd
		for {
			if _, err := os.Stat(filepath.Join(dir, ".dizz")); err == nil {
				projectRoot = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				return
			}
			dir = parent
		}
	})
	return projectRoot
}

func RelPath(absPath string) string {
	root := getProjectRoot()
	if root == "" {
		return absPath
	}
	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return absPath
	}
	return rel
}
