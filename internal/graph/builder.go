package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/signals"
	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/store"
)

// BuildOptions controls how much the graph derives. Options only affect the
// graph command itself — the underlying analysis pipeline is never invoked.
type BuildOptions struct {
	ProjectRoot        string
	IncludeCoChange    bool
	MinJaccard         float64
	CoChangeMinCommits int
	CoChangeMaxCommits int
	IncludeSnapshots   bool
	IncludeGuardrails  bool
	IncludeModules     bool
	IncludeTodos       bool
}

// DefaultBuildOptions returns the default derivation options.
func DefaultBuildOptions(projectRoot string) BuildOptions {
	return BuildOptions{
		ProjectRoot:        projectRoot,
		IncludeCoChange:    false,
		MinJaccard:         0.3,
		CoChangeMinCommits: 3,
		CoChangeMaxCommits: 1000,
		IncludeSnapshots:   true,
		IncludeGuardrails:  true,
		IncludeModules:     true,
		IncludeTodos:       true,
	}
}

// ErrNoState is returned when there is no persisted project state to derive
// the graph from. Running any dizz analysis command (dizz context, dizz log)
// once creates it.
var ErrNoState = fmt.Errorf("no project state found; run 'dizz context' or 'dizz log' once to build it")

type symbolRange struct {
	id      string
	line    int
	endLine int
	span    int
}

