package operator

import (
	agentcli "github.com/gh-xj/agentcli-go"
)

// TemplateOperator handles template rendering and project file generation.
type TemplateOperator interface {
	RenderTemplate(path, body string, data TemplateData) error
	KebabToCamel(in string) string
	DetectLocalReplaceLine() string
	ParseModulePath(goMod string) string
	ResolveParentModule(targetRoot string) (modulePath, moduleRoot string, err error)
}

// TemplateData is the data passed to scaffold templates.
type TemplateData struct {
	Module           string
	Name             string
	Description      string
	Preset           string
	GokitReplaceLine string
}

// ComplianceOperator checks project structure compliance.
type ComplianceOperator interface {
	CheckFileExists(rootDir, relPath string) *agentcli.DoctorFinding
	CheckFileContains(rootDir, relPath, code, want, msg string) *agentcli.DoctorFinding
	ValidateCommandName(name string) error
}

// ArgsOperator handles CLI argument parsing.
type ArgsOperator interface {
	Parse(args []string) map[string]string
	Require(args map[string]string, key, usage string) (string, error)
	Get(args map[string]string, key, defaultVal string) string
	HasFlag(args map[string]string, key string) bool
}
