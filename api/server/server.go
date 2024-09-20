package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/routes"
	"github.com/goledgerdev/goprocess-api/certs"
	"github.com/goledgerdev/goprocess-api/env"
	"github.com/goledgerdev/goprocess-api/websocket"
)

func defaultServer(r *gin.Engine, port string) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}
}

// Serve starts the server with gin's default engine.
// Server gracefully shut's down
func Serve(r *gin.Engine, ctx context.Context, wsServer *websocket.WebSocketServer) {
	// Initialize and start WebSocket server
	go wsServer.Run()

	// Register routes and handlers
	routes.AddRoutesToEngine(r, wsServer)

	// Get port
	var port string
	if port = os.Getenv(env.SERVER_PORT); port == "" {
		port = "8080"
	}

	// Returns a http.Server from provided handler
	srv := defaultServer(r, port)

	// Init CA manager
	err := initCAMngr()
	if err != nil {
		log.Panic(err)
	}

	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	go func(server *http.Server) {
		log.Println("Listening on port " + port)
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Panic(err)
		}
	}(srv)

	// Graceful shutdown
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
	}
	log.Println("Shutting down")
}

// Serve sync starts the server with a given wait group.
// When server starts, the wait group counter is increased and processes
// that depend on server can be ran synchronously with it
func ServeSync(ctx context.Context, wg *sync.WaitGroup) {
	// Initialize and start WebSocket server
	wsServer := websocket.NewWebSocketServer()
	go wsServer.Run()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	routes.AddRoutesToEngine(r, wsServer)

	srv := defaultServer(r, "8080")

	go func(server *http.Server) {
		log.Println("Listening on port 8080")
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Panic(err)
		}
		// finish wait group
		time.Sleep(1 * time.Second)
		wg.Done()
	}(srv)

	wg.Add(1)
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
	}
	log.Println("Shutting down")
}

func initCAMngr() error {

	caMngr, err := certs.InitCAMngr(os.Getenv("SDK_CONFIG_PATH"), os.Getenv("CA_URL"))
	if caMngr == nil {
		log.Printf("Error initializing CA manager: %v", err)
		return errors.New("could not init CA manager")
	}
	return nil
}
