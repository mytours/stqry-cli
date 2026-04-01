package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

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
