package agentcli

// Hook allows commands to run standardized pre/post execution steps.
type Hook interface {
	Preflight(*AppContext) error
	Postflight(*AppContext) error
}
