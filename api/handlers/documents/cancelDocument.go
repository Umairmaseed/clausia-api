package documents

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/utils"
)

type CancelDocumentRequest struct {
	Key string `json:"@key" binding:"required"`
}

func CancelDocument(c *gin.Context) {
	var request CancelDocumentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	documentKey := request.Key

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

	doc, err := chaincode.GetDoc(documentKey)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ownerMap, ok := doc["owner"].(map[string]interface{})
	if !ok {
		logger.Error("Invalid owner data format")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid owner data format"})
		return
	}

	ownerKey, ok := ownerMap["@key"].(string)
	if !ok {
		logger.Error("Owner key not found")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Owner key not found"})
		return
	}

	if ownerKey != signerKey {
		logger.Error("User not authorized to changed the status of the document")
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not authorized to changed the status of the document"})
		return
	}

	documentMAp := map[string]interface{}{
		"@assetType": "document",
		"@key":       documentKey,
	}

	updatedDocument, err := chaincode.CancelDocument(documentMAp, float64(1))
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Document canceled successfully",
		"documentKey": updatedDocument,
	})
}
