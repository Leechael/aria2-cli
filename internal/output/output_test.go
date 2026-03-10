package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Out: &buf, Err: &bytes.Buffer{}}

	data := map[string]any{"gid": "abc123", "status": "active"}
	if err := p.Print(data); err != nil {
		t.Fatal(err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result["gid"] != "abc123" {
		t.Errorf("expected abc123, got %v", result["gid"])
	}
}

func TestPrintPlainString(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatPlain, Out: &buf, Err: &bytes.Buffer{}}

	p.Print("hello")
	if strings.TrimSpace(buf.String()) != "hello" {
		t.Errorf("expected 'hello', got %q", buf.String())
	}
}

func TestPrintPlainMap(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatPlain, Out: &buf, Err: &bytes.Buffer{}}

	data := map[string]any{"status": "active", "gid": "abc"}
	p.Print(data)

	out := buf.String()
	if !strings.Contains(out, "gid") || !strings.Contains(out, "abc") {
		t.Errorf("expected gid/abc in output: %s", out)
	}
	if !strings.Contains(out, "status") || !strings.Contains(out, "active") {
		t.Errorf("expected status/active in output: %s", out)
	}
}

func TestPrintPlainSlice(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatPlain, Out: &buf, Err: &bytes.Buffer{}}

	data := []any{
		map[string]any{"gid": "1"},
		map[string]any{"gid": "2"},
	}
	p.Print(data)

	out := buf.String()
	if !strings.Contains(out, "1") || !strings.Contains(out, "2") {
		t.Errorf("expected both items in output: %s", out)
	}
}

func TestHint(t *testing.T) {
	var errBuf bytes.Buffer
	p := &Printer{Format: FormatPlain, Out: &bytes.Buffer{}, Err: &errBuf}

	p.Hint("downloading...")
	if !strings.Contains(errBuf.String(), "downloading...") {
		t.Errorf("expected hint on stderr: %s", errBuf.String())
	}
}
