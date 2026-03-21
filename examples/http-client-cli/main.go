package main

import (
	"os"

	"github.com/gh-xj/agentops/examples/http-client-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
