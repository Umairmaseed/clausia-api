package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

// this handler will use clause key to get the contract dates

func GetDatesWithCLause(c *gin.Context) {
	clauseKey := c.DefaultQuery("clauseKey", "")
	if clauseKey == "" {
		errorhandler.ReturnError(c, fmt.Errorf("clauseKey not found in query parameters"), "clauseKey not found", http.StatusBadRequest)
		return
	}

	query := map[string]interface{}{
		"@assetType": "autoExecutableContract",
		"clauses": map[string]interface{}{
			"$elemMatch": map[string]interface{}{
				"@assetType": "clause",
				"@key":       clauseKey,
			},
		},
	}

	result, err := chaincode.SearchAssetTx(query)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for contract", http.StatusInternalServerError)
		return
	}

	response := gin.H{}

	fmt.Println("result", result)
	if len(result) > 0 {
		contractAsset := result[0]
		response["dates"] = contractAsset["dates"]
	} else {
		response["dates"] = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, response)
}
