package documents

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type expectedUserform struct {
	Status *string `json:"status" form:"status"`
}

func ExpectedUserSignatures(c *gin.Context) {

	var form expectedUserform

	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind form data", http.StatusBadRequest)
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

	signer := map[string]interface{}{
		"@assetType": "user",
		"@key":       signerKey,
	}

	reqMap := map[string]interface{}{
		"signer": signer,
	}

	if form.Status != nil {
		statusFloat, err := strconv.ParseFloat(*form.Status, 64)
		if err != nil {
			errorhandler.ReturnError(c, err, "Invalid status value", http.StatusBadRequest)
			return
		}
		reqMap["status"] = statusFloat
	}

	documents, err := chaincode.GetExpectedUserDoc(reqMap)
	if err != nil {
		log.Fatalf("Error getting expected user documents: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
	})
}
