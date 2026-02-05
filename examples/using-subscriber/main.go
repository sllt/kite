package main

import (
	"github.com/sllt/kite/examples/using-subscriber/migrations"
	"github.com/sllt/kite/pkg/kite"
)

func main() {
	app := kite.New()

	app.Migrate(migrations.All())

	app.Subscribe("products", productHandler)

	app.Subscribe("order-logs", orderHandler)

	app.Run()
}

func productHandler(c *kite.Context) error {
	var productInfo struct {
		ProductId string `json:"productId"`
		Price     string `json:"price"`
	}

	err := c.Bind(&productInfo)
	if err != nil {
		c.Logger.Error(err)

		return nil
	}

	c.Logger.Info("Received product", productInfo)

	return nil
}

func orderHandler(c *kite.Context) error {
	var orderStatus struct {
		OrderId string `json:"orderId"`
		Status  string `json:"status"`
	}

	err := c.Bind(&orderStatus)
	if err != nil {
		c.Logger.Error(err)

		return nil
	}

	c.Logger.Info("Received order", orderStatus)

	return nil
}
