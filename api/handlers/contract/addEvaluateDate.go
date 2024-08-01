package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

type addEvaluateDateForm struct {
	Clause       map[string]interface{} `form:"clause" binding:"required"`
	EvaluateDate string                 `form:"evaluateDate" binding:"required"`
}

func AddEvaluateDate(c *gin.Context) {
	var form addEvaluateDateForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"clause":        form.Clause,
		"evaluatedDate": form.EvaluateDate,
	}

	updatedClause, err := chaincode.AddEvaluateDate(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add evaluate date to clause", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"clause": updatedClause})
}
