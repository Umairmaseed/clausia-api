package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	router *gin.Engine
)

// InitServer saves a server instance in package
// not that sophisticated but it'll work
func InitServer(r *gin.Engine) {
	if router == nil {
		router = r
	}
}

func TestMain(m *testing.M) {
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
	apiGroup.POST("/signdocs", SignDocument)
	apiGroup.POST("/saltPdf", SaltPdf)
	apiGroup.POST("/verifydocs", VerifyPDF)
	apiGroup.POST("/getkey", GetLedgerKey)

	InitServer(r)

	os.Exit(m.Run())
}

func TestPingRoute(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"message\":\"pong\"}", w.Body.String())
}
