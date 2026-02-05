# GoFR Command Line Interface

Managing repetitive tasks and maintaining consistency across large-scale applications is challenging!

**Kite CLI provides the following:**

* All-in-one command-line tool designed specifically for Kite applications
* Simplifies **database migrations** management
* Abstracts **tracing**, **metrics** and structured **logging** for Kite's gRPC server/client
* Enforces standard **Kite conventions** in new projects

## Prerequisites

- Go 1.22 or above. To check Go version use the following command:
```bash
  go version
```

## **Installation**
To get started with Kite CLI, use the below commands

```bash
  go install github.com/sllt/kite/cli/kite@latest
```

To check the installation:
```bash
  kite version
```
---

## Usage

The CLI can be run directly from the terminal after installation. Here’s the general syntax:

```bash
  kite <subcommand> [flags]=[arguments]
```
---

## **Commands**

## 1. ***`init`***

   The init command initializes a new Kite project. It sets up the foundational structure for the project and generates a basic "Hello World!" program as a starting point. This allows developers to quickly dive into building their application with a ready-made structure.

### Command Usage
```bash
  kite init
```
---

## 2. ***`migrate create`***

   The migrate create command generates a migration template file with pre-defined structure in your migrations directory.
   This boilerplate code helps you maintain consistent patterns when writing database schema modifications across your project.


### Command Usage
```bash
  kite migrate create -name=<migration-name>
```

### Example Usage

```bash
kite migrate create -name=create_employee_table
```
This command generates a migration directory which has the below files:

1. A new migration file with timestamp prefix (e.g., `20250127152047_create_employee_table.go`) containing:
```go
package migrations

import (
    "github.com/sllt/kite/pkg/kite/migration"
)

func create_employee_table() migration.Migrate {
    return migration.Migrate{
        UP: func(d migration.Datasource) error {
            // write your migrations here
            return nil
        },
    }
}
```
2. An auto-generated all.go file that maintains a registry of all migrations:
```go
// This is auto-generated file using 'kite migrate' tool. DO NOT EDIT.
package migrations

import (
    "github.com/sllt/kite/pkg/kite/migration"
)

func All() map[int64]migration.Migrate {
    return map[int64]migration.Migrate {
        20250127152047: create_employee_table(),
    }
}
```

> **💡 Best Practice:** Learn about [organizing migrations by feature](../../docs/advanced-guide/handling-data-migrations#organizing-migrations-by-feature) to avoid creating one migration per table or operation.

For detailed instructions on handling database migrations, see the [handling-data-migrations documentation](../../docs/advanced-guide/handling-data-migrations)
For more examples, see the [using-migrations](https://github.com/kite-dev/kite/tree/main/examples/using-migrations)
---

## 3. ***`wrap grpc`***

   * The kite wrap grpc command streamlines gRPC integration in a Kite project by generating Kite's context-aware structures.
   * It simplifies setting up gRPC handlers with minimal steps, and accessing datasources, adding tracing as well as custom metrics. Based on the proto file it creates the handler/client with Kite's context.
   For detailed instructions on using grpc with Kite see the [gRPC documentation](../../advanced-guide/grpc/page.md)

### Command Usage
**gRPC Server**
```bash
  kite wrap grpc server --proto=<path_to_the_proto_file>
```
### Generated Files
**Server**
- ```{serviceName}_gofr.go (auto-generated; do not modify)```
- ```{serviceName}_server.go (example structure below)```

### Example Usage
**gRPC Server**

The command generates a server implementation template similar to this:
```go
package server

import (
   "github.com/sllt/kite/pkg/kite"
)

// Register the gRPC service in your app using the following code in your main.go:
//
// service.Register{ServiceName}ServerWithGofr(app, &server.{ServiceName}Server{})
//
// {ServiceName}Server defines the gRPC server implementation.
// Customize the struct with required dependencies and fields as needed.
type {ServiceName}Server struct {
}

// Example method (actual methods will depend on your proto file)
func (s *MyServiceServer) MethodName(ctx *kite.Context) (any, error) {
   // Replace with actual logic if needed
   return &ServiceResponse{
   }, nil
}
```
For detailed instruction on setting up a gRPC server with Kite see the [gRPC Server Documentation](https://github.com/sllt/kite/docs/advanced-guide/grpc#generating-g-rpc-server-handler-template-using)

**gRPC Client**
```bash
  kite wrap grpc client --proto=<path_to_the_proto_file>
```

**Client**
- ```{serviceName}_client.go (example structure below)```


### Example Usage:
Assuming the service is named hello, after generating the hello_client.go file, you can seamlessly register and access the gRPC service using the following steps:

```go
type GreetHandler struct {
	helloGRPCClient client.HelloKiteClient
}

func NewGreetHandler(helloClient client.HelloKiteClient) *GreetHandler {
    return &GreetHandler{
        helloGRPCClient: helloClient,
    }
}

func (g GreetHandler) Hello(ctx *kite.Context) (any, error) {
    userName := ctx.Param("name")
    helloResponse, err := g.helloGRPCClient.SayHello(ctx, &client.HelloRequest{Name: userName})
    if err != nil {
        return nil, err
    }

    return helloResponse, nil
}

func main() {
    app := kite.New()

// Create a gRPC client for the Hello service
    helloGRPCClient, err := client.NewHelloKiteClient(app.Config.Get("GRPC_SERVER_HOST"), app.Metrics())
    if err != nil {
		app.Logger().Errorf("Failed to create Hello gRPC client: %v", err)
    return
}

    greetHandler := NewGreetHandler(helloGRPCClient)

    // Register HTTP endpoint for Hello service
    app.GET("/hello", greetHandler.Hello)

    // Run the application
    app.Run()
}
```
For detailed instruction on setting up a gRPC server with Kite see the [gRPC Client Documentation](https://github.com/sllt/kite/docs/advanced-guide/grpc#generating-tracing-enabled-g-rpc-client-using)
For more examples refer [gRPC Examples](https://github.com/kite-dev/kite/tree/main/examples/grpc)
