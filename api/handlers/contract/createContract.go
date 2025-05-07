package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/db"
	"github.com/umairmaseed/clausia-api/utils"
)

type createContractForm struct {
	Name          string                   `form:"name" binding:"required"`
	SignatureDate string                   `form:"signatureDate" binding:"required"`
	Clauses       []map[string]interface{} `form:"clauses"`
	Data          map[string]interface{}   `form:"data"`
	Participants  []map[string]interface{} `form:"participants"`
}

func CreateContract(c *gin.Context) {
	var form createContractForm
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

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	signerAsset := map[string]interface{}{
		"@assetType": "user",
		"@key":       signerKey,
	}

	req := map[string]interface{}{
		"name":          form.Name,
		"signatureDate": form.SignatureDate,
		"owner":         signerAsset,
	}

	if len(form.Clauses) > 0 {
		req["clauses"] = form.Clauses
	}
	if len(form.Participants) > 0 {
		req["participants"] = form.Participants
	}
	if form.Data != nil {
		req["data"] = form.Data
	}

	contract, err := chaincode.CreateAutoExecutableContract(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var notifications []db.Notification

	notifications = append(notifications, db.Notification{
		UserID:   signerKey,
		Type:     "contract",
		Message:  "You have created a new contract",
		Metadata: map[string]string{"contractID": contract["@key"].(string)},
	})

	for _, participant := range form.Participants {
		notifications = append(notifications, db.Notification{
			UserID:   participant["@key"].(string),
			Type:     "contract",
			Message:  "You have been invited to sign a contract",
			Metadata: map[string]string{"contractID": contract["@key"].(string)},
		})
	}

	_, err = db.NewNotificationService(db.GetDB().Database()).CreateNotification(c.Request.Context(), &notifications)
	if err != nil {
		errorhandler.ReturnError(c, err, "failed to generate notification", http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{"contract": contract})

}
