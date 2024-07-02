package documents

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type updateDoc struct {
	DocKey  string `form:"dockey" binding:"required"`
	Name    string `form:"name"`
	Timeout string `form:"timeout"`
}

func UpdateDocNameOrTimeout(c *gin.Context) {
	var form updateDoc

	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind form data", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		errorhandler.ReturnError(c, fmt.Errorf("email not found in headers"), "email not found in headers", http.StatusBadRequest)
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		errorhandler.ReturnError(c, err, err.Error(), http.StatusInternalServerError)
		return
	}

	doc, err := chaincode.GetDoc(form.DocKey)
	if err != nil {
		errorhandler.ReturnError(c, err, err.Error(), http.StatusInternalServerError)
		return
	}

	ownerMap, ok := doc["owner"].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid owner data format"), "invalid owner data format", http.StatusInternalServerError)
		return
	}

	ownerKey, ok := ownerMap["@key"].(string)
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("owner key not found"), "owner key not found", http.StatusInternalServerError)
		return
	}

	if ownerKey != signerKey {
		errorhandler.ReturnError(c, fmt.Errorf("user not authorized to changed the status of the document"), "user not authorized to changed the status of the document", http.StatusBadRequest)
		return
	}

	if doc == nil {
		errorhandler.ReturnError(c, nil, "No document found", http.StatusNotFound)
		return
	}

	updatesMap := make(map[string]interface{})
	if form.Name != "" {
		updatesMap["name"] = form.Name
	}
	if form.Timeout != "" {
		updatesMap["timeout"] = form.Timeout
	}

	if len(updatesMap) == 0 {
		errorhandler.ReturnError(c, nil, "No fields to update", http.StatusBadRequest)
		return
	}

	documentMAp := map[string]interface{}{
		"@assetType": "document",
		"@key":       form.DocKey,
	}

	updateDoc, err := chaincode.UpdateDocument(documentMAp, updatesMap)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Document updated successfully",
		"documentKey": updateDoc,
	})

}
