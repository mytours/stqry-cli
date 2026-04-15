package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/itchyny/gojq"
)

type Meta struct {
	Site    string `json:"site"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Total   int    `json:"total"`
}

type JSONFormatter struct {
	Writer io.Writer
}

func (f *JSONFormatter) Write(data interface{}, meta *Meta) error {
	envelope := map[string]interface{}{
		"data": data,
		"meta": meta,
	}
	enc := json.NewEncoder(f.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(envelope)
}

type QuietFormatter struct {
	Writer io.Writer
}

func (f *QuietFormatter) Write(data interface{}, meta *Meta) error {
	enc := json.NewEncoder(f.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

type HumanFormatter struct {
	Writer io.Writer
}

func (f *HumanFormatter) WriteTable(columns []string, rows []map[string]interface{}) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = strings.ToUpper(c)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, row := range rows {
		vals := make([]string, len(columns))
		for i, c := range columns {
			v := row[c]
			vals[i] = formatValue(v)
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
	return w.Flush()
}

func (f *HumanFormatter) WriteKeyValue(data map[string]interface{}) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "%s:\t%s\n", k, formatValue(data[k]))
	}
	return w.Flush()
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func isScalar(v interface{}) bool {
	switch val := v.(type) {
	case nil, bool, float64, string:
		return true
	case []interface{}:
		for _, elem := range val {
			if !isScalar(elem) {
				return false
			}
		}
		return true
	default:
		_ = val
		return false
	}
}

func formatValue(v interface{}) string {
	if v == nil {
		return "-"
	}
	switch val := v.(type) {
	case map[string]interface{}:
		return FormatTranslatedString(val)
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%.2f", val)
	case []interface{}:
		if len(val) == 0 {
			return "-"
		}
		parts := make([]string, 0, len(val))
		for _, elem := range val {
			parts = append(parts, formatValue(elem))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprintf("%v", val)
	}
}

func FormatTranslatedString(ts map[string]interface{}) string {
	if len(ts) == 0 {
		return "-"
	}
	langs := make([]string, 0, len(ts))
	for k := range ts {
		langs = append(langs, k)
	}
	sort.Strings(langs)
	parts := make([]string, 0, len(langs))
	for _, lang := range langs {
		if ts[lang] != nil {
			parts = append(parts, fmt.Sprintf("[%s] %v", lang, ts[lang]))
		}
	}
	return strings.Join(parts, " ")
}

func applyJQ(w io.Writer, code *gojq.Code, data interface{}) error {
	iter := code.Run(data)
	enc := json.NewEncoder(w)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return fmt.Errorf("jq: %w", err)
		}
		if err := enc.Encode(v); err != nil {
			return fmt.Errorf("jq: encoding output: %w", err)
		}
	}
	return nil
}

type Printer struct {
	JSON   bool
	Quiet  bool
	JQCode *gojq.Code
}

func (p *Printer) PrintOne(data interface{}, meta *Meta) error {
	if p.JQCode != nil {
		if p.JSON {
			return applyJQ(os.Stdout, p.JQCode, map[string]interface{}{"data": data, "meta": meta})
		}
		return applyJQ(os.Stdout, p.JQCode, data)
	}
	if p.Quiet {
		f := &QuietFormatter{Writer: os.Stdout}
		return f.Write(data, nil)
	}
	if p.JSON {
		f := &JSONFormatter{Writer: os.Stdout}
		return f.Write(data, meta)
	}
	if m, ok := data.(map[string]interface{}); ok {
		f := &HumanFormatter{Writer: os.Stdout}
		return f.WriteKeyValue(m)
	}
	f := &QuietFormatter{Writer: os.Stdout}
	return f.Write(data, nil)
}

func (p *Printer) PrintList(columns []string, rows []map[string]interface{}, meta *Meta) error {
	if p.JQCode != nil {
		irows := make([]interface{}, len(rows))
		for i, r := range rows {
			irows[i] = r
		}
		if p.JSON {
			return applyJQ(os.Stdout, p.JQCode, map[string]interface{}{"data": irows, "meta": meta})
		}
		return applyJQ(os.Stdout, p.JQCode, irows)
	}
	if p.Quiet {
		f := &QuietFormatter{Writer: os.Stdout}
		return f.Write(rows, nil)
	}
	if p.JSON {
		f := &JSONFormatter{Writer: os.Stdout}
		return f.Write(rows, meta)
	}
	f := &HumanFormatter{Writer: os.Stdout}
	return f.WriteTable(columns, rows)
}

func (p *Printer) PrintError(err error) {
	if p.JSON || p.Quiet || p.JQCode != nil {
		enc := json.NewEncoder(os.Stderr)
		enc.Encode(map[string]string{"error": err.Error()})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}
