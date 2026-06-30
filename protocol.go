package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// TYPE: CHAT
// what a client is ENTITLED to assert
type MessageType string

const (
	MessageTypeChat MessageType = "message"
	// future: MessageTypeJoin, MessageTypeLeave, MessageTypeCreateRoom
)

type IncomingMessage struct {
	MessageType MessageType `json:"type"`
	RoomID      RoomID      `json:"roomID"`
	Text        string      `json:"text"`
}

// the authoritative broadcast record the server constructs
type OutgoingMessage struct {
	MessageType MessageType `json:"type"`
	RoomID      RoomID      `json:"roomID"`
	RoomName    string      `json:"roomName"`
	Text        string      `json:"text"`
	Sender      string      `json:"sender"`    // server fills from session ID
	Timestamp   int64       `json:"timestamp"` // server stamps on receiptIncomingMessage
}

// connection.SetReadDeadline(deadline)
// connection.SetPongHandler(callback) - register a function to handle when a callback arrives, just reset the deadline. runs inside readloop
// connection.WriteControl
const (
	pongWait   = 60 * time.Second    // max silence tolerated before declaring the client dead
	pingPeriod = (pongWait * 9) / 10 // = 54s. send a ping this often
	writeWait  = 10 * time.Second    // max time a single write may take before erroring
)

func (client *Client) readLoop(hub *Hub) {
	defer client.disconnect(hub)

	// ping pong (presence)
	client.connection.SetReadDeadline(time.Now().Add(pongWait))
	client.connection.SetPongHandler(func(appData string) error {
		client.connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	fmt.Printf("Client %s now active.\n", client.displayName)

	for {
		_, rawMessageBytes, readError := client.connection.ReadMessage()
		if readError != nil {
			return // ends connection
		}

		// parse JSON, continue if badJSON.
		var incomingJSON IncomingMessage
		if parseError := json.Unmarshal(rawMessageBytes, &incomingJSON); parseError != nil {
			log.Println("bad JSON from :", client.connection.RemoteAddr(), parseError)
			continue // skip iteration, dont return.
		}
		fmt.Printf("Recieved %d bytes from %s: %s\n", len(rawMessageBytes), client.displayName, rawMessageBytes)

		switch incomingJSON.MessageType {
		case MessageTypeChat:
			client.handleChatMessage(hub, incomingJSON)
			// next case MessageTypeJoin:
		}

	}
}
func (client *Client) writeLoop() {
	pingTicker := time.NewTicker(pingPeriod) // pingTicker's channel fires every pingPeriod (s)
	defer pingTicker.Stop()
	defer client.connection.Close()

	// infinite loop, checks for:
	// 1. write to client
	// 2. ping client for presence
	// 3. end connection signal
	for {
		select {

		// write bytes to websocket connection, used to broadcast message.
		case payloadBytes := <-client.writeChannel:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			err := client.connection.WriteMessage(websocket.TextMessage, payloadBytes)
			if err != nil {
				return
			}
			fmt.Printf("wrote to %s\n", client.displayName)
		// pingPeriod elapsed.
		case <-pingTicker.C:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))                                // cap how long the write can block.
			err := client.connection.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)) // send ping message
			if err != nil {                                                                              // write failed
				return // ends connection
			}
		// close connection
		case <-client.done:
			return
		}

	}
}

func (client *Client) handleChatMessage(hub *Hub, chatMessage IncomingMessage) {
	hub.mutex.RLock()
	room, roomExists := hub.roomsByID[chatMessage.RoomID]
	hub.mutex.RUnlock()
	if !roomExists {
		return
	}
	if !client.rooms[room] {
		return // not a member; never joined
	}
	outgoingMessage := OutgoingMessage{
		MessageType: MessageTypeChat,
		RoomID:      room.roomID,        // identity, for client-side routing
		RoomName:    room.name,          // cosmetic, for display
		Sender:      client.displayName, // server-stamped
		Text:        chatMessage.Text,
		Timestamp:   time.Now().Unix(),
	}
	payloadBytes, marshalError := json.Marshal(outgoingMessage)
	if marshalError != nil {
		return
	}
	room.broadcast(payloadBytes)
}
