package main

import (
	"context"
	"os"

	agentops "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/cobrax"
	"github.com/gh-xj/agentops/dal"
	"github.com/gh-xj/agentops/resource"
	caseresource "github.com/gh-xj/agentops/resource/case"
	projectresource "github.com/gh-xj/agentops/resource/project"
	slotresource "github.com/gh-xj/agentops/resource/slot"
	"github.com/gh-xj/agentops/strategy"
)

var appMeta = agentops.AppMeta{
	Name:    "agentops",
	Version: "dev",
	Commit:  "none",
	Date:    "unknown",
}

func main() {
	ctx := agentops.NewAppContext(context.Background())
	fs := dal.NewFileSystem()
	exec := dal.NewExecutor()

	// Strategy loading is optional (commands like "new" don't need it).
	strat, _ := strategy.Discover(".")

	reg := resource.NewRegistry()
	reg.Register(caseresource.New(fs, exec, strat))
	reg.Register(slotresource.New(fs, exec))
	reg.Register(projectresource.New(fs, exec))

	root := cobrax.BuildRoot(cobrax.RootSpec{
		Use:   "agentops",
		Short: "Agent operations toolkit",
		Meta:  appMeta,
	}, reg, ctx)

	root.AddCommand(newInitCmd(fs))
	root.AddCommand(newDoctorCmd(reg, ctx))
	root.AddCommand(newNewCmd(reg, ctx))
	root.AddCommand(newVersionCmd())
	root.AddCommand(newLoopCmd())
	root.AddCommand(newLoopServerCmd())

	os.Exit(cobrax.ExecuteRoot(root, os.Args[1:]))
}
