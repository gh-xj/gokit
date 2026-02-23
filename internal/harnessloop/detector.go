package harnessloop

import "strings"

func DetectFindings(r ScenarioResult) []Finding {
	findings := make([]Finding, 0)
	if !r.OK {
		for _, step := range r.Steps {
			if step.ExitCode == 0 {
				continue
			}
			combined := strings.ToLower(step.CombinedTail)
			findings = append(findings, Finding{
				Code:     "step_failed",
				Severity: "high",
				Message:  "scenario step failed: " + step.Name,
				Source:   step.Name,
			})
			if strings.Contains(combined, "fmt:check") && strings.Contains(combined, "exit status 1") {
				findings = append(findings, Finding{
					Code:     "generated_go_not_formatted",
					Severity: "high",
					Message:  "generated go file is not gofmt-clean",
					Source:   step.Name,
				})
			}
			if strings.Contains(combined, "failed to run task") && strings.Contains(combined, "exit status") {
				findings = append(findings, Finding{
					Code:     "counter_intuitive_abort",
					Severity: "medium",
					Message:  "flow aborts without clear user-oriented guidance",
					Source:   step.Name,
				})
			}
			break
		}
	}
	return findings
}
