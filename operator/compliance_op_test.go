package operator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gh-xj/agentcli-go/dal"
)

func TestComplianceOperator_CheckFileExists_Present(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewComplianceOperator(fs)

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	finding := op.CheckFileExists(tmpDir, "exists.txt")
	if finding != nil {
		t.Errorf("expected nil finding for existing file, got %+v", finding)
	}
}

func TestComplianceOperator_CheckFileExists_Missing(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewComplianceOperator(fs)

	tmpDir := t.TempDir()

	finding := op.CheckFileExists(tmpDir, "missing.txt")
	if finding == nil {
		t.Fatal("expected finding for missing file")
	}
	if finding.Code != "missing_file" {
		t.Errorf("expected code missing_file, got %q", finding.Code)
	}
	if finding.Path != "missing.txt" {
		t.Errorf("expected path missing.txt, got %q", finding.Path)
	}
}

func TestComplianceOperator_CheckFileContains_Present(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewComplianceOperator(fs)

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "root.go"), []byte(`import "cobrax"`), 0644); err != nil {
		t.Fatal(err)
	}

	finding := op.CheckFileContains(tmpDir, "root.go", "missing_contract", "cobrax", "cobrax import missing")
	if finding != nil {
		t.Errorf("expected nil finding, got %+v", finding)
	}
}

func TestComplianceOperator_CheckFileContains_Missing(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewComplianceOperator(fs)

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "root.go"), []byte(`import "fmt"`), 0644); err != nil {
		t.Fatal(err)
	}

	finding := op.CheckFileContains(tmpDir, "root.go", "missing_contract", "cobrax", "cobrax import missing")
	if finding == nil {
		t.Fatal("expected finding for missing content")
	}
	if finding.Code != "missing_contract" {
		t.Errorf("expected code missing_contract, got %q", finding.Code)
	}
	if finding.Message != "cobrax import missing" {
		t.Errorf("expected message 'cobrax import missing', got %q", finding.Message)
	}
}

func TestComplianceOperator_ValidateCommandName_Valid(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewComplianceOperator(fs)

	for _, name := range []string{"foo", "foo-bar", "a1-b2"} {
		if err := op.ValidateCommandName(name); err != nil {
			t.Errorf("ValidateCommandName(%q) unexpected error: %v", name, err)
		}
	}
}

func TestComplianceOperator_ValidateCommandName_Invalid(t *testing.T) {
	fs := dal.NewFileSystem()
	op := NewComplianceOperator(fs)

	for _, name := range []string{"FooBar", "", "1abc", "-start"} {
		if err := op.ValidateCommandName(name); err == nil {
			t.Errorf("ValidateCommandName(%q) expected error", name)
		}
	}
}
