package benchmarks

// language_conformance_test.go — golden-fixture tests that verify signal
// extraction works correctly for every major language supported by the
// language registry.  These tests do NOT require a git repository.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/analyzer/regex"
	"github.com/TheShiveshNetwork/dizz/internal/discover"
	"github.com/TheShiveshNetwork/dizz/internal/language"
	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// writeTemp creates a temp file with the given content and extension, returning
// its path.  The caller is responsible for removing it.
func writeTemp(t *testing.T, ext, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "dizz_test_*"+ext)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

func definedNames(sigSet *signals.SignalSet) []string {
	var names []string
	for _, s := range sigSet.ByType(signals.FunctionDefined) {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return names
}

func calledNames(sigSet *signals.SignalSet) []string {
	seen := map[string]bool{}
	var names []string
	for _, s := range sigSet.ByType(signals.FunctionCalled) {
		if !seen[s.Name] {
			seen[s.Name] = true
			names = append(names, s.Name)
		}
	}
	sort.Strings(names)
	return names
}

func containsAll(haystack, needles []string) bool {
	set := make(map[string]bool, len(haystack))
	for _, h := range haystack {
		set[h] = true
	}
	for _, n := range needles {
		if !set[n] {
			return false
		}
	}
	return true
}

// analyzeContent is a convenience wrapper that writes content to a temp file
// with the given extension, then runs the regex analyzer on it.
func analyzeContent(t *testing.T, ext, content string) *signals.SignalSet {
	t.Helper()
	path := writeTemp(t, ext, content)
	a := regex.NewAnalyzer()
	sigSet, err := a.Analyze([]string{path})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	return sigSet
}

// ──────────────────────────────────────────────────────────────────────────────
// Language registry tests
// ──────────────────────────────────────────────────────────────────────────────

// TestLanguageRegistryCompleteness ensures the registry has entries for all
// extensions that historically mattered and several new ones.
func TestLanguageRegistryCompleteness(t *testing.T) {
	required := []string{
		".go", ".js", ".ts", ".jsx", ".tsx",
		".py", ".rs", ".java", ".rb", ".php",
		".c", ".cpp", ".h", ".hpp",
		".kt", ".swift", ".cs", ".scala",
		".lua", ".sh", ".hs", ".ex", ".exs",
		".r", ".R", ".jl", ".dart", ".pl",
		".nim", ".zig", ".clj", ".erl",
	}
	for _, ext := range required {
		t.Run(ext, func(t *testing.T) {
			_, ok := language.Detect("dummy" + ext)
			if !ok {
				t.Errorf("extension %s not registered", ext)
			}
		})
	}
}

// TestLanguageDetectionByExtension exercises the extension-based detection path.
func TestLanguageDetectionByExtension(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"server.ts", "typescript"},
		{"lib.rs", "rust"},
		{"Foo.java", "java"},
		{"helper.rb", "ruby"},
		{"script.lua", "lua"},
		{"main.zig", "zig"},
		{"init.ex", "elixir"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			lc, ok := language.Detect(tc.path)
			if !ok {
				t.Fatalf("no language detected for %s", tc.path)
			}
			if lc.ID != tc.want {
				t.Errorf("got %q, want %q", lc.ID, tc.want)
			}
		})
	}
}

