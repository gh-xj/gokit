package cobrax

import (
	"testing"

	agentops "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/resource"
	"github.com/spf13/cobra"
)

// mockResource implements the core Resource interface.
type mockResource struct{}

func (m *mockResource) Schema() resource.ResourceSchema {
	return resource.ResourceSchema{
		Kind:        "mock",
		Description: "A mock resource for testing",
		Fields: []resource.FieldDef{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "status", Type: "string"},
		},
		CreateArgs: []resource.ArgDef{
			{Name: "slug", Description: "resource slug", Required: true},
		},
	}
}

func (m *mockResource) Create(ctx *agentops.AppContext, slug string, opts map[string]string) (*resource.Record, error) {
	return &resource.Record{
		Kind:   "mock",
		ID:     "mock-001",
		Fields: map[string]any{"id": "mock-001", "name": slug, "status": "active"},
	}, nil
}

func (m *mockResource) List(ctx *agentops.AppContext, filter resource.Filter) ([]resource.Record, error) {
	return []resource.Record{
		{Kind: "mock", ID: "mock-001", Fields: map[string]any{"id": "mock-001", "name": "test", "status": "active"}},
	}, nil
}

func (m *mockResource) Get(ctx *agentops.AppContext, id string) (*resource.Record, error) {
	return &resource.Record{
		Kind:   "mock",
		ID:     id,
		Fields: map[string]any{"id": id, "name": "test", "status": "active"},
	}, nil
}

// mockDeleterResource implements Resource + Deleter.
type mockDeleterResource struct {
	mockResource
}

func (m *mockDeleterResource) Delete(ctx *agentops.AppContext, id string) error {
	return nil
}

// mockFullResource implements Resource + Validator + Deleter + Syncer + Transitioner.
type mockFullResource struct {
	mockResource
}

func (m *mockFullResource) Schema() resource.ResourceSchema {
	s := m.mockResource.Schema()
	s.Kind = "full"
	s.Statuses = []string{"active", "inactive"}
	return s
}

func (m *mockFullResource) Validate(ctx *agentops.AppContext, id string) (*agentops.DoctorReport, error) {
	return &agentops.DoctorReport{SchemaVersion: "1", OK: true}, nil
}

func (m *mockFullResource) Delete(ctx *agentops.AppContext, id string) error {
	return nil
}

func (m *mockFullResource) Sync(ctx *agentops.AppContext, id string) error {
	return nil
}

func (m *mockFullResource) Transition(ctx *agentops.AppContext, id string, action string) (*resource.Record, error) {
	return &resource.Record{
		Kind:   "full",
		ID:     id,
		Fields: map[string]any{"id": id, "name": "test", "status": action},
	}, nil
}

// mockDoctorPrunerResource implements Resource + Doctor + Pruner.
type mockDoctorPrunerResource struct {
	mockResource
}

func (m *mockDoctorPrunerResource) Schema() resource.ResourceSchema {
	s := m.mockResource.Schema()
	s.Kind = "healthcheck"
	return s
}

func (m *mockDoctorPrunerResource) Doctor(ctx *agentops.AppContext) ([]resource.DoctorCheck, error) {
	return []resource.DoctorCheck{{Name: "test", Status: "ok", Message: "all good", Severity: "ok"}}, nil
}

func (m *mockDoctorPrunerResource) Prune(ctx *agentops.AppContext, confirm bool) ([]resource.PruneResult, error) {
	return []resource.PruneResult{{Name: "test", Path: "/tmp/test", Action: "would_remove"}}, nil
}

