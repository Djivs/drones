package main

import (
	"drones/internal/pkg/app"
	"log"
)

// @title drones
// @version 0.0-0
// @description UAV route control applications

// @host 127.0.0.1:8000
// @schemes http
// @BasePath /

func main() {
	log.Println("Application start!")

	a := app.New()
	a.StartServer()

	log.Println("Application terminated!")
}
