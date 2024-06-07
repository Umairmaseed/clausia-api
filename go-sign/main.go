package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/go-sign/api"
	"github.com/goledgerdev/go-sign/logger"
)

var log = logger.Logger().Sugar()

func main() {
	r := setupRouter()
	fmt.Println("Server init")

	r.Run(":8082")
}

func setupRouter() *gin.Engine {
	// gin configuration
	r := gin.Default()
	r.MaxMultipartMemory = 100 << 20 // 100MiB

	r.Use(cors.New(cors.Config{
		// AllowAllOrigins: true,
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT"},
		AllowHeaders:  []string{"Origin", "Content-Type", "authorization"},
		ExposeHeaders: []string{"Content-Length", "Content-Disposition"},
		MaxAge:        12 * time.Hour,
	}))

	// healthcheck
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// api endpoints
	apiGroup := r.Group("/api")
	apiGroup.POST("/signdocs", api.SignDocument)
	apiGroup.POST("/saltPdf", api.SaltPdf)
	apiGroup.POST("/verifydocs", api.VerifyPDF)
	apiGroup.POST("/getkey", api.GetLedgerKey)

	return r
}
