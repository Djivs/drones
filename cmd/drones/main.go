package main

import (
	"log"

	"drones/internal/pkg/app"
)

func main() {
	log.Println("Application start!")
	//api.StartServer()
	a := app.New()
	a.StartServer()

	log.Println("Application terminated!")
}