// Build derives the knowledge graph purely from persisted state: the project
// state file, the intent file, the per-file signal cache produced by normal
// analysis, snapshots, and configuration. No code analysis is performed here —
// the graph is a live, pure function of the state on disk. Re-running it after
// any analysis produces an up-to-date graph.
func Build(opts BuildOptions) (*Graph, error) {
	projectRoot := opts.ProjectRoot
	if !filepath.IsAbs(projectRoot) {
		abs, err := filepath.Abs(projectRoot)
		if err != nil {
			return nil, err
		}
		projectRoot = abs
	}
	projectRoot = filepath.Clean(projectRoot)
	trackDir := config.TrackDirPath(projectRoot)

	stateStore := store.NewStateStore(trackDir)
	ps, err := stateStore.LoadProjectState()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoState, err)
	}

	intentStore := store.NewIntentStore(trackDir)
	is, err := intentStore.LoadIntentState()
	if err != nil {
		is = state.NewIntentState()
	}

	cfg, _ := store.NewConfigStore(trackDir).LoadConfig()

	g := NewGraph()
	g.ProjectRoot = projectRoot

	// 1. Symbol and file nodes.
	rangesByFile := make(map[string][]symbolRange)
	for i := range ps.Symbols {
		sym := &ps.Symbols[i]
		rel := relPathOf(projectRoot, sym.File)
		addFileNode(g, rel, sym.Language, 0)
		symID := SymbolID(rel, sym.Name)
		attrs := map[string]interface{}{
			"file":        rel,
			"state":       string(sym.State),
			"type":        sym.Type,
			"language":    sym.Language,
			"line":        strconv.Itoa(sym.Line),
			"end_line":    strconv.Itoa(sym.EndLine),
			"churn":       strconv.Itoa(sym.ChurnCount),
			"instability": strconv.FormatFloat(sym.InstabilityScore, 'f', 3, 64),
			"is_called":   strconv.FormatBool(sym.IsCalled),
			"is_defined":  strconv.FormatBool(sym.IsDefined),
			"has_todo":    strconv.FormatBool(sym.HasTodo),
			"confidence":  strconv.FormatFloat(sym.Confidence, 'f', 2, 64),
		}
		if sym.SignalSource != "" {
			attrs["source"] = sym.SignalSource
		}
		if sym.IntentMarker != "" {
			attrs["marker"] = sym.IntentMarker
		}
		var lastTouched int64
		if sym.LastTouched != nil {
			lastTouched = sym.LastTouched.Unix()
			attrs["last_touched"] = sym.LastTouched.Format(time.RFC3339)
		}
		symNode := g.AddNode(&Node{
			ID:    symID,
			Type:  NodeSymbol,
			Label: sym.Name,
			Attrs: attrs,
			Rationale: Rationale{
				SourceTier: sym.SignalSource,
				Confidence: sym.Confidence,
				LineRange:  fmt.Sprintf("%d-%d", sym.Line, sym.EndLine),
				Evidence:   churnEvidence(sym.ChurnCount),
				Timestamp:  lastTouched,
				SourceType: "static_analysis",
			},
		})
		_ = symNode
		fileNode := FileID(rel)
		g.AddEdge(&Edge{
			Type:   EdgeDefinedIn,
			From:   symID,
			To:     fileNode,
			Weight: 1.0,
			Rationale: Rationale{
				SourceTier: sym.SignalSource,
				Confidence: 1.0,
				SourceType: "static_analysis",
			},
		})
		g.AddEdge(&Edge{
			Type:   EdgeContains,
			From:   fileNode,
			To:     symID,
			Weight: 1.0,
			Rationale: Rationale{
				SourceTier: sym.SignalSource,
				Confidence: 1.0,
				SourceType: "static_analysis",
			},
		})
		rangesByFile[rel] = append(rangesByFile[rel], symbolRange{
			id: symID, line: sym.Line, endLine: sym.EndLine, span: sym.EndLine - sym.Line,
		})
	}
	for rel := range rangesByFile {
		sort.Slice(rangesByFile[rel], func(i, j int) bool {
			return rangesByFile[rel][i].line < rangesByFile[rel][j].line
		})
	}

	// 2. Cached signals: CALLS and IMPORTS/DEPENDS_ON edges. These come from
	// the per-file cache populated by normal analysis — never a fresh scan.
	cacheDir := config.CacheDirPath(projectRoot)
	signalCache := store.NewSignalCache(projectRoot, cacheDir)
	_ = signalCache.LoadManifest()
	cached := signalCache.AllSignals()

	var allRelFiles []string
	seenFiles := make(map[string]bool)
	for rel := range cached {
		addFileNode(g, rel, "", 0)
		if !seenFiles[rel] {
			seenFiles[rel] = true
			allRelFiles = append(allRelFiles, rel)
		}
	}
	for _, fc := range ps.Files {
		rel := relPathOf(projectRoot, fc.Path)
		addFileNode(g, rel, fc.Language, fc.ChurnCount)
		if !seenFiles[rel] {
			seenFiles[rel] = true
			allRelFiles = append(allRelFiles, rel)
		}
	}
	sort.Strings(allRelFiles)
	resolver := NewImportResolver(projectRoot, allRelFiles)

	for rel, sigs := range cached {
		for _, sig := range sigs {
			switch sig.Type {
			case signals.FunctionCalled:
				addCallsEdge(g, rel, sig, rangesByFile)
			case signals.ImportFound:
				addImportEdge(g, rel, sig, resolver)
			}
		}
	}

	// 3. Intents + SCOPE_MATCH.
	addIntentNodes(g, is, allRelFiles)

	// 4. Tests.
	g.LinkTests(resolver, cached)

	// 5. Todos.
	if opts.IncludeTodos {
		addTodoNodes(g, ps, rangesByFile)
	}

	// 6. Snapshots.
	if opts.IncludeSnapshots {
		addSnapshotNodes(g, trackDir)
	}

	// 7. Guardrails.
	if opts.IncludeGuardrails && cfg != nil {
		addGuardrailNodes(g, cfg, allRelFiles)
	}

	// 8. Modules.
	if opts.IncludeModules {
		addModuleNode(g, projectRoot, allRelFiles)
	}

	// 9. Co-change coupling (opt-in; requires git).
	if opts.IncludeCoChange {
		addCoChangeEdges(g, projectRoot, opts)
	}

	return g, nil
}

