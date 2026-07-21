package ton

import (
	"bytes"
	"testing"
)

func TestWriterWriteHeader(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("id", "name", "value")
	w.Flush()

	expected := "id|name|value\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWriterWriteRecord(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("id", "name")
	w.WriteRecord("1", "foo")
	w.WriteRecord("2", "bar")
	w.Flush()

	expected := "id|name\n1|foo\n2|bar\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("id", "name", "value")
	w.WriteRecord("1", "foo", "bar")
	w.WriteRecord("2", "baz", "qux")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	header := r.Header()
	if len(header) != 3 || header[0] != "id" || header[1] != "name" || header[2] != "value" {
		t.Fatalf("unexpected header: %v", header)
	}

	rec1, ok := r.Next()
	if !ok {
		t.Fatal("expected first record")
	}
	if len(rec1) != 3 || rec1[0] != "1" || rec1[1] != "foo" || rec1[2] != "bar" {
		t.Fatalf("unexpected record 1: %v", rec1)
	}

	rec2, ok := r.Next()
	if !ok {
		t.Fatal("expected second record")
	}
	if len(rec2) != 3 || rec2[0] != "2" || rec2[1] != "baz" || rec2[2] != "qux" {
		t.Fatalf("unexpected record 2: %v", rec2)
	}

	_, ok = r.Next()
	if ok {
		t.Fatal("expected no more records")
	}
}

func TestEscapePipe(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("msg")
	w.WriteRecord("hello|world|test")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if rec[0] != "hello|world|test" {
		t.Fatalf("unexpected value: %q", rec[0])
	}
}

func TestEscapeBackslash(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("path")
	w.WriteRecord(`C:\Users\test`)
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if rec[0] != `C:\Users\test` {
		t.Fatalf("unexpected value: %q", rec[0])
	}
}

func TestEscapeNewline(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("desc")
	w.WriteRecord("line1\nline2")
	w.Flush()

	content := buf.String()
	if !bytes.Contains([]byte(content), []byte(`\n`)) {
		t.Fatalf("expected escaped newline in output, got: %q", content)
	}

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if rec[0] != "line1\nline2" {
		t.Fatalf("unexpected value: %q", rec[0])
	}
}

func TestEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("a", "b", "c")
	w.WriteRecord("x", "", "z")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if len(rec) != 3 || rec[0] != "x" || rec[1] != "" || rec[2] != "z" {
		t.Fatalf("unexpected record: %v", rec)
	}
}

func TestTrailingEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("a", "b", "c")
	w.WriteRecord("x", "y", "")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if len(rec) != 3 || rec[0] != "x" || rec[1] != "y" || rec[2] != "" {
		t.Fatalf("unexpected record: %v", rec)
	}
}

func TestSkipEmptyLines(t *testing.T) {
	data := []byte("id|name\n\n1|foo\n\n\n2|bar\n")
	r, err := NewReader(data)
	if err != nil {
		t.Fatal(err)
	}

	rec1, _ := r.Next()
	if rec1[1] != "foo" {
		t.Fatalf("unexpected record: %v", rec1)
	}

	rec2, _ := r.Next()
	if rec2[1] != "bar" {
		t.Fatalf("unexpected record: %v", rec2)
	}
}

