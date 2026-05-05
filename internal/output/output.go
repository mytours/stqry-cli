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

func (f *HumanFormatter) WriteTable(columns []string, rows []map[string]interface{}, meta *Meta) error {
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
	if err := w.Flush(); err != nil {
		return err
	}

	if footer := paginationFooter(meta, len(rows)); footer != "" {
		fmt.Fprintln(f.Writer)
		fmt.Fprintln(f.Writer, footer)
	}
	return nil
}

// paginationFooter returns a one-line summary like
// "Showing 30 of 1017 (page 1 of 34) — pass --page / --per-page to see more"
// when the response is paginated and more rows exist on the server.
// Returns "" when meta is nil or the visible rows already cover the total.
func paginationFooter(meta *Meta, visibleRows int) string {
	if meta == nil {
		return ""
	}
	if meta.Total == 0 || meta.Total <= visibleRows {
		return ""
	}
	if meta.Page > 0 && meta.PerPage > 0 {
		pages := (meta.Total + meta.PerPage - 1) / meta.PerPage
		return fmt.Sprintf("Showing %d of %d (page %d of %d) — pass --page / --per-page to see more",
			visibleRows, meta.Total, meta.Page, pages)
	}
	return fmt.Sprintf("Showing %d of %d — pass --page / --per-page to see more",
		visibleRows, meta.Total)
}

func (f *HumanFormatter) WriteKeyValue(data map[string]interface{}) error {
	keys := sortedKeys(data)

	// Two-pass rendering: scalars first so tabwriter can align them together,
	// then complex fields as indented blocks below. This means complex fields
	// always appear after scalars regardless of alphabetical order.

	// Pass 1: scalar fields, tabwriter-aligned
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	for _, k := range keys {
		if isScalar(data[k]) {
			fmt.Fprintf(w, "%s:\t%s\n", k, formatValue(data[k]))
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}

	// Pass 2: complex fields, indented blocks
	for _, k := range keys {
		if !isScalar(data[k]) {
			if _, err := fmt.Fprintf(f.Writer, "%s:\n", k); err != nil {
				return err
			}
			if err := writeComplexValue(f.Writer, data[k], "  "); err != nil {
				return err
			}
		}
	}
	return nil
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func writeComplexValue(w io.Writer, v interface{}, indent string) error {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := sortedKeys(val)
		for _, k := range keys {
			if isScalar(val[k]) {
				if _, err := fmt.Fprintf(w, "%s%s: %s\n", indent, k, formatValue(val[k])); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s%s:\n", indent, k); err != nil {
					return err
				}
				if err := writeComplexValue(w, val[k], indent+"  "); err != nil {
					return err
				}
			}
		}
	case []interface{}:
		for _, elem := range val {
			if m, ok := elem.(map[string]interface{}); ok {
				keys := sortedKeys(m)
				first := true
				for _, k := range keys {
					prefix := indent + "  "
					if first {
						prefix = indent + "- "
						first = false
					}
					if isScalar(m[k]) {
						if _, err := fmt.Fprintf(w, "%s%s: %s\n", prefix, k, formatValue(m[k])); err != nil {
							return err
						}
					} else {
						if _, err := fmt.Fprintf(w, "%s%s:\n", prefix, k); err != nil {
							return err
						}
						if err := writeComplexValue(w, m[k], indent+"    "); err != nil {
							return err
						}
					}
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s- %s\n", indent, formatValue(elem)); err != nil {
					return err
				}
			}
		}
	default:
		if _, err := fmt.Fprintf(w, "%s%s\n", indent, formatValue(v)); err != nil {
			return err
		}
	}
	return nil
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
	return f.WriteTable(columns, rows, meta)
}

func (p *Printer) PrintError(err error) {
	if p.JSON || p.Quiet || p.JQCode != nil {
		enc := json.NewEncoder(os.Stderr)
		enc.Encode(map[string]string{"error": err.Error()})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}
