package main

import (
	"github.com/hashicorp/go-hclog"
)

func main() {
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "dtn",
		Level: hclog.LevelFromString("DEBUG"),
	})

	appLogger.Info("Starting Up")
}
