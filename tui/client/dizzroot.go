package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	dizzRootCache string
	dizzRootOnce  sync.Once
	dizzBinCache  string
	dizzBinOnce   sync.Once
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

func FindDizzBinary() string {
	dizzBinOnce.Do(func() {
		if root := getDizzRoot(); root != "" {
			localBin := filepath.Join(root, "bin", "dizz")
			if _, err := os.Stat(localBin); err == nil {
				dizzBinCache = localBin
				return
			}
		}
		if path, err := exec.LookPath("dizz"); err == nil {
			dizzBinCache = path
			return
		}
		dizzBinCache = "dizz"
	})
	return dizzBinCache
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
