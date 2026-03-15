package operator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	agentcli "github.com/gh-xj/agentcli-go"
	"github.com/gh-xj/agentcli-go/dal"
)

var validCommandNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ComplianceOperatorImpl implements ComplianceOperator.
type ComplianceOperatorImpl struct {
	fs dal.FileSystem
}

// NewComplianceOperator returns a new ComplianceOperatorImpl.
func NewComplianceOperator(fs dal.FileSystem) *ComplianceOperatorImpl {
	return &ComplianceOperatorImpl{fs: fs}
}

// CheckFileExists returns nil if the file exists, or a DoctorFinding if missing.
func (c *ComplianceOperatorImpl) CheckFileExists(rootDir, relPath string) *agentcli.DoctorFinding {
	abs := filepath.Join(rootDir, relPath)
	if c.fs.Exists(abs) {
		return nil
	}
	return &agentcli.DoctorFinding{
		Code:    "missing_file",
		Path:    relPath,
		Message: "required file is missing",
	}
}

// CheckFileContains returns nil if the file contains want, or a DoctorFinding otherwise.
// If the file cannot be read, nil is returned (the file-exists check handles that case).
func (c *ComplianceOperatorImpl) CheckFileContains(rootDir, relPath, code, want, msg string) *agentcli.DoctorFinding {
	abs := filepath.Join(rootDir, relPath)
	content, err := c.fs.ReadFile(abs)
	if err != nil {
		return nil
	}
	if !strings.Contains(string(content), want) {
		return &agentcli.DoctorFinding{
			Code:    code,
			Path:    relPath,
			Message: msg,
		}
	}
	return nil
}

// ValidateCommandName checks that a command name matches kebab-case [a-z][a-z0-9-]*.
func (c *ComplianceOperatorImpl) ValidateCommandName(name string) error {
	if !validCommandNameRe.MatchString(name) {
		return fmt.Errorf("invalid command name %q: use kebab-case [a-z0-9-]", name)
	}
	return nil
}
