package main

import (
	"github.com/sllt/kite/pkg/kite"
)

func main() {
	app := kite.New()

	app.WebSocket("/ws", WSHandler)

	app.Run()
}

func WSHandler(ctx *kite.Context) (any, error) {
	var message string

	err := ctx.Bind(&message)
	if err != nil {
		ctx.Logger.Errorf("Error binding message: %v", err)
		return nil, err
	}

	ctx.Logger.Infof("Received message: %s", message)

	err = ctx.WriteMessageToSocket("Hello! Kite")
	if err != nil {
		return nil, err
	}

	return message, nil
}
