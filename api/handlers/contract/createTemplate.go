package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/db"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type createTemplateForm struct {
	Id          string                   `form:"id" binding:"required"`
	Name        string                   `form:"name" binding:"required"`
	Description string                   `form:"description"`
	Public      bool                     `form:"public" binding:"required"`
	Clauses     []map[string]interface{} `form:"clauses"`
}

func CreateTemplate(c *gin.Context) {
	var form createTemplateForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Error("Failed to bind request form: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		logger.Error("Email not found in headers")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in headers"})
		return
	}

	userKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	userAsset := map[string]interface{}{
		"@assetType": "user",
		"@key":       userKey,
	}

	req := map[string]interface{}{
		"name":    form.Name,
		"creator": userAsset,
		"public":  form.Public,
		"id":      form.Id,
	}

	if len(form.Clauses) > 0 {
		req["clauses"] = form.Clauses
	}
	if form.Description != "" {
		req["description"] = form.Description
	}

	contract, err := chaincode.CreateTemplate(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var notifications []db.Notification
	notifications = append(notifications, db.Notification{
		UserID:   userKey,
		Type:     "template",
		Message:  "Template " + form.Name + " created",
		Metadata: map[string]string{"templateId": form.Id},
	})

	_, err = db.NewNotificationService(db.GetDB().Database()).CreateNotification(c.Request.Context(), &notifications)
	if err != nil {
		errorhandler.ReturnError(c, err, "failed to generate notification", http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{"template": contract})

}
