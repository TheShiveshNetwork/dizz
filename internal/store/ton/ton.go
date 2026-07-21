package ton

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Writer serializes structured data as TON (Token-Optimized Notation).
// TON is a line-oriented, pipe-delimited format optimized for
// token-efficient storage and agent consumption.
type Writer struct {
	w             io.Writer
	headerWritten bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) WriteHeader(fields ...string) {
	if w.headerWritten {
		panic("ton: header already written")
	}
	fmt.Fprintln(w.w, encodeLine(fields...))
	w.headerWritten = true
}

func (w *Writer) WriteRecord(fields ...string) {
	fmt.Fprintln(w.w, encodeLine(fields...))
}

func (w *Writer) Flush() {
	if f, ok := w.w.(interface{ Flush() error }); ok {
		f.Flush()
	}
}

func encodeLine(fields ...string) string {
	escaped := make([]string, len(fields))
	for i, f := range fields {
		escaped[i] = escape(f)
	}
	return strings.Join(escaped, "|")
}

// Reader deserializes TON data.
// Comments (lines starting with #) and empty lines are skipped.
// The first non-comment, non-empty line is treated as the header.
type Reader struct {
	lines  [][]string
	pos    int
	header []string
}

func NewReader(data []byte) (*Reader, error) {
	rawLines := strings.Split(string(data), "\n")

	var contentLines []string
	for _, line := range rawLines {
		line = strings.TrimRight(line, "\r")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		contentLines = append(contentLines, line)
	}

	if len(contentLines) == 0 {
		return nil, fmt.Errorf("ton: empty or comment-only input")
	}

	header := parseLine(contentLines[0])
	lines := make([][]string, len(contentLines)-1)
	for i, line := range contentLines[1:] {
		lines[i] = parseLine(line)
	}

	return &Reader{
		lines:  lines,
		pos:    0,
		header: header,
	}, nil
}

func (r *Reader) Header() []string {
	return r.header
}

func (r *Reader) Next() ([]string, bool) {
	if r.pos >= len(r.lines) {
		return nil, false
	}
	record := r.lines[r.pos]
	r.pos++
	return record, true
}

func parseLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return []string{""}
	}

	var fields []string
	var current strings.Builder
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '\\' && i+1 < len(line) {
			next := line[i+1]
			switch next {
			case 'n':
				current.WriteByte('\n')
				i++
			case '|':
				current.WriteByte('|')
				i++
			case '\\':
				current.WriteByte('\\')
				i++
			default:
				current.WriteByte(ch)
			}
		} else if ch == '|' {
			fields = append(fields, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	fields = append(fields, strings.TrimSpace(current.String()))
	return fields
}

func escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '|':
			b.WriteString(`\|`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// BytesWriter is a convenience wrapper that writes to a bytes.Buffer.
type BytesWriter struct {
	*Writer
	buf bytes.Buffer
}

func NewBytesWriter() *BytesWriter {
	bw := &BytesWriter{}
	bw.Writer = NewWriter(&bw.buf)
	return bw
}

func (bw *BytesWriter) Bytes() []byte {
	return bw.buf.Bytes()
}

func (bw *BytesWriter) String() string {
	return bw.buf.String()
}
