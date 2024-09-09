package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

type EditTemplateClauseForm struct {
	TemplateClause    map[string]interface{}   `form:"templateClause" binding:"required"`
	Name              string                   `form:"name"`
	Number            *float64                 `form:"number"`
	Description       string                   `form:"description"`
	Category          string                   `form:"category"`
	ActionType        *float64                 `form:"actionType"`
	Dependencies      []map[string]interface{} `form:"dependencies"`
	DefaultInputs     map[string]interface{}   `form:"defaultInputs"`
	DefaultParameters map[string]interface{}   `form:"defaultParameters"`
	Optional          *bool                    `form:"optional"`
}

func EditTemplateClause(c *gin.Context) {
	var form EditTemplateClauseForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	_, err := chaincode.SearchAsset(form.TemplateClause)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find template asset", http.StatusInternalServerError)
		return
	}

	req := map[string]interface{}{
		"templateClause": form.TemplateClause,
	}

	if len(form.Dependencies) > 0 {
		req["dependencies"] = form.Dependencies
	}
	if form.Description != "" {
		req["description"] = form.Description
	}
	if form.Name != "" {
		req["name"] = form.Name
	}
	if form.ActionType != nil {
		req["actionType"] = *form.ActionType
	}
	if form.Number != nil {
		req["number"] = *form.Number
	}
	if form.Category != "" {
		req["category"] = form.Category
	}
	if form.DefaultInputs != nil {
		req["defaultInputs"] = form.DefaultInputs
	}
	if form.DefaultParameters != nil {
		req["defaultParameters"] = form.DefaultParameters
	}
	if form.Optional != nil {
		req["optional"] = *form.Optional
	}

	updatedContractAsset, err := chaincode.EditTemplateClause(req)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to edit template clause", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"templateClause": updatedContractAsset})
}
