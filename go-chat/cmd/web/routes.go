package main

import (
	"net/http"

	"go-chat/internal/handlers"

	"github.com/bmizerany/pat"
)

func routes() http.Handler {
	mux := pat.New() // create new router

	mux.Get("/", http.HandlerFunc(handlers.Home))         // render home page
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint)) // connect websockets

	// upload static files on the file server
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return mux
}
