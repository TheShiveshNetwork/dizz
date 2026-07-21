package state

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TheShiveshNetwork/dizz/internal/store/ton"
)

const (
	intentTONColumns = "id|type|sev|status|msg|scope|tags|created_by|resolution"
	intentTONPrefix  = "# intents"
	intentTONVersion = "1.0"
)

func (is *IntentState) MarshalTON() ([]byte, error) {
	var buf bytes.Buffer
	w := ton.NewWriter(&buf)

	w.WriteHeader("id", "type", "sev", "status", "msg", "scope", "tags", "created_by", "resolution")

	for _, intent := range is.Intents {
		resolution := ""
		if intent.Resolution != nil {
			resolution = fmt.Sprintf("%s:%s:%d:%s",
				intent.Resolution.Method,
				intent.Resolution.Description,
				intent.Resolution.ResolvedAt.Unix(),
				intent.Resolution.ResolvedBy,
			)
		}

		tags := strings.Join(intent.Tags, ",")

		w.WriteRecord(
			intent.ID,
			string(intent.Type),
			strconv.Itoa(intent.Severity),
			string(intent.Status),
			intent.Message,
			intent.Scope,
			tags,
			intent.CreatedBy,
			resolution,
		)
	}

	return buf.Bytes(), nil
}

func UnmarshalIntentStateTON(data []byte) (*IntentState, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("ton: cannot unmarshal empty intent state")
	}

	r, err := ton.NewReader(data)
	if err != nil {
		return nil, fmt.Errorf("ton: %w", err)
	}

	header := r.Header()
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[h] = i
	}

	required := []string{"id", "type", "sev", "status", "msg"}
	for _, col := range required {
		if _, ok := colMap[col]; !ok {
			return nil, fmt.Errorf("ton: missing required column %q in header %v", col, header)
		}
	}

	is := NewIntentState()

	for {
		rec, ok := r.Next()
		if !ok {
			break
		}

		intent := Intent{
			ID:        getCol(rec, colMap, "id"),
			Message:   getCol(rec, colMap, "msg"),
			Scope:     getCol(rec, colMap, "scope"),
			CreatedBy: getCol(rec, colMap, "created_by"),
		}

		intent.Type = IntentType(getCol(rec, colMap, "type"))

		sevStr := getCol(rec, colMap, "sev")
		if sev, err := strconv.Atoi(sevStr); err == nil {
			intent.Severity = sev
		}

		intent.Status = IntentStatus(getCol(rec, colMap, "status"))

		tagsStr := getCol(rec, colMap, "tags")
		if tagsStr != "" {
			intent.Tags = strings.Split(tagsStr, ",")
		}

		resolutionStr := getCol(rec, colMap, "resolution")
		if resolutionStr != "" {
			intent.Resolution = parseResolution(resolutionStr)
		}

		intent.Confidence = 0.8
		intent.CreatedAt = time.Now()
		intent.UpdatedAt = time.Now()

		is.Intents = append(is.Intents, intent)
	}

	return is, nil
}

func getCol(rec []string, colMap map[string]int, col string) string {
	idx, ok := colMap[col]
	if !ok || idx >= len(rec) {
		return ""
	}
	return rec[idx]
}

func parseResolution(s string) *Resolution {
	parts := strings.SplitN(s, ":", 4)
	if len(parts) < 4 {
		return nil
	}

	var resolvedAt time.Time
	if unixSec, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
		resolvedAt = time.Unix(unixSec, 0)
	} else {
		resolvedAt = time.Now()
	}

	return &Resolution{
		Method:      parts[0],
		Description: parts[1],
		ResolvedAt:  resolvedAt,
		ResolvedBy:  parts[3],
	}
}
