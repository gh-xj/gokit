package main

import (
	"fmt"

	agentops "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/resource"
	"github.com/spf13/cobra"
)

func newNewCmd(reg *resource.Registry, ctx *agentops.AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new project (alias for project create)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRes, ok := reg.Get("project")
			if !ok {
				return fmt.Errorf("project resource not registered")
			}

			module, _ := cmd.Flags().GetString("module")
			mode, _ := cmd.Flags().GetString("mode")
			baseDir, _ := cmd.Flags().GetString("dir")

			opts := map[string]string{}
			if module != "" {
				opts["module"] = module
			}
			if mode != "" {
				opts["mode"] = mode
			}
			if baseDir != "" {
				opts["base_dir"] = baseDir
			}

			record, err := projectRes.Create(ctx, args[0], opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created project: %s\n", record.RawPath)
			return nil
		},
	}
	cmd.Flags().String("module", "", "Go module path (defaults to project name)")
	cmd.Flags().String("mode", "", "Scaffold mode: minimal|lean|full (default lean)")
	return cmd
}
