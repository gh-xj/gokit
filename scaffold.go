package gokit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
)

const rootCommandMarker = "// gokit:add-command"

var validCommandName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

type templateData struct {
	Module string
	Name   string
}

// DoctorFinding describes a single compliance issue in a scaffolded project.
type DoctorFinding struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// DoctorReport summarizes scaffold compliance checks.
type DoctorReport struct {
	OK       bool            `json:"ok"`
	Findings []DoctorFinding `json:"findings"`
}

func (r DoctorReport) JSON() (string, error) {
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ScaffoldNew creates a new CLI project using the golden gokit layout.
func ScaffoldNew(baseDir, name, module string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("project name is required")
	}
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "."
	}
	if strings.TrimSpace(module) == "" {
		module = name
	}

	root := filepath.Join(baseDir, name)
	if err := ensureEmptyDir(root); err != nil {
		return "", err
	}

	d := templateData{Module: module, Name: name}
	for path, body := range map[string]string{
		"go.mod":                    goModTpl,
		"main.go":                   mainTpl,
		"cmd/root.go":               rootCmdTpl,
		"internal/app/bootstrap.go": appBootstrapTpl,
		"internal/app/lifecycle.go": appLifecycleTpl,
		"internal/app/errors.go":    appErrorsTpl,
		"internal/config/schema.go": configSchemaTpl,
		"internal/config/load.go":   configLoadTpl,
		"internal/io/output.go":     outputTpl,
		"pkg/version/version.go":    versionTpl,
		"test/e2e/cli_test.go":      e2eTestTpl,
		"Taskfile.yml":              taskfileTpl,
		"README.md":                 readmeTpl,
	} {
		if err := writeTemplate(filepath.Join(root, path), body, d); err != nil {
			return "", err
		}
	}
	return root, nil
}

// ScaffoldAddCommand creates a command file and wires it into cmd/root.go.
func ScaffoldAddCommand(rootDir, commandName string) error {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = "."
	}
	if !validCommandName.MatchString(commandName) {
		return fmt.Errorf("invalid command name %q: use kebab-case [a-z0-9-]", commandName)
	}

	cmdFile := filepath.Join(rootDir, "cmd", commandName+".go")
	if FileExists(cmdFile) {
		return fmt.Errorf("command file already exists: %s", cmdFile)
	}
	funcName := kebabToCamel(commandName)
	if err := writeTemplate(cmdFile, addCommandTpl, templateData{
		Name:   commandName,
		Module: funcName,
	}); err != nil {
		return err
	}

	rootFile := filepath.Join(rootDir, "cmd", "root.go")
	content, err := os.ReadFile(rootFile)
	if err != nil {
		return err
	}
	registerLine := fmt.Sprintf("registerCommand(%q, %sCommand())", commandName, funcName)
	text := string(content)
	if strings.Contains(text, registerLine) {
		return nil
	}
	idx := strings.Index(text, rootCommandMarker)
	if idx < 0 {
		return fmt.Errorf("marker %q not found in %s", rootCommandMarker, rootFile)
	}
	updated := text[:idx] + registerLine + "\n\t" + text[idx:]
	return os.WriteFile(rootFile, []byte(updated), 0644)
}

