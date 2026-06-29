package main

import (
	"net/http"
)

func main() {
	hub := newHub()

	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir("static")))
	serveMux.HandleFunc("/ws", hub.handleUpgradeToWebsocket)

	// create server object connected on port 9000, requests handled by serveMux router.
	server := &http.Server{Addr: ":9000", Handler: serveMux}
	server.ListenAndServe()
}
