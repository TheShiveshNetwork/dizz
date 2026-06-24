// Package language provides language detection and per-language analysis
// configuration for the dizz signal-extraction pipeline.
package language

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Tier represents the accuracy of a signal extracted for a given language.
// It is stored in signal metadata (key "source_tier") so the scorer can
// weight classification decisions accordingly.
type Tier int

const (
	TierAST     Tier = 1 // Parser-backed (AST) — highest accuracy
	TierLexical Tier = 2 // Structural/lexical regex — good accuracy
	TierRegex   Tier = 3 // Regex fallback — lower accuracy
)

// CommentStyle describes how inline and block comments are delimited in a
// language. A language may have multiple styles (e.g. PHP supports both
// // and # line comments).
type CommentStyle struct {
	LinePrefix string // e.g. "//", "#", "--", ";"
	BlockStart string // e.g. "/*", "{-", "(*", "--[["
	BlockEnd   string // e.g. "*/", "-}", "*)". "]]"
}

// LanguageConfig holds all analysis settings for a single language.
type LanguageConfig struct {
	ID   string
	Name string

	// File matching
	Extensions []string // e.g. [".py", ".pyw"]
	Shebangs   []string // interpreter names in shebang lines, e.g. ["python3"]

	// Comment syntax — used for TODO / @dizz / @ignore-* extraction
	CommentStyles []CommentStyle

	// Regex patterns for structure extraction.
	// Every pattern MUST capture the symbol name in capture group 1.
	FunctionPatterns []string
	TypePatterns     []string // class / struct / interface declarations
	ImportPatterns   []string // import/require/use statements

	// Regex patterns for usage extraction (call-site detection).
	// Each pattern MUST capture the called name in capture group 1.
	CallPatterns []string

	// Keywords to exclude when a call pattern matches a keyword name.
	Keywords map[string]bool

	// DefaultTier is the accuracy level of signals produced for this language.
	DefaultTier Tier
}

// ──────────────────────────────────────────────────────────────────────────────
// keyword sets
// ──────────────────────────────────────────────────────────────────────────────

var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

var jsKeywords = map[string]bool{
	"if": true, "for": true, "while": true, "switch": true, "function": true,
	"class": true, "new": true, "typeof": true, "instanceof": true, "return": true,
	"throw": true, "catch": true, "import": true, "export": true, "await": true,
	"async": true, "delete": true, "void": true, "in": true, "of": true,
	"let": true, "var": true, "const": true, "do": true, "else": true,
	"try": true, "finally": true, "yield": true, "from": true, "case": true,
	"break": true, "continue": true, "default": true, "with": true, "static": true,
	"super": true, "extends": true, "get": true, "set": true, "debugger": true,
}

var tsKeywords = jsKeywords // TypeScript extends the JS keyword set

var pythonKeywords = map[string]bool{
	"if": true, "elif": true, "else": true, "for": true, "while": true,
	"with": true, "def": true, "class": true, "return": true, "raise": true,
	"except": true, "import": true, "from": true, "lambda": true, "yield": true,
	"assert": true, "del": true, "pass": true, "break": true, "continue": true,
	"global": true, "nonlocal": true, "try": true, "finally": true, "and": true,
	"or": true, "not": true, "in": true, "is": true, "as": true, "async": true,
	"await": true, "print": true, "exec": true, "type": true,
}

var rustKeywords = map[string]bool{
	"as": true, "async": true, "await": true, "break": true, "const": true,
	"continue": true, "crate": true, "dyn": true, "else": true, "enum": true,
	"extern": true, "false": true, "fn": true, "for": true, "if": true,
	"impl": true, "in": true, "let": true, "loop": true, "match": true,
	"mod": true, "move": true, "mut": true, "pub": true, "ref": true,
	"return": true, "self": true, "Self": true, "static": true, "struct": true,
	"super": true, "trait": true, "true": true, "type": true, "union": true,
	"unsafe": true, "use": true, "where": true, "while": true, "yield": true,
}

var javaKeywords = map[string]bool{
	"abstract": true, "assert": true, "boolean": true, "break": true, "byte": true,
	"case": true, "catch": true, "char": true, "class": true, "const": true,
	"continue": true, "default": true, "do": true, "double": true, "else": true,
	"enum": true, "extends": true, "final": true, "finally": true, "float": true,
	"for": true, "goto": true, "if": true, "implements": true, "import": true,
	"instanceof": true, "int": true, "interface": true, "long": true, "native": true,
	"new": true, "package": true, "private": true, "protected": true, "public": true,
	"return": true, "short": true, "static": true, "strictfp": true, "super": true,
	"switch": true, "synchronized": true, "this": true, "throw": true, "throws": true,
	"transient": true, "try": true, "void": true, "volatile": true, "while": true,
}

