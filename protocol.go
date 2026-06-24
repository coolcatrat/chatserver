package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// TYPE: CHAT
// what a client is ENTITLED to assert
type IncomingMessage struct {
	MessageType string `json:"type"`
	Text        string `json:"text"`
	// room later, when rooms exist
}

// the authoritative broadcast record the server constructs
type OutgoingMessage struct {
	MessageType string `json:"type"`
	Text        string `json:"text"`
	Sender      string `json:"sender"`    // server fills from session ID
	Timestamp   int64  `json:"timestamp"` // server stamps on receipt
	// room later
}

func handleChatMessage(hub *Hub, connection *websocket.Conn, message []byte) {
	var incomingMessage IncomingMessage
	parseError := json.Unmarshal(message, &incomingMessage)
	if parseError != nil {
		log.Println("Error parsing CHAT MESSAGE:", parseError)
		return
	}

	// translate across the trust boundary: copy the client's CLAIMS,
	// then ADD the fields only the server can vouch for.
	outgoingMessage := OutgoingMessage{
		MessageType: "message",                        // server asserts the type
		Text:        incomingMessage.Text,             // trusted claim, copied
		Sender:      connection.RemoteAddr().String(), // PLACEHOLDER — session ID later
		Timestamp:   time.Now().Unix(),                // server stamps on receipt
	}

	serializedBytes, marshalError := json.Marshal(outgoingMessage)
	if marshalError != nil {
		log.Println("Error marshaling outgoing CHAT MESSAGE:", marshalError)
		return
	}

	hub.broadcast(serializedBytes)
}
