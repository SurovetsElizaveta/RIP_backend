package main

import (
	"log"

	"Lab1/internal/api"
)

func main() {
	log.Println("Application start")
	api.StartServer()
	log.Println("Application terminated")
}
