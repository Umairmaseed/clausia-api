package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goledgerdev/goprocess-api/api/handlers/auth"
	"github.com/goledgerdev/goprocess-api/api/handlers/documents"
	"github.com/goledgerdev/goprocess-api/api/routes/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Register routes and handlers used by engine
func AddRoutesToEngine(r *gin.Engine) {

	a := auth.NewAuth()

	r.POST("/login", a.SignIn)
	r.POST("/signup", a.SignUp)
	r.POST("/otp", a.VerifyAccount)
	r.POST("/logout", a.SignOut)
	r.POST("/changepw", a.ChangePassword)
	r.POST("/forgotpw", a.ForgotPassword)
	r.POST("/confirmforgotpw", a.ConfirmForgotPassword)
	r.POST("/resend", a.ResendCode)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api-docs/index.html")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.Use(a.AuthMiddleware())
	r.POST("/uploaddocument", documents.UploadDocument)
	r.POST("/signdocument", documents.SignDocument)
	r.POST("/canceldocument", documents.CancelDocument)
	r.POST("/updatedocnameortimeout", documents.UpdateDocNameOrTimeout)
	r.POST("/updateemailorphone", a.UpdateEmailOrPhone)
	r.POST("/confirmnewemail", a.ConfirmNewEmail)

	// serve swagger files
	docs.SwaggerInfo.BasePath = "/api"
	r.StaticFile("/swagger.yaml", "./api/routes/docs/swagger.yaml")

	url := ginSwagger.URL("/swagger.yaml")
	r.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))
}
