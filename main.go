package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/resinstack/dtn/pkg/nomad"
	"github.com/resinstack/dtn/pkg/web"
)

func main() {
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "dtn",
		Level: hclog.LevelFromString("DEBUG"),
	})

	appLogger.Info("Starting Up")

	n := nomad.New()
	n.SetParentLogger(appLogger)
	n.Connect()

	ws := web.New()
	ws.SetParentLogger(appLogger)
	ws.SetNomadProvider(n)

	go func() {
		if err := ws.Start(":1323"); err != nil && err != http.ErrServerClosed {
			appLogger.Error("Error starting server", "error", err)
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	appLogger.Info("Shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ws.Shutdown(ctx); err != nil {
		appLogger.Error("Error shutting down", "error", err)
	}
}
