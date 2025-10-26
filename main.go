package main

import (
	"Focogram/config"
	"Focogram/router"
)

func main() {
	config.InitConfig()
	r := router.SetRouter()
	port := config.AppConfig.App.Port
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