func findSubCommand(cmd *cobra.Command, path ...string) *cobra.Command {
	current := cmd
	for _, name := range path {
		found := false
		for _, child := range current.Commands() {
			if child.Name() == name {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}

func TestGenerateCommands(t *testing.T) {
	reg := resource.NewRegistry()
	reg.Register(&mockResource{})

	root := &cobra.Command{Use: "test"}
	root.PersistentFlags().String("json", "", "JSON field selection")
	root.PersistentFlags().String("jq", "", "jq expression")
	ctx := agentops.NewAppContext(nil)

	GenerateResourceCommands(reg, root, ctx)

	// Verify noun command exists
	mockCmd := findSubCommand(root, "mock")
	if mockCmd == nil {
		t.Fatal("expected 'mock' subcommand to exist")
	}

	// Verify core verb commands exist
	for _, verb := range []string{"create", "list", "get"} {
		cmd := findSubCommand(root, "mock", verb)
		if cmd == nil {
			t.Fatalf("expected 'mock %s' subcommand to exist", verb)
		}
	}
}

func TestGenerateOptionalCommands(t *testing.T) {
	t.Run("with Deleter", func(t *testing.T) {
		reg := resource.NewRegistry()
		reg.Register(&mockDeleterResource{})

		root := &cobra.Command{Use: "test"}
		root.PersistentFlags().String("json", "", "JSON field selection")
		root.PersistentFlags().String("jq", "", "jq expression")
		ctx := agentops.NewAppContext(nil)

		GenerateResourceCommands(reg, root, ctx)

		// Should have "remove" command
		removeCmd := findSubCommand(root, "mock", "remove")
		if removeCmd == nil {
			t.Fatal("expected 'mock remove' subcommand to exist for Deleter")
		}

		// Should NOT have "validate", "sync", "transition"
		for _, verb := range []string{"validate", "sync", "transition"} {
			cmd := findSubCommand(root, "mock", verb)
			if cmd != nil {
				t.Fatalf("expected 'mock %s' subcommand NOT to exist", verb)
			}
		}
	})

	t.Run("without Deleter", func(t *testing.T) {
		reg := resource.NewRegistry()
		reg.Register(&mockResource{})

		root := &cobra.Command{Use: "test"}
		root.PersistentFlags().String("json", "", "JSON field selection")
		root.PersistentFlags().String("jq", "", "jq expression")
		ctx := agentops.NewAppContext(nil)

		GenerateResourceCommands(reg, root, ctx)

		// Should NOT have "remove" command
		removeCmd := findSubCommand(root, "mock", "remove")
		if removeCmd != nil {
			t.Fatal("expected 'mock remove' subcommand NOT to exist for non-Deleter")
		}
	})

	t.Run("full resource", func(t *testing.T) {
		reg := resource.NewRegistry()
		reg.Register(&mockFullResource{})

		root := &cobra.Command{Use: "test"}
		root.PersistentFlags().String("json", "", "JSON field selection")
		root.PersistentFlags().String("jq", "", "jq expression")
		ctx := agentops.NewAppContext(nil)

		GenerateResourceCommands(reg, root, ctx)

		// Should have all commands
		for _, verb := range []string{"create", "list", "get", "validate", "remove", "sync", "transition"} {
			cmd := findSubCommand(root, "full", verb)
			if cmd == nil {
				t.Fatalf("expected 'full %s' subcommand to exist", verb)
			}
		}
	})
}

func TestGenerateDoctorPruneCommands(t *testing.T) {
	reg := resource.NewRegistry()
	reg.Register(&mockDoctorPrunerResource{})

	root := &cobra.Command{Use: "test"}
	root.PersistentFlags().String("json", "", "JSON field selection")
	root.PersistentFlags().String("jq", "", "jq expression")
	ctx := agentops.NewAppContext(nil)

	GenerateResourceCommands(reg, root, ctx)

	// Should have "doctor" and "prune" commands
	doctorCmd := findSubCommand(root, "healthcheck", "doctor")
	if doctorCmd == nil {
		t.Fatal("expected 'healthcheck doctor' subcommand to exist")
	}

	pruneCmd := findSubCommand(root, "healthcheck", "prune")
	if pruneCmd == nil {
		t.Fatal("expected 'healthcheck prune' subcommand to exist")
	}

	// Verify prune has --confirm flag
	confirmFlag := pruneCmd.Flags().Lookup("confirm")
	if confirmFlag == nil {
		t.Fatal("expected prune to have --confirm flag")
	}

	// Verify base resource without Doctor/Pruner does NOT have these commands
	reg2 := resource.NewRegistry()
	reg2.Register(&mockResource{})
	root2 := &cobra.Command{Use: "test2"}
	root2.PersistentFlags().String("json", "", "")
	root2.PersistentFlags().String("jq", "", "")
	GenerateResourceCommands(reg2, root2, ctx)

	if findSubCommand(root2, "mock", "doctor") != nil {
		t.Fatal("expected 'mock doctor' NOT to exist for non-Doctor resource")
	}
	if findSubCommand(root2, "mock", "prune") != nil {
		t.Fatal("expected 'mock prune' NOT to exist for non-Pruner resource")
	}
}

func TestBuildRootHasGlobalFlags(t *testing.T) {
	spec := RootSpec{
		Use:   "testapp",
		Short: "A test app",
	}
	reg := resource.NewRegistry()
	ctx := agentops.NewAppContext(nil)

	root := BuildRoot(spec, reg, ctx)

	requiredFlags := []string{"json", "jq", "dir", "verbose", "no-color"}
	for _, name := range requiredFlags {
		if root.PersistentFlags().Lookup(name) == nil {
			t.Fatalf("missing persistent flag: %s", name)
		}
	}
}

func TestExecuteRootReturnsExitCode(t *testing.T) {
	spec := RootSpec{
		Use:   "testapp",
		Short: "A test app",
		Commands: []CommandSpec{
			{
				Use:   "ping",
				Short: "Ping command",
				Run: func(*agentops.AppContext, []string) error {
					return nil
				},
			},
		},
	}
	reg := resource.NewRegistry()
	reg.Register(&mockResource{})
	ctx := agentops.NewAppContext(nil)

	root := BuildRoot(spec, reg, ctx)

	// Running with unknown command should return usage error
	code := ExecuteRoot(root, []string{"nonexistent"})
	if code != agentops.ExitUsage {
		t.Fatalf("expected exit code %d for unknown command, got %d", agentops.ExitUsage, code)
	}
}
