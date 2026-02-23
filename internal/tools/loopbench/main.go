package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	harnessloop "github.com/gh-xj/agentcli-go/internal/harnessloop"
)

type benchSummary struct {
	SchemaVersion string      `json:"schema_version"`
	Classic       benchResult `json:"classic"`
	Committee     benchResult `json:"committee"`
}

type benchResult struct {
	Score float64 `json:"score"`
	Pass  bool    `json:"pass"`
}

type baseline struct {
	SchemaVersion            string  `json:"schema_version"`
	MinClassicScore          float64 `json:"min_classic_score"`
	MinCommitteeScore        float64 `json:"min_committee_score"`
	MinCommitteeMinusClassic float64 `json:"min_committee_minus_classic"`
}

func main() {
	mode := flag.String("mode", "run", "run|check")
	repoRoot := flag.String("repo-root", ".", "repo root")
	threshold := flag.Float64("threshold", 9.0, "judge threshold")
	output := flag.String("output", ".docs/onboarding-loop/benchmarks/latest.json", "summary output")
	baselinePath := flag.String("baseline", "testdata/benchmarks/loop-benchmark-baseline.json", "baseline file")
	flag.Parse()

	switch *mode {
	case "run":
		if err := run(*repoRoot, *threshold, *output); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "check":
		if err := check(*output, *baselinePath); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "invalid mode")
		os.Exit(2)
	}
}

func run(repoRoot string, threshold float64, output string) error {
	classic, err := harnessloop.RunLoop(harnessloop.Config{RepoRoot: repoRoot, Threshold: threshold, MaxIterations: 1, Mode: "classic", AutoFix: false, AutoCommit: false})
	if err != nil {
		return err
	}
	committee, err := harnessloop.RunLoop(harnessloop.Config{RepoRoot: repoRoot, Threshold: threshold, MaxIterations: 1, Mode: "committee", RoleConfigPath: "./configs/committee.roles.example.json", AutoFix: false, AutoCommit: false})
	if err != nil {
		return err
	}
	summary := benchSummary{
		SchemaVersion: "v1",
		Classic:       benchResult{Score: classic.Judge.Score, Pass: classic.Judge.Pass},
		Committee:     benchResult{Score: committee.Judge.Score, Pass: committee.Judge.Pass},
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	return writeJSON(output, summary)
}

func check(summaryPath, baselinePath string) error {
	var s benchSummary
	if err := readJSON(summaryPath, &s); err != nil {
		return err
	}
	var b baseline
	if err := readJSON(baselinePath, &b); err != nil {
		return err
	}
	if s.Classic.Score < b.MinClassicScore {
		return fmt.Errorf("classic score regression: got %.2f < min %.2f", s.Classic.Score, b.MinClassicScore)
	}
	if s.Committee.Score < b.MinCommitteeScore {
		return fmt.Errorf("committee score regression: got %.2f < min %.2f", s.Committee.Score, b.MinCommitteeScore)
	}
	delta := s.Committee.Score - s.Classic.Score
	if delta < b.MinCommitteeMinusClassic {
		return fmt.Errorf("committee delta regression: got %.2f < min %.2f", delta, b.MinCommitteeMinusClassic)
	}
	fmt.Printf("benchmark check passed (classic=%.2f committee=%.2f delta=%.2f)\n", s.Classic.Score, s.Committee.Score, delta)
	return nil
}

func readJSON(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func writeJSON(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0644)
}
