package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

func GetClause(c *gin.Context) {
	clauseKey := c.DefaultQuery("clauseKey", "")
	if clauseKey == "" {
		errorhandler.ReturnError(c, fmt.Errorf("clauseKey not found in query parameters"), "clauseKey not found", http.StatusBadRequest)
		return
	}

	queryMapClause := map[string]interface{}{
		"@assetType": "clause",
		"@key":       clauseKey,
	}

	clauseAsset, err := chaincode.SearchAssetTx(queryMapClause)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for clause", http.StatusInternalServerError)
		return
	}

	response := gin.H{}

	if len(clauseAsset) > 0 {
		response["clause"] = clauseAsset[0]
	} else {
		response["clause"] = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, response)
}
