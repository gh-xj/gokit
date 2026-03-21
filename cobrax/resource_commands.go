package cobrax

import (
	"fmt"
	"os"
	"strings"

	agentcli "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/resource"
	"github.com/spf13/cobra"
)

// GenerateResourceCommands walks the registry and creates noun-verb commands.
// For each resource:
//   - Always: create, list, get
//   - If Validator: validate
//   - If Deleter: remove
//   - If Syncer: sync
//   - If Transitioner: transition
func GenerateResourceCommands(reg *resource.Registry, root *cobra.Command, ctx *agentcli.AppContext) {
	for _, res := range reg.All() {
		schema := res.Schema()
		nounCmd := &cobra.Command{
			Use:   schema.Kind,
			Short: schema.Description,
		}

		// Always add core commands
		nounCmd.AddCommand(makeCreateCmd(res, schema, ctx))
		nounCmd.AddCommand(makeListCmd(res, schema, ctx))
		nounCmd.AddCommand(makeGetCmd(res, schema, ctx))

		// Optional: validate
		if v, ok := res.(resource.Validator); ok {
			nounCmd.AddCommand(makeValidateCmd(v, schema, ctx))
		}

		// Optional: remove (Deleter)
		if d, ok := res.(resource.Deleter); ok {
			nounCmd.AddCommand(makeRemoveCmd(d, schema, ctx))
		}

		// Optional: sync
		if s, ok := res.(resource.Syncer); ok {
			nounCmd.AddCommand(makeSyncCmd(s, schema, ctx))
		}

		// Optional: transition
		if tr, ok := res.(resource.Transitioner); ok {
			nounCmd.AddCommand(makeTransitionCmd(tr, schema, ctx))
		}

		root.AddCommand(nounCmd)
	}
}

// resolveOutputMode reads --json and --jq flags from the command and returns
// the appropriate output mode, field list, and jq expression.
func resolveOutputMode(cmd *cobra.Command) (OutputMode, []string, string) {
	jsonFields, _ := cmd.Flags().GetString("json")
	jqExpr, _ := cmd.Flags().GetString("jq")

	if jqExpr != "" {
		return OutputJQ, nil, jqExpr
	}
	if jsonFields != "" {
		fields := parseFieldList(jsonFields)
		return OutputJSON, fields, ""
	}
	return OutputAuto, nil, ""
}

// parseFieldList splits a comma-separated field list.
func parseFieldList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	fields := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			fields = append(fields, p)
		}
	}
	return fields
}

func makeCreateCmd(res resource.Resource, schema resource.ResourceSchema, ctx *agentcli.AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "create <slug>",
		Short: fmt.Sprintf("Create a new %s", schema.Kind),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := res.Create(ctx, args[0], nil)
			if err != nil {
				return err
			}
			mode, fields, jqExpr := resolveOutputMode(cmd)
			records := []resource.Record{*record}
			return RenderRecords(cmd.OutOrStdout(), records, schema, mode, fields, jqExpr)
		},
	}
}

func makeListCmd(res resource.Resource, schema resource.ResourceSchema, ctx *agentcli.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s resources", schema.Kind),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := resource.Filter{}
			if status, _ := cmd.Flags().GetString("status"); status != "" {
				filter["status"] = status
			}
			if slot, _ := cmd.Flags().GetString("slot"); slot != "" {
				filter["slot"] = slot
			}

			records, err := res.List(ctx, filter)
			if err != nil {
				return err
			}
			mode, fields, jqExpr := resolveOutputMode(cmd)
			return RenderRecords(cmd.OutOrStdout(), records, schema, mode, fields, jqExpr)
		},
	}
	cmd.Flags().String("status", "", "filter by status")
	cmd.Flags().String("slot", "", "filter by slot")
	return cmd
}

func makeGetCmd(res resource.Resource, schema resource.ResourceSchema, ctx *agentcli.AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: fmt.Sprintf("Get a %s by ID", schema.Kind),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := res.Get(ctx, args[0])
			if err != nil {
				return err
			}
			mode, fields, jqExpr := resolveOutputMode(cmd)
			records := []resource.Record{*record}
			return RenderRecords(cmd.OutOrStdout(), records, schema, mode, fields, jqExpr)
		},
	}
}

func makeValidateCmd(v resource.Validator, schema resource.ResourceSchema, ctx *agentcli.AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "validate <id>",
		Short: fmt.Sprintf("Validate a %s", schema.Kind),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := v.Validate(ctx, args[0])
			if err != nil {
				return err
			}
			jsonFields, _ := cmd.Flags().GetString("json")
			jqExpr, _ := cmd.Flags().GetString("jq")
			jsonMode := jsonFields != "" || jqExpr != ""
			return RenderDoctorReport(cmd.OutOrStdout(), *report, jsonMode)
		},
	}
}

func makeRemoveCmd(d resource.Deleter, schema resource.ResourceSchema, ctx *agentcli.AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <id>",
		Short: fmt.Sprintf("Remove a %s", schema.Kind),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := d.Delete(ctx, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %s %s\n", schema.Kind, args[0])
			return nil
		},
	}
}

func makeSyncCmd(s resource.Syncer, schema resource.ResourceSchema, ctx *agentcli.AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "sync <id>",
		Short: fmt.Sprintf("Sync a %s", schema.Kind),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.Sync(ctx, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Synced %s %s\n", schema.Kind, args[0])
			return nil
		},
	}
}

func makeTransitionCmd(tr resource.Transitioner, schema resource.ResourceSchema, ctx *agentcli.AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "transition <id> <action>",
		Short: fmt.Sprintf("Transition a %s to a new state", schema.Kind),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := tr.Transition(ctx, args[0], args[1])
			if err != nil {
				return err
			}
			mode, fields, jqExpr := resolveOutputMode(cmd)
			records := []resource.Record{*record}
			return RenderRecords(cmd.OutOrStdout(), records, schema, mode, fields, jqExpr)
		},
	}
}

// BuildRoot creates a root command with global flags and auto-generated resource commands.
func BuildRoot(spec RootSpec, reg *resource.Registry, ctx *agentcli.AppContext) *cobra.Command {
	root := &cobra.Command{
		Use:          spec.Use,
		Short:        spec.Short,
		SilenceUsage: true,
	}

	// Global persistent flags
	root.PersistentFlags().BoolP("verbose", "v", false, "enable debug logs")
	root.PersistentFlags().Bool("no-color", false, "disable colorized output")
	root.PersistentFlags().String("json", "", "output as JSON with optional field selection (comma-separated)")
	root.PersistentFlags().String("jq", "", "filter JSON output with a jq expression")
	root.PersistentFlags().String("dir", "", "working directory path")

	// Add any manually-specified commands from RootSpec
	for _, c := range spec.Commands {
		cmdSpec := c
		child := &cobra.Command{
			Use:   cmdSpec.Use,
			Short: cmdSpec.Short,
			RunE: func(cmd *cobra.Command, args []string) error {
				if cmdSpec.Run == nil {
					return nil
				}
				return cmdSpec.Run(ctx, args)
			},
		}
		root.AddCommand(child)
	}

	// Auto-generate resource commands
	GenerateResourceCommands(reg, root, ctx)

	return root
}

// ExecuteRoot runs a pre-built root command with exit code handling.
func ExecuteRoot(root *cobra.Command, args []string) int {
	root.SetArgs(args)
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return resolveCode(err)
	}
	return agentcli.ExitSuccess
}