// addCallsEdge creates a CALLS edge from the symbol enclosing the call site to
// every local symbol matching the called name, splitting confidence among
// ambiguous matches.
func addCallsEdge(g *Graph, rel string, sig signals.Signal, rangesByFile map[string][]symbolRange) {
	caller := enclosingSymbol(rangesByFile[rel], sig.Line)
	if caller == "" {
		return
	}
	targets := g.SymbolsNamed(sig.Name)
	if len(targets) == 0 {
		return
	}
	weight := sig.Confidence / float64(len(targets))
	for _, t := range targets {
		g.AddEdge(&Edge{
			Type:   EdgeCalls,
			From:   caller,
			To:     t.ID,
			Weight: weight,
			Attrs: map[string]string{
				"call_line": strconv.Itoa(sig.Line),
				"name":      sig.Name,
			},
			Rationale: Rationale{
				SourceTier: signalTier(sig),
				Confidence: weight,
				LineRange:  strconv.Itoa(sig.Line),
				Evidence:   "call at line " + strconv.Itoa(sig.Line),
				SourceType: "static_analysis",
			},
		})
	}
}

func enclosingSymbol(ranges []symbolRange, line int) string {
	best := ""
	bestSpan := -1
	for _, r := range ranges {
		if line >= r.line && line <= r.endLine {
			if bestSpan == -1 || r.span < bestSpan {
				best = r.id
				bestSpan = r.span
			}
		}
	}
	return best
}

// addImportEdge creates either an IMPORTS edge to a local file or a DEPENDS_ON
// edge to an external dependency node.
func addImportEdge(g *Graph, rel string, sig signals.Signal, resolver *ImportResolver) {
	fromFile := FileID(rel)
	imp := strings.Trim(sig.Name, `"`)
	if resolved, ok := resolver.Resolve(imp, rel); ok {
		g.AddEdge(&Edge{
			Type:   EdgeImports,
			From:   fromFile,
			To:     FileID(resolved),
			Weight: sig.Confidence,
			Attrs:  map[string]string{"import": imp},
			Rationale: Rationale{
				SourceTier: signalTier(sig),
				Confidence: sig.Confidence,
				Evidence:   "imports " + imp,
				SourceType: "static_analysis",
			},
		})
		return
	}
	depID := DepID(imp)
	g.AddNode(&Node{
		ID:    depID,
		Type:  NodeDep,
		Label: imp,
		Attrs: map[string]interface{}{"import_path": imp},
		Rationale: Rationale{
			Confidence: sig.Confidence,
			SourceType: "static_analysis",
		},
	})
	g.AddEdge(&Edge{
		Type:   EdgeDependsOn,
		From:   fromFile,
		To:     depID,
		Weight: sig.Confidence,
		Attrs:  map[string]string{"import": imp},
		Rationale: Rationale{
			SourceTier: signalTier(sig),
			Confidence: sig.Confidence,
			Evidence:   "depends on " + imp,
			SourceType: "static_analysis",
		},
	})
}

func signalTier(sig signals.Signal) string {
	if v, ok := sig.Metadata["source_tier"].(string); ok && v != "" {
		return v
	}
	return "regex"
}

// addIntentNodes adds intent nodes and SCOPE_MATCH edges to matching files.
func addIntentNodes(g *Graph, is *state.IntentState, allRelFiles []string) {
	if is == nil {
		return
	}
	for _, intent := range is.Intents {
		intentID := IntentID(intent.ID)
		g.AddNode(&Node{
			ID:    intentID,
			Type:  NodeIntent,
			Label: intent.Message,
			Attrs: map[string]interface{}{
				"id":         intent.ID,
				"type":       string(intent.Type),
				"severity":   strconv.Itoa(intent.Severity),
				"status":     string(intent.Status),
				"scope":      intent.Scope,
				"tags":       strings.Join(intent.Tags, ","),
				"created_by": intent.CreatedBy,
			},
			Rationale: Rationale{
				Confidence: 1.0,
				Timestamp:  intent.CreatedAt.Unix(),
				SourceType: "intent",
			},
		})
		for _, rel := range scopeMatchedFiles(intent.Scope, allRelFiles) {
			g.AddEdge(&Edge{
				Type:   EdgeScopeMatch,
				From:   intentID,
				To:     FileID(rel),
				Weight: 0.7,
				Attrs:  map[string]string{"scope": intent.Scope},
				Rationale: Rationale{
					Confidence: 0.7,
					Evidence:   "scope " + intent.Scope,
					SourceType: "intent",
				},
			})
		}
	}
}

