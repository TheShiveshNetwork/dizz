package benchmarks

import (
	"testing"

	"github.com/TheShiveshNetwork/dizz/internal/language"
	"github.com/TheShiveshNetwork/dizz/internal/signals"
)

// ──────────────────────────────────────────────────────────────────────────────
// Extended language conformance tests — covers all 34 registered languages
// ──────────────────────────────────────────────────────────────────────────────

func TestGoFunctionDefined(t *testing.T) {
	src := `
package main

func greet(name string) string {
	return "Hello " + name
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
`
	sigSet := analyzeContent(t, ".go", src)
	defs := definedNames(sigSet)
	t.Logf("Go defs: %v", defs)
	for _, want := range []string{"greet", "ServeHTTP"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestJavaFunctionDefined(t *testing.T) {
	src := `
public class Hello {
	public String greet(String name) {
		return "Hello " + name;
	}

	private static int calculate(int x, int y) {
		return x + y;
	}
}
`
	sigSet := analyzeContent(t, ".java", src)
	defs := definedNames(sigSet)
	t.Logf("Java defs: %v", defs)
	for _, want := range []string{"greet", "calculate"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestSwiftFunctionDefined(t *testing.T) {
	src := `
func greet(name: String) -> String {
	return "Hello \(name)"
}

private func helper(x: Int) -> Int {
	return x * 2
}

public func fetchData(url: String) async throws -> Data {
	return Data()
}
`
	sigSet := analyzeContent(t, ".swift", src)
	defs := definedNames(sigSet)
	t.Logf("Swift defs: %v", defs)
	for _, want := range []string{"greet", "helper", "fetchData"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestCSharpFunctionDefined(t *testing.T) {
	src := `
public class Greeter {
	public string Greet(string name) {
		return "Hello " + name;
	}

	private static int Calculate(int x, int y) {
		return x + y;
	}

	public async Task<string> FetchData(string url) {
		return await httpClient.GetStringAsync(url);
	}
}
`
	sigSet := analyzeContent(t, ".cs", src)
	defs := definedNames(sigSet)
	t.Logf("C# defs: %v", defs)
	for _, want := range []string{"Greet", "Calculate", "FetchData"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestPHPFunctionDefined(t *testing.T) {
	src := `<?php

function greet(string $name): string {
	return "Hello " . $name;
}

class MyClass {
	public function calculate(int $x, int $y): int {
		return $x + $y;
	}

	private static function helper(): void {
		// noop
	}
}
`
	sigSet := analyzeContent(t, ".php", src)
	defs := definedNames(sigSet)
	t.Logf("PHP defs: %v", defs)
	for _, want := range []string{"greet", "calculate", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestShellFunctionDefined(t *testing.T) {
	src := `
#!/bin/bash

greet() {
	echo "Hello $1"
}

helper() {
	return 0
}

function longform() {
	echo "ok"
}
`
	sigSet := analyzeContent(t, ".sh", src)
	defs := definedNames(sigSet)
	t.Logf("Shell defs: %v", defs)
	for _, want := range []string{"greet", "helper", "longform"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestHaskellFunctionDefined(t *testing.T) {
	src := `
module Main where

greet :: String -> String
greet name = "Hello " ++ name

helper :: Int -> Int
helper x = x * 2
`
	sigSet := analyzeContent(t, ".hs", src)
	defs := definedNames(sigSet)
	t.Logf("Haskell defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestRFunctionDefined(t *testing.T) {
	src := `
greet <- function(name) {
	paste("Hello", name)
}

helper <- function(x, y) {
	x + y
}
`
	sigSet := analyzeContent(t, ".r", src)
	defs := definedNames(sigSet)
	t.Logf("R defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestJuliaFunctionDefined(t *testing.T) {
	src := `
function greet(name)
	return "Hello " * name
end

helper(x) = x * 2

function compute(x, y)
	return x + y
end
`
	sigSet := analyzeContent(t, ".jl", src)
	defs := definedNames(sigSet)
	t.Logf("Julia defs: %v", defs)
	for _, want := range []string{"greet", "helper", "compute"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestDartFunctionDefined(t *testing.T) {
	src := `
String greet(String name) {
	return 'Hello $name';
}

int helper(int x) => x * 2;

Future<String> fetchData(String url) async {
	return url;
}
`
	sigSet := analyzeContent(t, ".dart", src)
	defs := definedNames(sigSet)
	t.Logf("Dart defs: %v", defs)
	for _, want := range []string{"greet", "helper", "fetchData"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestPerlFunctionDefined(t *testing.T) {
	src := `
sub greet {
	my ($name) = @_;
	return "Hello $name";
}

sub helper {
	my ($x) = @_;
	return $x * 2;
}
`
	sigSet := analyzeContent(t, ".pl", src)
	defs := definedNames(sigSet)
	t.Logf("Perl defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestClojureFunctionDefined(t *testing.T) {
	src := `
(defn greet [name]
	(str "Hello " name))

(defn- helper [x]
	(* x 2))
`
	sigSet := analyzeContent(t, ".clj", src)
	defs := definedNames(sigSet)
	t.Logf("Clojure defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestErlangFunctionDefined(t *testing.T) {
	src := `
-module(hello).

greet(Name) ->
	"Hello " ++ Name.

helper(X) ->
	X * 2.
`
	sigSet := analyzeContent(t, ".erl", src)
	defs := definedNames(sigSet)
	t.Logf("Erlang defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestOCamlFunctionDefined(t *testing.T) {
	src := `
let greet name = "Hello " ^ name

let rec helper x = x * 2
`
	sigSet := analyzeContent(t, ".ml", src)
	defs := definedNames(sigSet)
	t.Logf("OCaml defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestFSharpFunctionDefined(t *testing.T) {
	src := `
let greet name = "Hello " + name

let rec helper x = x * 2
`
	sigSet := analyzeContent(t, ".fs", src)
	defs := definedNames(sigSet)
	t.Logf("F# defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestSQLFunctionDefined(t *testing.T) {
	src := `
CREATE FUNCTION calculate_tax(amount DECIMAL) RETURNS DECIMAL
BEGIN
	RETURN amount * 0.2;
END;

CREATE OR REPLACE PROCEDURE update_inventory(item_id INT)
BEGIN
	UPDATE stock SET count = count - 1 WHERE id = item_id;
END;
`
	sigSet := analyzeContent(t, ".sql", src)
	defs := definedNames(sigSet)
	t.Logf("SQL defs: %v", defs)
	for _, want := range []string{"calculate_tax", "update_inventory"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestMATLABFunctionDefined(t *testing.T) {
	src := `
function result = greet(name)
	result = ['Hello ', name];
end

function helper(x)
	disp(x);
end
`
	sigSet := analyzeContent(t, ".m", src)
	defs := definedNames(sigSet)
	t.Logf("MATLAB defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestTerraformResourceDefined(t *testing.T) {
	src := `
resource "aws_s3_bucket" "my_bucket" {
	bucket = "my-bucket"
}

resource "aws_instance" "web_server" {
	ami = "ami-123"
}

module "networking" {
	source = "./networking"
}
`
	sigSet := analyzeContent(t, ".tf", src)
	defs := definedNames(sigSet)
	t.Logf("Terraform defs: %v", defs)
	for _, want := range []string{"my_bucket", "web_server", "networking"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestCFunctionDefined(t *testing.T) {
	src := `
int greet(char *name) {
	return 0;
}

static void helper(int x) {
	return;
}
`
	sigSet := analyzeContent(t, ".c", src)
	defs := definedNames(sigSet)
	t.Logf("C defs: %v", defs)
	for _, want := range []string{"greet", "helper"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

func TestCppFunctionDefined(t *testing.T) {
	src := `
int greet(const std::string& name) {
	return 0;
}

template<typename T>
T helper(T x) {
	return x;
}

class Foo {
	void bar() const {}
};
`
	sigSet := analyzeContent(t, ".cpp", src)
	defs := definedNames(sigSet)
	t.Logf("C++ defs: %v", defs)
	if !containsAll(defs, []string{"greet"}) {
		t.Errorf("expected greet in defs %v", defs)
	}
}

func TestScalaFunctionDefined(t *testing.T) {
	src := `
def greet(name: String): String = {
	"Hello " + name
}

private def helper(x: Int): Int = x * 2

def compute(x: Int)(y: Int): Int = x + y
`
	sigSet := analyzeContent(t, ".scala", src)
	defs := definedNames(sigSet)
	t.Logf("Scala defs: %v", defs)
	for _, want := range []string{"greet", "helper", "compute"} {
		if !containsAll(defs, []string{want}) {
			t.Errorf("expected %q in defs %v", want, defs)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Comprehensive: every registered extension detects correctly
// ──────────────────────────────────────────────────────────────────────────────

func TestAllRegisteredExtensions(t *testing.T) {
	all := language.All()
	extCount := 0
	for _, lc := range all {
		extCount += len(lc.Extensions)
		for _, ext := range lc.Extensions {
			t.Run(lc.ID+":"+ext, func(t *testing.T) {
				detected, ok := language.Detect("file" + ext)
				if !ok {
					t.Errorf("extension %s not detected for language %s", ext, lc.ID)
				}
				if detected.ID != lc.ID {
					t.Errorf("detected %q, want %q for extension %s", detected.ID, lc.ID, ext)
				}
			})
		}
	}
	t.Logf("Total registered language entries: %d", len(all))
	t.Logf("Total registered extensions: %d", extCount)
}

// ──────────────────────────────────────────────────────────────────────────────
// TODO/FIXME extraction across the remaining comment styles
// ──────────────────────────────────────────────────────────────────────────────

func TestTodoExtractionPercent(t *testing.T) {
	src := `
-module(hello).
% TODO: implement this
% FIXME: handle edge case
`
	sigSet := analyzeContent(t, ".erl", src)
	todos := sigSet.ByType(signals.TodoFound)
	if len(todos) < 2 {
		t.Errorf("expected at least 2 TODOs for Erlang %%, got %d", len(todos))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// @ignore-* and @dizz:* markers for additional languages
// ──────────────────────────────────────────────────────────────────────────────

func TestIgnoreMarkerJava(t *testing.T) {
	src := `
// @ignore-unused
public class TempHelper {
	public void helper() {}
}
`
	path := writeTemp(t, ".java", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for Java //")
	}
}

func TestIgnoreMarkerShellHash(t *testing.T) {
	src := `#!/bin/bash
# @ignore-unused
helper() {
	return 0
}
`
	path := writeTemp(t, ".sh", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for Shell #")
	}
}

func TestIgnoreMarkerCpp(t *testing.T) {
	src := `
// @ignore-unused
void temp_helper() {}
`
	path := writeTemp(t, ".cpp", src)
	sigs := extractIgnoreSignals(path)
	if len(sigs) == 0 {
		t.Error("expected at least one @ignore-unused signal for C++ //")
	}
}