var kotlinKeywords = map[string]bool{
	"abstract": true, "actual": true, "as": true, "break": true, "by": true,
	"catch": true, "class": true, "companion": true, "const": true, "constructor": true,
	"continue": true, "crossinline": true, "data": true, "delegate": true, "do": true,
	"dynamic": true, "else": true, "enum": true, "expect": true, "external": true,
	"false": true, "final": true, "finally": true, "for": true, "fun": true,
	"get": true, "if": true, "import": true, "in": true, "infix": true,
	"init": true, "inline": true, "inner": true, "interface": true, "internal": true,
	"is": true, "it": true, "lateinit": true, "noinline": true, "null": true,
	"object": true, "open": true, "operator": true, "out": true, "override": true,
	"package": true, "private": true, "protected": true, "public": true, "reified": true,
	"return": true, "sealed": true, "set": true, "super": true, "suspend": true,
	"tailrec": true, "this": true, "throw": true, "true": true, "try": true,
	"typealias": true, "typeof": true, "val": true, "var": true, "vararg": true,
	"when": true, "where": true, "while": true,
}

var swiftKeywords = map[string]bool{
	"as": true, "associatedtype": true, "break": true, "case": true, "catch": true,
	"class": true, "continue": true, "default": true, "defer": true, "deinit": true,
	"do": true, "else": true, "enum": true, "extension": true, "fallthrough": true,
	"false": true, "fileprivate": true, "final": true, "for": true, "func": true,
	"guard": true, "if": true, "import": true, "in": true, "init": true,
	"inout": true, "internal": true, "is": true, "lazy": true, "let": true,
	"mutating": true, "nil": true, "open": true, "operator": true, "override": true,
	"private": true, "protocol": true, "public": true, "repeat": true, "required": true,
	"rethrows": true, "return": true, "self": true, "Self": true, "static": true,
	"struct": true, "subscript": true, "super": true, "switch": true, "throw": true,
	"throws": true, "true": true, "try": true, "typealias": true, "unowned": true,
	"var": true, "weak": true, "where": true, "while": true,
}

var csharpKeywords = map[string]bool{
	"abstract": true, "as": true, "base": true, "bool": true, "break": true,
	"byte": true, "case": true, "catch": true, "char": true, "checked": true,
	"class": true, "const": true, "continue": true, "decimal": true, "default": true,
	"delegate": true, "do": true, "double": true, "else": true, "enum": true,
	"event": true, "explicit": true, "extern": true, "false": true, "finally": true,
	"fixed": true, "float": true, "for": true, "foreach": true, "goto": true,
	"if": true, "implicit": true, "in": true, "int": true, "interface": true,
	"internal": true, "is": true, "lock": true, "long": true, "namespace": true,
	"new": true, "null": true, "object": true, "operator": true, "out": true,
	"override": true, "params": true, "private": true, "protected": true, "public": true,
	"readonly": true, "ref": true, "return": true, "sbyte": true, "sealed": true,
	"short": true, "sizeof": true, "stackalloc": true, "static": true, "string": true,
	"struct": true, "switch": true, "this": true, "throw": true, "true": true,
	"try": true, "typeof": true, "uint": true, "ulong": true, "unchecked": true,
	"unsafe": true, "ushort": true, "using": true, "virtual": true, "void": true,
	"volatile": true, "while": true,
}

var rubyKeywords = map[string]bool{
	"BEGIN": true, "END": true, "alias": true, "and": true, "begin": true,
	"break": true, "case": true, "class": true, "def": true, "defined?": true,
	"do": true, "else": true, "elsif": true, "end": true, "ensure": true,
	"false": true, "for": true, "if": true, "in": true, "module": true,
	"next": true, "nil": true, "not": true, "or": true, "redo": true,
	"rescue": true, "retry": true, "return": true, "self": true, "super": true,
	"then": true, "true": true, "undef": true, "unless": true, "until": true,
	"when": true, "while": true, "yield": true, "puts": true, "print": true,
	"require": true, "require_relative": true, "include": true, "extend": true,
	"raise": true, "fail": true, "lambda": true, "proc": true, "p": true,
}

var phpKeywords = map[string]bool{
	"abstract": true, "and": true, "array": true, "as": true, "break": true,
	"callable": true, "case": true, "catch": true, "class": true, "clone": true,
	"const": true, "continue": true, "declare": true, "default": true, "die": true,
	"do": true, "echo": true, "else": true, "elseif": true, "empty": true,
	"enddeclare": true, "endfor": true, "endforeach": true, "endif": true,
	"endswitch": true, "endwhile": true, "eval": true, "exit": true, "extends": true,
	"final": true, "finally": true, "for": true, "foreach": true, "function": true,
	"global": true, "goto": true, "if": true, "implements": true, "include": true,
	"include_once": true, "instanceof": true, "insteadof": true, "interface": true,
	"isset": true, "list": true, "match": true, "namespace": true, "new": true,
	"or": true, "print": true, "private": true, "protected": true, "public": true,
	"readonly": true, "require": true, "require_once": true, "return": true,
	"static": true, "switch": true, "throw": true, "trait": true, "try": true,
	"unset": true, "use": true, "var": true, "while": true, "xor": true, "yield": true,
}

