package harnessloop

import "fmt"

func DefaultOnboardingScenario(agentcliBin string) Scenario {
	return Scenario{
		Name:        "default-onboarding",
		Description: "Mimic user onboarding flow in a clean temp project",
		WorkDir:     "",
		Steps: []Step{
			{Name: "scaffold", Command: fmt.Sprintf("%s new --module example.com/demo demo", agentcliBin)},
			{Name: "add-command", Command: fmt.Sprintf("%s add command --dir ./demo --preset file-sync sync-data", agentcliBin)},
			{Name: "doctor", Command: fmt.Sprintf("%s doctor --dir ./demo --json", agentcliBin)},
			{Name: "verify", Command: "cd demo && task verify"},
		},
	}
}
