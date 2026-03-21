package cobrax

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	agentcli "github.com/gh-xj/agentcli-go"
	"github.com/gh-xj/agentcli-go/resource"
	"github.com/itchyny/gojq"
)

// OutputMode for rendering records.
type OutputMode int

const (
	OutputAuto OutputMode = iota // TTY→table, pipe→TSV
	OutputJSON                   // --json field1,field2
	OutputJQ                     // --jq expression
)

// maxFieldWidth is the maximum character width for table cell values.
const maxFieldWidth = 40

// Envelope is the JSON output wrapper.
type Envelope struct {
	OK       bool             `json:"ok"`
	Kind     string           `json:"kind"`
	Data     []map[string]any `json:"data"`
	Warnings []string         `json:"warnings,omitempty"`
}

// RenderRecords renders records based on output mode.
func RenderRecords(w io.Writer, records []resource.Record, schema resource.ResourceSchema, mode OutputMode, fields []string, jqExpr string) error {
	switch mode {
	case OutputJSON:
		return renderJSON(w, records, schema, fields)
	case OutputJQ:
		return renderJQ(w, records, schema, jqExpr)
	default:
		// OutputAuto: always render as table when called directly.
		// The caller should detect TTY vs pipe if needed.
		return renderTable(w, records, schema, fields)
	}
}

// renderJSON outputs records as a JSON envelope with optional field selection.
func renderJSON(w io.Writer, records []resource.Record, schema resource.ResourceSchema, fields []string) error {
	env := buildEnvelope(records, schema, fields)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}

// renderJQ outputs records as JSON filtered through a jq expression.
func renderJQ(w io.Writer, records []resource.Record, schema resource.ResourceSchema, jqExpr string) error {
	env := buildEnvelope(records, schema, nil)

	// Marshal to generic interface for gojq
	raw, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal for jq: %w", err)
	}
	var input any
	if err := json.Unmarshal(raw, &input); err != nil {
		return fmt.Errorf("unmarshal for jq: %w", err)
	}

	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return fmt.Errorf("parse jq expression: %w", err)
	}

	iter := query.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return fmt.Errorf("jq evaluation: %w", err)
		}
		out, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("marshal jq result: %w", err)
		}
		fmt.Fprintln(w, string(out))
	}
	return nil
}

// renderTable renders records as a text table using tabwriter.
func renderTable(w io.Writer, records []resource.Record, schema resource.ResourceSchema, fields []string) error {
	cols := fieldNames(schema, fields)
	if len(cols) == 0 {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	// Header
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = strings.ToUpper(c)
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	// Rows
	for _, rec := range records {
		vals := make([]string, len(cols))
		for i, col := range cols {
			vals[i] = truncate(formatField(rec.Fields[col]), maxFieldWidth)
		}
		fmt.Fprintln(tw, strings.Join(vals, "\t"))
	}

	return tw.Flush()
}

// renderTSV renders records as tab-separated values (no alignment padding).
func renderTSV(w io.Writer, records []resource.Record, schema resource.ResourceSchema, fields []string) error {
	cols := fieldNames(schema, fields)
	if len(cols) == 0 {
		return nil
	}

	// Header
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = strings.ToUpper(c)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Rows
	for _, rec := range records {
		vals := make([]string, len(cols))
		for i, col := range cols {
			vals[i] = formatField(rec.Fields[col])
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
	return nil
}

// RenderDoctorReport renders a DoctorReport as JSON or table.
func RenderDoctorReport(w io.Writer, report agentcli.DoctorReport, jsonMode bool) error {
	if jsonMode {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if report.OK {
		fmt.Fprintln(tw, "STATUS\tOK")
	} else {
		fmt.Fprintln(tw, "STATUS\tFAIL")
	}

	if len(report.Findings) > 0 {
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "CODE\tPATH\tMESSAGE")
		for _, f := range report.Findings {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", f.Code, f.Path, f.Message)
		}
	}
	return tw.Flush()
}

// buildEnvelope constructs a JSON envelope from records.
func buildEnvelope(records []resource.Record, schema resource.ResourceSchema, fields []string) Envelope {
	data := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		entry := filterFields(rec.Fields, fields)
		data = append(data, entry)
	}
	return Envelope{
		OK:   true,
		Kind: schema.Kind,
		Data: data,
	}
}

// filterFields returns a copy of fields, keeping only the specified ones.
// If fields is nil or empty, all fields are returned.
func filterFields(src map[string]any, fields []string) map[string]any {
	if len(fields) == 0 {
		out := make(map[string]any, len(src))
		for k, v := range src {
			out[k] = v
		}
		return out
	}
	out := make(map[string]any, len(fields))
	for _, f := range fields {
		if v, ok := src[f]; ok {
			out[f] = v
		}
	}
	return out
}

// fieldNames returns the list of field names to display.
// If fields is non-empty, only those are returned (preserving order).
// Otherwise, all schema fields are used.
func fieldNames(schema resource.ResourceSchema, fields []string) []string {
	if len(fields) > 0 {
		return fields
	}
	names := make([]string, len(schema.Fields))
	for i, f := range schema.Fields {
		names[i] = f.Name
	}
	return names
}

// formatField converts a field value to its string representation.
func formatField(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// truncate shortens a string to max characters, appending "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
