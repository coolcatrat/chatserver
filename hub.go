// Contains Hub struct and its mehtod

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	// generate session and room ID's
	"github.com/gorilla/websocket"
)

type SessionID string
type RoomID string
type Hub struct {
	mutex              sync.RWMutex
	clientsBySessionID map[SessionID]*Client
	roomsByID          map[RoomID]*Room
}
type Client struct {
	sessionID    SessionID // server assigned value
	displayName  string    // cosmetic value
	connection   *websocket.Conn
	writeChannel chan []byte
	done         chan struct{}
	rooms        map[*Room]bool // for cleanup
}

const GlobalRoomID RoomID = "global"

type Room struct {
	roomID     RoomID           // server assigned value
	name       string           // cosmetic value
	visibility string           // global, public, private
	members    map[*Client]bool // for broadcast
	mutex      sync.RWMutex
}

func (client *Client) disconnect(hub *Hub) {
	// 1. wakes writeLoop, defers connection.Close(),
	// readLoop errors on ReadMessage() and returns.
	close(client.done)
	// 2. update every room's registry
	for room := range client.rooms {
		// write operation, so use .Lock(). (shared map in .broadcast())
		room.mutex.Lock()
		delete(room.members, client)
		room.mutex.Unlock()
	}
	// 3. update hub's registry
	hub.mutex.Lock()
	delete(hub.clientsBySessionID, client.sessionID)
	hub.mutex.Unlock()
}

// protects from a cascading hang.
func (client *Client) send(payloadBytes []byte) {
	select {
	case client.writeChannel <- payloadBytes:
	case <-client.done:
	}
}
func newSessionID() SessionID {
	randomBytes := make([]byte, 16)
	if _, randomError := rand.Read(randomBytes); randomError != nil {
		panic(randomError) // crypto/rand not working
	}
	return SessionID(hex.EncodeToString(randomBytes))
}
func newClient(connection *websocket.Conn, sessionID SessionID) *Client {
	return &Client{
		sessionID:    sessionID,
		displayName:  "user-" + string(sessionID)[:8], // placeholder name
		connection:   connection,
		writeChannel: make(chan []byte, 16),
		done:         make(chan struct{}),
		rooms:        map[*Room]bool{},
	}
}

// =======================
//
//	methods for room
//
// =======================
func newRoom(roomID RoomID, name string, visibility string) *Room {
	return &Room{
		roomID:     roomID,
		name:       name,
		visibility: visibility,
		members:    make(map[*Client]bool),
	}
}
func (room *Room) broadcast(payloadBytes []byte) {
	room.mutex.RLock()
	recipients := make([]*Client, 0, len(room.members))
	for client := range room.members {
		recipients = append(recipients, client)
	}
	room.mutex.RUnlock() // lock released BEFORE any send

	for _, client := range recipients {
		client.send(payloadBytes) // blocking/slow work happens outside the lock
	}
}
func (room *Room) addMember(client *Client) {
	room.mutex.Lock()
	room.members[client] = true // shared across all clients → locked
	room.mutex.Unlock()

	client.rooms[room] = true // owned by one goroutine → unlocked
}

// methods for Hub

func newHub() *Hub {
	hub := &Hub{
		clientsBySessionID: make(map[SessionID]*Client),
		roomsByID:          make(map[RoomID]*Room),
	}
	globalRoom := newRoom(GlobalRoomID, "Global", "global")
	hub.roomsByID[GlobalRoomID] = globalRoom
	return hub
}

// websocketUpgrader holds the config for turning HTTP request into a websocket
var websocketUpgrader = websocket.Upgrader{CheckOrigin: func(request *http.Request) bool { return true }}

// sets up Client object for new connnection.
func (hub *Hub) handleUpgradeToWebsocket(writer http.ResponseWriter, request *http.Request) {
	connection, err := websocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println("upgrade failed: ", err)
		return
	}
	sessionID := newSessionID()
	client := newClient(connection, sessionID)
	hub.register(client) // fully registered + joined global

	go client.writeLoop() // write before any IO starts.
	go client.readLoop(hub)
}

func (hub *Hub) register(client *Client) {
	hub.mutex.Lock()
	hub.clientsBySessionID[client.sessionID] = client
	globalRoom := hub.roomsByID[GlobalRoomID]
	hub.mutex.Unlock()

	globalRoom.addMember(client)
}
