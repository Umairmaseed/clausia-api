package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/utils"
)

func GetUserContracts(c *gin.Context) {

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

	queryMapUserContract := map[string]interface{}{
		"@assetType": "autoExecutableContract",
		"owner": map[string]interface{}{
			"@assetType": "user",
			"@key":       signerKey,
		},
	}

	userContractAsset, err := chaincode.SearchAssetTx(queryMapUserContract)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for contracts", http.StatusInternalServerError)
		return
	}

	queryMapParticipantContract := map[string]interface{}{
		"@assetType": "autoExecutableContract",
		"participants": map[string]interface{}{
			"$elemMatch": map[string]interface{}{
				"@assetType": "user",
				"@key":       signerKey,
			},
		},
	}

	participantContractsAsset, err := chaincode.SearchAssetTx(queryMapParticipantContract)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for participant contracts", http.StatusInternalServerError)
		return
	}

	response := gin.H{}

	if len(userContractAsset) > 0 {
		response["userCreatedContracts"] = userContractAsset
	} else {
		response["userCreatedContracts"] = []interface{}{}
	}

	if len(participantContractsAsset) > 0 {
		response["participantContract"] = participantContractsAsset
	} else {
		response["participantContract"] = []interface{}{}
	}

	c.JSON(http.StatusOK, response)

}
