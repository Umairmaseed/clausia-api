package documents

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

func ListSuccessfulSignatures(c *gin.Context) {

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
		"successfulSignatures": map[string]interface{}{
			"$elemMatch": map[string]interface{}{
				"@assetType": "user",
				"@key":       signerKey,
			},
		},
	}

	docAsset, err := chaincode.SearchAssetTx(queryMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for documents", http.StatusInternalServerError)
		return
	}

	var response interface{}
	if len(docAsset) > 0 {
		response = docAsset
	} else if len(docAsset) == 0 {
		response = []interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": response,
	})

}
