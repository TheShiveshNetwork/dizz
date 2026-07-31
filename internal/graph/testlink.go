package graph

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// IsTestFile detects test files by filename convention across the common
// language families. Pure string matching — no source scanning.
func IsTestFile(relPath string) bool {
	base := filepath.Base(relPath)
	lower := strings.ToLower(base)
	switch {
	case strings.HasSuffix(lower, "_test.go"), strings.HasSuffix(lower, "_test.py"),
		strings.HasSuffix(lower, "_test.rb"), strings.HasSuffix(lower, "_spec.rb"),
		strings.HasSuffix(lower, "_test.rs"), strings.HasSuffix(lower, "_spec.rs"):
		return true
	case strings.HasSuffix(lower, ".test.ts"), strings.HasSuffix(lower, ".test.tsx"),
		strings.HasSuffix(lower, ".test.js"), strings.HasSuffix(lower, ".test.jsx"),
		strings.HasSuffix(lower, ".spec.ts"), strings.HasSuffix(lower, ".spec.tsx"),
		strings.HasSuffix(lower, ".spec.js"), strings.HasSuffix(lower, ".spec.jsx"),
		strings.HasSuffix(lower, ".test.go"):
		return true
	case strings.HasPrefix(lower, "test_"), strings.HasPrefix(lower, "spec_"),
		strings.HasPrefix(lower, "test-"):
		return true
	}
	return false
}

// testCandidates derives the set of production symbol names a test file could
// target from its own name and from the test functions defined inside it
// (which come from the cached signals, never from a fresh scan).
func testCandidates(testFile string, testFnNames []string) []string {
	seen := make(map[string]bool)
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}

	base := filepath.Base(testFile)
	stem := trimExt(base)
	stem = strings.TrimSuffix(stem, ".test")
	stem = strings.TrimSuffix(stem, ".spec")
	if strings.HasPrefix(stem, "test_") {
		stem = strings.TrimPrefix(stem, "test_")
	}
	if strings.HasPrefix(stem, "spec_") {
		stem = strings.TrimPrefix(stem, "spec_")
	}
	if strings.HasSuffix(stem, "_test") {
		stem = strings.TrimSuffix(stem, "_test")
	}
	if strings.HasSuffix(stem, "_spec") {
		stem = strings.TrimSuffix(stem, "_spec")
	}
	add(stem)
	add(strings.ToLower(stem))
	if stem != "" {
		add(exported(stem))
	}

	for _, fn := range testFnNames {
		add(fn)
		lower := strings.ToLower(fn)
		switch {
		case strings.HasPrefix(lower, "test_") && len(fn) > len("test_"):
			add(fn[len("test_"):])
		case strings.HasPrefix(fn, "Test") && len(fn) > 4:
			add(fn[4:])
		case strings.HasSuffix(lower, "_test") && len(fn) > len("_test"):
			add(strings.TrimSuffix(fn, "_test"))
		case strings.HasSuffix(fn, "Test") && len(fn) > 4:
			add(strings.TrimSuffix(fn, "Test"))
		}
	}
	return out
}

func exported(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// LinkTests adds test nodes and TESTS edges for every test file found in the
// cached signals. Two methods are used, in priority order:
//
//  1. Naming convention (confidence 0.9): test function names (TestLogin,
//     test_login, login_test) and the test file stem resolve to production
//     symbol names.
//  2. Import proximity (confidence 0.6): the test file imports a production
//     file, so every symbol in that file is treated as covered.
func (g *Graph) LinkTests(resolver *ImportResolver, cachedSignals map[string][]signals.Signal) {
	testFiles := make(map[string]bool)
	for rel := range cachedSignals {
		if IsTestFile(rel) {
			testFiles[rel] = true
		}
	}
	for _, rel := range sortedKeys(testFiles) {
		g.linkTestFile(rel, cachedSignals[rel], resolver)
	}
}

func (g *Graph) linkTestFile(rel string, sigs []signals.Signal, resolver *ImportResolver) {
	id := TestID(rel)
	g.AddNode(&Node{
		ID:    id,
		Type:  NodeTest,
		Label: filepath.Base(rel),
		Attrs: map[string]interface{}{"path": rel},
		Rationale: Rationale{
			Confidence: 1.0,
			SourceType: "static_analysis",
		},
	})

	var fnNames []string
	importedFiles := make(map[string]bool)
	for _, sig := range sigs {
		switch sig.Type {
		case signals.FunctionDefined:
			fnNames = append(fnNames, sig.Name)
		case signals.ImportFound:
			if resolver == nil {
				continue
			}
			fromRel := relPath(sig.File)
			if resolved, ok := resolver.Resolve(sig.Name, fromRel); ok {
				importedFiles[resolved] = true
			}
		}
	}

	// Method 1: naming convention.
	candidates := testCandidates(rel, fnNames)
	seen := make(map[string]bool)
	for _, name := range candidates {
		for _, sym := range g.SymbolsNamed(name) {
			if seen[sym.ID] {
				continue
			}
			seen[sym.ID] = true
			g.AddEdge(&Edge{
				Type:   EdgeTests,
				From:   id,
				To:     sym.ID,
				Weight: 0.9,
				Attrs: map[string]string{
					"match_method": "naming",
					"name":         name,
				},
				Rationale: Rationale{
					SourceTier: "lexical",
					Confidence: 0.9,
					Evidence:   "test name " + name,
					SourceType: "static_analysis",
				},
			})
		}
	}

	// Method 2: import proximity.
	for file := range importedFiles {
		for _, sym := range g.SymbolsInFile(file) {
			if seen[sym.ID] {
				continue
			}
			seen[sym.ID] = true
			g.AddEdge(&Edge{
				Type:   EdgeTests,
				From:   id,
				To:     sym.ID,
				Weight: 0.6,
				Attrs: map[string]string{
					"match_method": "imports",
					"imports":      file,
				},
				Rationale: Rationale{
					SourceTier: "lexical",
					Confidence: 0.6,
					Evidence:   "test imports " + file,
					SourceType: "static_analysis",
				},
			})
		}
	}
}

// relPath normalizes a path to forward slashes. Callers are expected to pass
// project-relative paths.
func relPath(p string) string {
	return filepath.ToSlash(p)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
