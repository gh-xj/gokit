package service

import (
	"context"
	"errors"
	"testing"

	agentcli "github.com/gh-xj/agentcli-go"
)

// testHook records which phases were called.
type testHook struct {
	preflightCalled  bool
	postflightCalled bool
	preflightErr     error
	postflightErr    error
}

func (h *testHook) Preflight(app *agentcli.AppContext) error {
	h.preflightCalled = true
	return h.preflightErr
}

func (h *testHook) Postflight(app *agentcli.AppContext) error {
	h.postflightCalled = true
	return h.postflightErr
}

func TestLifecycleService_HappyPath(t *testing.T) {
	svc := NewLifecycleService()
	hook := &testHook{}
	runCalled := false

	err := svc.Run(agentcli.NewAppContext(context.Background()), hook, func(app *agentcli.AppContext) error {
		runCalled = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hook.preflightCalled {
		t.Error("preflight was not called")
	}
	if !runCalled {
		t.Error("run was not called")
	}
	if !hook.postflightCalled {
		t.Error("postflight was not called")
	}
}

func TestLifecycleService_PreflightError(t *testing.T) {
	svc := NewLifecycleService()
	hook := &testHook{preflightErr: errors.New("preflight failed")}
	runCalled := false

	err := svc.Run(agentcli.NewAppContext(context.Background()), hook, func(app *agentcli.AppContext) error {
		runCalled = true
		return nil
	})

	if err == nil || err.Error() != "preflight failed" {
		t.Fatalf("expected preflight error, got: %v", err)
	}
	if runCalled {
		t.Error("run should not have been called after preflight error")
	}
	if hook.postflightCalled {
		t.Error("postflight should not have been called after preflight error")
	}
}

func TestLifecycleService_NilHook(t *testing.T) {
	svc := NewLifecycleService()
	runCalled := false

	err := svc.Run(agentcli.NewAppContext(context.Background()), nil, func(app *agentcli.AppContext) error {
		runCalled = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runCalled {
		t.Error("run should still execute with nil hook")
	}
}

func TestLifecycleService_NilApp(t *testing.T) {
	svc := NewLifecycleService()
	hook := &testHook{}
	var receivedApp *agentcli.AppContext

	err := svc.Run(nil, hook, func(app *agentcli.AppContext) error {
		receivedApp = app
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedApp == nil {
		t.Error("expected a default AppContext to be created")
	}
	if receivedApp.Context == nil {
		t.Error("expected default context to be non-nil")
	}
}

func TestLifecycleService_PostflightError(t *testing.T) {
	svc := NewLifecycleService()
	hook := &testHook{postflightErr: errors.New("postflight failed")}

	err := svc.Run(agentcli.NewAppContext(context.Background()), hook, func(app *agentcli.AppContext) error {
		return nil
	})

	if err == nil || err.Error() != "postflight failed" {
		t.Fatalf("expected postflight error, got: %v", err)
	}
}

func TestLifecycleService_NilRun(t *testing.T) {
	svc := NewLifecycleService()
	hook := &testHook{}

	err := svc.Run(agentcli.NewAppContext(context.Background()), hook, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hook.preflightCalled {
		t.Error("preflight should be called even with nil run")
	}
	if !hook.postflightCalled {
		t.Error("postflight should be called even with nil run")
	}
}
