package cobrax

import (
	"testing"

	"github.com/gh-xj/gokit"
)

func TestNewRootHasRequiredPersistentFlags(t *testing.T) {
	root := NewRoot(RootSpec{Use: "demo"})
	required := []string{"verbose", "config", "json", "no-color"}
	for _, name := range required {
		if root.PersistentFlags().Lookup(name) == nil {
			t.Fatalf("missing persistent flag: %s", name)
		}
	}
}

func TestExecuteSuccess(t *testing.T) {
	spec := RootSpec{
		Use: "demo",
		Meta: gokit.AppMeta{
			Name: "demo",
		},
		Commands: []CommandSpec{
			{
				Use: "ping",
				Run: func(*gokit.AppContext, []string) error {
					return nil
				},
			},
		},
	}
	code := Execute(spec, []string{"ping"})
	if code != gokit.ExitSuccess {
		t.Fatalf("unexpected exit code: %d", code)
	}
}

func TestExecuteUsageErrorForUnknownCommand(t *testing.T) {
	spec := RootSpec{
		Use: "demo",
		Commands: []CommandSpec{
			{
				Use: "ping",
				Run: func(*gokit.AppContext, []string) error {
					return nil
				},
			},
		},
	}
	code := Execute(spec, []string{"unknown"})
	if code != gokit.ExitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
}

func TestExecuteTypedExitCode(t *testing.T) {
	spec := RootSpec{
		Use: "demo",
		Commands: []CommandSpec{
			{
				Use: "fail",
				Run: func(*gokit.AppContext, []string) error {
					return gokit.NewCLIError(gokit.ExitPreflightDependency, "preflight", "missing dependency", nil)
				},
			},
		},
	}
	code := Execute(spec, []string{"fail"})
	if code != gokit.ExitPreflightDependency {
		t.Fatalf("expected %d, got %d", gokit.ExitPreflightDependency, code)
	}
}
