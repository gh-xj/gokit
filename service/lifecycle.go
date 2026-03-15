package service

import (
	"context"

	agentcli "github.com/gh-xj/agentcli-go"
)

// LifecycleService orchestrates preflight/run/postflight execution.
type LifecycleService struct{}

// NewLifecycleService returns a new LifecycleService.
func NewLifecycleService() *LifecycleService {
	return &LifecycleService{}
}

// Run executes the lifecycle: preflight, run, postflight.
// If app is nil a default AppContext is created.
// If hook is nil the preflight/postflight phases are skipped.
func (s *LifecycleService) Run(app *agentcli.AppContext, hook agentcli.Hook, run func(*agentcli.AppContext) error) error {
	if app == nil {
		app = agentcli.NewAppContext(context.TODO())
	}
	if hook != nil {
		if err := hook.Preflight(app); err != nil {
			return err
		}
	}
	if run != nil {
		if err := run(app); err != nil {
			return err
		}
	}
	if hook != nil {
		if err := hook.Postflight(app); err != nil {
			return err
		}
	}
	return nil
}
