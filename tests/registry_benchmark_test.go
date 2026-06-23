package benchmarks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/discover"
	"github.com/TheShiveshNetwork/dizz/internal/language"
)

// ──────────────────────────────────────────────────────────────────────────────
// Language registry benchmarks
// ──────────────────────────────────────────────────────────────────────────────

func BenchmarkDetectByExtension(b *testing.B) {
	exts := []string{
		".go", ".js", ".ts", ".py", ".rs", ".java", ".rb", ".php", ".c", ".cpp",
		".kt", ".swift", ".cs", ".scala", ".lua", ".sh", ".hs", ".ex", ".jl",
		".dart", ".pl", ".nim", ".zig", ".clj", ".erl", ".ml", ".fs", ".sql",
		".tf",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ext := range exts {
			language.Detect("some/path/to/file" + ext)
		}
	}
}

func BenchmarkGetAllLanguages(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = language.All()
	}
}

func BenchmarkGetLanguageByID(b *testing.B) {
	ids := []string{"go", "python", "rust", "javascript", "typescript", "java", "csharp", "cpp"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, id := range ids {
			language.Get(id)
		}
	}
}

func BenchmarkAllExtensions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = language.AllExtensions()
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// File discovery benchmarks
// ──────────────────────────────────────────────────────────────────────────────

func BenchmarkCodeFiles(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkProject(b, tmpDir, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := discover.CodeFiles(tmpDir, []string{".git/**"})
		if err != nil {
			b.Fatalf("CodeFiles failed: %v", err)
		}
		if len(files) == 0 {
			b.Fatal("no files discovered")
		}
	}
}

func BenchmarkCodeFilesWithExcludes(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkProject(b, tmpDir, 100)

	excludes := []string{
		".git/**",
		"vendor/**",
		"node_modules/**",
		"dist/**",
		"build/**",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := discover.CodeFiles(tmpDir, excludes)
		if err != nil {
			b.Fatalf("CodeFiles failed: %v", err)
		}
		if len(files) == 0 {
			b.Fatal("no files discovered")
		}
	}
}

func BenchmarkCodeFilesLargeProject(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkProject(b, tmpDir, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := discover.CodeFiles(tmpDir, []string{".git/**"})
		if err != nil {
			b.Fatalf("CodeFiles failed: %v", err)
		}
		if len(files) == 0 {
			b.Fatal("no files discovered")
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Shebang detection benchmarks
// ──────────────────────────────────────────────────────────────────────────────

func BenchmarkDetectByShebang(b *testing.B) {
	tmpDir := b.TempDir()
	script := filepath.Join(tmpDir, "runner")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env python3\nprint('hello')\n"), 0755); err != nil {
		b.Fatalf("write script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		language.Detect(script)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Helper
// ──────────────────────────────────────────────────────────────────────────────

// setupBenchmarkProject creates a temporary project directory with n source
// files spread across multiple language extensions and directory depths.
func setupBenchmarkProject(b *testing.B, root string, n int) {
	b.Helper()

	exts := []string{".go", ".js", ".ts", ".py", ".rs", ".java", ".rb", ".php", ".c", ".cpp"}
	dirs := []string{"src", "lib", "cmd", "pkg", "internal"}

	for i := 0; i < n; i++ {
		dir := filepath.Join(root, dirs[i%len(dirs)])
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatalf("mkdir: %v", err)
		}
		f := filepath.Join(dir, "file_"+itoa(i)+exts[i%len(exts)])
		content := "package main\nfunc fn_" + itoa(i) + "() {}\n"
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			b.Fatalf("write file: %v", err)
		}
	}
}
