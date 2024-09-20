package websocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
	"github.com/gorilla/websocket"
)

type Client struct {
	socket *websocket.Conn
	send   chan []byte
	server *WebSocketServer
	userID string
}

func (client *Client) ReadPump() {
	defer func() {
		client.server.unregister <- client
		client.socket.Close()
	}()
	for {
		_, message, err := client.socket.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		// Optionally handle incoming messages from the client here
		fmt.Println("Received message: ", string(message))
	}
}

func (client *Client) WritePump() {
	defer func() {
		client.socket.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				client.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := client.socket.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error writing message:", err)
				return
			}
		}
	}
}

func (server *WebSocketServer) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}

	email := r.Header.Get("Email")
	if email == "" {
		log.Println("Email not found in headers")
		socket.Close()
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		logger.Error(err)
		return
	}

	client := &Client{socket: socket, send: make(chan []byte, 256), server: server, userID: signerKey}
	server.register <- client

	go client.WritePump()
	go client.ReadPump()
}
