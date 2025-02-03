package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

func GetContract(c *gin.Context) {
	contractKey := c.DefaultQuery("contractKey", "")
	if contractKey == "" {
		errorhandler.ReturnError(c, fmt.Errorf("contractKey not found in query parameters"), "contractKey not found", http.StatusBadRequest)
		return
	}

	queryMapContract := map[string]interface{}{
		"@assetType": "autoExecutableContract",
		"@key":       contractKey,
	}

	contractAsset, err := chaincode.SearchAssetTx(queryMapContract)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for contract", http.StatusInternalServerError)
		return
	}

	response := gin.H{}

	if len(contractAsset) > 0 {
		response["contract"] = contractAsset[0]
	} else {
		response["contract"] = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, response)
}
