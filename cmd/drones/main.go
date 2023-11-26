package main

import (
	"context"
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

	a, err := app.New(context.Background())
	if err != nil {
		log.Println(err)

		return
	}

	a.StartServer()

	log.Println("Application terminated!")
}
