---
name: agentcli-go
description: "agentcli-go framework reference for building Go CLI tools. Use when working on agentcli-go itself, scaffolding new CLI projects, adding commands, integrating the library, or debugging framework behavior. Triggers on: agentcli-go, scaffold new CLI, add command, cobrax, configx, AppContext, RunLifecycle, agentcli."
version: "2.0"
last_updated: 2026-03-15
---

# agentcli-go

Shared Go CLI helpers and framework modules.

**Module:** `github.com/gh-xj/agentcli-go`
**Repo:** `github.com/gh-xj/agentcli-go` | **Versioning:** `v0.x.y` (pre-1.0)

---

## Architecture: DAG Layers

```
cmd/ (handler) → service → operator → dal
                    ↓          ↓        ↓
               root (model: AppContext, Hook, CLIError)
```

### Root Package (Model Layer)
| File | Exported Symbols |
|------|-----------------|
| `core_context.go` | `AppContext`, `IOStreams`, `AppMeta`, `NewAppContext(ctx)` |
| `lifecycle.go` | `Hook` interface (`Preflight`, `Postflight`) |
| `errors.go` | `CLIError`, `ExitCoder`, `ResolveExitCode(err)`, `NewCLIError()`, exit code constants |
| `scaffold.go` | `DoctorFinding`, `DoctorReport`, `ScaffoldNewOptions` (model types only) |

### DAL Layer (`dal/`)
| File | Exported Symbols |
|------|-----------------|
| `dal/interfaces.go` | `FileSystem`, `Executor`, `Logger` interfaces, `DirEntry` |
| `dal/filesystem.go` | `FileSystemImpl`, `NewFileSystem()` — Exists, EnsureDir, ReadFile, WriteFile, ReadDir, BaseName |
| `dal/exec.go` | `ExecutorImpl`, `NewExecutor()` — Run, RunInDir, RunOsascript, Which |
| `dal/logger.go` | `LoggerImpl`, `NewLogger()` — Init(verbose, writer) |

### Operator Layer (`operator/`)
| File | Exported Symbols |
|------|-----------------|
| `operator/interfaces.go` | `TemplateOperator`, `ComplianceOperator`, `ArgsOperator`, `TemplateData` |
| `operator/args_op.go` | `ArgsOperatorImpl`, `NewArgsOperator()` — Parse, Require, Get, HasFlag |
| `operator/template_op.go` | `TemplateOperatorImpl`, `NewTemplateOperator(fs)` — RenderTemplate, KebabToCamel, etc. |
| `operator/compliance_op.go` | `ComplianceOperatorImpl`, `NewComplianceOperator(fs)` — CheckFileExists, CheckFileContains, ValidateCommandName |

### Service Layer (`service/`)
| File | Exported Symbols |
|------|-----------------|
| `service/container.go` | `Container`, `NewContainer()`, `Get()`, `Reset()` |
| `service/wire.go` | `ProviderSet`, `InitializeContainer()` (Wire DI) |
| `service/scaffold.go` | `ScaffoldService`, `NewScaffoldService()` — New, AddCommand |
| `service/doctor.go` | `DoctorService`, `NewDoctorService()` — Run (with DAG compliance checks) |
| `service/lifecycle.go` | `LifecycleService`, `NewLifecycleService()` — Run |
| `service/templates.go` | All scaffold template constants |

### Adapters
| File | Exported Symbols |
|------|-----------------|
| `cobrax/cobrax.go` | `Execute(RootSpec, args) int`, `NewRoot(RootSpec) *cobra.Command`, `CommandSpec`, `RootSpec` |
| `configx/configx.go` | `Load(Options) map[string]any`, `Decode[T](raw)`, `NormalizeEnv(prefix, environ)` |

---

## Scaffold Workflows

