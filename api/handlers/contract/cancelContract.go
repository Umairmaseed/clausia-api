package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/chaincode"
)

type CancelContractForm struct {
	Clause                map[string]interface{} `form:"clause" binding:"required"`
	ForceCancellation     bool                   `form:"forceCancellation"`
	RequestedCancellation bool                   `form:"requestedCancellation"`
}

func CancelContract(c *gin.Context) {
	var form CancelContractForm

	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"clause": form.Clause,
	}

	if form.ForceCancellation {
		reqMap["forceCancellation"] = form.ForceCancellation
	}

	if form.RequestedCancellation {
		reqMap["requestedCancellation"] = form.RequestedCancellation
	}

	updatedClause, err := chaincode.CancelContract(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to complete cancel request for contract", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"clause": updatedClause})
}
