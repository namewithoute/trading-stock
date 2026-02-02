package main

import (
	"trading-stock/internal/bootstrap"
	"trading-stock/internal/global"
)

func main() {
	bootstrap.Setup()
	bootstrap.Run()
	defer func() {
		if global.Logger != nil {
			global.Logger.Sync()
		}
	}()

	global.Logger.Info("System started successfully!")
}