// TestLanguageDetectionByShebang exercises shebang-based detection.
func TestLanguageDetectionByShebang(t *testing.T) {
	cases := []struct {
		shebang string
		wantID  string
	}{
		{"#!/usr/bin/env python3\n", "python"},
		{"#!/usr/bin/python\n", "python"},
		{"#!/usr/bin/env node\n", "javascript"},
		{"#!/bin/bash\n", "shell"},
		{"#!/usr/bin/env ruby\n", "ruby"},
		{"#!/usr/bin/perl\n", "perl"},
	}
	for _, tc := range cases {
		t.Run(tc.wantID, func(t *testing.T) {
			// Write a file with no recognised extension but with a shebang.
			path := writeTemp(t, "", tc.shebang+"echo hello\n")
			lc, ok := language.Detect(path)
			if !ok {
				t.Fatalf("no language detected via shebang for %q", tc.shebang)
			}
			if lc.ID != tc.wantID {
				t.Errorf("got %q, want %q", lc.ID, tc.wantID)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Regex analyzer — function definition extraction
// ──────────────────────────────────────────────────────────────────────────────

func TestPythonFunctionDefined(t *testing.T) {
	src := `
def greet(name):
    return "Hello " + name

async def fetch_data(url):
    pass

class MyClass:
    def method(self):
        self.value = 1
`
	sigSet := analyzeContent(t, ".py", src)
	defs := definedNames(sigSet)
	t.Logf("Python defs: %v", defs)
	for _, want := range []string{"greet", "fetch_data", "method"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestJavaScriptFunctionDefined(t *testing.T) {
	src := `
function greet(name) {
    return 'Hello ' + name;
}

async function fetchData(url) {
    const res = await fetch(url);
    return res.json();
}

const arrowFn = (x) => x * 2;
const asyncArrow = async (id) => {
    return id;
};
export default function exported() {}
`
	sigSet := analyzeContent(t, ".js", src)
	defs := definedNames(sigSet)
	t.Logf("JavaScript defs: %v", defs)
	for _, want := range []string{"greet", "fetchData"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestTypeScriptFunctionDefined(t *testing.T) {
	src := `
export async function loadUser(id: string): Promise<User> {
    return fetch('/users/' + id).then(r => r.json());
}

const transform = (data: Data): string => {
    return JSON.stringify(data);
};
`
	sigSet := analyzeContent(t, ".ts", src)
	defs := definedNames(sigSet)
	t.Logf("TypeScript defs: %v", defs)
	if !containsAll(defs, []string{"loadUser"}) {
		t.Errorf("expected loadUser in defs %v", defs)
	}
}

func TestRustFunctionDefined(t *testing.T) {
	src := `
pub fn compute(x: i32) -> i32 {
    x * 2
}

async fn fetch(url: &str) -> Result<String, Error> {
    Ok(url.to_string())
}

pub(crate) fn helper() {}
`
	sigSet := analyzeContent(t, ".rs", src)
	defs := definedNames(sigSet)
	t.Logf("Rust defs: %v", defs)
	for _, want := range []string{"compute", "fetch", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestRubyFunctionDefined(t *testing.T) {
	src := `
def greet(name)
  puts "Hello #{name}"
end

def valid?
  @value != nil
end

private

def internal_helper(x)
  x.to_s
end
`
	sigSet := analyzeContent(t, ".rb", src)
	defs := definedNames(sigSet)
	t.Logf("Ruby defs: %v", defs)
	for _, want := range []string{"greet", "valid?", "internal_helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestLuaFunctionDefined(t *testing.T) {
	src := `
function greet(name)
    print("Hello " .. name)
end

local function helper(x)
    return x * 2
end

myModule.method = function(self)
    return self.value
end
`
	sigSet := analyzeContent(t, ".lua", src)
	defs := definedNames(sigSet)
	t.Logf("Lua defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestElixirFunctionDefined(t *testing.T) {
	src := `
defmodule MyModule do
  def greet(name) do
    "Hello #{name}"
  end

  defp validate?(value) do
    value != nil
  end
end
`
	sigSet := analyzeContent(t, ".ex", src)
	defs := definedNames(sigSet)
	t.Logf("Elixir defs: %v", defs)
	for _, want := range []string{"greet", "validate?"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestKotlinFunctionDefined(t *testing.T) {
	src := `
fun greet(name: String): String {
    return "Hello $name"
}

suspend fun fetchData(url: String): String {
    return url
}

private fun helper(x: Int) = x * 2
`
	sigSet := analyzeContent(t, ".kt", src)
	defs := definedNames(sigSet)
	t.Logf("Kotlin defs: %v", defs)
	for _, want := range []string{"greet", "fetchData"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestZigFunctionDefined(t *testing.T) {
	src := `
pub fn greet(name: []const u8) void {
    std.debug.print("Hello {s}\n", .{name});
}

fn helper(x: i32) i32 {
    return x * 2;
}
`
	sigSet := analyzeContent(t, ".zig", src)
	defs := definedNames(sigSet)
	t.Logf("Zig defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestNimFunctionDefined(t *testing.T) {
	src := `
proc greet(name: string): string =
  "Hello " & name

func helper(x: int): int =
  x * 2
`
	sigSet := analyzeContent(t, ".nim", src)
	defs := definedNames(sigSet)
	t.Logf("Nim defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Regex analyzer — FunctionCalled extraction (the previously missing feature)
// ──────────────────────────────────────────────────────────────────────────────

func TestPythonFunctionCalled(t *testing.T) {
	src := `
def main():
    result = greet("world")
    print(result)
    data = fetch_data("http://example.com")
    process(data)
`
	sigSet := analyzeContent(t, ".py", src)
	called := calledNames(sigSet)
	t.Logf("Python called: %v", called)
	for _, want := range []string{"greet", "fetch_data", "process"} {
		if !containsAll(called, []string{want}) {
			t.Errorf("expected call to %q in %v", want, called)
		}
	}
}

func TestJavaScriptFunctionCalled(t *testing.T) {
	src := `
function main() {
    const msg = greet("world");
    console.log(msg);
    const data = fetchData("http://example.com");
    render(data);
}
`
	sigSet := analyzeContent(t, ".js", src)
	called := calledNames(sigSet)
	t.Logf("JavaScript called: %v", called)
	for _, want := range []string{"greet", "fetchData", "render"} {
		if !containsAll(called, []string{want}) {
			t.Errorf("expected call to %q in %v", want, called)
		}
	}
}

func TestRustFunctionCalled(t *testing.T) {
	src := `
fn main() {
    let result = compute(42);
    println!("{}", result);
    let s = to_string(result);
    save(s);
}
`
	sigSet := analyzeContent(t, ".rs", src)
	called := calledNames(sigSet)
	t.Logf("Rust called: %v", called)
	for _, want := range []string{"compute", "to_string", "save"} {
		if !containsAll(called, []string{want}) {
			t.Errorf("expected call to %q in %v", want, called)
		}
	}
}

// TestFunctionCallsNotCountedAsDefinitions ensures that a line containing a
// function definition is not also emitted as a call site.
func TestFunctionCallsNotCountedAsDefinitions(t *testing.T) {
	src := `
def greet(name):
    return "Hello " + name

def main():
    greet("world")
`
	sigSet := analyzeContent(t, ".py", src)
	// "greet" should appear as a definition; "main" as a definition.
	// "greet" should also appear as a call (from inside main).
	// "def" must NOT appear as a call.
	called := calledNames(sigSet)
	t.Logf("calls: %v", called)
	for _, bad := range []string{"def", "async"} {
		if containsAll(called, []string{bad}) {
			t.Errorf("keyword %q should not appear as a called name", bad)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// TODO extraction across comment styles
// ──────────────────────────────────────────────────────────────────────────────

func TestTodoExtractionSlashSlash(t *testing.T) {
	src := `
func main() {
    // TODO: implement this
    // FIXME: broken when x < 0
}
`
	sigSet := analyzeContent(t, ".go", src)
	todos := sigSet.ByType(signals.TodoFound)
	if len(todos) < 2 {
		t.Errorf("expected at least 2 TODOs, got %d", len(todos))
	}
}

func TestTodoExtractionHash(t *testing.T) {
	src := `
def main():
    # TODO: add input validation
    # FIXME: off-by-one error
    pass
`
	sigSet := analyzeContent(t, ".py", src)
	todos := sigSet.ByType(signals.TodoFound)
	if len(todos) < 2 {
		t.Errorf("expected at least 2 TODOs for Python, got %d", len(todos))
	}
}

func TestTodoExtractionDashDash(t *testing.T) {
	src := `
local function main()
    -- TODO: implement error handling
    -- FIXME: nil check missing
end
`
	sigSet := analyzeContent(t, ".lua", src)
	todos := sigSet.ByType(signals.TodoFound)
	if len(todos) < 2 {
		t.Errorf("expected at least 2 TODOs for Lua, got %d", len(todos))
	}
}

func TestTodoExtractionHashRuby(t *testing.T) {
	src := `
def main
  # TODO: refactor this
  # HACK: temporary fix
  puts "hello"
end
`
	sigSet := analyzeContent(t, ".rb", src)
	todos := sigSet.ByType(signals.TodoFound)
	if len(todos) < 2 {
		t.Errorf("expected at least 2 TODOs for Ruby, got %d", len(todos))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// @ignore-* extraction across comment styles
// ──────────────────────────────────────────────────────────────────────────────

func TestIgnoreMarkerGoSlashSlash(t *testing.T) {
	src := `package main

// @ignore-unused
func unusedHelper() {
    _ = 1
}
`
	path := writeTemp(t, ".go", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for Go //")
	}
}

func TestIgnoreMarkerPythonHash(t *testing.T) {
	src := `
# @ignore-unused
def unused_helper():
    pass
`
	path := writeTemp(t, ".py", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for Python #")
	}
}

func TestIgnoreMarkerRubyHash(t *testing.T) {
	src := `
# @ignore-unused
def unused_method
  42
end
`
	path := writeTemp(t, ".rb", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for Ruby #")
	}
}

func TestIgnoreMarkerLuaDashDash(t *testing.T) {
	src := `
-- @ignore-unused
local function unused_helper()
    return 42
end
`
	path := writeTemp(t, ".lua", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for Lua --")
	}
}

func TestIgnoreMarkerElixirHash(t *testing.T) {
	src := `
# @ignore-unused
def unused_fn do
  :ok
end
`
	path := writeTemp(t, ".ex", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for Elixir #")
	}
}

// extractIgnoreSignals is a test helper that reads a file and runs the same
// ignore-marker extraction that the main analysis pipeline uses.
func extractIgnoreSignals(path string) []signals.Signal {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lc, _ := language.Detect(path)
	raw := signals.ExtractIgnoreMarkers(string(content), path, lc.ID)

	var result []signals.Signal
	for _, ig := range raw {
		if sym, ok := ig.Metadata["symbol_name"].(string); ok {
			sig := signals.NewSignal(signals.IgnoreFlag, path).
				WithName(sym)
			result = append(result, *sig)
		}
	}
	return result
}

// ──────────────────────────────────────────────────────────────────────────────
// @dizz intent marker extraction
// ──────────────────────────────────────────────────────────────────────────────

func TestIntentMarkerPython(t *testing.T) {
	src := `
# @dizz:state planned
def planned_feature():
    pass
`
	sigSet := analyzeContent(t, ".py", src)
	markers := sigSet.ByType(signals.IntentMarker)
	if len(markers) == 0 {
		t.Error("expected @dizz:state marker in Python source")
	}
	for _, m := range markers {
		if v, _ := m.Metadata["value"].(string); v != "planned" {
			t.Errorf("expected value=planned, got %q", v)
		}
	}
}

func TestIntentMarkerRust(t *testing.T) {
	src := `
// @dizz:state unstable
fn fragile_function() {
    unimplemented!()
}
`
	sigSet := analyzeContent(t, ".rs", src)
	markers := sigSet.ByType(signals.IntentMarker)
	if len(markers) == 0 {
		t.Error("expected @dizz:state marker in Rust source")
	}
}

func TestIntentMarkerLua(t *testing.T) {
	src := `
-- @dizz:feature auth
local function login(user, pass)
    return true
end
`
	sigSet := analyzeContent(t, ".lua", src)
	markers := sigSet.ByType(signals.IntentMarker)
	if len(markers) == 0 {
		t.Error("expected @dizz:feature marker in Lua source")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Discover: dynamic extension list
// ──────────────────────────────────────────────────────────────────────────────

// TestCodeFilesWithIncludesCustom verifies that explicitly passed include
// patterns are respected, allowing analysis of any extension.
func TestCodeFilesWithIncludesCustom(t *testing.T) {
	dir := t.TempDir()
	// Write a handful of files with different extensions
	extensions := []string{".py", ".rb", ".lua", ".ex", ".kt", ".swift", ".zig"}
	for _, ext := range extensions {
		f := filepath.Join(dir, "sample"+ext)
		if err := os.WriteFile(f, []byte("# sample\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}

	// Using the default (registry-based) extension list — all files should be found.
	files, err := discover.CodeFiles(dir, nil)
	if err != nil {
		t.Fatalf("CodeFiles: %v", err)
	}
	if len(files) != len(extensions) {
		t.Errorf("expected %d files, got %d: %v", len(extensions), len(files), files)
	}
}

// TestCodeFilesCustomIncludes verifies that custom include patterns override
// the default list.
func TestCodeFilesCustomIncludes(t *testing.T) {
	dir := t.TempDir()
	// Write a .myext file (not in the registry)
	f := filepath.Join(dir, "sample.myext")
	if err := os.WriteFile(f, []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Also write a .py file
	f2 := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(f2, []byte("# py\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Without custom includes: only .py should be found
	files, _ := discover.CodeFiles(dir, nil)
	for _, fn := range files {
		if strings.HasSuffix(fn, ".myext") {
			t.Errorf(".myext should not be found without custom includes")
		}
	}

	// With custom includes: .myext should also be found
	all, err := discover.CodeFilesWithIncludes(dir, []string{"**/*.myext", "**/*.py"}, nil)
	if err != nil {
		t.Fatalf("CodeFilesWithIncludes: %v", err)
	}
	foundMyExt := false
	for _, fn := range all {
		if strings.HasSuffix(fn, ".myext") {
			foundMyExt = true
		}
	}
	if !foundMyExt {
		t.Error("expected .myext to be found with custom includes")
	}
}
