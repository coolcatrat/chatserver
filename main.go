package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func handleConn(connection *websocket.Conn, hub *Hub) {
	// add connection to hub, call the closers to be run after function returns.
	defer connection.Close()
	hub.add(connection)
	defer hub.remove(connection)
	fmt.Println("Connected: ", connection.RemoteAddr())

	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			fmt.Println("disconnected: ", connection.RemoteAddr(), err)
			return
		}
		hub.broadcast(connection, message)
		fmt.Printf("[%s] received %d bytes: %q\n", connection.RemoteAddr(), len(message), message)
	}
}

func main() {
	hub := newHub()

	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir("static")))
	serveMux.HandleFunc("/ws", hub.handleUpgradeToWebsocket)

	// create server object connected on port 9000, requests handled by serveMux router.
	server := &http.Server{Addr: ":9000", Handler: serveMux}
	server.ListenAndServe()
}