// scopeMatchedFiles matches an intent scope against the project's files.
// Supported forms: exact path, file:line-line, path/with/* glob.
func scopeMatchedFiles(scope string, allRelFiles []string) []string {
	scope = strings.TrimSpace(scope)
	if scope == "" || scope == "project" {
		return nil
	}
	clean := scope
	if idx := strings.Index(scope, ":"); idx > 0 {
		clean = scope[:idx]
	}
	if clean == "" {
		return nil
	}
	clean = strings.TrimPrefix(clean, "./")
	var out []string
	for _, rel := range allRelFiles {
		if rel == clean {
			out = append(out, rel)
			continue
		}
		if strings.ContainsAny(clean, "*?") {
			if ok, _ := path.Match(clean, rel); ok {
				out = append(out, rel)
			}
		}
	}
	return out
}

// addTodoNodes adds todo nodes for unresolved todos and HAS_TODO edges from
// the file (and enclosing symbol) that contains them.
func addTodoNodes(g *Graph, ps *state.ProjectState, rangesByFile map[string][]symbolRange) {
	for i := range ps.Todos {
		todo := &ps.Todos[i]
		if todo.Resolved {
			continue
		}
		rel := relPathOf(g.ProjectRoot, todo.File)
		todoID := TodoID(rel, todo.Line)
		g.AddNode(&Node{
			ID:    todoID,
			Type:  NodeTodo,
			Label: todo.Text,
			Attrs: map[string]interface{}{
				"file": rel,
				"line": strconv.Itoa(todo.Line),
				"type": todo.Type,
			},
			Rationale: Rationale{
				Confidence: 1.0,
				SourceType: "static_analysis",
			},
		})
		g.AddEdge(&Edge{
			Type:   EdgeHasTodo,
			From:   FileID(rel),
			To:     todoID,
			Weight: 1.0,
			Rationale: Rationale{
				Confidence: 1.0,
				SourceType: "static_analysis",
			},
		})
		if owner := enclosingSymbol(rangesByFile[rel], todo.Line); owner != "" {
			g.AddEdge(&Edge{
				Type:   EdgeHasTodo,
				From:   owner,
				To:     todoID,
				Weight: 1.0,
				Rationale: Rationale{
					Confidence: 1.0,
					SourceType: "static_analysis",
				},
			})
		}
	}
}

// addSnapshotNodes adds snapshot nodes from .dizz/objects.
func addSnapshotNodes(g *Graph, trackDir string) {
	for _, info := range listSnapshots(trackDir) {
		g.AddNode(&Node{
			ID:    SnapshotID(info.hash),
			Type:  NodeSnapshot,
			Label: info.hash[:minLen(info.hash, 8)],
			Attrs: map[string]interface{}{
				"hash": info.hash,
				"size": strconv.FormatInt(info.size, 10),
			},
			Rationale: Rationale{
				Timestamp:  info.mtime.Unix(),
				SourceType: "snapshot",
			},
		})
	}
}

