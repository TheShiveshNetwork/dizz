package graph

import (
	"path/filepath"
	"strings"
)

// ImportResolver resolves import paths to local files. It is purely derived:
// it matches import strings against the set of files already present in the
// graph/state, so it never reads or analyzes source code on its own.
type ImportResolver struct {
	projectRoot string
	dirSet      map[string]bool     // relpath of directories containing a file
	byStem      map[string][]string // relpath without extension of every file
	byBasename  map[string][]string
}

// NewImportResolver builds a resolver over the given relative file paths.
func NewImportResolver(projectRoot string, relFiles []string) *ImportResolver {
	r := &ImportResolver{
		projectRoot: projectRoot,
		dirSet:      make(map[string]bool),
		byStem:      make(map[string][]string),
		byBasename:  make(map[string][]string),
	}
	for _, rel := range relFiles {
		dir := filepath.ToSlash(filepath.Dir(rel))
		if dir != "." {
			r.dirSet[dir] = true
		}
		stem := trimExt(rel)
		r.byStem[stem] = append(r.byStem[stem], rel)
		base := filepath.Base(rel)
		r.byBasename[base] = append(r.byBasename[base], rel)
		r.byBasename[trimExt(base)] = append(r.byBasename[trimExt(base)], rel)
	}
	return r
}

// Resolve maps an import path to a local relative file path.
// Returns ok=false when the import cannot be attributed to a local file (i.e.
// it is an external dependency).
func (r *ImportResolver) Resolve(importPath, fromFile string) (string, bool) {
	imp := strings.Trim(importPath, `"'`)
	if imp == "" {
		return "", false
	}
	imp = filepath.ToSlash(imp)
	fromDir := filepath.ToSlash(filepath.Dir(fromFile))

	// 1. Relative imports: "./foo", "../shared/foo".
	if strings.HasPrefix(imp, "./") || strings.HasPrefix(imp, "../") {
		joined := filepath.ToSlash(filepath.Clean(filepath.Join(fromDir, imp)))
		if rel, ok := r.matchPath(joined); ok {
			return rel, true
		}
		return "", false
	}

	// 2. Absolute filesystem imports.
	if filepath.IsAbs(imp) {
		if rel, err := filepath.Rel(r.projectRoot, imp); err == nil && !strings.HasPrefix(rel, "..") {
			if rel, ok := r.matchPath(filepath.ToSlash(rel)); ok {
				return rel, true
			}
		}
		return "", false
	}

	// 3. Module-style imports: try directory suffix match first, then the
	//    longest file-path suffix match.
	parts := strings.Split(imp, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.Join(parts[i:], "/")
		if r.dirSet[candidate] {
			if rel, ok := r.pickFileInDir(candidate); ok {
				return rel, true
			}
		}
		if rels := r.byStem[candidate]; len(rels) > 0 {
			return rels[0], true
		}
	}

	// 4. Bare name matching a known file basename (e.g. "store").
	if base := parts[len(parts)-1]; base != "" {
		if rels := r.byBasename[base]; len(rels) > 0 {
			return rels[0], true
		}
	}

	return "", false
}

// matchPath resolves a cleaned path to an existing file, checking both the
// exact path and directory-level imports (importing "internal/store" links to
// the first file in that directory).
func (r *ImportResolver) matchPath(path string) (string, bool) {
	if rels := r.byStem[path]; len(rels) > 0 {
		return rels[0], true
	}
	if r.dirSet[path] {
		return r.pickFileInDir(path)
	}
	return "", false
}

// pickFileInDir returns the first known file inside the given relative dir.
func (r *ImportResolver) pickFileInDir(dir string) (string, bool) {
	best := ""
	for _, rels := range r.byBasename {
		for _, rel := range rels {
			if filepath.ToSlash(filepath.Dir(rel)) == dir {
				if best == "" || rel < best {
					best = rel
				}
			}
		}
	}
	if best != "" {
		return best, true
	}
	return "", false
}

func trimExt(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path
	}
	return strings.TrimSuffix(path, ext)
}
