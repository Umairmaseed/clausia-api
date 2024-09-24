package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/db"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type notificationForm struct {
	limits int `form:"limits" binding:"required"`
}

func GetNotifications(c *gin.Context) {
	var form notificationForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		logger.Error("Email not found in headers")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in headers"})
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	_, err = db.NewNotificationService(db.GetDB().Database()).GetNotificationsByUser(c.Request.Context(), signerKey, form.limits)
	if err != nil {
		errorhandler.ReturnError(c, err, "failed to generate notification", http.StatusInternalServerError)
	}
}
