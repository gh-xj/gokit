//go:build agentcli_core
// +build agentcli_core

package main

import (
	"fmt"
	"os"

	agentcli "github.com/gh-xj/agentcli-go"
)

func runLoop(_ []string) int {
	fmt.Fprintln(os.Stderr, "loop features are disabled in agentcli_core build; rebuild without -tags agentcli_core to enable loop/harness commands")
	return agentcli.ExitUsage
}
