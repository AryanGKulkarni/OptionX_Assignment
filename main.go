package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"strings"
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

type Message struct {
	ID      string `json:"id"`      // Receiver's ID
	Message string `json:"message"` // Actual message
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

	sendWelcomeMessage(client)

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

func sendWelcomeMessage(client *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// Create a list of connected client IDs
	var clientList []string
	for id := range clients {
		if id == client.ID {
			clientList = append(clientList, fmt.Sprintf("%s (You)", id))
		} else {
			clientList = append(clientList, id)
		}
	}

	// Prepare and send the message
	welcomeMessage := fmt.Sprintf(
		"Welcome! Your Client ID: %s\nConnected Clients:\n%s",
		client.ID, strings.Join(clientList, "\n"),
	)

	if err := client.Conn.WriteMessage(websocket.TextMessage, []byte(welcomeMessage)); err != nil {
		fmt.Println("Error sending welcome message to", client.ID, ":", err)
	}
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

		// Parse the JSON message
		var msgData Message
		if err := json.Unmarshal(message, &msgData); err != nil {
			fmt.Println("Invalid message format from", client.ID, ":", err)
			continue
		}

		fmt.Printf("Message received from %s to %s: %s\n", client.ID, msgData.ID, msgData.Message)

		// Send the message only to the specified client
		err = sendMessageToClient(messageType, []byte(msgData.Message), client.ID, msgData.ID)
		if err != nil {
			fmt.Println("Failed to send message to", msgData.ID, ":", err)
		}
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
// func broadcastMessage(messageType int, message []byte, senderID string) {
// 	clientsMu.Lock()
// 	defer clientsMu.Unlock()

// 	for id, client := range clients {
// 		if id == senderID {
// 			continue
// 		}
// 		if err := client.Conn.WriteMessage(messageType, message); err != nil {
// 			fmt.Println("Error broadcasting to", id, ":", err)
// 			disconnectClient(client)
// 		}
// 	}
// }

func sendMessageToClient(messageType int, message []byte, senderID string, receiverID string) error {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	client, exists := clients[receiverID]
	if !exists {
		return fmt.Errorf("client %s not found", receiverID)
	}


	formattedMessage := fmt.Sprintf("You have message from %s: %s", senderID, string(message))

	if err := client.Conn.WriteMessage(messageType, []byte(formattedMessage)); err != nil {
		fmt.Println("Error sending message to", receiverID, ":", err)
		disconnectClient(client)
		return err
	}

	return nil
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
