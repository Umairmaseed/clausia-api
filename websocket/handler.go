package websocket

import (
	"net/http"
)

// WebSocketHandler is a function that can be registered with a route
func WebSocketHandler(server *WebSocketServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.ServeWebSocket(w, r)
	}
}
