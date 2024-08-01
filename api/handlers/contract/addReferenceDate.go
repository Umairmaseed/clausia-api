package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

type addReferenceDateForm struct {
	Clause        map[string]interface{} `form:"clause" binding:"required"`
	ReferenceDate string                 `form:"referenceDate" binding:"required"`
}

func AddReferenceDate(c *gin.Context) {
	var form addReferenceDateForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"clause":        form.Clause,
		"referenceDate": form.ReferenceDate,
	}

	updatedClause, err := chaincode.AddReferenceDate(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add reference date to clause", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"clause": updatedClause})
}
