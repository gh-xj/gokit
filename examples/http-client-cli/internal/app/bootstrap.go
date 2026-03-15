package app

import (
	"os"
	"slices"

	"github.com/gh-xj/agentcli-go/dal"
)

func Bootstrap() {
	verbose := slices.Contains(os.Args, "-v") || slices.Contains(os.Args, "--verbose")
	dal.NewLogger().Init(verbose, os.Stderr)
}
