package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/db"
	"github.com/goledgerdev/goprocess-api/utils"
)

type addClauseForm struct {
	AutoExecutableContract map[string]interface{}   `form:"autoExecutableContract" binding:"required"`
	Id                     string                   `form:"id" binding:"required"`
	Description            string                   `form:"description"`
	Category               string                   `form:"category"`
	Parameters             map[string]interface{}   `form:"parameters"`
	Input                  map[string]interface{}   `form:"input"`
	Dependencies           []map[string]interface{} `form:"dependencies"`
	ActionType             float64                  `form:"actionType" binding:"required"`
}

func AddClause(c *gin.Context) {
	var form addClauseForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		errorhandler.ReturnError(c, fmt.Errorf("email not found in headers"), "email not found in headers", http.StatusBadRequest)
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find user key", http.StatusInternalServerError)
		return
	}

	contractAsset, err := chaincode.SearchAsset(form.AutoExecutableContract)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find contract asset", http.StatusInternalServerError)
		return
	}

	results, ok := contractAsset["result"].([]interface{})
	if !ok || len(results) == 0 {
		errorhandler.ReturnError(c, fmt.Errorf("no results found in contract asset"), "no results found in contract asset", http.StatusInternalServerError)
		return
	}

	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid result format"), "invalid result format", http.StatusInternalServerError)
		return
	}

	contractOwner, ok := firstResult["owner"].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("could not find owner of th contract"), "could not find owner of th contract", http.StatusInternalServerError)
		return
	}

	if contractOwner["@key"] != signerKey {
		errorhandler.ReturnError(c, fmt.Errorf("only the owner of the contract can add the clause"), "only the owner of the contract can add the clause", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"autoExecutableContract": form.AutoExecutableContract,
		"id":                     form.Id,
		"actionType":             form.ActionType,
	}

	if form.Description != "" {
		reqMap["description"] = form.Description
	}
	if form.Category != "" {
		reqMap["category"] = form.Category
	}
	if form.Parameters != nil {
		reqMap["parameters"] = form.Parameters
	}
	if form.Input != nil {
		reqMap["input"] = form.Input
	}
	if form.Dependencies != nil {
		reqMap["dependencies"] = form.Dependencies
	}

	updatedContractAsset, err := chaincode.AddClause(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add clause to contract", http.StatusInternalServerError)
		return
	}

	var notifications []db.Notification

	notifications = append(notifications, db.Notification{
		UserID:  signerKey,
		Type:    "contract",
		Message: "Clause added successfully",
	})

	participants, ok := firstResult["participants"].([]interface{})
	if ok {
		for _, participant := range participants {
			participantMap, ok := participant.(map[string]interface{})
			if ok {
				if userID, ok := participantMap["@key"].(string); ok {
					notifications = append(notifications, db.Notification{
						UserID:  userID,
						Type:    "contract",
						Message: "A new clause has been added to the contract you are participating in.",
						Metadata: map[string]string{
							"contractId": form.Id,
						},
					})
				}
			}
		}
	}

	_, err = db.NewNotificationService(db.GetDB().Database()).CreateNotification(c.Request.Context(), &notifications)
	if err != nil {
		errorhandler.ReturnError(c, err, "failed to generate notification", http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{"contract": updatedContractAsset})

}
