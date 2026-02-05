# Solr

## Configuration
To connect to `Solr` DB, you need to provide the following environment variables:
- `HOST`: The hostname or IP address of your Solr DB server.
- `PORT`: The port number.

## Setup
Kite supports injecting Solr database that supports the following interface. Any driver that implements the interface can be added
using `app.AddSolr()` method, and user's can use Solr DB across application with `kite.Context`.

```go
type Solr interface {
	Search(ctx context.Context, collection string, params map[string]any) (any, error)
	Create(ctx context.Context, collection string, document *bytes.Buffer, params map[string]any) (any, error)
	Update(ctx context.Context, collection string, document *bytes.Buffer, params map[string]any) (any, error)
	Delete(ctx context.Context, collection string, document *bytes.Buffer, params map[string]any) (any, error)

	Retrieve(ctx context.Context, collection string, params map[string]any) (any, error)
	ListFields(ctx context.Context, collection string, params map[string]any) (any, error)
	AddField(ctx context.Context, collection string, document *bytes.Buffer) (any, error)
	UpdateField(ctx context.Context, collection string, document *bytes.Buffer) (any, error)
	DeleteField(ctx context.Context, collection string, document *bytes.Buffer) (any, error)
}
```

User's can easily inject a driver that supports this interface, this provides usability
without compromising the extensibility to use multiple databases.

Import the kite's external driver for Solr:

```shell
go get github.com/sllt/kite/pkg/kite/datasource/solr@latest
```
Note : This datasource package requires the user to create the collection before performing any operations.
While testing the below code create a collection using :
`curl --location 'http://localhost:2020/solr/admin/collections?action=CREATE&name=test&numShards=2&replicationFactor=1&wt=xml'`

```go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/sllt/kite/pkg/kite"
	"github.com/sllt/kite/pkg/kite/datasource/solr"
)

func main() {
	app := kite.New()

	app.AddSolr(solr.New(solr.Config{
		Host: app.Config.Get("HOST"),
		Port: app.Config.Get("PORT"),
	}))

	app.POST("/solr", post)
	app.GET("/solr", get)

	app.Run()
}

type Person struct {
	Name string
	Age  int
}

func post(c *kite.Context) (any, error) {
	p := []Person{{Name: "Srijan", Age: 24}}
	body, _ := json.Marshal(p)

	resp, err := c.Solr.Create(c, "test", bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func get(c *kite.Context) (any, error) {
	resp, err := c.Solr.Search(c, "test", nil)
	if err != nil {
		return nil, err
	}

	res, ok := resp.(solr.Response)
	if !ok {
		return nil, errors.New("invalid response type")
	}

	b, _ := json.Marshal(res.Data)
	err = json.Unmarshal(b, &Person{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
```
