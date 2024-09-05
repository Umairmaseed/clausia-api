package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type createTemplateForm struct {
	Id          string                   `form:"id" binding:"required"`
	Name        string                   `form:"name" binding:"required"`
	Description string                   `form:"description"`
	Public      bool                     `form:"public" binding:"required"`
	Clauses     []map[string]interface{} `form:"clauses"`
}

func CreateTemplate(c *gin.Context) {
	var form createTemplateForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Error("Failed to bind request form: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		logger.Error("Email not found in headers")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in headers"})
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	signerAsset := map[string]interface{}{
		"@assetType": "user",
		"@key":       signerKey,
	}

	req := map[string]interface{}{
		"name":    form.Name,
		"creator": signerAsset,
		"public":  form.Public,
		"id":      form.Id,
	}

	if len(form.Clauses) > 0 {
		req["clauses"] = form.Clauses
	}
	if form.Description != "" {
		req["description"] = form.Description
	}

	contract, err := chaincode.CreateTemplate(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": contract})

}
