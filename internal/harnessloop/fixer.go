package harnessloop

import (
	"fmt"
	"os/exec"
)

func ApplyFixes(repoRoot string, findings []Finding) ([]string, error) {
	applied := make([]string, 0)
	for _, f := range findings {
		switch f.Code {
		case "generated_go_not_formatted":
			cmd := exec.Command("zsh", "-lc", "gofmt -w scaffold.go")
			cmd.Dir = repoRoot
			if out, err := cmd.CombinedOutput(); err != nil {
				return applied, fmt.Errorf("apply gofmt fix: %w\n%s", err, string(out))
			}
			applied = append(applied, "gofmt scaffold.go")
		case "counter_intuitive_abort":
			applied = append(applied, "recorded counter-intuitive issue")
		}
	}
	return applied, nil
}
