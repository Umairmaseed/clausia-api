package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

type EditTemplateForm struct {
	Template    map[string]interface{} `form:"template" binding:"required"`
	Name        string                 `form:"name"`
	Description string                 `form:"description"`
	Public      *bool                  `form:"public"`
}

func EditTemplate(c *gin.Context) {
	var form EditTemplateForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		errorhandler.ReturnError(c, fmt.Errorf("email not found in headers"), "email not found in headers", http.StatusBadRequest)
		return
	}

	userKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find user key", http.StatusInternalServerError)
		return
	}

	contractAsset, err := chaincode.SearchAsset(form.Template)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find template asset", http.StatusInternalServerError)
		return
	}

	results, ok := contractAsset["result"].([]interface{})
	if !ok || len(results) == 0 {
		errorhandler.ReturnError(c, fmt.Errorf("no results found in template asset"), "no results found in template asset", http.StatusInternalServerError)
		return
	}

	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid result format"), "invalid result format", http.StatusInternalServerError)
		return
	}

	contractOwner, ok := firstResult["creator"].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("could not find creator of the template"), "could not find creator of the template", http.StatusInternalServerError)
		return
	}

	if contractOwner["@key"] != userKey {
		errorhandler.ReturnError(c, fmt.Errorf("only the creator of the template can edit"), "only the creator of the template can edit", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"template": form.Template,
	}

	if form.Description != "" {
		reqMap["description"] = form.Description
	}
	if form.Name != "" {
		reqMap["name"] = form.Name
	}

	if form.Public != nil {
		reqMap["public"] = *form.Public
	}
	updatedContractAsset, err := chaincode.EditTemplate(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to edit template", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": updatedContractAsset})
}
