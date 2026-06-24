// Contains Hub struct and its mehtod

package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mutex       sync.Mutex
	connections map[*websocket.Conn]bool // registery
}

// methods for Hub
func newHub() *Hub {
	return &Hub{connections: make(map[*websocket.Conn]bool)}
}
func (hub *Hub) add(connection *websocket.Conn) {
	hub.mutex.Lock()
	hub.connections[connection] = true
	hub.mutex.Unlock()
}
func (hub *Hub) remove(connection *websocket.Conn) {
	hub.mutex.Lock()
	delete(hub.connections, connection)
	hub.mutex.Unlock()
}
func (hub *Hub) broadcast(outgoingJSON []byte) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	for connection := range hub.connections {
		// send message to (connection)
		connection.WriteMessage(websocket.TextMessage, outgoingJSON)
	}
}

// websocketUpgrader holds the config for turning HTTP request into a websocket
var websocketUpgrader = websocket.Upgrader{CheckOrigin: func(request *http.Request) bool { return true }}

func (hub *Hub) handleUpgradeToWebsocket(writer http.ResponseWriter, request *http.Request) {
	connection, err := websocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println("upgrade failed: ", err)
		return
	}
	handleConn(connection, hub)
}