func TestSkipCommentLines(t *testing.T) {
	data := []byte("# this is a comment\n# another comment\nid|name\n# inline comment\n1|foo\n")
	r, err := NewReader(data)
	if err != nil {
		t.Fatal(err)
	}

	header := r.Header()
	if header[0] != "id" {
		t.Fatalf("unexpected header: %v", header)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if rec[1] != "foo" {
		t.Fatalf("unexpected record: %v", rec)
	}

	_, ok = r.Next()
	if ok {
		t.Fatal("expected no more records after comments at end of section, but comments are data lines")
	}
}

func TestOnlyHeader(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("a", "b")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	_, ok := r.Next()
	if ok {
		t.Fatal("expected no records")
	}
}

func TestMultipleWrites(t *testing.T) {
	var sections bytes.Buffer

	w := NewWriter(&sections)
	w.WriteHeader("id", "name")
	w.WriteRecord("1", "foo")
	w.WriteRecord("2", "bar")
	w.Flush()

	// Reader handles the whole stream as one section.
	// The "# second section" comment is skipped, and "x|y" becomes the
	// first record of a SECOND section — but our Reader is single-section
	// and treats the first non-comment line as header, rest as records.
	// So all 4 non-comment lines after the header become records.
	sections.WriteString("# second section\n")
	w2 := NewWriter(&sections)
	w2.WriteHeader("x", "y")
	w2.WriteRecord("a", "b")
	w2.Flush()

	r, err := NewReader(sections.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if r.Header()[0] != "id" {
		t.Fatalf("expected 'id' header, got %v", r.Header())
	}

	count := 0
	for {
		_, ok := r.Next()
		if !ok {
			break
		}
		count++
	}
	// The second section's header and records are treated as records
	if count != 4 {
		t.Fatalf("expected 4 total records (2 from each section), got %d", count)
	}
}

func TestEmptyInput(t *testing.T) {
	_, err := NewReader([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestOnlyComments(t *testing.T) {
	_, err := NewReader([]byte("# comment\n# another\n"))
	if err == nil {
		t.Fatal("expected error for comments-only input")
	}
}

func TestNoNewlineAtEnd(t *testing.T) {
	r, err := NewReader([]byte("id|name\n1|foo"))
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if rec[0] != "1" || rec[1] != "foo" {
		t.Fatalf("unexpected record: %v", rec)
	}
}

func TestRecordWithAllEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("a", "b")
	w.WriteRecord("", "")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if rec[0] != "" || rec[1] != "" {
		t.Fatalf("unexpected record: %v", rec)
	}
}

func TestUnicodeFields(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("desc")
	w.WriteRecord("日本語 | English | 😊")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if rec[0] != "日本語 | English | 😊" {
		t.Fatalf("unexpected value: %q", rec[0])
	}
}

func TestFieldsWithOnlySpecialChars(t *testing.T) {
	cases := []struct {
		input string
	}{
		{`\`},
		{`\\`},
		{`\|`},
		{`\n`},
		{`\\\|\\`},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewWriter(&buf)
			w.WriteHeader("v")
			w.WriteRecord(tc.input)
			w.Flush()

			r, err := NewReader(buf.Bytes())
			if err != nil {
				t.Fatal(err)
			}

			rec, ok := r.Next()
			if !ok {
				t.Fatal("expected record")
			}
			if rec[0] != tc.input {
				t.Fatalf("roundtrip: got %q, want %q", rec[0], tc.input)
			}
		})
	}
}

func TestHeaderThenRecord_CorrectFieldCount(t *testing.T) {
	w := NewWriter(&bytes.Buffer{})
	// We can't detect wrong field count at write time since we just write strings
	// But the reader should handle variable-length records

	var buf bytes.Buffer
	w = NewWriter(&buf)
	w.WriteHeader("a", "b", "c")
	w.WriteRecord("1", "2")
	w.Flush()

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rec, ok := r.Next()
	if !ok {
		t.Fatal("expected record")
	}
	if len(rec) != 2 {
		t.Fatalf("expected 2 fields in record (we don't pad), got %v", rec)
	}
}

func TestFlushIdempotent(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteHeader("a")
	w.WriteRecord("1")
	w.Flush()
	first := buf.String()

	w.WriteRecord("2")
	w.Flush()
	second := buf.String()

	if first != "a\n1\n" {
		t.Fatalf("unexpected first flush: %q", first)
	}
	// second flush adds more data
	if second != "a\n1\n2\n" {
		t.Fatalf("unexpected second flush: %q", second)
	}
}