var cKeywords = map[string]bool{
	"auto": true, "break": true, "case": true, "char": true, "const": true,
	"continue": true, "default": true, "do": true, "double": true, "else": true,
	"enum": true, "extern": true, "float": true, "for": true, "goto": true,
	"if": true, "int": true, "long": true, "register": true, "return": true,
	"short": true, "signed": true, "sizeof": true, "static": true, "struct": true,
	"switch": true, "typedef": true, "union": true, "unsigned": true, "void": true,
	"volatile": true, "while": true,
}

var cppKeywords = map[string]bool{
	"alignas": true, "alignof": true, "and": true, "and_eq": true, "asm": true,
	"auto": true, "bitand": true, "bitor": true, "bool": true, "break": true,
	"case": true, "catch": true, "char": true, "char8_t": true, "char16_t": true,
	"char32_t": true, "class": true, "compl": true, "concept": true, "const": true,
	"consteval": true, "constexpr": true, "constinit": true, "const_cast": true,
	"continue": true, "co_await": true, "co_return": true, "co_yield": true,
	"decltype": true, "default": true, "delete": true, "do": true, "double": true,
	"dynamic_cast": true, "else": true, "enum": true, "explicit": true,
	"export": true, "extern": true, "false": true, "float": true, "for": true,
	"friend": true, "goto": true, "if": true, "inline": true, "int": true,
	"long": true, "mutable": true, "namespace": true, "new": true, "noexcept": true,
	"not": true, "not_eq": true, "nullptr": true, "operator": true, "or": true,
	"or_eq": true, "private": true, "protected": true, "public": true,
	"register": true, "reinterpret_cast": true, "requires": true, "return": true,
	"short": true, "signed": true, "sizeof": true, "static": true,
	"static_assert": true, "static_cast": true, "struct": true, "switch": true,
	"template": true, "this": true, "thread_local": true, "throw": true,
	"true": true, "try": true, "typedef": true, "typeid": true, "typename": true,
	"union": true, "unsigned": true, "using": true, "virtual": true, "void": true,
	"volatile": true, "wchar_t": true, "while": true, "xor": true, "xor_eq": true,
}

var scalaKeywords = map[string]bool{
	"abstract": true, "case": true, "catch": true, "class": true, "def": true,
	"do": true, "else": true, "extends": true, "false": true, "final": true,
	"finally": true, "for": true, "forSome": true, "if": true, "implicit": true,
	"import": true, "lazy": true, "match": true, "new": true, "null": true,
	"object": true, "override": true, "package": true, "private": true,
	"protected": true, "return": true, "sealed": true, "super": true, "this": true,
	"throw": true, "trait": true, "try": true, "true": true, "type": true,
	"val": true, "var": true, "while": true, "with": true, "yield": true,
}

var luaKeywords = map[string]bool{
	"and": true, "break": true, "do": true, "else": true, "elseif": true,
	"end": true, "false": true, "for": true, "function": true, "goto": true,
	"if": true, "in": true, "local": true, "nil": true, "not": true,
	"or": true, "repeat": true, "return": true, "then": true, "true": true,
	"until": true, "while": true, "require": true, "print": true,
}

var shellKeywords = map[string]bool{
	"if": true, "then": true, "else": true, "elif": true, "fi": true, "for": true,
	"while": true, "do": true, "done": true, "case": true, "esac": true,
	"function": true, "return": true, "exit": true, "break": true, "continue": true,
	"export": true, "local": true, "declare": true, "readonly": true, "echo": true,
	"printf": true, "read": true, "test": true, "true": true, "false": true,
	"set": true, "unset": true, "shift": true, "source": true, "exec": true,
}

var haskellKeywords = map[string]bool{
	"case": true, "class": true, "data": true, "default": true, "deriving": true,
	"do": true, "else": true, "foreign": true, "if": true, "import": true,
	"in": true, "infix": true, "infixl": true, "infixr": true, "instance": true,
	"let": true, "module": true, "newtype": true, "of": true, "then": true,
	"type": true, "where": true, "forall": true,
}

