package main

import (
	"github.com/sllt/kite/examples/grpc/grpc-unary-server/server"
	"github.com/sllt/kite/pkg/kite"
)

func main() {
	app := kite.New()

	server.RegisterHelloServerWithKite(app, server.NewHelloKiteServer())

	app.Run()
}
