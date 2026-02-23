package harnessloop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RenderCompareMarkdown(report CompareReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Loop Compare Report\n\n")
	fmt.Fprintf(&b, "- Run A: `%s`\n", report.RunA.RunID)
	fmt.Fprintf(&b, "- Run B: `%s`\n", report.RunB.RunID)
	fmt.Fprintf(&b, "- Score delta (B-A): `%.2f`\n", report.Delta.Score)
	fmt.Fprintf(&b, "- Pass delta (B-A): `%d`\n", report.Delta.PassDelta)
	fmt.Fprintf(&b, "- Findings delta (B-A): `%d`\n", report.Delta.FindingsDelta)
	fmt.Fprintf(&b, "- Iterations delta (B-A): `%d`\n", report.Delta.IterationsDelta)
	fmt.Fprintf(&b, "- Fixes delta (B-A): `%d`\n", report.Delta.FixesAppliedDelta)
	fmt.Fprintf(&b, "\n## Judge Breakdown\n\n")
	fmt.Fprintf(&b, "| Metric | Run A | Run B |\n")
	fmt.Fprintf(&b, "|---|---:|---:|\n")
	fmt.Fprintf(&b, "| Total score | %.2f | %.2f |\n", report.RunA.Judge.Score, report.RunB.Judge.Score)
	fmt.Fprintf(&b, "| UX | %.2f | %.2f |\n", report.RunA.Judge.UXScore, report.RunB.Judge.UXScore)
	fmt.Fprintf(&b, "| Quality | %.2f | %.2f |\n", report.RunA.Judge.QualityScore, report.RunB.Judge.QualityScore)
	fmt.Fprintf(&b, "| Penalty | %.2f | %.2f |\n", report.RunA.Judge.PenaltyScore, report.RunB.Judge.PenaltyScore)
	fmt.Fprintf(&b, "| Planner | %.2f | %.2f |\n", report.RunA.Judge.PlannerScore, report.RunB.Judge.PlannerScore)
	fmt.Fprintf(&b, "| Fixer | %.2f | %.2f |\n", report.RunA.Judge.FixerScore, report.RunB.Judge.FixerScore)
	fmt.Fprintf(&b, "| Judger | %.2f | %.2f |\n", report.RunA.Judge.JudgerScore, report.RunB.Judge.JudgerScore)
	return b.String()
}

func WriteCompareOutput(repoRoot string, report CompareReport, format, outPath string) (string, error) {
	f := strings.ToLower(strings.TrimSpace(format))
	if f == "" {
		f = "json"
	}
	if f != "json" && f != "md" {
		return "", fmt.Errorf("unsupported compare format: %s", format)
	}

	if strings.TrimSpace(outPath) == "" {
		if f == "json" {
			return "", nil
		}
		ts := time.Now().UTC().Format("20060102-150405")
		outPath = filepath.Join(repoRoot, ".docs", "onboarding-loop", "compare", ts+".md")
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return "", err
	}
	if f == "md" {
		if err := os.WriteFile(outPath, []byte(RenderCompareMarkdown(report)), 0644); err != nil {
			return "", err
		}
		return outPath, nil
	}

	if err := writeJSON(outPath, report); err != nil {
		return "", err
	}
	return outPath, nil
}