var elixirKeywords = map[string]bool{
	"after": true, "and": true, "catch": true, "case": true, "cond": true,
	"def": true, "defexception": true, "defmacro": true, "defmodule": true,
	"defp": true, "defprotocol": true, "defstruct": true, "do": true,
	"else": true, "end": true, "false": true, "fn": true, "for": true,
	"if": true, "import": true, "in": true, "nil": true, "not": true,
	"or": true, "quote": true, "raise": true, "receive": true, "rescue": true,
	"require": true, "return": true, "super": true, "throw": true, "true": true,
	"try": true, "unless": true, "unquote": true, "use": true, "when": true,
	"with": true,
}

var rKeywords = map[string]bool{
	"if": true, "else": true, "for": true, "while": true, "repeat": true,
	"function": true, "return": true, "next": true, "break": true,
	"TRUE": true, "FALSE": true, "NULL": true, "Inf": true, "NaN": true,
	"NA": true, "in": true, "library": true, "require": true,
}

var juliaKeywords = map[string]bool{
	"abstract": true, "baremodule": true, "begin": true, "break": true,
	"catch": true, "const": true, "continue": true, "do": true, "else": true,
	"elseif": true, "end": true, "export": true, "false": true, "finally": true,
	"for": true, "function": true, "global": true, "if": true, "import": true,
	"in": true, "let": true, "local": true, "macro": true, "module": true,
	"mutable": true, "primitive": true, "quote": true, "return": true,
	"struct": true, "true": true, "try": true, "type": true, "using": true,
	"where": true, "while": true,
}

var dartKeywords = map[string]bool{
	"abstract": true, "as": true, "assert": true, "async": true, "await": true,
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "covariant": true, "default": true, "deferred": true,
	"do": true, "dynamic": true, "else": true, "enum": true, "export": true,
	"extends": true, "extension": true, "external": true, "factory": true,
	"false": true, "final": true, "finally": true, "for": true, "Function": true,
	"get": true, "hide": true, "if": true, "implements": true, "import": true,
	"in": true, "inout": true, "interface": true, "is": true, "late": true,
	"library": true, "mixin": true, "new": true, "null": true, "on": true,
	"operator": true, "part": true, "required": true, "rethrow": true,
	"return": true, "sealed": true, "set": true, "show": true, "static": true,
	"super": true, "switch": true, "sync": true, "this": true, "throw": true,
	"true": true, "try": true, "typedef": true, "var": true, "void": true,
	"when": true, "while": true, "with": true, "yield": true,
}

var perlKeywords = map[string]bool{
	"if": true, "elsif": true, "else": true, "unless": true, "for": true,
	"foreach": true, "while": true, "do": true, "until": true, "sub": true,
	"my": true, "our": true, "local": true, "use": true, "no": true,
	"require": true, "return": true, "last": true, "next": true, "redo": true,
	"die": true, "exit": true, "print": true, "say": true, "warn": true,
	"push": true, "pop": true, "shift": true, "unshift": true, "splice": true,
}

var nimKeywords = map[string]bool{
	"addr": true, "and": true, "as": true, "asm": true, "bind": true,
	"block": true, "break": true, "case": true, "cast": true, "concept": true,
	"const": true, "continue": true, "converter": true, "defer": true,
	"discard": true, "distinct": true, "div": true, "do": true, "elif": true,
	"else": true, "end": true, "enum": true, "except": true, "export": true,
	"finally": true, "for": true, "from": true, "func": true, "if": true,
	"import": true, "in": true, "include": true, "interface": true, "is": true,
	"isnot": true, "iterator": true, "let": true, "macro": true, "method": true,
	"mixin": true, "mod": true, "nil": true, "not": true, "notin": true,
	"object": true, "of": true, "or": true, "out": true, "proc": true,
	"ptr": true, "raise": true, "ref": true, "return": true, "shl": true,
	"shr": true, "static": true, "template": true, "try": true, "tuple": true,
	"type": true, "using": true, "var": true, "when": true, "while": true,
	"xor": true, "yield": true,
}

var zigKeywords = map[string]bool{
	"addrspace": true, "align": true, "allowzero": true, "and": true,
	"anyframe": true, "anytype": true, "asm": true, "async": true, "await": true,
	"break": true, "callconv": true, "catch": true, "comptime": true, "const": true,
	"continue": true, "defer": true, "else": true, "enum": true, "errdefer": true,
	"error": true, "export": true, "extern": true, "false": true, "fn": true,
	"for": true, "if": true, "inline": true, "linksection": true, "noalias": true,
	"noinline": true, "nosuspend": true, "null": true, "opaque": true, "or": true,
	"orelse": true, "packed": true, "pub": true, "resume": true, "return": true,
	"struct": true, "suspend": true, "switch": true, "test": true, "threadlocal": true,
	"true": true, "try": true, "type": true, "undefined": true, "union": true,
	"unreachable": true, "usingnamespace": true, "var": true, "volatile": true,
	"while": true,
}

