package main

import (
	"github.com/sllt/kite/examples/grpc/grpc-streaming-server/server"
	"github.com/sllt/kite/pkg/kite"
)

func main() {
	app := kite.New()

	server.RegisterChatServiceServerWithKite(app, server.NewChatServiceKiteServer())

	app.Run()
}
