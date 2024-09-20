package websocket

import (
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var allowedOrigins = map[string]bool{
	"http://localhost:3000":      true,
	"http://localhost":           true,
	os.Getenv("FRONTEND_ORIGIN"): true,
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin: func(r *http.Request) bool {
	// 	origin := r.Header.Get("Origin")
	// 	return allowedOrigins[origin]
	// },
}

type WebSocketServer struct {
	clientsByUserID map[string][]*Client
	register        chan *Client
	unregister      chan *Client
	Broadcast       chan NotificationMessage
	mu              sync.Mutex
}

// Struct to represent a notification message
type NotificationMessage struct {
	UserID  string
	Message []byte
}

// NewWebSocketServer creates a new WebSocketServer
func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clientsByUserID: make(map[string][]*Client),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		Broadcast:       make(chan NotificationMessage),
	}
}

func (server *WebSocketServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.mu.Lock()
			server.clientsByUserID[client.userID] = append(server.clientsByUserID[client.userID], client)
			server.mu.Unlock()

		case client := <-server.unregister:
			server.mu.Lock()
			if clients, ok := server.clientsByUserID[client.userID]; ok {
				for i, c := range clients {
					if c == client {
						// Remove client from the slice
						server.clientsByUserID[client.userID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(server.clientsByUserID[client.userID]) == 0 {
					delete(server.clientsByUserID, client.userID)
				}
			}
			server.mu.Unlock()

		case notification := <-server.Broadcast:
			server.mu.Lock()
			if clients, ok := server.clientsByUserID[notification.UserID]; ok {
				for _, client := range clients {
					select {
					case client.send <- notification.Message:
					default:
						// If the client can't receive messages, close the connection
						close(client.send)
						server.unregister <- client
					}
				}
			}
			server.mu.Unlock()
		}
	}
}