var clojureKeywords = map[string]bool{
	"def": true, "defn": true, "defn-": true, "defmacro": true, "defmethod": true,
	"defmulti": true, "defprotocol": true, "defrecord": true, "defstruct": true,
	"deftype": true, "fn": true, "if": true, "let": true, "loop": true,
	"recur": true, "do": true, "cond": true, "case": true, "when": true,
	"or": true, "and": true, "not": true, "for": true, "doseq": true,
	"dotimes": true, "while": true, "import": true, "require": true, "use": true,
	"ns": true, "throw": true, "try": true, "catch": true, "finally": true,
	"nil": true, "true": true, "false": true, "new": true, "this": true,
}

var erlangKeywords = map[string]bool{
	"after": true, "and": true, "andalso": true, "band": true, "begin": true,
	"bnot": true, "bor": true, "bsl": true, "bsr": true, "bxor": true,
	"case": true, "catch": true, "cond": true, "div": true, "end": true,
	"fun": true, "if": true, "let": true, "not": true, "of": true,
	"or": true, "orelse": true, "query": true, "receive": true, "rem": true,
	"try": true, "when": true, "xor": true,
}

var ocamlKeywords = map[string]bool{
	"and": true, "as": true, "assert": true, "asr": true, "begin": true,
	"class": true, "constraint": true, "do": true, "done": true, "downto": true,
	"else": true, "end": true, "exception": true, "external": true, "false": true,
	"for": true, "fun": true, "function": true, "functor": true, "if": true,
	"in": true, "include": true, "inherit": true, "initializer": true, "land": true,
	"lazy": true, "let": true, "lor": true, "lsl": true, "lsr": true,
	"lxor": true, "match": true, "method": true, "mod": true, "module": true,
	"mutable": true, "new": true, "nonrec": true, "object": true, "of": true,
	"open": true, "or": true, "private": true, "rec": true, "sig": true,
	"struct": true, "then": true, "to": true, "true": true, "try": true,
	"type": true, "val": true, "virtual": true, "when": true, "while": true,
	"with": true,
}

var fsharpKeywords = map[string]bool{
	"abstract": true, "and": true, "as": true, "assert": true, "base": true,
	"begin": true, "class": true, "default": true, "delegate": true, "do": true,
	"done": true, "downcast": true, "downto": true, "elif": true, "else": true,
	"end": true, "exception": true, "extern": true, "false": true, "finally": true,
	"fixed": true, "for": true, "fun": true, "function": true, "global": true,
	"if": true, "in": true, "inherit": true, "inline": true, "interface": true,
	"internal": true, "lazy": true, "let": true, "match": true, "member": true,
	"module": true, "mutable": true, "namespace": true, "new": true, "not": true,
	"null": true, "of": true, "open": true, "or": true, "override": true,
	"private": true, "public": true, "rec": true, "return": true, "sig": true,
	"static": true, "struct": true, "then": true, "to": true, "true": true,
	"try": true, "type": true, "upcast": true, "use": true, "val": true,
	"virtual": true, "void": true, "when": true, "while": true, "with": true,
	"yield": true,
}

var sqlKeywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true, "UPDATE": true,
	"DELETE": true, "CREATE": true, "DROP": true, "ALTER": true, "TABLE": true,
	"INDEX": true, "VIEW": true, "DATABASE": true, "SCHEMA": true, "JOIN": true,
	"LEFT": true, "RIGHT": true, "INNER": true, "OUTER": true, "FULL": true,
	"ON": true, "AND": true, "OR": true, "NOT": true, "IN": true, "NULL": true,
	"IS": true, "AS": true, "BY": true, "GROUP": true, "ORDER": true,
	"HAVING": true, "LIMIT": true, "OFFSET": true, "UNION": true, "ALL": true,
	"DISTINCT": true, "CASE": true, "WHEN": true, "THEN": true, "ELSE": true,
	"END": true, "BEGIN": true, "COMMIT": true, "ROLLBACK": true, "SET": true,
	"INTO": true, "VALUES": true, "RETURNING": true, "WITH": true,
}

var matlabKeywords = map[string]bool{
	"break": true, "case": true, "catch": true, "classdef": true, "continue": true,
	"else": true, "elseif": true, "end": true, "for": true, "function": true,
	"global": true, "if": true, "otherwise": true, "parfor": true, "persistent": true,
	"return": true, "spmd": true, "switch": true, "try": true, "while": true,
}

// ──────────────────────────────────────────────────────────────────────────────
// Language definitions
// ──────────────────────────────────────────────────────────────────────────────

