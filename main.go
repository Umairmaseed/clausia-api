package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/umairmaseed/clausia-api/api/handlers/contract"
	"github.com/umairmaseed/clausia-api/api/handlers/documents"
	"github.com/umairmaseed/clausia-api/api/server"
	"github.com/umairmaseed/clausia-api/db"
	"github.com/umairmaseed/clausia-api/websocket"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Create gin handler and start server
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",      // Dev address
			"http://localhost",           // Test addresses
			os.Getenv("FRONTEND_ORIGIN"), // Address of where the front-end is deployed
		},

		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Get the unit from environment variable
	unitStr := os.Getenv("CHECK_INTERVAL_UNIT")
	if unitStr == "" {
		unitStr = "minute"
	}

	var duration time.Duration
	switch unitStr {
	case "second":
		duration = 1 * time.Second
	case "minute":
		duration = 1 * time.Minute
	case "hour":
		duration = 1 * time.Hour
	case "day":
		duration = 24 * time.Hour
	default:
		duration = 1 * time.Minute
	}

	// Start the routine to check for expired documents and contracts to be executed
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				documents.CheckExpiredDocs()
				contract.ExecuteContract()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Initialize and start WebSocket server
	wsServer := websocket.NewWebSocketServer()

	go server.Serve(r, ctx, wsServer)

	mongo := db.GetDB()
	if mongo == nil {
		log.Fatal("Could not init database")
		return
	}

	// Watch for changes in MongoDB and notify users via WebSockets
	db.WatchForNotifications(mongo, wsServer, ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	cancel()
}