// Doctor checks whether a project follows the golden scaffold contract.
func Doctor(rootDir string) DoctorReport {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = "."
	}

	required := []string{
		"main.go",
		"go.mod",
		"cmd/root.go",
		"internal/app/bootstrap.go",
		"internal/app/lifecycle.go",
		"internal/app/errors.go",
		"internal/config/schema.go",
		"internal/config/load.go",
		"internal/io/output.go",
		"pkg/version/version.go",
		"test/e2e/cli_test.go",
		"Taskfile.yml",
	}

	report := DoctorReport{
		OK:       true,
		Findings: make([]DoctorFinding, 0),
	}

	for _, p := range required {
		abs := filepath.Join(rootDir, p)
		if !FileExists(abs) {
			report.OK = false
			report.Findings = append(report.Findings, DoctorFinding{
				Code:    "missing_file",
				Path:    p,
				Message: "required file is missing",
			})
		}
	}

	checkContains := func(relPath, code, want, msg string) {
		content, err := os.ReadFile(filepath.Join(rootDir, relPath))
		if err != nil {
			return
		}
		if !strings.Contains(string(content), want) {
			report.OK = false
			report.Findings = append(report.Findings, DoctorFinding{
				Code:    code,
				Path:    relPath,
				Message: msg,
			})
		}
	}

	checkContains("cmd/root.go", "missing_contract", "--verbose", "required root flag --verbose missing")
	checkContains("cmd/root.go", "missing_contract", "--config", "required root flag --config missing")
	checkContains("cmd/root.go", "missing_contract", "--json", "required root flag --json missing")
	checkContains("cmd/root.go", "missing_contract", "--no-color", "required root flag --no-color missing")
	checkContains("cmd/root.go", "missing_contract", rootCommandMarker, "missing scaffold command marker")
	checkContains("Taskfile.yml", "missing_contract", "ci:", "canonical CI task missing")
	checkContains("Taskfile.yml", "missing_contract", "verify:", "local verification task missing")
	checkContains("internal/app/lifecycle.go", "missing_contract", "Preflight", "lifecycle preflight hook missing")
	checkContains("internal/app/lifecycle.go", "missing_contract", "Postflight", "lifecycle postflight hook missing")

	slices.SortFunc(report.Findings, func(a, b DoctorFinding) int {
		if c := strings.Compare(a.Path, b.Path); c != 0 {
			return c
		}
		return strings.Compare(a.Code, b.Code)
	})
	return report
}

func ensureEmptyDir(root string) error {
	if FileExists(root) {
		entries, err := os.ReadDir(root)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("target directory is not empty: %s", root)
		}
		return nil
	}
	return os.MkdirAll(root, 0755)
}

func writeTemplate(path, body string, data templateData) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
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