// languages is the authoritative registry of all supported languages.
var languages = []LanguageConfig{
	{
		ID: "go", Name: "Go",
		Extensions:    []string{".go"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		// Go is analysed by the AST analyzer; these patterns are a fallback only.
		FunctionPatterns: []string{
			`func\s+(\w+)\s*\(`,
			`func\s+\(\w+\s+\*?\w+\)\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     goKeywords,
		DefaultTier:  TierAST,
	},
	{
		ID: "javascript", Name: "JavaScript",
		Extensions: []string{".js", ".mjs", ".cjs"},
		Shebangs:   []string{"node"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"},
		},
		FunctionPatterns: []string{
			`(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s*\*?\s+(\w+)\s*\(`,
			`(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     jsKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "typescript", Name: "TypeScript",
		Extensions: []string{".ts", ".mts", ".cts"},
		Shebangs:   []string{"ts-node"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"},
		},
		FunctionPatterns: []string{
			`(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s*\*?\s+(\w+)\s*\(`,
			`(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     tsKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "jsx", Name: "JavaScript (JSX)",
		Extensions:    []string{".jsx"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s*\*?\s+(\w+)\s*\(`,
			`(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     jsKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "tsx", Name: "TypeScript (TSX)",
		Extensions:    []string{".tsx"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s*\*?\s+(\w+)\s*\(`,
			`(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     tsKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "python", Name: "Python",
		Extensions: []string{".py", ".pyw"},
		Shebangs:   []string{"python", "python2", "python3"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "#"},
			{BlockStart: `"""`, BlockEnd: `"""`},
		},
		FunctionPatterns: []string{
			`(?:async\s+)?def\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     pythonKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "rust", Name: "Rust",
		Extensions:    []string{".rs"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:pub\s+)?(?:pub\s*\([^)]*\)\s+)?(?:async\s+)?(?:unsafe\s+)?fn\s+(\w+)\s*[<(]`,
		},
		TypePatterns: []string{
			`(?:pub\s+)?const\s+(\w+)\s*:`,
			`(?:pub\s+)?static\s+(?:mut\s+)?(\w+)\s*:`,
			`(?:pub\s+)?(?:struct|enum|trait|union|type)\s+(\w+)\s*[<{;]`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     rustKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "java", Name: "Java",
		Extensions:    []string{".java"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:(?:public|private|protected|static|final|synchronized|abstract|native|strictfp)\s+)+[\w<>\[\],\s]+\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     javaKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "kotlin", Name: "Kotlin",
		Extensions:    []string{".kt", ".kts"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:(?:private|public|internal|protected|open|override|inline|suspend|operator|infix|tailrec)\s+)*fun\s+(\w+)\s*[<(]`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     kotlinKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "swift", Name: "Swift",
		Extensions:    []string{".swift"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:(?:public|private|internal|open|fileprivate|override|static|class|mutating|nonmutating|final|required)\s+)*func\s+(\w+)\s*[<(]`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     swiftKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "csharp", Name: "C#",
		Extensions:    []string{".cs"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:(?:public|private|protected|internal|static|virtual|abstract|override|sealed|async|new|partial|extern|unsafe)\s+)+[\w<>\[\],\s]+\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     csharpKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "ruby", Name: "Ruby",
		Extensions: []string{".rb", ".rake", ".ru"},
		Shebangs:   []string{"ruby"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "#"},
			{BlockStart: "=begin", BlockEnd: "=end"},
		},
		FunctionPatterns: []string{
			// Ruby method names can end with ? or ! and may or may not have params
			`def\s+(\w+[?!]?)`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     rubyKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "php", Name: "PHP",
		Extensions: []string{".php", ".phtml"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "//"},
			{LinePrefix: "#"},
			{BlockStart: "/*", BlockEnd: "*/"},
		},
		FunctionPatterns: []string{
			`(?:(?:public|private|protected|static|abstract|final)\s+)*function\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     phpKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "c", Name: "C",
		Extensions:    []string{".c"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		// C function defs: return-type name(params) {
		FunctionPatterns: []string{`^\w[\w\s\*]*\s(\w+)\s*\([^)]*\)\s*\{`},
		CallPatterns:     []string{`\b(\w+)\s*\(`},
		Keywords:         cKeywords,
		DefaultTier:      TierRegex,
	},
	{
		ID: "cpp", Name: "C++",
		Extensions:    []string{".cpp", ".cc", ".cxx", ".c++"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`^\w[\w\s\*:<>]*\s(\w+)\s*\([^)]*\)\s*(?:const\s*)?\{`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     cppKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "header", Name: "C/C++ Header",
		Extensions:    []string{".h", ".hpp", ".hxx"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`^\w[\w\s\*:<>]*\s(\w+)\s*\([^)]*\)\s*;`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     cppKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "scala", Name: "Scala",
		Extensions:    []string{".scala", ".sc"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:(?:private|public|protected|override|implicit|final|abstract|lazy|sealed)\s+)*def\s+(\w+)\s*[(\[]`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     scalaKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "lua", Name: "Lua",
		Extensions: []string{".lua"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "--", BlockStart: "--[[", BlockEnd: "]]"},
		},
		FunctionPatterns: []string{
			`(?:local\s+)?function\s+(\w+)\s*\(`,
			`(\w+)\s*=\s*function\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     luaKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "shell", Name: "Shell/Bash",
		Extensions: []string{".sh", ".bash", ".zsh", ".ksh", ".fish"},
		Shebangs:   []string{"sh", "bash", "zsh", "fish"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "#"},
		},
		FunctionPatterns: []string{
			`(?:function\s+)?(\w+)\s*\(\s*\)\s*\{`,
		},
		CallPatterns: []string{`\b(\w+)\b`},
		Keywords:     shellKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "haskell", Name: "Haskell",
		Extensions: []string{".hs", ".lhs"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "--", BlockStart: "{-", BlockEnd: "-}"},
		},
		FunctionPatterns: []string{
			`^(\w+)\s+::`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     haskellKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "elixir", Name: "Elixir",
		Extensions:    []string{".ex", ".exs"},
		CommentStyles: []CommentStyle{{LinePrefix: "#"}},
		FunctionPatterns: []string{
			// def/defp, name may end with ? or !, followed by anything (params or do)
			`defp?\s+(\w+[?!]?)`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     elixirKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "r", Name: "R",
		Extensions:    []string{".r", ".R"},
		CommentStyles: []CommentStyle{{LinePrefix: "#"}},
		FunctionPatterns: []string{
			`(\w+)\s*<-\s*function\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     rKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "julia", Name: "Julia",
		Extensions: []string{".jl"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "#", BlockStart: "#=", BlockEnd: "=#"},
		},
		FunctionPatterns: []string{
			`function\s+(\w+)\s*\(`,
			`(\w+)\s*\([^)]*\)\s*=`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     juliaKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "dart", Name: "Dart",
		Extensions:    []string{".dart"},
		CommentStyles: []CommentStyle{{LinePrefix: "//", BlockStart: "/*", BlockEnd: "*/"}},
		FunctionPatterns: []string{
			`(?:(?:static|async|external|factory|const|final)\s+)*\w[\w<>\[\]? ]*\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     dartKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "perl", Name: "Perl",
		Extensions: []string{".pl", ".pm", ".perl"},
		Shebangs:   []string{"perl"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "#"},
		},
		FunctionPatterns: []string{`sub\s+(\w+)\s*[{(]`},
		CallPatterns:     []string{`\b(\w+)\s*\(`},
		Keywords:         perlKeywords,
		DefaultTier:      TierRegex,
	},
	{
		ID: "nim", Name: "Nim",
		Extensions:    []string{".nim"},
		CommentStyles: []CommentStyle{{LinePrefix: "#"}},
		FunctionPatterns: []string{
			`(?:proc|func|method|template|macro|iterator)\s+(\w+)\s*[(*]`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     nimKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "zig", Name: "Zig",
		Extensions:    []string{".zig"},
		CommentStyles: []CommentStyle{{LinePrefix: "//"}},
		FunctionPatterns: []string{
			`(?:pub\s+)?fn\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     zigKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "clojure", Name: "Clojure",
		Extensions:    []string{".clj", ".cljs", ".cljc", ".edn"},
		CommentStyles: []CommentStyle{{LinePrefix: ";"}},
		FunctionPatterns: []string{
			`\(defn-?\s+([\w\-?!.*+/<>=]+)\s`,
		},
		CallPatterns: []string{`\(([\w\-?!.*+/<>=]+)\s`},
		Keywords:     clojureKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "erlang", Name: "Erlang",
		Extensions:    []string{".erl", ".hrl"},
		CommentStyles: []CommentStyle{{LinePrefix: "%"}},
		FunctionPatterns: []string{
			`^(\w+)\s*\([^)]*\)\s*->`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     erlangKeywords,
		DefaultTier:  TierLexical,
	},
	{
		ID: "ocaml", Name: "OCaml",
		Extensions: []string{".ml", ".mli"},
		CommentStyles: []CommentStyle{
			{BlockStart: "(*", BlockEnd: "*)"},
		},
		FunctionPatterns: []string{
			`(?:let|and)\s+(?:rec\s+)?(\w+)\s+[^=]*=`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     ocamlKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "fsharp", Name: "F#",
		Extensions: []string{".fs", ".fsi", ".fsx"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "//", BlockStart: "(*", BlockEnd: "*)"},
		},
		FunctionPatterns: []string{
			`(?:let|and)\s+(?:rec\s+)?(?:inline\s+)?(\w+)\s+[^=]*=`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     fsharpKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "sql", Name: "SQL",
		Extensions: []string{".sql"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "--", BlockStart: "/*", BlockEnd: "*/"},
		},
		FunctionPatterns: []string{
			`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?(?:FUNCTION|PROCEDURE)\s+(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     sqlKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "matlab", Name: "MATLAB",
		Extensions:    []string{".m"},
		CommentStyles: []CommentStyle{{LinePrefix: "%"}},
		FunctionPatterns: []string{
			`function\s+(?:[\w,\[\]\s]*=\s*)?(\w+)\s*\(`,
		},
		CallPatterns: []string{`\b(\w+)\s*\(`},
		Keywords:     matlabKeywords,
		DefaultTier:  TierRegex,
	},
	{
		ID: "terraform", Name: "Terraform/HCL",
		Extensions: []string{".tf", ".tfvars"},
		CommentStyles: []CommentStyle{
			{LinePrefix: "//"},
			{LinePrefix: "#"},
			{BlockStart: "/*", BlockEnd: "*/"},
		},
		FunctionPatterns: []string{
			`resource\s+"[^"]+"\s+"(\w+)"\s*\{`,
			`module\s+"(\w+)"\s*\{`,
		},
		CallPatterns: []string{},
		Keywords:     map[string]bool{},
		DefaultTier:  TierRegex,
	},
}

// ──────────────────────────────────────────────────────────────────────────────
// Registry lookup helpers
// ──────────────────────────────────────────────────────────────────────────────

// extIndex maps file extension → language ID.  Built once at init time.
var extIndex = func() map[string]string {
	m := make(map[string]string, 128)
	for _, lc := range languages {
		for _, ext := range lc.Extensions {
			m[strings.ToLower(ext)] = lc.ID
		}
	}
	return m
}()

// langIndex maps language ID → LanguageConfig. Built once at init time.
var langIndex = func() map[string]LanguageConfig {
	m := make(map[string]LanguageConfig, len(languages))
	for _, lc := range languages {
		m[lc.ID] = lc
	}
	return m
}()

// allExtensions is the cached result of AllExtensions, computed at init time.
var allExtensions []string

func init() {
	allExtensions = make([]string, 0, len(extIndex))
	for ext := range extIndex {
		allExtensions = append(allExtensions, ext)
	}
}

// All returns the full list of registered language configs.
func All() []LanguageConfig {
	return languages
}

// Get returns the LanguageConfig for a given language ID.
// The second return value is false when the ID is not registered.
func Get(id string) (LanguageConfig, bool) {
	lc, ok := langIndex[id]
	return lc, ok
}

// AllExtensions returns every file extension that has a registered language.
func AllExtensions() []string {
	return allExtensions
}

// ──────────────────────────────────────────────────────────────────────────────
// Detection
// ──────────────────────────────────────────────────────────────────────────────

// Detect returns the best-matching LanguageConfig for filePath.
//
// Detection order:
//  1. File extension (fast, covers the vast majority of cases).
//  2. Shebang line in the file content.
//  3. Zero value (unknown language).
func Detect(filePath string) (LanguageConfig, bool) {
	// 1. Extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if id, ok := extIndex[ext]; ok {
		if lc, ok := Get(id); ok {
			return lc, true
		}
	}

	// 2. Shebang
	if lc, ok := detectByShebang(filePath); ok {
		return lc, true
	}

	return LanguageConfig{}, false
}

// detectByShebang opens the file and reads the first line to check for a
// shebang.  It matches against the Shebangs field of each language.
func detectByShebang(filePath string) (LanguageConfig, bool) {
	f, err := os.Open(filePath)
	if err != nil {
		return LanguageConfig{}, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return LanguageConfig{}, false
	}
	line := scanner.Text()

	if !strings.HasPrefix(line, "#!") {
		return LanguageConfig{}, false
	}

	// Extract the interpreter name, handling both:
	//   #!/usr/bin/python3
	//   #!/usr/bin/env python3
	parts := strings.Fields(line[2:]) // strip "#!"
	if len(parts) == 0 {
		return LanguageConfig{}, false
	}

	interpreter := filepath.Base(parts[0])
	if interpreter == "env" && len(parts) > 1 {
		interpreter = parts[1]
	}
	// Strip version suffix: "python3.11" → "python3" → match "python3"
	// Try exact first, then strip trailing digits
	for _, lc := range languages {
		for _, sb := range lc.Shebangs {
			if interpreter == sb ||
				strings.HasPrefix(interpreter, sb) {
				return lc, true
			}
		}
	}

	return LanguageConfig{}, false
}
