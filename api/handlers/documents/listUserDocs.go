package documents

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/utils"
)

type statusParam struct {
	Status *string `json:"status" form:"status"`
}

func ListUserDocs(c *gin.Context) {
	var form statusParam

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

	queryMap := map[string]interface{}{
		"@assetType": "document",
		"owner": map[string]interface{}{
			"@assetType": "user",
			"@key":       signerKey,
		},
	}

	if form.Status != nil {
		statusFloat, err := strconv.ParseFloat(*form.Status, 64)
		if err != nil {
			errorhandler.ReturnError(c, err, "Invalid status value", http.StatusBadRequest)
			return
		}
		queryMap["status"] = statusFloat
	}

	docAsset, err := chaincode.SearchAssetTx(queryMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for documents", http.StatusInternalServerError)
		return
	}

	if len(docAsset) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"documents": []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": docAsset,
	})

}
