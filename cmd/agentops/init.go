package main

import (
	"fmt"

	"github.com/gh-xj/agentops/dal"
	"github.com/gh-xj/agentops/strategy"
	"github.com/spf13/cobra"
)

func newInitCmd(fs dal.FileSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap .agentops/ with default strategy files",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			if dir == "" {
				dir = "."
			}
			if err := strategy.Bootstrap(dir); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized .agentops/ in %s\n", dir)
			return nil
		},
	}
	return cmd
}
