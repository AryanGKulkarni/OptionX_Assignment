package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client structure
type Client struct {
	ID         string
	Conn       *websocket.Conn
	LastActive time.Time
}

var (
	clients   = make(map[string]*Client)
	clientsMu sync.Mutex
)

const (
	pingInterval = 5 * time.Second   // Send a ping every 5 seconds
	pongTimeout  = 10 * time.Second  // Disconnect if no pong within 10 seconds
)

// Handles new WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}

	// Assign a unique ID to the client
	clientID := uuid.New().String()
	client := &Client{ID: clientID, Conn: conn, LastActive: time.Now()}

	// Add client to the map
	clientsMu.Lock()
	clients[clientID] = client
	clientsMu.Unlock()

	fmt.Println("Client connected:", clientID)

	// Set pong handler
	conn.SetPongHandler(func(appData string) error {
		// fmt.Println("Received PONG from client:", client.ID)
		clientsMu.Lock()
		if c, exists := clients[clientID]; exists {
			c.LastActive = time.Now()
		}
		clientsMu.Unlock()
		return nil
	})

	// Start listening for messages
	go listenForMessages(client)

	// Start pinging client
	go pingClient(client)
}

// Listens for incoming messages from a client
func listenForMessages(client *Client) {
	defer disconnectClient(client)

	for {
		messageType, message, err := client.Conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message from", client.ID, ":", err)
			break
		}

		fmt.Printf("Message received from %s: %s\n", client.ID, string(message))

		// Broadcast message to all connected clients
		broadcastMessage(messageType, message, client.ID)
	}
}

// Sends ping messages to the client
func pingClient(client *Client) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		clientsMu.Lock()
		if time.Since(client.LastActive) > pongTimeout {
			fmt.Println("Client inactive, disconnecting:", client.ID)
			clientsMu.Unlock()
			disconnectClient(client)
			return
		}
		clientsMu.Unlock()

		// Send ping message
		if err := client.Conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
			fmt.Println("Error sending ping to", client.ID, ":", err)
			disconnectClient(client)
			return
		}
	}
}

// Broadcasts messages to all connected clients
func broadcastMessage(messageType int, message []byte, senderID string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for id, client := range clients {
		if id == senderID {
			continue
		}
		if err := client.Conn.WriteMessage(messageType, message); err != nil {
			fmt.Println("Error broadcasting to", id, ":", err)
			disconnectClient(client)
		}
	}
}

// Disconnects and removes a client from the list
func disconnectClient(client *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	fmt.Println("Client disconnected:", client.ID)
	client.Conn.Close()
	delete(clients, client.ID)
}

// Main function
func main() {
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
