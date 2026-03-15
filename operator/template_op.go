package operator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gh-xj/agentcli-go/dal"
)

// TemplateOperatorImpl implements TemplateOperator.
type TemplateOperatorImpl struct {
	fs dal.FileSystem
}

// NewTemplateOperator returns a new TemplateOperatorImpl.
func NewTemplateOperator(fs dal.FileSystem) *TemplateOperatorImpl {
	return &TemplateOperatorImpl{fs: fs}
}

// RenderTemplate ensures the parent directory exists, then parses and executes
// the Go template body into the file at path.
func (t *TemplateOperatorImpl) RenderTemplate(path, body string, data TemplateData) error {
	if err := t.fs.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	tpl, err := template.New(filepath.Base(path)).Parse(body)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tpl.Execute(f, data)
}

// KebabToCamel converts kebab-case to PascalCase (e.g., "foo-bar" -> "FooBar").
func (t *TemplateOperatorImpl) KebabToCamel(in string) string {
	parts := strings.Split(in, "-")
	for i := range parts {
		if len(parts[i]) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

// DetectLocalReplaceLine walks up from the current working directory looking
// for a go.mod containing the agentcli-go module, returning a replace directive
// line if found.
func (t *TemplateOperatorImpl) DetectLocalReplaceLine() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for i := 0; i < 16; i++ {
		modFile := filepath.Join(dir, "go.mod")
		data, readErr := t.fs.ReadFile(modFile)
		if readErr == nil && strings.Contains(string(data), "module github.com/gh-xj/agentcli-go") {
			return fmt.Sprintf("replace github.com/gh-xj/agentcli-go => %s", dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// ParseModulePath extracts the module path from go.mod content.
func (t *TemplateOperatorImpl) ParseModulePath(goMod string) string {
	lines := strings.Split(goMod, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "module ") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
		value = strings.Trim(value, `"`)
		return value
	}
	return ""
}

// ResolveParentModule walks up from targetRoot looking for a go.mod, returning
// the module path and the directory containing go.mod.
func (t *TemplateOperatorImpl) ResolveParentModule(targetRoot string) (modulePath, moduleRoot string, err error) {
	dir := filepath.Clean(targetRoot)
	for {
		modFile := filepath.Join(dir, "go.mod")
		raw, readErr := t.fs.ReadFile(modFile)
		if readErr == nil {
			module := t.ParseModulePath(string(raw))
			if module == "" {
				return "", "", fmt.Errorf("invalid go.mod in parent module root: %s", modFile)
			}
			return module, dir, nil
		}
		if !errors.Is(readErr, os.ErrNotExist) {
			return "", "", fmt.Errorf("read parent go.mod: %w", readErr)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", fmt.Errorf("parent go.mod not found for --in-existing-module target: %s", targetRoot)
}
