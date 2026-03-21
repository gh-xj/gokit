package main

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print agentops version information",
		Run: func(cmd *cobra.Command, args []string) {
			v, c, d := resolveVersion()
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s %s)\n", appMeta.Name, v, c, d)
		},
	}
}

func resolveVersion() (string, string, string) {
	outVersion := appMeta.Version
	outCommit := appMeta.Commit
	outDate := appMeta.Date

	if info, ok := debug.ReadBuildInfo(); ok {
		if outVersion == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			outVersion = info.Main.Version
		}
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if outCommit == "none" && setting.Value != "" {
					outCommit = setting.Value
				}
			case "vcs.time":
				if outDate == "unknown" && setting.Value != "" {
					outDate = strings.TrimSpace(strings.SplitN(setting.Value, " ", 2)[0])
				}
			}
		}
	}
	if outVersion == "" {
		outVersion = "dev"
	}
	return outVersion, outCommit, outDate
}
