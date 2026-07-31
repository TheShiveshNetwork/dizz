package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

// MarshalTON serializes the graph in TON format for export. The graph itself
// is never stored — this is a lossless snapshot of the current derivation.
func (g *Graph) MarshalTON() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintln(&buf, "# graph")
	w := ton.NewWriter(&buf)
	w.WriteHeader("version", "nodes", "edges")
	w.WriteRecord("1", itoa(len(g.nodes)), itoa(len(g.edges)))

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "# nodes")
	nw := ton.NewWriter(&buf)
	nw.WriteHeader("id", "type", "label", "attrs", "rationale")
	for _, n := range g.Nodes() {
		nw.WriteRecord(n.ID, string(n.Type), n.Label, encodeAttrs(n.Attrs), encodeRationale(n.Rationale))
	}

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "# edges")
	ew := ton.NewWriter(&buf)
	ew.WriteHeader("id", "type", "from", "to", "weight", "attrs", "rationale")
	for _, e := range g.Edges() {
		ew.WriteRecord(e.ID, string(e.Type), e.From, e.To,
			strconv.FormatFloat(e.Weight, 'f', 4, 64), encodeAttrsStr(e.Attrs), encodeRationale(e.Rationale))
	}

	return buf.Bytes(), nil
}

func encodeAttrs(attrs map[string]interface{}) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+fmt.Sprint(attrs[k]))
	}
	return strings.Join(parts, ",")
}

func encodeAttrsStr(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+attrs[k])
	}
	return strings.Join(parts, ",")
}

func encodeRationale(r Rationale) string {
	b, _ := json.Marshal(r)
	return string(b)
}
