package projectresource

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	agentcli "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/dal"
	"github.com/gh-xj/agentops/resource"
)

// ProjectResource implements the Resource interface for scaffolding new Go CLI
// projects. It is a degenerate resource: only Create is meaningful. List returns
// an empty slice and Get returns an error.
type ProjectResource struct {
	fs   dal.FileSystem
	exec dal.Executor
}

// New creates a ProjectResource with the given filesystem and executor.
func New(fs dal.FileSystem, exec dal.Executor) *ProjectResource {
	return &ProjectResource{fs: fs, exec: exec}
}

// Schema returns the resource schema for projects.
func (p *ProjectResource) Schema() resource.ResourceSchema {
	return resource.ResourceSchema{
		Kind:        "project",
		Description: "Scaffold a new Go CLI project",
		Fields: []resource.FieldDef{
			{Name: "path", Type: "string", Required: true},
		},
		CreateArgs: []resource.ArgDef{
			{Name: "module", Description: "Go module path (defaults to slug)", Required: false},
			{Name: "mode", Description: "Scaffold mode: minimal|lean|full (default lean)", Required: false},
			{Name: "base_dir", Description: "Parent directory for the project (default .)", Required: false},
		},
	}
}

// templateData holds values passed to scaffold templates.
type templateData struct {
	Module string
	Name   string
}

// Create scaffolds a new Go CLI project directory.
// slug is the project name. opts can contain "module", "mode", and "base_dir".
func (p *ProjectResource) Create(ctx *agentcli.AppContext, slug string, opts map[string]string) (*resource.Record, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("project name (slug) is required")
	}

	baseDir := "."
	if v, ok := opts["base_dir"]; ok && strings.TrimSpace(v) != "" {
		baseDir = v
	}

	module := strings.TrimSpace(slug)
	if v, ok := opts["module"]; ok && strings.TrimSpace(v) != "" {
		module = strings.TrimSpace(v)
	}

	root := filepath.Join(baseDir, slug)
	if err := p.ensureEmptyDir(root); err != nil {
		return nil, err
	}

	cliName := filepath.Base(slug)

	files := map[string]string{
		"main.go":     projectMainTpl,
		"go.mod":      projectGoModTpl,
		"README.md":   projectReadmeTpl,
		"cmd/root.go": projectRootCmdTpl,
	}

	d := templateData{
		Module: module,
		Name:   cliName,
	}

	for relPath, body := range files {
		fullPath := filepath.Join(root, relPath)
		if err := p.renderTemplate(fullPath, body, d); err != nil {
			return nil, fmt.Errorf("render %s: %w", relPath, err)
		}
	}

	return &resource.Record{
		Kind: "project",
		ID:   slug,
		Fields: map[string]any{
			"path": root,
		},
		RawPath: root,
	}, nil
}

// List returns an empty slice. Projects are not tracked after creation.
func (p *ProjectResource) List(_ *agentcli.AppContext, _ resource.Filter) ([]resource.Record, error) {
	return []resource.Record{}, nil
}

// Get returns an error. Individual project lookup is not supported.
func (p *ProjectResource) Get(_ *agentcli.AppContext, id string) (*resource.Record, error) {
	return nil, fmt.Errorf("project resource does not support Get (requested %q)", id)
}

// ensureEmptyDir checks that root either doesn't exist (and creates it) or is
// an empty directory.
func (p *ProjectResource) ensureEmptyDir(root string) error {
	if p.fs.Exists(root) {
		entries, err := p.fs.ReadDir(root)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("target directory is not empty: %s", root)
		}
		return nil
	}
	return p.fs.EnsureDir(root)
}

// renderTemplate parses a Go text/template body and writes the result to path,
// ensuring the parent directory exists.
func (p *ProjectResource) renderTemplate(path, body string, data templateData) error {
	if err := p.fs.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	tpl, err := template.New(filepath.Base(path)).Parse(body)
	if err != nil {
		return err
	}
	var buf strings.Builder
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}
	return p.fs.WriteFile(path, []byte(buf.String()), 0o644)
}
