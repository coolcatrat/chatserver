package main

import (
	"fmt"
	"net"
	"sync"
)

type Hub struct {
	mutex       sync.Mutex
	connections map[net.Conn]bool // registery
}

// methods for Hub
func newHub() *Hub {
	return &Hub{connections: make(map[net.Conn]bool)}
}
func (hub *Hub) add(connection net.Conn) {
	hub.mutex.Lock()
	hub.connections[connection] = true
	hub.mutex.Unlock()
}
func (hub *Hub) remove(connection net.Conn) {
	hub.mutex.Lock()
	delete(hub.connections, connection)
	hub.mutex.Unlock()
}
func (hub *Hub) broadcast(sender net.Conn, message []byte) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	for connection := range hub.connections {
		if connection == sender {
			continue
		}
		// add prefix to message to mention sender
		fmt.Fprintf(connection, "FROM: %s: %s\n", sender.RemoteAddr(), message)
	}
}

func handleConn(connection net.Conn, hub *Hub) {
	// add connection to hub, call the closers to be run after function returns.
	defer connection.Close()
	hub.add(connection)
	defer hub.remove(connection)
	fmt.Println("Connected: ", connection.RemoteAddr())

	buffer := make([]byte, 1024) // buffer for reading bytes from a connection
	for {
		bytesRead, err := connection.Read(buffer) // read bytes from connection into buffer.
		if err != nil {
			fmt.Println("disconnected: ", connection.RemoteAddr(), err)
			return
		}
		hub.broadcast(connection, buffer[:bytesRead])
		fmt.Printf("[%s] recieved %d bytes: %q\n", connection.RemoteAddr(), bytesRead, buffer[:bytesRead])
	}
}

func main() {
	hub := newHub()
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("could not listen:", err)
		return
	}
	fmt.Println("listening on :9000")

	for {
		newConnection, err := listener.Accept()
		if err != nil {
			fmt.Println("connection failed.", err)
			continue
		}
		go handleConn(newConnection, hub)
	}
}
