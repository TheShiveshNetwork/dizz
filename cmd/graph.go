package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	graphpkg "github.com/TheShiveshNetwork/dizz/internal/graph"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/spf13/cobra"
)

var (
	graphCoChange   bool
	graphMinJaccard float64
	graphMinCommits int
	graphMaxCommits int
	graphDepth      int
	graphJSON       bool
	graphFile       string
	graphPort       int
	graphOpen       bool
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Query the derived project knowledge graph",
	Long: `Derives and queries a knowledge graph purely from persisted dizz state
(state.ton.gz, intent.ton, the per-file signal cache, snapshots, config) and,
opt-in, git history. No code analysis is ever performed here, and the graph is
never materialized: every invocation re-derives it in-memory so it is always
live. Run 'dizz context' or 'dizz log' at least once so state exists.`,
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.PersistentFlags().BoolVar(&graphCoChange, "cochange", false, "Include co-change coupling analysis (requires git)")
	graphCmd.PersistentFlags().Float64Var(&graphMinJaccard, "min-jaccard", 0.3, "Minimum Jaccard similarity for co-change edges")
	graphCmd.PersistentFlags().IntVar(&graphMinCommits, "min-commits", 3, "Minimum commits a file must appear in for co-change analysis")
	graphCmd.PersistentFlags().IntVar(&graphMaxCommits, "max-commits", 1000, "Maximum git history depth for co-change analysis")
	graphCmd.PersistentFlags().IntVar(&graphDepth, "depth", 3, "Traversal depth for blast radius")
	graphCmd.PersistentFlags().BoolVar(&graphJSON, "json", false, "Emit machine-readable JSON")
	graphCmd.PersistentFlags().StringVar(&graphFile, "file", "", "Disambiguate a symbol by its file path")
	graphCmd.PersistentFlags().IntVar(&graphPort, "port", 0, "Port for the web visualizer (0 picks a free port)")
	graphCmd.PersistentFlags().BoolVar(&graphOpen, "open", true, "Open the web visualizer in the default browser")

	graphCmd.AddCommand(graphBuildCmd)
	graphCmd.AddCommand(graphStatsCmd)
	graphCmd.AddCommand(graphQueryCmd)
	graphCmd.AddCommand(graphTraceCmd)
	graphCmd.AddCommand(graphCoChangesCmd)
	graphCmd.AddCommand(graphTestsCmd)
	graphCmd.AddCommand(graphPathCmd)
	graphCmd.AddCommand(graphScopeCmd)
	graphCmd.AddCommand(graphTONCmd)
	graphCmd.AddCommand(graphDumpCmd)
	graphCmd.AddCommand(graphVisualizeCmd)
}

// loadGraph derives the graph for the current project. Co-change analysis is
// opt-in because it is the only step that touches git.
func loadGraph(includeCoChange bool) (*graphpkg.Graph, *graphpkg.QueryEngine, error) {
	projectRoot, err := commonPkg.FindProjectRoot()
	if err != nil {
		return nil, nil, err
	}
	opts := graphpkg.DefaultBuildOptions(projectRoot)
	opts.IncludeCoChange = includeCoChange
	opts.MinJaccard = graphMinJaccard
	opts.CoChangeMinCommits = graphMinCommits
	opts.CoChangeMaxCommits = graphMaxCommits
	g, err := graphpkg.Build(opts)
	if err != nil {
		return nil, nil, err
	}
	return g, graphpkg.NewQueryEngine(g), nil
}

// resolve loads the graph and resolves an entity specifier to a node.
func resolve(includeCoChange bool, term string) (*graphpkg.Node, *graphpkg.QueryEngine, error) {
	_, qe, err := loadGraph(includeCoChange)
	if err != nil {
		return nil, nil, err
	}
	if graphFile != "" && !strings.Contains(term, "@") && !strings.Contains(term, "::") {
		term = "symbol:" + term + "@" + graphFile
	}
	node, err := qe.ResolveEntity(term)
	if err != nil {
		return nil, nil, err
	}
	return node, qe, nil
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

var graphBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Derive the graph from persisted state",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, _, err := loadGraph(graphCoChange)
		if err != nil {
			return err
		}
		stats := g.ComputeStats()
		if graphJSON {
			return printJSON(stats)
		}
		printStats(stats)
		return nil
	},
}

var graphStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Summarize graph shape",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, _, err := loadGraph(graphCoChange)
		if err != nil {
			return err
		}
		stats := g.ComputeStats()
		if graphJSON {
			return printJSON(stats)
		}
		printStats(stats)
		return nil
	},
}

