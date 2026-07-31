package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindProjectRoot_RequiresConfig(t *testing.T) {
	root := t.TempDir()

	t.Run("stray .dizz without config errors", func(t *testing.T) {
		stray := filepath.Join(root, "stray")
		if err := os.MkdirAll(filepath.Join(stray, ".dizz"), 0755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(stray)

		got, err := FindProjectRoot()
		if err == nil {
			t.Fatalf("expected error for .dizz without config.json, got root %q", got)
		}
		if !strings.Contains(err.Error(), "config.json") {
			t.Fatalf("error should mention config.json, got: %v", err)
		}
	})

	t.Run("valid project root detected", func(t *testing.T) {
		validRoot := filepath.Join(root, "valid")
		trackDir := filepath.Join(validRoot, ".dizz")
		if err := os.MkdirAll(trackDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(trackDir, "config.json"), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
		t.Chdir(validRoot)

		got, err := FindProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != validRoot {
			t.Fatalf("expected root %q, got %q", validRoot, got)
		}
	})

	t.Run("subdirectory of valid project resolved", func(t *testing.T) {
		sub := filepath.Join(root, "valid", "cmd")
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(sub)

		got, err := FindProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != filepath.Join(root, "valid") {
			t.Fatalf("expected project root %q, got %q", filepath.Join(root, "valid"), got)
		}
	})

	t.Run("no .dizz anywhere errors", func(t *testing.T) {
		empty := filepath.Join(root, "empty")
		if err := os.MkdirAll(empty, 0755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(empty)

		if _, err := FindProjectRoot(); err == nil {
			t.Fatal("expected error when no .dizz directory exists")
		}
	})
}
