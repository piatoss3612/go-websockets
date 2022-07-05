package main

import (
	"go-chat/internal/handlers"
	"log"
	"net/http"
)

func main() {
	routes := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWSChannel()

	log.Println("Starting web server on port 8080")

	if err := http.ListenAndServe(":8080", routes); err != nil {
		log.Panic(err)
	}
}
