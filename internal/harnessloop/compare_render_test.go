package harnessloop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteCompareOutputMarkdown(t *testing.T) {
	repo := t.TempDir()
	report := CompareReport{
		SchemaVersion: "v1",
		RunA:          RunResult{RunID: "a", Judge: JudgeScore{Score: 9.0}},
		RunB:          RunResult{RunID: "b", Judge: JudgeScore{Score: 9.5}},
		Delta:         CompareDelta{Score: 0.5},
	}
	path, err := WriteCompareOutput(repo, report, "md", "")
	if err != nil {
		t.Fatalf("write md output: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read md output: %v", err)
	}
	if !strings.Contains(string(raw), "# Loop Compare Report") {
		t.Fatalf("unexpected markdown: %s", string(raw))
	}
}

func TestWriteCompareOutputJSON(t *testing.T) {
	repo := t.TempDir()
	report := CompareReport{SchemaVersion: "v1"}
	out := filepath.Join(repo, "compare.json")
	path, err := WriteCompareOutput(repo, report, "json", out)
	if err != nil {
		t.Fatalf("write json output: %v", err)
	}
	if path != out {
		t.Fatalf("unexpected output path: %s", path)
	}
}
