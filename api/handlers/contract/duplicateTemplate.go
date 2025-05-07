package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/utils"
)

type DuplicateTemplateForm struct {
	Id       string                 `form:"id" binding:"required"`
	Name     string                 `form:"name" binding:"required"`
	Template map[string]interface{} `form:"Template" binding:"required"`
}

func DuplicateTemplate(c *gin.Context) {
	var form DuplicateTemplateForm
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

	userKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	userAsset := map[string]interface{}{
		"@assetType": "user",
		"@key":       userKey,
	}

	req := map[string]interface{}{
		"name":     form.Name,
		"creator":  userAsset,
		"id":       form.Id,
		"template": form.Template,
	}

	contract, err := chaincode.DuplicateTemplate(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": contract})

}
