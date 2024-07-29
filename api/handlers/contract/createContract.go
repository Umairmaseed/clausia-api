package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type createContractForm struct {
	Name          string                   `form:"name" binding:"required"`
	SignatureDate string                   `form:"signatureDate" binding:"required"`
	Clauses       []map[string]interface{} `form:"clauses"`
	Data          map[string]interface{}   `form:"data"`
}

func CreateContract(c *gin.Context) {
	var form createContractForm
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
		"name":          form.Name,
		"signatureDate": form.SignatureDate,
		"owner":         signerAsset,
	}

	if len(form.Clauses) > 0 {
		req["clauses"] = form.Clauses
	}
	if form.Data != nil {
		req["data"] = form.Data
	}

	contract, err := chaincode.CreateAutoExecutableContract(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"contract": contract})

}
