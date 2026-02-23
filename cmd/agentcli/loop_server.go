package main

import (
	"fmt"
	"os"

	agentcli "github.com/gh-xj/agentcli-go"
	"github.com/gh-xj/agentcli-go/internal/loopapi"
)

func runLoopServer(args []string) int {
	addr := "127.0.0.1:7878"
	repoRoot := "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--addr":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--addr requires a value")
				return agentcli.ExitUsage
			}
			addr = args[i+1]
			i++
		case "--repo-root":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--repo-root requires a value")
				return agentcli.ExitUsage
			}
			repoRoot = args[i+1]
			i++
		default:
			fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", args[i])
			return agentcli.ExitUsage
		}
	}

	if err := loopapi.Serve(addr, repoRoot); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return agentcli.ExitFailure
	}
	return agentcli.ExitSuccess
}