var graphQueryCmd = &cobra.Command{
	Use:   "query <entity>",
	Short: "Blast radius of changing an entity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		node, qe, err := resolve(graphCoChange, args[0])
		if err != nil {
			return err
		}
		affected := qe.BlastRadius(node.ID, graphDepth)
		if graphJSON {
			return printJSON(affected)
		}
		fmt.Fprintf(os.Stdout, "%s\n", ui.Header(fmt.Sprintf("blast radius of %s (depth %d)", node.ID, graphDepth)))
		if len(affected) == 0 {
			fmt.Fprintln(os.Stdout, ui.Muted("  no entities affected"))
			return nil
		}
		files := map[string]bool{}
		for _, a := range affected {
			if a.Node != nil && a.Node.Type == graphpkg.NodeFile {
				files[a.Node.ID] = true
			}
			if a.Node != nil && a.Node.Type == graphpkg.NodeSymbol {
				if f := a.Node.Attr("file"); f != "" {
					files["file:"+f] = true
				}
			}
		}
		fmt.Fprintf(os.Stdout, "%s\n", ui.Muted(fmt.Sprintf("  %d entities affected across %d files", len(affected), len(files))))
		for _, a := range affected {
			via := make([]string, 0, len(a.Via))
			for _, v := range a.Via {
				via = append(via, string(v))
			}
			fmt.Fprintf(os.Stdout, "  %s score=%.3f depth=%d via=%s\n",
				ui.Highlight(a.NodeID), a.Score, a.Depth, strings.Join(via, ","))
		}
		return nil
	},
}

var graphTraceCmd = &cobra.Command{
	Use:   "trace <entity>",
	Short: "Full trace of an entity: callers, callees, tests, intents, co-changes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		node, qe, err := resolve(graphCoChange, args[0])
		if err != nil {
			return err
		}
		tr := qe.Trace(node.ID)
		if graphJSON {
			return printJSON(tr)
		}
		fmt.Fprintf(os.Stdout, "%s (%s)\n", ui.Header(node.ID), node.Label)
		if tr.DefinedIn != nil {
			fmt.Fprintf(os.Stdout, "  defined_in: %s\n", tr.DefinedIn.To)
		}
		printRefs("  callers", tr.Callers)
		printRefs("  callees", tr.Callees)
		printRefs("  imports", tr.Imports)
		printRefs("  imported_by", tr.ImportedBy)
		printRefs("  tests", tr.Tests)
		printRefs("  tested", tr.Tested)
		printRefs("  intents", tr.Intents)
		printRefs("  todos", tr.Todo)
		printRefs("  co_changes", tr.CoChanges)
		printRefs("  contains", tr.Contains)
		printRefs("  protected_by", tr.Protects)
		printRefs("  protects", tr.Protected)
		return nil
	},
}

func printRefs(label string, refs []graphpkg.EdgeRef) {
	if len(refs) == 0 {
		return
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].From+refs[i].To < refs[j].From+refs[j].To })
	fmt.Fprintf(os.Stdout, "%s (%d):\n", ui.Info(label), len(refs))
	for _, r := range refs {
		weight := ""
		if r.Weight != 1.0 {
			weight = fmt.Sprintf(" (%.3f)", r.Weight)
		}
		if r.Attrs != nil {
			if v, ok := r.Attrs["match_method"]; ok {
				weight += " [" + v + "]"
			}
		}
		fmt.Fprintf(os.Stdout, "    %s -> %s%s\n", r.From, r.To, weight)
	}
}

var graphCoChangesCmd = &cobra.Command{
	Use:   "cochanges <entity>",
	Short: "Files that historically change together with the entity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		node, qe, err := resolve(true, args[0])
		if err != nil {
			return err
		}
		refs := qe.CoChanges(node.ID, graphMinJaccard)
		if graphJSON {
			return printJSON(refs)
		}
		fmt.Fprintf(os.Stdout, "%s\n", ui.Header(fmt.Sprintf("co-change coupling for %s (min-jaccard %.2f)", node.ID, graphMinJaccard)))
		if len(refs) == 0 {
			fmt.Fprintln(os.Stdout, ui.Muted("  no hidden coupling found"))
			return nil
		}
		seen := map[string]bool{}
		for _, r := range refs {
			counterpart := r.From
			if counterpart == node.ID {
				counterpart = r.To
			}
			if seen[counterpart] {
				continue
			}
			seen[counterpart] = true
			fmt.Fprintf(os.Stdout, "  %s jaccard=%.3f %s\n",
				ui.Highlight(counterpart), r.Weight, ui.Muted(r.Rationale.Evidence))
		}
		return nil
	},
}

