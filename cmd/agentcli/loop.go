//go:build !agentcli_core
// +build !agentcli_core

package main

import (
	"fmt"
	"os"
	"time"

	harnessloop "github.com/gh-xj/agentcli-go/internal/harnessloop"
	"github.com/gh-xj/agentcli-go/internal/loopapi"
	harness "github.com/gh-xj/agentcli-go/tools/harness"
	loopcommands "github.com/gh-xj/agentcli-go/tools/harness/commands"
)

type loopProfile struct {
	mode             string
	roleConfig       string
	maxIterations    int
	threshold        float64
	budget           int
	verboseArtifacts bool
}

var loopProfiles = fromLoopProfiles(loopcommands.LoopProfiles)

type loopRuntimeFlags struct {
	Format      string
	SummaryPath string
	NoColor     bool
	DryRun      bool
	Explain     bool
}

type loopRegressionFlags struct {
	Profile       string
	BaselinePath  string
	WriteBaseline bool
}

type loopFlags struct {
	RepoRoot      string
	Threshold     float64
	MaxIterations int
	Branch        string
	APIURL        string
	Markdown      bool
}

type loopProfileFlags struct {
	loopFlags
	RoleConfig         string
	VerboseArtifacts   bool
	NoVerboseArtifacts bool
}

type loopLabFlags struct {
	loopFlags
	Mode               string
	RoleConfig         string
	Seed               int64
	Budget             int
	RunA               string
	RunB               string
	RunID              string
	Iteration          int
	Format             string
	Out                string
	VerboseArtifacts   bool
	NoVerboseArtifacts bool
}

func runLoop(args []string) int {
	runtime := loopRuntimeFlags{Format: "text"}
	if len(args) == 0 {
		return emitLoopFailureSummary(
			"loop",
			runtime,
			harness.NewFailure(
				harness.CodeUsage,
				"usage: agentcli loop [global flags] [run|judge|autofix|doctor|quality|profiles|profile|<profile>|regression|capabilities|lab] [command flags]",
				"",
				false,
			),
		)
	}

	parsedRuntime, remaining, err := parseLoopRuntimeFlags(args)
	runtime = parsedRuntime
	if err != nil {
		return emitLoopFailureSummary("loop", runtime, err)
	}
	if len(remaining) == 0 {
		return emitLoopFailureSummary(
			"loop",
			runtime,
			harness.NewFailure(
				harness.CodeUsage,
				"usage: agentcli loop [global flags] [run|judge|autofix|doctor|quality|profiles|profile|<profile>|regression|capabilities|lab] [command flags]",
				"",
				false,
			),
		)
	}

	handler := loopcommands.NewLoopHandler(runLoopWithOptionalAPI)
	action := remaining[0]
	actionArgs := remaining[1:]
	summary, execErr := harness.Run(harness.CommandInput{
		Name:        "loop " + action,
		SummaryPath: runtime.SummaryPath,
		DryRun:      runtime.DryRun,
		Explain:     runtime.Explain,
		Execute: func(ctx harness.Context) (harness.CommandOutcome, error) {
			return handler.ExecuteLoopAction(action, actionArgs, ctx)
		},
	})

	rendered, renderErr := harness.RenderSummary(summary, runtime.Format, runtime.NoColor)
	if renderErr != nil {
		fmt.Fprintln(os.Stderr, renderErr.Error())
		return harness.ExitCodeFor(renderErr)
	}
	fmt.Fprint(os.Stdout, rendered)
	return harness.ExitCodeFor(execErr)
}

func runLoopWithOptionalAPI(apiURL, action string, cfg harnessloop.Config) (harnessloop.RunResult, error) {
	if apiURL != "" {
		return loopapi.Run(apiURL, loopapi.RunRequest{
			Action:           action,
			RepoRoot:         cfg.RepoRoot,
			Threshold:        cfg.Threshold,
			MaxIterations:    cfg.MaxIterations,
			Branch:           cfg.Branch,
			Mode:             cfg.Mode,
			RoleConfig:       cfg.RoleConfigPath,
			VerboseArtifacts: cfg.VerboseArtifacts,
			Seed:             cfg.Seed,
			Budget:           cfg.Budget,
		})
	}
	return harnessloop.RunLoop(cfg)
}

func emitLoopFailureSummary(command string, runtime loopRuntimeFlags, err error) int {
	now := time.Now().UTC()
	summary := harness.CommandSummary{
		SchemaVersion: harness.SummarySchemaVersion,
		Command:       command,
		Status:        harness.StatusFail,
		StartedAt:     now,
		FinishedAt:    now,
		DurationMs:    0,
		Failures: []harness.Failure{
			harness.FailureFromError(err),
		},
	}
	format := runtime.Format
	if format == "" {
		format = "text"
	}
	rendered, renderErr := harness.RenderSummary(summary, format, runtime.NoColor)
	if renderErr != nil {
		fmt.Fprintln(os.Stderr, renderErr.Error())
		return harness.ExitCodeFor(renderErr)
	}
	fmt.Fprint(os.Stdout, rendered)
	return harness.ExitCodeFor(err)
}

