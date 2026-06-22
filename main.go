package main

import (
	"fmt"
	"net"
)

func handleConn(connection net.Conn) {
	defer connection.Close()
	fmt.Println("got a connection from: ", connection.RemoteAddr())

	buffer := make([]byte, 1024) // buffer for reading bytes from a connection
	for {
		n, err := connection.Read(buffer) // read bytes from connection into buffer.
		if err != nil {
			fmt.Println("connection closed: ", connection.RemoteAddr(), err)
			return
		}
		fmt.Printf("[%s] recieved %d bytes: %q\n", connection.RemoteAddr(), n, buffer[:n])
	}
}

func main() {

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
		go handleConn(newConnection)
	}
}
