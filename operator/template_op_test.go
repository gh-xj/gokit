package operator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gh-xj/agentcli-go/dal"
)

func TestTemplateOperator_RenderTemplate(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewTemplateOperator(fs)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "sub", "out.txt")

	err := op.RenderTemplate(outPath, "Hello {{.Name}}!", TemplateData{Name: "World"})
	if err != nil {
		t.Fatalf("RenderTemplate error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if got := string(data); got != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", got)
	}
}

func TestTemplateOperator_KebabToCamel(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewTemplateOperator(fs)

	tests := []struct {
		in   string
		want string
	}{
		{"foo-bar", "FooBar"},
		{"hello", "Hello"},
		{"a-b-c", "ABC"},
	}

	for _, tc := range tests {
		got := op.KebabToCamel(tc.in)
		if got != tc.want {
			t.Errorf("KebabToCamel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTemplateOperator_ParseModulePath(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewTemplateOperator(fs)

	goMod := "module github.com/example/myproject\n\ngo 1.25\n"
	got := op.ParseModulePath(goMod)
	if got != "github.com/example/myproject" {
		t.Errorf("ParseModulePath = %q, want %q", got, "github.com/example/myproject")
	}
}

func TestTemplateOperator_ParseModulePath_Empty(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewTemplateOperator(fs)

	got := op.ParseModulePath("no module line here")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestTemplateOperator_ResolveParentModule(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewTemplateOperator(fs)

	tmpDir := t.TempDir()
	goModContent := "module github.com/test/parent\n\ngo 1.25\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	childDir := filepath.Join(tmpDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	modPath, modRoot, err := op.ResolveParentModule(childDir)
	if err != nil {
		t.Fatalf("ResolveParentModule error: %v", err)
	}
	if modPath != "github.com/test/parent" {
		t.Errorf("modulePath = %q, want github.com/test/parent", modPath)
	}
	// Resolve symlinks for comparison (macOS /var -> /private/var)
	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	gotRoot, _ := filepath.EvalSymlinks(modRoot)
	if gotRoot != expectedRoot {
		t.Errorf("moduleRoot = %q, want %q", gotRoot, expectedRoot)
	}
}
