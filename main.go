package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/server"
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

	go server.Serve(r, ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	cancel()
}
