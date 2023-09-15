package main

import (
	"drones/internal/pkg/app"
	"log"
)

func main() {
	log.Println("Application start!")

	a := app.New()
	a.StartServer()

	log.Println("Application terminated!")
}
