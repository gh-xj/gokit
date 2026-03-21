package main

import (
	"os"

	"github.com/gh-xj/agentops/examples/deploy-helper-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