// addGuardrailNodes adds guardrail nodes and PROTECTS edges to matched files.
func addGuardrailNodes(g *Graph, cfg *config.Config, allRelFiles []string) {
	for _, gr := range cfg.Guardrails {
		id := gr.ID
		if id == "" {
			id = strings.Join(gr.Paths, ",") + ":" + string(gr.Action)
		}
		guardID := GuardrailID(id)
		g.AddNode(&Node{
			ID:    guardID,
			Type:  NodeGuardrail,
			Label: string(gr.Action),
			Attrs: map[string]interface{}{
				"action": string(gr.Action),
				"reason": gr.Reason,
				"paths":  strings.Join(gr.Paths, ","),
			},
			Rationale: Rationale{
				Confidence: 1.0,
				SourceType: "config",
			},
		})
		var targets []string
		if len(gr.Paths) == 0 {
			targets = allRelFiles
		} else {
			for _, rel := range allRelFiles {
				if matchesAnyGlob(rel, gr.Paths) {
					targets = append(targets, rel)
				}
			}
		}
		for _, rel := range targets {
			g.AddEdge(&Edge{
				Type:   EdgeProtects,
				From:   guardID,
				To:     FileID(rel),
				Weight: 1.0,
				Rationale: Rationale{
					Confidence: 1.0,
					Evidence:   gr.Reason,
					SourceType: "config",
				},
			})
		}
	}
}

// matchesAnyGlob reports whether a relative path matches any of the glob
// patterns, supporting the /** suffix common in path scopes.
func matchesAnyGlob(rel string, patterns []string) bool {
	for _, p := range patterns {
		if ok, _ := path.Match(p, rel); ok {
			return true
		}
		if strings.HasSuffix(p, "/**") {
			prefix := strings.TrimSuffix(p, "/**")
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				return true
			}
		}
	}
	return false
}

// detectModule reads the module/package name from a root manifest without any
// code analysis: go.mod, package.json or Cargo.toml.
func detectModule(projectRoot string) string {
	if data, err := os.ReadFile(filepath.Join(projectRoot, "go.mod")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if rest, ok := strings.CutPrefix(line, "module "); ok {
				return strings.TrimSpace(rest)
			}
		}
	}
	if data, err := os.ReadFile(filepath.Join(projectRoot, "package.json")); err == nil {
		var pkg struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(data, &pkg) == nil && pkg.Name != "" {
			return pkg.Name
		}
	}
	if data, err := os.ReadFile(filepath.Join(projectRoot, "Cargo.toml")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name =") {
				return strings.Trim(strings.TrimPrefix(line, "name ="), `"`)
			}
		}
	}
	return ""
}

// addModuleNode adds a module node from a root manifest (go.mod, package.json,
// Cargo.toml) and links it to all files with CONTAINS edges.
func addModuleNode(g *Graph, projectRoot string, allRelFiles []string) {
	modulePath := detectModule(projectRoot)
	if modulePath == "" || len(allRelFiles) == 0 {
		return
	}
	modID := ModuleID(modulePath)
	g.AddNode(&Node{
		ID:    modID,
		Type:  NodeModule,
		Label: modulePath,
		Attrs: map[string]interface{}{"path": modulePath},
		Rationale: Rationale{
			Confidence: 1.0,
			SourceType: "manifest",
		},
	})
	for _, rel := range allRelFiles {
		g.AddEdge(&Edge{
			Type:   EdgeContains,
			From:   modID,
			To:     FileID(rel),
			Weight: 1.0,
			Rationale: Rationale{
				Confidence: 1.0,
				SourceType: "manifest",
			},
		})
	}
}

