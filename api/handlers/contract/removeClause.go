package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

type removeClauseForm struct {
	AutoExecutableContract map[string]interface{} `form:"autoExecutableContract" binding:"required"`
	Clause                 string                 `form:"clause" binding:"required"`
}

func RemoveClause(c *gin.Context) {
	var form removeClauseForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		errorhandler.ReturnError(c, fmt.Errorf("email not found in headers"), "email not found in headers", http.StatusBadRequest)
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find signer key", http.StatusInternalServerError)
		return
	}

	contractAsset, err := chaincode.SearchAsset(form.AutoExecutableContract)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find contract asset", http.StatusInternalServerError)
		return
	}

	results, ok := contractAsset["result"].([]interface{})
	if !ok || len(results) == 0 {
		errorhandler.ReturnError(c, fmt.Errorf("no results found in contract asset"), "no results found in contract asset", http.StatusInternalServerError)
		return
	}

	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid result format"), "invalid result format", http.StatusInternalServerError)
		return
	}

	contractOwner, ok := firstResult["owner"].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("could not find owner of th contract"), "could not find owner of th contract", http.StatusInternalServerError)
		return
	}

	if contractOwner["@key"] != signerKey {
		errorhandler.ReturnError(c, fmt.Errorf("only the owner of the contract can remove the clause"), "only the owner of the contract can remove the clause", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"autoExecutableContract": form.AutoExecutableContract,
		"clause": map[string]interface{}{
			"@key":       form.Clause,
			"@assetType": "clause",
		},
	}

	updatedContractAsset, err := chaincode.RemoveClause(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to remove clause from contract", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"contract": updatedContractAsset})
}
