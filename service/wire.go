//go:build wireinject

package service

import (
	"github.com/google/wire"
	"github.com/gh-xj/agentcli-go/dal"
	"github.com/gh-xj/agentcli-go/operator"
)

// ProviderSet is the Wire provider set for the service layer.
var ProviderSet = wire.NewSet(
	dal.NewFileSystem,
	wire.Bind(new(dal.FileSystem), new(*dal.FileSystemImpl)),
	dal.NewExecutor,
	wire.Bind(new(dal.Executor), new(*dal.ExecutorImpl)),
	dal.NewLogger,
	wire.Bind(new(dal.Logger), new(*dal.LoggerImpl)),

	operator.NewTemplateOperator,
	wire.Bind(new(operator.TemplateOperator), new(*operator.TemplateOperatorImpl)),
	operator.NewComplianceOperator,
	wire.Bind(new(operator.ComplianceOperator), new(*operator.ComplianceOperatorImpl)),
	operator.NewArgsOperator,
	wire.Bind(new(operator.ArgsOperator), new(*operator.ArgsOperatorImpl)),

	NewScaffoldService,
	NewDoctorService,
	NewLifecycleService,
	NewContainer,
)

// InitializeContainer is the Wire injector that constructs a fully wired Container.
func InitializeContainer() *Container {
	wire.Build(ProviderSet)
	return nil
}
