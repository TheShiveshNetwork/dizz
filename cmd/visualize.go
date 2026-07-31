package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/TheShiveshNetwork/dizz/internal/ui"
	"github.com/TheShiveshNetwork/dizz/internal/webview"
	"github.com/spf13/cobra"
)

const visualizeLong = `Starts a local HTTP server rendering the graph as an interactive 3D
force-directed view and opens it in the default browser. The visualization is
served on 127.0.0.1; --port selects the port (0 picks a free one) and --open
controls whether the browser is launched automatically.`

// runVisualize is the single implementation behind both `dizz visualize` and
// `dizz graph visualize`. Changes made here apply to both commands.
func runVisualize(includeCoChange bool, port int, open bool) error {
	g, _, err := loadGraph(includeCoChange)
	if err != nil {
		return err
	}
	if port == 0 {
		port = webview.FreePort()
	}
	if port == 0 {
		return fmt.Errorf("no free port available")
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)
	if open {
		_ = webview.OpenBrowser(url)
	}
	fmt.Fprintf(os.Stdout, "%s\n", ui.Header(fmt.Sprintf("dizz graph: %d nodes, %d edges", len(g.Nodes()), len(g.Edges()))))
	fmt.Fprintf(os.Stdout, "  serving at %s (Ctrl+C to stop)\n", url)
	server := &http.Server{Addr: addr, Handler: webview.Handler(g)}
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Serve an interactive 3D web view of the graph",
	Long:  visualizeLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVisualize(graphCoChange, graphPort, graphOpen)
	},
}

func init() {
	rootCmd.AddCommand(visualizeCmd)
	visualizeCmd.Flags().BoolVar(&graphCoChange, "cochange", false, "Include co-change coupling analysis (requires git)")
	visualizeCmd.Flags().IntVar(&graphPort, "port", 0, "Port for the web visualizer (0 picks a free port)")
	visualizeCmd.Flags().BoolVar(&graphOpen, "open", true, "Open the web visualizer in the default browser")
	visualizeCmd.Flags().Float64Var(&graphSimThreshold, "similarity-threshold", 0.4, "Minimum text similarity for RELATED_TO edges")
	visualizeCmd.Flags().IntVar(&graphSimTopK, "similarity-topk", 6, "Max RELATED_TO edges per intent")
}
