package kite

import (
	"context"
	"fmt"
	"net/http"
)

func (a *App) hasTelemetry() bool {
	return a.Config.GetOrDefault("KITE_TELEMETRY", defaultTelemetry) == "true"
}

func (a *App) sendTelemetry(client *http.Client, isStart bool) {
	url := fmt.Sprint(kiteHost, shutServerPing)

	if isStart {
		url = fmt.Sprint(kiteHost, startServerPing)

		a.container.Info("Kite records the number of active servers. Set KITE_TELEMETRY=false in configs to disable it.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return
	}

	req.Header.Set("Connection", "close")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	resp.Body.Close()
}