### New project
```bash
agentcli new --dir . --module github.com/me/my-tool my-tool
# or programmatically:
service.Get().ScaffoldSvc.New(".", "my-tool", "github.com/me/my-tool", service.ScaffoldNewOptions{})
```
Generates: `main.go`, `cmd/root.go`, `service/`, `dal/`, `operator/`, `internal/app/`, `internal/config/`, `internal/io/`, `internal/tools/smokecheck/`, `pkg/version/`, `test/`, `Taskfile.yml`, `README.md`

### Add command
```bash
agentcli add command --name sync --preset file-sync
agentcli add command --name deploy --desc "run deploy checks"
```
Presets: `file-sync`, `http-client`, `deploy-helper`, `task-replay-orchestrator`

### Doctor check
```bash
agentcli doctor [--dir ./my-tool] [--json]
# checks scaffold compliance + DAG layer presence
```

---

## Golden Project Layout

```
my-tool/
├── main.go                          # os.Exit(cmd.Execute(os.Args[1:]))
├── cmd/
│   ├── root.go                      # cobrax.Execute(RootSpec{...}) → calls service.Get()
│   └── <command>.go                 # func <Name>Command() command
├── service/
│   ├── container.go                 # Wire DI container + Get() singleton
│   └── wire.go                      # Wire provider set
├── operator/
│   ├── interfaces.go                # Business logic contracts
│   └── example_op.go               # Stub operator
├── dal/
│   ├── interfaces.go                # Data access contracts
│   └── filesystem.go               # Default filesystem impl
├── internal/
│   ├── app/{bootstrap,lifecycle,errors}.go
│   ├── config/{schema,load}.go
│   ├── io/output.go
│   └── tools/smokecheck/main.go
├── pkg/version/version.go
├── test/
│   ├── e2e/cli_test.go
│   └── smoke/version.schema.json
└── Taskfile.yml                     # includes wire task
```

**Data flow:** `cmd/ → service.Get().XxxSvc.Method() → operator → dal`

---

## cobrax Pattern

```go
// cmd/root.go
return cobrax.Execute(cobrax.RootSpec{
    Use:   "my-tool",
    Short: "my-tool CLI",
    Meta:  agentcli.AppMeta{Name: "my-tool", Version: version.Version},
    Commands: []cobrax.CommandSpec{
        {Use: "sync", Short: "sync files", Run: SyncCommand().Run},
    },
}, args)
```

Persistent flags auto-wired: `--verbose/-v`, `--config`, `--json`, `--no-color`
Values accessible via `app.Values["json"]`, `app.Values["config"]`, etc.

---

## configx Pattern

```go
raw, err := configx.Load(configx.Options{
    Defaults: map[string]any{"env": "default"},
    FilePath: configPath,
    Env:      configx.NormalizeEnv("MYTOOL_", os.Environ()),
    Flags:    map[string]string{"env": flagVal},
})
cfg, err := configx.Decode[config.Config](raw)
// Precedence: Defaults < File < Env < Flags
```

---

## Wire DI Pattern

```go
// service/wire.go
var ProviderSet = wire.NewSet(
    dal.NewFileSystem,
    wire.Bind(new(dal.FileSystem), new(*dal.FileSystemImpl)),
    operator.NewMyOperator,
    wire.Bind(new(operator.MyOperator), new(*operator.MyOperatorImpl)),
    NewMyService,
    NewContainer,
)
```

Generate: `wire ./service/` or `task wire`

---

## Dependency Direction Rules

- `dal` imports root only — no knowledge of operator/service
- `operator` imports root + dal — transforms data, applies rules
- `service` imports root + operator + dal — orchestrates multi-step flows
- `cmd` imports service + root — CLI wiring only
- Operators return errors (never `log.Fatal`)

---

## Rules

- **Root = model only** — shared contracts (types, interfaces, error codes)
- **New code → DAG layer** — dal for I/O, operator for logic, service for orchestration
- **Exported only** — all types/functions PascalCase
- **Minimal deps** — zerolog, cobra, wire; justify new additions
- **No business logic in root** — generic model types only

## Out of Scope

- Project-specific logic (put that in consuming projects)
- Adding functions used by only one project
