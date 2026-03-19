package service

import (
	"strings"
	"testing"
)

func TestAppLifecycleTemplateMatchesHookInterface(t *testing.T) {
	// The Hook interface requires Preflight(*AppContext) error and
	// Postflight(*AppContext) error. Verify the template generates
	// methods with the correct signature.
	if !strings.Contains(appLifecycleTpl, "Preflight(_ *agentcli.AppContext) error") {
		t.Error("appLifecycleTpl Preflight signature does not match Hook interface: expected (*agentcli.AppContext) error")
	}
	if !strings.Contains(appLifecycleTpl, "Postflight(_ *agentcli.AppContext) error") {
		t.Error("appLifecycleTpl Postflight signature does not match Hook interface: expected (*agentcli.AppContext) error")
	}
}
