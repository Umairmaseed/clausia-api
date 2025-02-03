package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goledgerdev/goprocess-api/api/handlers/auth"
	"github.com/goledgerdev/goprocess-api/api/handlers/contract"
	"github.com/goledgerdev/goprocess-api/api/handlers/documents"
	"github.com/goledgerdev/goprocess-api/api/handlers/notification"
	"github.com/goledgerdev/goprocess-api/api/handlers/user"
	"github.com/goledgerdev/goprocess-api/api/routes/docs"
	"github.com/goledgerdev/goprocess-api/websocket"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Register routes and handlers used by engine
func AddRoutesToEngine(r *gin.Engine, wsServer *websocket.WebSocketServer) {

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
	r.POST("/checkpw", a.CheckPw)

	r.POST("/uploaddocument", documents.UploadDocument)
	r.POST("/signdocument", documents.SignDocument)
	r.POST("/canceldocument", documents.CancelDocument)
	r.POST("/updatedocnameortimeout", documents.UpdateDocNameOrTimeout)
	r.POST("/updateemailorphone", a.UpdateEmailOrPhone)
	r.POST("/confirmnewemail", a.ConfirmNewEmail)
	r.GET("/listdocuments", documents.ListUserDocs)
	r.POST("/downloaddocument", documents.DownloadDocument)
	r.GET("/expectedsignatures", documents.ExpectedUserSignatures)
	r.GET("/getdocument", documents.GetDoc)
	r.GET("/listsuccessfulsignatures", documents.ListSuccessfulSignatures)
	r.GET("/pendingsignatures", documents.PendingSignatures)

	r.POST("/createcontract", contract.CreateContract)
	r.GET("/getusercontracts", contract.GetUserContracts)
	r.GET("/getcontract", contract.GetContract)
	r.POST("/addclause", contract.AddClause)
	r.POST("/removeclause", contract.RemoveClause)
	r.POST("/addclauses", contract.AddMultipleClauses)
	r.POST("/addparticipants", contract.AddParticipants)
	r.POST("/addreferencedate", contract.AddReferenceDate)
	r.POST("/addevaluatedate", contract.AddEvaluateDate)
	r.POST("/addinputstocheckfine", contract.AddInputsToCheckFine)
	r.POST("/addreviewtocontract", contract.AddReviewToContract)
	r.POST("/addinputstomakepayment", contract.AddInputsToMakePayment)
	r.POST("/cancelcontract", contract.CancelContract)
	r.POST("/createtemplate", contract.CreateTemplate)
	r.POST("/createtemplateclause", contract.CreateTemplateClause)
	r.POST("/edittemplate", contract.EditTemplate)
	r.POST("/edittemplateclause", contract.EditTemplateClause)
	r.POST("/duplicatetemplate", contract.DuplicateTemplate)
	r.POST("/removetemplate", contract.RemoveTemplate)
	r.POST("/removetemplateclause", contract.RemoveTemplateClause)
	r.POST("/addparticipantrequest", contract.AddParticipantRequest)
	r.POST("/sharetemplate", contract.ShareTemplate)
	r.POST("/viewsharedtemplate", contract.ViewSharedTemplate)

	r.GET("/getnotifications", notification.GetNotifications)
	r.POST("/deletenotification", notification.DeleteNotification)
	r.POST("/readnotifications", notification.ReadNotifications)
	r.POST("/unreadnotifications", notification.UnreadNotifications)
	r.GET("/getunreadnotifications", notification.GetUnreadNotifications)

	r.GET("/user/info", user.GetUserInfo)
	r.GET("/confirmuser", user.ConfirmUser)

	// serve swagger files
	docs.SwaggerInfo.BasePath = "/api"
	r.StaticFile("/swagger.yaml", "./api/routes/docs/swagger.yaml")

	url := ginSwagger.URL("/swagger.yaml")
	r.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))

	// WebSocket route
	r.GET("/ws", func(c *gin.Context) {
		// Convert the Gin context to the http.ResponseWriter and *http.Request
		http.HandlerFunc(websocket.WebSocketHandler(wsServer)).ServeHTTP(c.Writer, c.Request)
	})
}
