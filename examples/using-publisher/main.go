package main

import (
	"encoding/json"

	"github.com/sllt/kite/examples/using-publisher/migrations"
	"github.com/sllt/kite/pkg/kite"
)

func main() {
	app := kite.New()

	app.Migrate(migrations.All())

	app.POST("/publish-order", order)
	app.POST("/publish-product", product)

	app.Run()
}

func order(ctx *kite.Context) (any, error) {
	type orderStatus struct {
		OrderId string `json:"orderId"`
		Status  string `json:"status"`
	}

	var data orderStatus

	err := ctx.Bind(&data)
	if err != nil {
		return nil, err
	}

	msg, _ := json.Marshal(data)

	err = ctx.GetPublisher().Publish(ctx, "order-logs", msg)
	if err != nil {
		return nil, err
	}

	return "Published", nil
}

func product(ctx *kite.Context) (any, error) {
	type productInfo struct {
		ProductId string `json:"productId"`
		Price     string `json:"price"`
	}

	var data productInfo

	err := ctx.Bind(&data)
	if err != nil {
		return nil, err
	}

	msg, _ := json.Marshal(data)

	err = ctx.GetPublisher().Publish(ctx, "products", msg)
	if err != nil {
		return nil, err
	}

	return "Published", nil
}
