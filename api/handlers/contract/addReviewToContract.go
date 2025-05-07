package contract

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/utils"
)

type Review struct {
	Rating                 int                    `json:"rating" binding:"required"`
	Comments               string                 `json:"comments"`
	Date                   time.Time              `json:"date" binding:"required"`
	AutoExecutableContract map[string]interface{} `json:"autoExecutableContract" binding:"required"`
}

func AddReviewToContract(c *gin.Context) {
	var form Review
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

	reviewReq := make(map[string]interface{})

	reviewReq["user"] = map[string]interface{}{
		"@assettype": "user",
		"@key":       signerKey,
	}
	reviewReq["rating"] = form.Rating
	reviewReq["date"] = form.Date

	if form.Comments != "" {
		reviewReq["comments"] = form.Comments
	}

	req := map[string]interface{}{
		"autoExecutableContract": firstResult,
		"review":                 reviewReq,
	}

	updatedContract, err := chaincode.AddReview(req)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add review to contract", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"contract": updatedContract})
}
