package main

import (
	"trading-stock/internal/bootstrap"
	"trading-stock/internal/global"
)

func main() {
	bootstrap.Setup()
	// Log message should be before Run() because Run() blocks until shutdown
	global.Logger.Info("System started successfully!")

	bootstrap.Run()
}
