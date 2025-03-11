package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

type addStoredValueToGetCreditForm struct {
	Clause          map[string]interface{} `form:"clause" binding:"required"`
	StoredValue  *float64               `form:"storedValue" binding:"required"`
}

func AddStoredValueToGetCredit(c *gin.Context) {
	var form addStoredValueToGetCreditForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"clause": form.Clause,
	}

	if form.StoredValue != nil {
		reqMap["storedValue"] = *form.StoredValue
	}

	updatedClause, err := chaincode.AddStoredValueToGetCredit(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add inputs to clause", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"clause": updatedClause})
}
