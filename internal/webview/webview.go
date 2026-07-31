// Package webview serves an interactive 3D web visualization of the derived
// knowledge graph. It embeds a single HTML page (which loads the 3d-force-graph
// renderer from a CDN at runtime) and exposes the graph as JSON. The package
// has no dependencies beyond the standard library.
package webview

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	graphpkg "github.com/TheShiveshNetwork/dizz/internal/graph"
)

//go:embed index.html
var assets embed.FS

// Handler returns the HTTP handler serving the visualization page and the
// graph data for the given derivation.
func Handler(g *graphpkg.Graph) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, err := assets.ReadFile("index.html")
		if err != nil {
			http.Error(w, "index not embedded", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/graph.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(g.DumpJSON())
	})
	return mux
}

// FreePort asks the OS for a currently-free TCP port.
func FreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// OpenBrowser opens a URL in the user's default browser, platform-agnostic.
// It returns an error only when the command could not be started.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("opening browser: %w", err)
	}
	return nil
}
