package main

import (
	"os"

	"github.com/gh-xj/agentops/examples/file-sync-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