func getLoopProfiles(repoRoot string) (map[string]loopProfile, error) {
	profiles, err := loopcommands.GetLoopProfiles(repoRoot)
	if err != nil {
		return nil, err
	}
	return fromLoopProfiles(profiles), nil
}

func parseLoopProfilesRepoRoot(args []string) (string, error) {
	return loopcommands.ParseLoopProfilesRepoRoot(args)
}

func parseLoopRuntimeFlags(args []string) (loopRuntimeFlags, []string, error) {
	flags, remaining, err := loopcommands.ParseLoopRuntimeFlags(args)
	if err != nil {
		return loopRuntimeFlags{}, nil, err
	}
	return loopRuntimeFlags{
		Format:      flags.Format,
		SummaryPath: flags.SummaryPath,
		NoColor:     flags.NoColor,
		DryRun:      flags.DryRun,
		Explain:     flags.Explain,
	}, remaining, nil
}

func parseLoopFlags(args []string) (loopFlags, error) {
	flags, err := loopcommands.ParseLoopFlags(args)
	if err != nil {
		return loopFlags{}, err
	}
	return fromLoopFlags(flags), nil
}

func parseLoopQualityFlags(profile loopProfile, args []string) (loopProfileFlags, error) {
	flags, err := loopcommands.ParseLoopQualityFlags(toLoopProfile(profile), args)
	if err != nil {
		return loopProfileFlags{}, err
	}
	return fromLoopProfileFlags(flags), nil
}

func parseLoopLabFlags(args []string) (loopLabFlags, error) {
	flags, err := loopcommands.ParseLoopLabFlags(args)
	if err != nil {
		return loopLabFlags{}, err
	}
	return fromLoopLabFlags(flags), nil
}

func parseLoopRegressionFlags(args []string) (loopRegressionFlags, []string, error) {
	flags, remaining, err := loopcommands.ParseLoopRegressionFlags(args)
	if err != nil {
		return loopRegressionFlags{}, nil, err
	}
	return loopRegressionFlags{Profile: flags.Profile, BaselinePath: flags.BaselinePath, WriteBaseline: flags.WriteBaseline}, remaining, nil
}

func resolveLoopRegressionBaselinePath(repoRoot, profileName, baselinePath string) string {
	return loopcommands.ResolveLoopRegressionBaselinePath(repoRoot, profileName, baselinePath)
}

func resolveVerboseArtifacts(defaultValue, forceEnable, forceDisable bool) (bool, error) {
	return loopcommands.ResolveVerboseArtifacts(defaultValue, forceEnable, forceDisable)
}

func toLoopProfile(p loopProfile) loopcommands.LoopProfile {
	return loopcommands.LoopProfile{
		Mode:             p.mode,
		RoleConfig:       p.roleConfig,
		MaxIterations:    p.maxIterations,
		Threshold:        p.threshold,
		Budget:           p.budget,
		VerboseArtifacts: p.verboseArtifacts,
	}
}

func fromLoopProfile(p loopcommands.LoopProfile) loopProfile {
	return loopProfile{
		mode:             p.Mode,
		roleConfig:       p.RoleConfig,
		maxIterations:    p.MaxIterations,
		threshold:        p.Threshold,
		budget:           p.Budget,
		verboseArtifacts: p.VerboseArtifacts,
	}
}

func fromLoopProfiles(in map[string]loopcommands.LoopProfile) map[string]loopProfile {
	out := make(map[string]loopProfile, len(in))
	for name, profile := range in {
		out[name] = fromLoopProfile(profile)
	}
	return out
}

func fromLoopFlags(in loopcommands.LoopFlags) loopFlags {
	return loopFlags{
		RepoRoot:      in.RepoRoot,
		Threshold:     in.Threshold,
		MaxIterations: in.MaxIterations,
		Branch:        in.Branch,
		APIURL:        in.APIURL,
		Markdown:      in.Markdown,
	}
}

func fromLoopProfileFlags(in loopcommands.LoopProfileFlags) loopProfileFlags {
	return loopProfileFlags{
		loopFlags:          fromLoopFlags(in.LoopFlags),
		RoleConfig:         in.RoleConfig,
		VerboseArtifacts:   in.VerboseArtifacts,
		NoVerboseArtifacts: in.NoVerboseArtifacts,
	}
}

func fromLoopLabFlags(in loopcommands.LoopLabFlags) loopLabFlags {
	return loopLabFlags{
		loopFlags:          fromLoopFlags(in.LoopFlags),
		Mode:               in.Mode,
		RoleConfig:         in.RoleConfig,
		Seed:               in.Seed,
		Budget:             in.Budget,
		RunA:               in.RunA,
		RunB:               in.RunB,
		RunID:              in.RunID,
		Iteration:          in.Iteration,
		Format:             in.Format,
		Out:                in.Out,
		VerboseArtifacts:   in.VerboseArtifacts,
		NoVerboseArtifacts: in.NoVerboseArtifacts,
	}
}