var graphTestsCmd = &cobra.Command{
	Use:   "tests <entity>",
	Short: "Which tests cover a given symbol or file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		node, qe, err := resolve(graphCoChange, args[0])
		if err != nil {
			return err
		}
		refs := qe.TestCoverage(node.ID)
		if graphJSON {
			return printJSON(refs)
		}
		fmt.Fprintf(os.Stdout, "%s\n", ui.Header(fmt.Sprintf("test coverage for %s", node.ID)))
		if len(refs) == 0 {
			fmt.Fprintln(os.Stdout, ui.Muted("  no tests found"))
			return nil
		}
		for _, r := range refs {
			method := ""
			if r.Attrs != nil {
				method = r.Attrs["match_method"]
			}
			fmt.Fprintf(os.Stdout, "  %s confidence=%.1f method=%s\n", ui.Highlight(r.From), r.Weight, method)
		}
		return nil
	},
}

var graphPathCmd = &cobra.Command{
	Use:   "path <from> <to>",
	Short: "Shortest path between two entities",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, qe, err := loadGraph(graphCoChange)
		if err != nil {
			return err
		}
		from, err := qe.ResolveEntity(args[0])
		if err != nil {
			return err
		}
		to, err := qe.ResolveEntity(args[1])
		if err != nil {
			return err
		}
		path := qe.ShortestPath(from.ID, to.ID)
		if graphJSON {
			return printJSON(path)
		}
		if path == nil {
			fmt.Fprintln(os.Stdout, ui.Muted("no path found"))
			return nil
		}
		for _, id := range path {
			fmt.Fprintln(os.Stdout, "  "+id)
		}
		return nil
	},
}

var graphScopeCmd = &cobra.Command{
	Use:   "scope <glob>",
	Short: "Files matching a path glob (e.g. internal/auth/**)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, qe, err := loadGraph(graphCoChange)
		if err != nil {
			return err
		}
		nodes := qe.ImpactScope(args[0])
		if graphJSON {
			return printJSON(nodes)
		}
		if len(nodes) == 0 {
			fmt.Fprintln(os.Stdout, ui.Muted("no files match "+args[0]))
			return nil
		}
		for _, n := range nodes {
			fmt.Fprintln(os.Stdout, "  "+n.ID)
		}
		return nil
	},
}

var graphTONCmd = &cobra.Command{
	Use:   "ton",
	Short: "Dump the full graph in TON format",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, _, err := loadGraph(graphCoChange)
		if err != nil {
			return err
		}
		data, err := g.MarshalTON()
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(data)
		return err
	},
}

var graphDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump the graph as JSON for visualizers",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, _, err := loadGraph(graphCoChange)
		if err != nil {
			return err
		}
		return printJSON(g.DumpJSON())
	},
}

var graphVisualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Serve an interactive 3D web view of the graph",
	Long:  visualizeLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVisualize(graphCoChange, graphPort, graphOpen)
	},
}

func printStats(s graphpkg.Stats) {
	fmt.Fprintf(os.Stdout, "%s\n", ui.Header(fmt.Sprintf("graph: %d nodes, %d edges", s.Nodes, s.Edges)))
	ntypes := make([]string, 0, len(s.ByNodeType))
	for t, c := range s.ByNodeType {
		ntypes = append(ntypes, fmt.Sprintf("%s=%d", t, c))
	}
	sort.Strings(ntypes)
	if len(ntypes) > 0 {
		fmt.Fprintf(os.Stdout, "  %s\n", ui.Info("by_node_type: "+strings.Join(ntypes, ", ")))
	}
	etypes := make([]string, 0, len(s.ByEdgeType))
	for t, c := range s.ByEdgeType {
		etypes = append(etypes, fmt.Sprintf("%s=%d", t, c))
	}
	sort.Strings(etypes)
	if len(etypes) > 0 {
		fmt.Fprintf(os.Stdout, "  %s\n", ui.Info("by_edge_type: "+strings.Join(etypes, ", ")))
	}
	fmt.Fprintf(os.Stdout, "  unused_deps: %d\n", s.UnusedDeps)
	if s.HasCoChange {
		fmt.Fprintln(os.Stdout, "  co_change: included")
	} else {
		fmt.Fprintln(os.Stdout, ui.Muted("  co_change: not included (use --cochange)"))
	}
}
