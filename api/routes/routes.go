package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goledgerdev/goprocess-api/api/routes/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Register routes and handlers used by engine
func AddRoutesToEngine(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api-docs/index.html")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// serve swagger files
	docs.SwaggerInfo.BasePath = "/api"
	r.StaticFile("/swagger.yaml", "./api/routes/docs/swagger.yaml")

	url := ginSwagger.URL("/swagger.yaml")
	r.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))
}
