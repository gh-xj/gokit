package harnessloop

func Judge(r ScenarioResult, findings []Finding, threshold float64) JudgeScore {
	passRate := 0.0
	if r.OK {
		passRate = 1.0
	}
	ux := 5.0 * passRate
	quality := 5.0 * passRate
	penalty := 0.0
	hardFailures := 0
	counter := 0
	for _, f := range findings {
		switch f.Code {
		case "step_failed", "generated_go_not_formatted":
			hardFailures++
			penalty += 0.7
		case "onboarding_install_missing":
			hardFailures++
			penalty += 2.5
		case "onboarding_install_verify_missing":
			counter++
			penalty += 1.2
		case "counter_intuitive_abort":
			counter++
			penalty += 0.3
		default:
			penalty += 0.1
		}
	}
	if penalty > 2.0 {
		penalty = 2.0
	}
	score := ux + quality - penalty
	if score < 0 {
		score = 0
	}
	return JudgeScore{
		Score:                score,
		Threshold:            threshold,
		Pass:                 score >= threshold,
		UXScore:              ux,
		QualityScore:         quality,
		PenaltyScore:         penalty,
		ScenarioPassRate:     passRate,
		CounterIntuitiveFind: counter,
		HardFailures:         hardFailures,
	}
}