// addCoChangeEdges adds hidden-coupling edges from git history, excluding
// pairs already connected by an import relationship.
func addCoChangeEdges(g *Graph, projectRoot string, opts BuildOptions) {
	commits, err := GitLogWithFiles(projectRoot, opts.CoChangeMaxCommits)
	if err != nil || len(commits) == 0 {
		return
	}
	relevant := make(map[string]bool)
	for _, rel := range g.fileRelPaths() {
		relevant[filepath.Join(projectRoot, rel)] = true
	}
	minJaccard := opts.MinJaccard
	if minJaccard <= 0 {
		minJaccard = 0.3
	}
	minCommits := opts.CoChangeMinCommits
	if minCommits <= 0 {
		minCommits = 3
	}
	for _, cc := range ComputeCoChanges(commits, relevant, minJaccard, minCommits) {
		relA := relPathOf(projectRoot, cc.FileA)
		relB := relPathOf(projectRoot, cc.FileB)
		idA, idB := FileID(relA), FileID(relB)
		if g.HasEdge(EdgeImports, idA, idB) || g.HasEdge(EdgeImports, idB, idA) {
			continue
		}
		if relA > relB {
			relA, relB = relB, relA
			idA, idB = idB, idA
		}
		attrs := map[string]string{
			"co_occurrences": strconv.Itoa(cc.CoOccurrences),
			"commits_f1":     strconv.Itoa(cc.CommitsA),
			"commits_f2":     strconv.Itoa(cc.CommitsB),
			"jaccard":        strconv.FormatFloat(cc.Jaccard, 'f', 3, 64),
		}
		rationale := Rationale{
			Confidence: cc.Jaccard,
			Evidence:   fmt.Sprintf("%d co-occurrences in %d commits", cc.CoOccurrences, cc.CommitsA+cc.CommitsB),
			SourceType: "git",
		}
		g.AddEdge(&Edge{
			Type: EdgeCoChangesWith, From: idA, To: idB,
			Weight: cc.Jaccard, Attrs: attrs, Rationale: rationale,
		})
		g.AddEdge(&Edge{
			Type: EdgeCoChangesWith, From: idB, To: idA,
			Weight: cc.Jaccard, Attrs: attrs, Rationale: rationale,
		})
	}
}

// fileRelPaths returns the relative paths of all file nodes in the graph.
func (g *Graph) fileRelPaths() []string {
	out := make([]string, 0, len(g.fileNodes))
	for rel := range g.fileNodes {
		out = append(out, rel)
	}
	sort.Strings(out)
	return out
}

// addFileNode creates a file node if missing and returns its ID.
func addFileNode(g *Graph, rel, language string, churn int) string {
	if existing := g.FileNode(rel); existing != nil {
		return existing.ID
	}
	id := FileID(rel)
	attrs := map[string]interface{}{"path": rel}
	if language != "" {
		attrs["language"] = language
	}
	if churn > 0 {
		attrs["churn"] = strconv.Itoa(churn)
	}
	g.AddNode(&Node{
		ID:    id,
		Type:  NodeFile,
		Label: filepath.Base(rel),
		Attrs: attrs,
		Rationale: Rationale{
			Confidence: 1.0,
			SourceType: "static_analysis",
		},
	})
	return id
}

// relPathOf converts an absolute (or already relative) file path to a
// project-relative, forward-slash path.
func relPathOf(projectRoot, file string) string {
	if filepath.IsAbs(file) {
		if rel, err := filepath.Rel(projectRoot, file); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(filepath.Clean(file))
}

func churnEvidence(churn int) string {
	if churn > 0 {
		return fmt.Sprintf("%d commits", churn)
	}
	return ""
}

func minLen(s string, n int) int {
	if len(s) < n {
		return len(s)
	}
	return n
}

type snapshotInfo struct {
	hash  string
	mtime time.Time
	size  int64
}

// listSnapshots returns the snapshot hashes present in .dizz/objects.
func listSnapshots(trackDir string) []snapshotInfo {
	objectsDir := config.ObjectsDirPath(trackDir)
	subdirs, err := os.ReadDir(objectsDir)
	if err != nil {
		return nil
	}
	var out []snapshotInfo
	for _, sub := range subdirs {
		if !sub.IsDir() {
			continue
		}
		files, err := os.ReadDir(filepath.Join(objectsDir, sub.Name()))
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".delta") {
				continue
			}
			ext := filepath.Ext(name)
			hash := sub.Name() + strings.TrimSuffix(name, ext)
			info, err := f.Info()
			if err != nil {
				continue
			}
			out = append(out, snapshotInfo{hash: hash, mtime: info.ModTime(), size: info.Size()})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].hash < out[j].hash })
	return out
}
