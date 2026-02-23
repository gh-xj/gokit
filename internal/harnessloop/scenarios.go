package harnessloop

import "fmt"

func DefaultOnboardingScenario(agentcliBin string) Scenario {
	verifyCmd := "cd demo && if command -v task >/dev/null 2>&1; then task verify; else go test ./... && go build ./... && mkdir -p test/smoke && rm -f test/smoke/version.output.json && ./demo --json version > test/smoke/version.output.json && go run ./internal/tools/smokecheck --schema test/smoke/version.schema.json --input test/smoke/version.output.json; fi"
	return Scenario{
		Name:        "default-onboarding",
		Description: "Mimic user onboarding flow in a clean temp project",
		WorkDir:     "",
		Steps: []Step{
			{Name: "scaffold", Command: fmt.Sprintf("%s new --module example.com/demo demo", agentcliBin)},
			{Name: "add-command", Command: fmt.Sprintf("%s add command --dir ./demo --preset file-sync sync-data", agentcliBin)},
			{Name: "doctor", Command: fmt.Sprintf("%s doctor --dir ./demo --json", agentcliBin)},
			{Name: "verify", Command: verifyCmd},
		},
	}
}
