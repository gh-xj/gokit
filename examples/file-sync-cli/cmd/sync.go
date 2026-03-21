package cmd

import (
	"fmt"
	"os"

	"github.com/gh-xj/agentops"
)

func init() {
}

func SyncCommand() command {
	return command{
		Description: "describe sync",
		Run: func(app *agentops.AppContext, args []string) error {
			if jsonOutput, _ := app.Values["json"].(bool); jsonOutput {
				_, err := fmt.Fprintln(os.Stdout, "{\"command\":\"sync\",\"ok\":true}")
				return err
			}
			_, err := fmt.Fprintf(os.Stdout, "sync executed with %d args\n", len(args))
			return err
		},
	}
}