func kebabToCamel(in string) string {
	parts := strings.Split(in, "-")
	for i := range parts {
		if len(parts[i]) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

const goModTpl = `module {{.Module}}

go 1.25.5
`

const mainTpl = `package main

import (
	"os"

	"{{.Module}}/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
`

const rootCmdTpl = `package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gh-xj/gokit"
)

type command struct {
	Description string
	Run         func(*gokit.AppContext, []string) error
}

var commandRegistry = map[string]command{}

func init() {
	registerBuiltins()
	// gokit:add-command
}

func registerCommand(name string, cmd command) {
	commandRegistry[name] = cmd
}

func registerBuiltins() {
	registerCommand("version", command{
		Description: "print build metadata",
		Run: func(app *gokit.AppContext, _ []string) error {
			data := map[string]string{
				"name": app.Meta.Name,
				"version": app.Meta.Version,
				"commit": app.Meta.Commit,
				"date": app.Meta.Date,
			}
			if jsonOutput, _ := app.Values["json"].(bool); jsonOutput {
				enc := json.NewEncoder(app.IO.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}
			_, err := fmt.Fprintf(app.IO.Stdout, "%s %s (%s %s)\n", data["name"], data["version"], data["commit"], data["date"])
			return err
		},
	})
}

func Execute(args []string) int {
	gokit.InitLogger()

	global, rest, err := parseGlobal(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		printUsage()
		return gokit.ExitUsage
	}
	if global.help || len(rest) == 0 {
		printUsage()
		return gokit.ExitUsage
	}
	cmdName := rest[0]
	cmd, ok := commandRegistry[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmdName)
		printUsage()
		return gokit.ExitUsage
	}

	app := gokit.NewAppContext(context.Background())
	app.Meta = gokit.AppMeta{Name: "{{.Name}}", Version: "dev", Commit: "none", Date: "unknown"}
	app.Values["json"] = global.json
	app.Values["config"] = global.config
	app.Values["no-color"] = global.noColor

	runErr := gokit.RunLifecycle(app, nil, func(app *gokit.AppContext) error {
		return cmd.Run(app, rest[1:])
	})
	if runErr != nil {
		fmt.Fprintln(os.Stderr, runErr.Error())
	}
	return gokit.ResolveExitCode(runErr)
}

type globalFlags struct {
	verbose bool
	config  string
	json    bool
	noColor bool
	help    bool
}

func parseGlobal(args []string) (globalFlags, []string, error) {
	out := globalFlags{}
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--verbose":
			out.verbose = true
		case "--json":
			out.json = true
		case "--no-color":
			out.noColor = true
		case "-h", "--help":
			out.help = true
		case "--config":
			if i+1 >= len(args) {
				return out, nil, fmt.Errorf("--config requires a value")
			}
			out.config = args[i+1]
			i++
		default:
			rest = append(rest, args[i:]...)
			return out, rest, nil
		}
	}
	return out, rest, nil
}

func printUsage() {
	names := make([]string, 0, len(commandRegistry))
	for name := range commandRegistry {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  {{.Name}} [--verbose] [--config path] [--json] [--no-color] <command> [args]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	for _, name := range names {
		desc := strings.TrimSpace(commandRegistry[name].Description)
		fmt.Fprintf(os.Stderr, "  %-16s %s\n", name, desc)
	}
}
`

const addCommandTpl = `package cmd

import (
	"fmt"

	"github.com/gh-xj/gokit"
)

func init() {
}

func {{.Module}}Command() command {
	return command{
		Description: "describe {{.Name}}",
		Run: func(app *gokit.AppContext, args []string) error {
			if jsonOutput, _ := app.Values["json"].(bool); jsonOutput {
				_, err := fmt.Fprintln(app.IO.Stdout, "{\"command\":\"{{.Name}}\",\"ok\":true}")
				return err
			}
			_, err := fmt.Fprintf(app.IO.Stdout, "{{.Name}} executed with %d args\n", len(args))
			return err
		},
	}
}
`

const appBootstrapTpl = `package app

import "github.com/gh-xj/gokit"

func Bootstrap() *gokit.AppContext {
	return gokit.NewAppContext(nil)
}
`

const appLifecycleTpl = `package app

import "github.com/gh-xj/gokit"

type Hooks struct{}

func (Hooks) Preflight(*gokit.AppContext) error {
	return nil
}

func (Hooks) Postflight(*gokit.AppContext) error {
	return nil
}
`

const appErrorsTpl = `package app

import "github.com/gh-xj/gokit"

func UsageError(message string) error {
	return gokit.NewCLIError(gokit.ExitUsage, "usage", message, nil)
}
`

const configSchemaTpl = `package config

type Config struct {
	Environment string ` + "`json:\"environment\"`" + `
}
`

const configLoadTpl = `package config

func Load() Config {
	return Config{Environment: "default"}
}
`

const outputTpl = `package appio

import (
	"encoding/json"
	"io"
)

func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
`

const versionTpl = `package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
`

const e2eTestTpl = `package e2e

import "testing"

func TestPlaceholder(t *testing.T) {
	t.Skip("add end-to-end command tests")
}
`

const taskfileTpl = `version: "3"

tasks:
  fmt:
    desc: Format Go files
    cmds:
      - gofmt -w $(find . -name '*.go' -type f)

  fmt:check:
    desc: Fail when gofmt would change files
    cmds:
      - test -z "$(gofmt -l $(find . -name '*.go' -type f))"

  lint:
    desc: Run static checks
    deps: [fmt:check]
    cmds:
      - go vet ./...

  test:
    desc: Run tests
    deps: [fmt:check]
    cmds:
      - go test ./...

  build:
    desc: Build binary
    deps: [fmt:check]
    cmds:
      - mkdir -p bin
      - go build -o bin/{{.Name}} .

  smoke:
    desc: Deterministic smoke checks
    deps: [build]
    cmds:
      - ./bin/{{.Name}} version --json >/tmp/{{.Name}}-smoke.json

  ci:
    desc: Canonical CI contract command
    deps: [lint, test, build, smoke]
    cmds:
      - echo "ci checks passed"

  verify:
    desc: Local aggregate verification
    deps: [ci]
    cmds:
      - echo "verify checks passed"
`

const readmeTpl = `# {{.Name}}

Generated by gokit scaffold.

## Commands

- ` + "`version`" + `: print build metadata

## Verification

- ` + "`task ci`" + `: canonical CI command
- ` + "`task verify`" + `: local aggregate verification
`
