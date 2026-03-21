package cobrax

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	agentcli "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/resource"
)

func testSchema() resource.ResourceSchema {
	return resource.ResourceSchema{
		Kind: "widget",
		Fields: []resource.FieldDef{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "status", Type: "string"},
		},
	}
}

func testRecords() []resource.Record {
	return []resource.Record{
		{
			Kind: "widget",
			ID:   "w-001",
			Fields: map[string]any{
				"id":     "w-001",
				"name":   "Alpha",
				"status": "active",
			},
		},
		{
			Kind: "widget",
			ID:   "w-002",
			Fields: map[string]any{
				"id":     "w-002",
				"name":   "Beta",
				"status": "pending",
			},
		},
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	records := testRecords()
	schema := testSchema()

	err := RenderRecords(&buf, records, schema, OutputJSON, nil, "")
	if err != nil {
		t.Fatalf("RenderRecords JSON: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if !env.OK {
		t.Fatal("expected envelope OK to be true")
	}
	if env.Kind != "widget" {
		t.Fatalf("expected kind 'widget', got %q", env.Kind)
	}
	if len(env.Data) != 2 {
		t.Fatalf("expected 2 data entries, got %d", len(env.Data))
	}
	if env.Data[0]["id"] != "w-001" {
		t.Fatalf("expected first id 'w-001', got %v", env.Data[0]["id"])
	}
}

func TestRenderJSONFieldSelection(t *testing.T) {
	var buf bytes.Buffer
	records := testRecords()
	schema := testSchema()

	err := RenderRecords(&buf, records, schema, OutputJSON, []string{"id", "name"}, "")
	if err != nil {
		t.Fatalf("RenderRecords JSON with fields: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if len(env.Data) != 2 {
		t.Fatalf("expected 2 data entries, got %d", len(env.Data))
	}
	// Only id and name should be present
	if _, ok := env.Data[0]["status"]; ok {
		t.Fatal("expected 'status' field to be excluded when fields are specified")
	}
	if env.Data[0]["name"] != "Alpha" {
		t.Fatalf("expected name 'Alpha', got %v", env.Data[0]["name"])
	}
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	records := testRecords()
	schema := testSchema()

	err := RenderRecords(&buf, records, schema, OutputAuto, nil, "")
	if err != nil {
		t.Fatalf("RenderRecords table: %v", err)
	}

	output := buf.String()

	// Table should have headers
	if !strings.Contains(output, "ID") {
		t.Fatal("expected table to contain 'ID' header")
	}
	if !strings.Contains(output, "NAME") {
		t.Fatal("expected table to contain 'NAME' header")
	}
	if !strings.Contains(output, "STATUS") {
		t.Fatal("expected table to contain 'STATUS' header")
	}

	// Table should have data
	if !strings.Contains(output, "w-001") {
		t.Fatal("expected table to contain 'w-001'")
	}
	if !strings.Contains(output, "Alpha") {
		t.Fatal("expected table to contain 'Alpha'")
	}
	if !strings.Contains(output, "Beta") {
		t.Fatal("expected table to contain 'Beta'")
	}
}

func TestRenderTSV(t *testing.T) {
	var buf bytes.Buffer
	records := testRecords()
	schema := testSchema()

	// OutputAuto with a non-TTY writer (buffer) should produce TSV
	// We force TSV mode by rendering to a buffer (non-TTY)
	err := renderTSV(&buf, records, schema, nil)
	if err != nil {
		t.Fatalf("renderTSV: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 { // header + 2 records
		t.Fatalf("expected 3 lines (header + 2 records), got %d", len(lines))
	}

	// Check header is tab-separated
	headers := strings.Split(lines[0], "\t")
	if len(headers) != 3 {
		t.Fatalf("expected 3 tab-separated headers, got %d: %q", len(headers), lines[0])
	}

	// Check data row is tab-separated
	row1 := strings.Split(lines[1], "\t")
	if len(row1) != 3 {
		t.Fatalf("expected 3 tab-separated fields in row 1, got %d", len(row1))
	}
	if row1[0] != "w-001" {
		t.Fatalf("expected first field 'w-001', got %q", row1[0])
	}
}

func TestRenderDoctorReportJSON(t *testing.T) {
	var buf bytes.Buffer
	report := agentcli.DoctorReport{
		SchemaVersion: "1",
		OK:            false,
		Findings: []agentcli.DoctorFinding{
			{Code: "E001", Path: "/foo/bar", Message: "missing file"},
		},
	}

	err := RenderDoctorReport(&buf, report, true)
	if err != nil {
		t.Fatalf("RenderDoctorReport JSON: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if out["ok"] != false {
		t.Fatal("expected ok=false")
	}
}

func TestRenderDoctorReportTable(t *testing.T) {
	var buf bytes.Buffer
	report := agentcli.DoctorReport{
		SchemaVersion: "1",
		OK:            true,
		Findings:      nil,
	}

	err := RenderDoctorReport(&buf, report, false)
	if err != nil {
		t.Fatalf("RenderDoctorReport table: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "OK") {
		t.Fatal("expected table output to contain OK status")
	}
}

func TestRenderJQ(t *testing.T) {
	var buf bytes.Buffer
	records := testRecords()
	schema := testSchema()

	err := RenderRecords(&buf, records, schema, OutputJQ, nil, ".data[0].id")
	if err != nil {
		t.Fatalf("RenderRecords JQ: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != `"w-001"` {
		t.Fatalf("expected '\"w-001\"', got %q", output)
	}
}

func TestRenderEmptyRecords(t *testing.T) {
	var buf bytes.Buffer
	schema := testSchema()

	err := RenderRecords(&buf, nil, schema, OutputJSON, nil, "")
	if err != nil {
		t.Fatalf("RenderRecords empty: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(env.Data) != 0 {
		t.Fatalf("expected empty data, got %d entries", len(env.Data))
	}
	if !env.OK {
		t.Fatal("expected OK=true for empty records")
	}
}
