package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/itchyny/gojq"
)

func mustCompileJQ(t *testing.T, expr string) *gojq.Code {
	t.Helper()
	q, err := gojq.Parse(expr)
	if err != nil {
		t.Fatalf("parse jq %q: %v", expr, err)
	}
	c, err := gojq.Compile(q)
	if err != nil {
		t.Fatalf("compile jq %q: %v", expr, err)
	}
	return c
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{Writer: &buf}

	data := map[string]string{"name": "test"}
	meta := &Meta{Site: "bobs", Page: 1, PerPage: 25, Total: 1}
	err := f.Write(data, meta)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(buf.Bytes(), &result)
	if result["data"] == nil {
		t.Error("expected data field")
	}
	m := result["meta"].(map[string]interface{})
	if m["site"] != "bobs" {
		t.Errorf("expected site bobs, got %v", m["site"])
	}
}

func TestQuietFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &QuietFormatter{Writer: &buf}

	data := map[string]string{"name": "test"}
	err := f.Write(data, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	var result map[string]string
	json.Unmarshal(buf.Bytes(), &result)
	if result["name"] != "test" {
		t.Errorf("expected name=test, got %s", result["name"])
	}
}

func TestHumanFormatterList(t *testing.T) {
	var buf bytes.Buffer
	f := &HumanFormatter{Writer: &buf}

	items := []map[string]interface{}{
		{"id": 1, "name": "Tour A"},
		{"id": 2, "name": "Tour B"},
	}
	err := f.WriteTable([]string{"id", "name"}, items)
	if err != nil {
		t.Fatalf("WriteTable: %v", err)
	}

	out := buf.String()
	if len(out) == 0 {
		t.Error("expected output")
	}
}

func TestFormatTranslatedString(t *testing.T) {
	ts := map[string]interface{}{
		"en": "Hello",
		"fr": "Bonjour",
	}
	result := FormatTranslatedString(ts)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestApplyJQ_List(t *testing.T) {
	rows := []interface{}{
		map[string]interface{}{"id": 1.0, "name": "alpha"},
		map[string]interface{}{"id": 2.0, "name": "beta"},
	}
	var buf bytes.Buffer
	err := applyJQ(&buf, mustCompileJQ(t, ".[].name"), rows)
	if err != nil {
		t.Fatalf("applyJQ: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"alpha"`) {
		t.Errorf("expected alpha in output, got: %s", out)
	}
	if !strings.Contains(out, `"beta"`) {
		t.Errorf("expected beta in output, got: %s", out)
	}
}

func TestApplyJQ_RuntimeError(t *testing.T) {
	// .foo on an array is a jq runtime error
	rows := []interface{}{1.0, 2.0}
	var buf bytes.Buffer
	err := applyJQ(&buf, mustCompileJQ(t, ".foo"), rows)
	if err == nil {
		t.Fatal("expected error for .foo on array")
	}
	if !strings.HasPrefix(err.Error(), "jq:") {
		t.Errorf("expected error to start with 'jq:', got: %s", err)
	}
}

func TestPrinterPrintList_JQ(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": 1.0, "name": "alpha"},
		{"id": 2.0, "name": "beta"},
	}
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p := &Printer{JQCode: mustCompileJQ(t, ".[].name")}
	err := p.PrintList([]string{"id", "name"}, rows, nil)

	w.Close()
	os.Stdout = origStdout
	var out bytes.Buffer
	out.ReadFrom(r)
	r.Close()

	if err != nil {
		t.Fatalf("PrintList: %v", err)
	}
	if !strings.Contains(out.String(), `"alpha"`) {
		t.Errorf("expected alpha in output, got: %s", out.String())
	}
}

func TestPrinterPrintOne_JQ(t *testing.T) {
	data := map[string]interface{}{"id": 1.0, "name": "alpha"}
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p := &Printer{JQCode: mustCompileJQ(t, ".name")}
	err := p.PrintOne(data, nil)

	w.Close()
	os.Stdout = origStdout
	var out bytes.Buffer
	out.ReadFrom(r)
	r.Close()

	if err != nil {
		t.Fatalf("PrintOne: %v", err)
	}
	if !strings.Contains(out.String(), `"alpha"`) {
		t.Errorf("expected alpha in output, got: %s", out.String())
	}
}

func TestPrinterPrintError_JQ(t *testing.T) {
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	p := &Printer{JQCode: mustCompileJQ(t, ".name")}
	p.PrintError(fmt.Errorf("something went wrong"))

	w.Close()
	os.Stderr = origStderr
	var out bytes.Buffer
	out.ReadFrom(r)
	r.Close()

	var result map[string]string
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("expected JSON error output, got: %s (%v)", out.String(), err)
	}
	if result["error"] != "something went wrong" {
		t.Errorf("expected error='something went wrong', got: %v", result)
	}
}

// TestPrinterPrintList_JQ_OverridesQuiet pins the documented behaviour:
// when --jq is set, --quiet is a no-op (jq runs instead).
func TestPrinterPrintList_JQ_OverridesQuiet(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": 1.0, "name": "alpha"},
	}
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p := &Printer{JQCode: mustCompileJQ(t, ".[].name"), Quiet: true}
	err := p.PrintList([]string{"id", "name"}, rows, nil)

	w.Close()
	os.Stdout = origStdout
	var out bytes.Buffer
	out.ReadFrom(r)
	r.Close()

	if err != nil {
		t.Fatalf("PrintList: %v", err)
	}
	// JQ output should contain the filtered value.
	if !strings.Contains(out.String(), `"alpha"`) {
		t.Errorf("expected jq output with alpha, got: %s", out.String())
	}
	// Should NOT be the quiet full-array envelope (pretty-printed array).
	if strings.HasPrefix(strings.TrimSpace(out.String()), "[") {
		t.Error("expected jq scalar output, not quiet array envelope")
	}
}
