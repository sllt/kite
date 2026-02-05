package main

import (
	"github.com/sllt/kite/pkg/kite"
	"github.com/sllt/kite/pkg/kite/http/response"
)

func main() {
	app := kite.New()
	app.GET("/list", listHandler)
	app.AddStaticFiles("/", "./static")
	app.Run()
}

type Todo struct {
	Title string
	Done  bool
}

type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

func listHandler(*kite.Context) (any, error) {
	// Get data from somewhere
	data := TodoPageData{
		PageTitle: "My TODO list",
		Todos: []Todo{
			{Title: "Expand on Kite documentation ", Done: false},
			{Title: "Add more examples", Done: true},
			{Title: "Write some articles", Done: false},
		},
	}

	return response.Template{Data: data, Name: "todo.html"}, nil
}
