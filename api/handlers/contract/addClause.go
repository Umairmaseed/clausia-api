package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

type addClauseForm struct {
	AutoExecutableContract map[string]interface{}   `form:"autoExecutableContract" binding:"required"`
	Id                     string                   `form:"id" binding:"required"`
	Description            string                   `form:"description"`
	Category               string                   `form:"category"`
	Parameters             map[string]interface{}   `form:"parameters"`
	Input                  map[string]interface{}   `form:"input"`
	Executable             bool                     `form:"executable" binding:"required"`
	Dependencies           []map[string]interface{} `form:"dependencies"`
	ActionType             float64                  `form:"actionType" binding:"required"`
	Finalized              bool                     `form:"finalized"`
	Result                 map[string]interface{}   `form:"result"`
}

func AddClause(c *gin.Context) {
	var form addClauseForm
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

	contractAsset, err := chaincode.CreateAutoExecutableContract(form.AutoExecutableContract)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find contract asset", http.StatusInternalServerError)
		return
	}

	contractOwner := contractAsset["owner"].(map[string]interface{})

	if contractOwner["@key"] != signerKey {
		errorhandler.ReturnError(c, fmt.Errorf("only the owner of the contract can add the clause"), "only the owner of the contract can add the clause", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"autoExecutableContract": form.AutoExecutableContract,
		"id":                     form.Id,
		"executable":             form.Executable,
		"actionType":             form.ActionType,
	}

	if form.Description != "" {
		reqMap["description"] = form.Description
	}
	if form.Category != "" {
		reqMap["category"] = form.Category
	}
	if form.Parameters != nil {
		reqMap["parameters"] = form.Parameters
	}
	if form.Input != nil {
		reqMap["input"] = form.Input
	}
	if form.Dependencies != nil {
		reqMap["dependencies"] = form.Dependencies
	}
	if form.Finalized {
		reqMap["finalized"] = form.Finalized
	}
	if form.Result != nil {
		reqMap["result"] = form.Result
	}

	updatedContractAsset, err := chaincode.AddClause(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add clause to contract", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"contract": updatedContractAsset})

}
