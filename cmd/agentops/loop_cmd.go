package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newLoopCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "loop",
		Short:              "Run harness loop commands (run, judge, autofix, doctor, ...)",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			code := runLoop(args)
			if code != 0 {
				os.Exit(code)
			}
		},
	}
}

func newLoopServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "loop-server",
		Short:              "Start the loop HTTP API server",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			code := runLoopServer(args)
			if code != 0 {
				os.Exit(code)
			}
		},
	}
}
