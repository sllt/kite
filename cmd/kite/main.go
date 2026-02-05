package main

import (
	"github.com/sllt/kite/pkg/kite"
	"github.com/sllt/kite/pkg/kite/cli/bootstrap"
	"github.com/sllt/kite/pkg/kite/cli/create"
	"github.com/sllt/kite/pkg/kite/cli/migration"
	"github.com/sllt/kite/pkg/kite/cli/wrap"
)

func main() {
	cli := kite.NewCMD()

	cli.SubCommand("init", bootstrap.Create,
		kite.AddDescription("Initialize a new Kite project"),
		kite.AddHelp("Usage: kite init <project-name>"),
	)

	cli.SubCommand("version",
		func(*kite.Context) (any, error) {
			return CLIVersion, nil
		},
		kite.AddDescription("Print the CLI version"),
	)

	cli.SubCommand("migrate create", migration.Migrate,
		kite.AddDescription("Create a new migration file"),
		kite.AddHelp("Usage: kite migrate create <migration-name>"),
	)

	cli.SubCommand("wrap grpc server", wrap.BuildGRPCKiteServer,
		kite.AddDescription("Generate Kite-integrated gRPC server from proto file"),
		kite.AddHelp("Usage: kite wrap grpc server -proto=<path-to-proto-file>"),
	)

	cli.SubCommand("wrap grpc client", wrap.BuildGRPCKiteClient,
		kite.AddDescription("Generate Kite-integrated gRPC client from proto file"),
		kite.AddHelp("Usage: kite wrap grpc client -proto=<path-to-proto-file>"),
	)

	// Create commands for generating layered architecture components
	cli.SubCommand("create handler", create.Handler,
		kite.AddDescription("Create a new handler"),
		kite.AddHelp("Usage: kite create handler <name>"),
	)

	cli.SubCommand("create service", create.Service,
		kite.AddDescription("Create a new service"),
		kite.AddHelp("Usage: kite create service <name>"),
	)

	cli.SubCommand("create repository", create.Repository,
		kite.AddDescription("Create a new repository"),
		kite.AddHelp("Usage: kite create repository <name>"),
	)

	cli.SubCommand("create model", create.Model,
		kite.AddDescription("Create a new model"),
		kite.AddHelp("Usage: kite create model <name>"),
	)

	cli.SubCommand("create all", create.All,
		kite.AddDescription("Create handler, service, repository, and model"),
		kite.AddHelp("Usage: kite create all <name>"),
	)

	cli.Run()
}
