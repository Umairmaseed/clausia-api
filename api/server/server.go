package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/routes"
	"github.com/goledgerdev/goprocess-api/env"
)

func defaultServer(r *gin.Engine, port string) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}
}

// Serve starts the server with gin's default engine.
// Server gracefully shut's down
func Serve(r *gin.Engine, ctx context.Context) {
	// Register routes and handlers
	routes.AddRoutesToEngine(r)

	// Get port
	var port string
	if port = os.Getenv(env.SERVER_PORT); port == "" {
		port = "8080"
	}

	// Returns a http.Server from provided handler
	srv := defaultServer(r, port)

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
	gin.SetMode(gin.TestMode)
	r := gin.New()

	routes.AddRoutesToEngine(r)

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
