# ClickHouse

## Configuration
To connect to `ClickHouse`, you need to provide the following environment variables and use it:
- `HOSTS`: The hostname or IP address of your `ClickHouse` server.
- `USERNAME`: The username for connecting to the database.
- `PASSWORD`: The password for the specified user.
- `DATABASE`: The name of the database to connect to.


## Setup
Kite supports injecting ClickHouse that supports the following interface. Any driver that implements the interface can be added
using `app.AddClickhouse()` method, and user's can use ClickHouse across application with `kite.Context`.
```go
type Clickhouse interface {
	Exec(ctx context.Context, query string, args ...any) error
	Select(ctx context.Context, dest any, query string, args ...any) error
	AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error
}
```

User's can easily inject a driver that supports this interface, this provides usability without
compromising the extensibility to use multiple databases.

Import the kite's external driver for ClickHouse:

```shell
go get github.com/sllt/kite/pkg/kite/datasource/clickhouse@latest
```

### Example
```go
package main

import (
	"github.com/sllt/kite/pkg/kite"
	"github.com/sllt/kite/pkg/kite/datasource/clickhouse"
)

type User struct {
	Id   string `ch:"id"`
	Name string `ch:"name"`
	Age  int    `ch:"age"`
}

func main() {
	app := kite.New()

	app.AddClickhouse(clickhouse.New(clickhouse.Config{
		Hosts:    app.Config.Get("HOSTS"),
		Username: app.Config.Get("USERNAME"),
		Password: app.Config.Get("PASSWORD"),
		Database: app.Config.Get("DATABASE"),
	}))

	app.POST("/user", Post)
	app.GET("/user", Get)

	app.Run()
}

func Post(ctx *kite.Context) (any, error) {
	err := ctx.Clickhouse.Exec(ctx, "INSERT INTO users (id, name, age) VALUES (?, ?, ?)", "8f165e2d-feef-416c-95f6-913ce3172e15", "aryan", 10)
	if err != nil {
		return nil, err
	}

	return "successfully inserted", nil
}

func Get(ctx *kite.Context) (any, error) {
	var user []User

	err := ctx.Clickhouse.Select(ctx, &user, "SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	return user, nil
}
```
