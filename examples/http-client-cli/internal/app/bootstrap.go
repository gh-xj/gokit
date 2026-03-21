package app

import (
	"os"
	"slices"

	"github.com/gh-xj/agentops/dal"
)

func Bootstrap() {
	verbose := slices.Contains(os.Args, "-v") || slices.Contains(os.Args, "--verbose")
	dal.NewLogger().Init(verbose, os.Stderr)
}
