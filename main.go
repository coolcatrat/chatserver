package main

import (
	"encoding/json"
	"fmt"
	"log"
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
		// read the bytes recieved from client, if unsucessful means client has disconnected.
		_, message, readError := connection.ReadMessage()
		if readError != nil {
			fmt.Println("disconnected: ", connection.RemoteAddr(), readError)
			return
		}

		// log the data recieved
		fmt.Printf("[%s] received %d bytes: %q\n", connection.RemoteAddr(), len(message), message)

		// parse the recieved JSON.
		var IncomingMessage struct {
			MessageType string `json:"type"`
		}
		parseError := json.Unmarshal(message, &IncomingMessage)

		if parseError != nil {
			// error in parsing
			log.Println("bad JSON from :", connection.RemoteAddr(), parseError)
			continue // skip iteration, dont return.
		}

		switch IncomingMessage.MessageType {
		case "message":
			handleChatMessage(hub, connection, message)
		default:
			log.Println("Unknown MessageType from", connection.RemoteAddr(), ":", IncomingMessage.MessageType)
		}
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
