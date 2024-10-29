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

type pendingSignatureForm struct {
	Status *string `json:"status" form:"status"`
}

func PendingSignatures(c *gin.Context) {

	var form pendingSignatureForm

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

	filteredDocuments := filterDocuments(documents, signerKey)

	c.JSON(http.StatusOK, gin.H{
		"documents": filteredDocuments,
	})
}

func filterDocuments(documents []map[string]interface{}, signerKey string) []map[string]interface{} {
	var filtered []map[string]interface{}

	for _, doc := range documents {
		rejectedSignatures, _ := doc["rejectedSignatures"].([]interface{})
		successfulSignatures, _ := doc["successfulSignatures"].([]interface{})

		if containsSignature(rejectedSignatures, signerKey) || containsSignature(successfulSignatures, signerKey) {
			continue
		}

		filtered = append(filtered, doc)
	}

	return filtered
}

func containsSignature(signatures []interface{}, signerKey string) bool {
	for _, sig := range signatures {
		signerMap, ok := sig.(map[string]interface{})
		if ok && signerMap["@key"] == signerKey {
			return true
		}
	}
	return false
}
