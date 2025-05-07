package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/chaincode"
)

type createTemplateClauseForm struct {
	Id                string                   `form:"id" binding:"required"`
	Template          map[string]interface{}   `form:"template" binding:"required"`
	Number            float64                  `form:"number" binding:"required"`
	Name              string                   `form:"name" binding:"required"`
	Description       string                   `form:"description"`
	Category          string                   `form:"category"`
	ActionType        float64                  `form:"actionType" binding:"required"`
	Dependencies      []map[string]interface{} `form:"dependencies"`
	DefaultInputs     map[string]interface{}   `form:"defaultInputs"`
	DefaultParameters map[string]interface{}   `form:"defaultParameters"`
	Optional          *bool                    `form:"optional"`
}

func CreateTemplateClause(c *gin.Context) {
	var form createTemplateClauseForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Error("Failed to bind request form: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := map[string]interface{}{
		"name":       form.Name,
		"template":   form.Template,
		"id":         form.Id,
		"number":     form.Number,
		"actionType": form.ActionType,
	}

	if len(form.Dependencies) > 0 {
		req["dependencies"] = form.Dependencies
	}
	if form.Description != "" {
		req["description"] = form.Description
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

	contract, err := chaincode.CreateTemplateClause(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"templateClause": contract})

}
